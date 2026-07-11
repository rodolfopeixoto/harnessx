// SPDX-License-Identifier: MIT

package audit

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// ErrRunNotFound is returned by Replay when the requested run id cannot be
// located, neither as a run directory under .harness/runs/<id>/ nor as
// audit events tagged with the same id.
var ErrRunNotFound = errors.New("audit: run not found")

// ReplayOptions controls the replay behaviour. DryRun defaults to true at
// the CLI layer; mutating replays are opt-in and still bounded to TmpDir.
type ReplayOptions struct {
	// DryRun, when true, only lists the actions that would be replayed.
	// When false, replay materialises artefacts into TmpDir (never
	// outside), but still touches zero state under the real project.
	DryRun bool
	// TmpDir, if empty, is created via os.MkdirTemp under the OS temp
	// dir. Replay never writes anywhere except this directory.
	TmpDir string
	// Now returns the wall clock used to stamp the report. Injected for
	// deterministic tests.
	Now func() time.Time
}

// ReplayStep is a single event scheduled for replay.
type ReplayStep struct {
	Index      int       `json:"index"`
	OccurredAt time.Time `json:"occurred_at"`
	Kind       string    `json:"kind"`
	Source     string    `json:"source"`
	Subject    string    `json:"subject"`
	// Action is the human-readable summary of what replay would (or
	// did) do for this event. Stable string, safe to snapshot in tests.
	Action string `json:"action"`
}

// ReplayReport is the structured output of a replay run.
type ReplayReport struct {
	RunID     string       `json:"run_id"`
	DryRun    bool         `json:"dry_run"`
	TmpDir    string       `json:"tmp_dir"`
	StartedAt time.Time    `json:"started_at"`
	Steps     []ReplayStep `json:"steps"`
	Events    int          `json:"events"`
	Notes     []string     `json:"notes,omitempty"`
}

// Replay reads the append-only audit event log for the given run and
// emits a ReplayReport enumerating what would (or did) be re-executed
// inside an isolated tmpdir. Behaviour is deliberately conservative: no
// hook, sensor, or agent is actually re-invoked from this function. The
// replay unit is the descriptive step; downstream code may hook into
// individual steps to implement side effects, but any such extension
// must still respect TmpDir isolation.
func Replay(ctx context.Context, projectRoot, runID string, opts ReplayOptions) (ReplayReport, error) {
	if projectRoot == "" {
		return ReplayReport{}, errors.New("audit: empty project root")
	}
	if strings.TrimSpace(runID) == "" {
		return ReplayReport{}, errors.New("audit: empty run id")
	}
	if opts.Now == nil {
		opts.Now = func() time.Time { return time.Now().UTC() }
	}

	events, err := collectRunEvents(ctx, projectRoot, runID)
	if err != nil {
		return ReplayReport{}, err
	}
	runDirExists := hasRunDir(projectRoot, runID)

	if len(events) == 0 && !runDirExists {
		return ReplayReport{}, fmt.Errorf("%w: %s", ErrRunNotFound, runID)
	}

	tmp, cleanup, err := ensureTmpDir(opts.TmpDir)
	if err != nil {
		return ReplayReport{}, err
	}
	_ = cleanup // caller inspects TmpDir via the report; retained for future use.

	report := ReplayReport{
		RunID:     runID,
		DryRun:    opts.DryRun,
		TmpDir:    tmp,
		StartedAt: opts.Now(),
		Events:    len(events),
	}
	if len(events) == 0 && runDirExists {
		report.Notes = append(report.Notes,
			"no audit events tagged with this run id; run directory exists but event log is empty")
	}

	// Chronological order — audit sinks return descending, replay wants
	// forward-time semantics so operators read cause before effect.
	sort.SliceStable(events, func(i, j int) bool {
		return events[i].OccurredAt.Before(events[j].OccurredAt)
	})
	for i, ev := range events {
		report.Steps = append(report.Steps, ReplayStep{
			Index:      i,
			OccurredAt: ev.OccurredAt,
			Kind:       ev.Kind,
			Source:     ev.Source,
			Subject:    ev.Subject,
			Action:     describeAction(ev, opts.DryRun),
		})
	}
	if !opts.DryRun {
		// Materialise a manifest of the replayed events inside the
		// tmpdir. This is the only side effect a non-dry-run replay
		// performs today; it stays sandboxed under TmpDir.
		if err := writeManifest(tmp, report); err != nil {
			return report, fmt.Errorf("audit: write manifest: %w", err)
		}
	}
	return report, nil
}

func collectRunEvents(ctx context.Context, projectRoot, runID string) ([]Event, error) {
	candidates := []string{
		filepath.Join(projectRoot, ".harness", "audit", "events.jsonl"),
		filepath.Join(projectRoot, ".harness", "logs", "events.jsonl"),
	}
	var merged []Event
	for _, p := range candidates {
		sink := &FileSink{Path: p}
		evs, err := sink.List(ctx)
		if err != nil {
			// A missing file is already normalised to (nil, nil) by
			// FileSink.List, so any error here is real (permissions,
			// malformed JSON) and should surface to the caller.
			return nil, fmt.Errorf("audit: list %s: %w", p, err)
		}
		for _, ev := range evs {
			if ev.ID == runID {
				merged = append(merged, ev)
			}
		}
	}
	return merged, nil
}

func hasRunDir(projectRoot, runID string) bool {
	st, err := os.Stat(filepath.Join(projectRoot, ".harness", "runs", runID))
	return err == nil && st.IsDir()
}

func ensureTmpDir(dir string) (string, func(), error) {
	if dir == "" {
		d, err := os.MkdirTemp("", "harness-audit-replay-*")
		if err != nil {
			return "", nil, fmt.Errorf("audit: mkdir tmp: %w", err)
		}
		return d, func() { _ = os.RemoveAll(d) }, nil
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", nil, fmt.Errorf("audit: mkdir tmp %s: %w", dir, err)
	}
	return dir, func() {}, nil
}

func describeAction(ev Event, dryRun bool) string {
	verb := "would replay"
	if !dryRun {
		verb = "replayed"
	}
	if ev.Subject == "" {
		return fmt.Sprintf("%s %s from %s", verb, ev.Kind, ev.Source)
	}
	return fmt.Sprintf("%s %s from %s: %s", verb, ev.Kind, ev.Source, ev.Subject)
}

func writeManifest(dir string, report ReplayReport) error {
	f, err := os.Create(filepath.Join(dir, "manifest.txt"))
	if err != nil {
		return err
	}
	defer f.Close()
	fmt.Fprintf(f, "run_id=%s\nevents=%d\nstarted_at=%s\n",
		report.RunID, report.Events, report.StartedAt.Format(time.RFC3339))
	for _, s := range report.Steps {
		fmt.Fprintf(f, "%03d\t%s\t%s\t%s\t%s\n",
			s.Index, s.OccurredAt.Format(time.RFC3339), s.Kind, s.Source, s.Subject)
	}
	return nil
}
