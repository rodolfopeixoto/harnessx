// SPDX-License-Identifier: MIT

// Package reportcmd renders the run-level report described in spec §28
// and exposes `harness report --last`.
package reportcmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/ropeixoto/harnessx/internal/adapters/sqlite"
	"github.com/ropeixoto/harnessx/internal/domain"
	"github.com/ropeixoto/harnessx/internal/platform/config"
	"github.com/ropeixoto/harnessx/internal/platform/ids"
	"github.com/ropeixoto/harnessx/internal/platform/paths"
)

// Input bundles everything Build needs to render a report end-to-end.
type Input struct {
	SessionID string
	RunID     string
	Mode      domain.Mode
	Intent    string
	SpecPath  string
	PlanPath  string
	Files     []string
	Tests     []string
	Sensors   []SensorLine
	Cost      CostLine
	Questions []string
	Risks     []string
	Rollback  []string
	Evidence  []string
}

type SensorLine struct {
	Name     string
	Status   string
	Duration time.Duration
}

type CostLine struct {
	TotalUSD     float64
	InputTokens  int
	OutputTokens int
}

func Build(root string, in Input) (string, error) {
	if in.RunID == "" {
		in.RunID = ids.New()
	}
	runDir := filepath.Join(paths.HarnessDir(root), "runs", in.RunID)
	if err := os.MkdirAll(runDir, 0o755); err != nil {
		return "", err
	}
	path := filepath.Join(runDir, "report.md")
	if _, err := os.Stat(path); err == nil {
		return path, nil
	}
	var b strings.Builder
	w := func(s string) { b.WriteString(s); b.WriteByte('\n') }

	w("# Summary")
	w(fmt.Sprintf("Mode `%s`. Run id `%s`. Session `%s`.", in.Mode, in.RunID, in.SessionID))
	w("")
	w("# Intent")
	w(in.Intent)
	w("")
	w("# Mode")
	w(string(in.Mode))
	w("")
	w("# Spec")
	if in.SpecPath != "" {
		w("- " + in.SpecPath)
	} else {
		w("_(none)_")
	}
	w("")
	w("# Plan")
	if in.PlanPath != "" {
		w("- " + in.PlanPath)
	} else {
		w("_(none)_")
	}
	w("")
	w("# Questions and Assumptions")
	if len(in.Questions) == 0 {
		w("_None._")
	} else {
		for _, q := range in.Questions {
			w("- " + q)
		}
	}
	w("")
	w("# Files Changed")
	if len(in.Files) == 0 {
		w("_None recorded._")
	} else {
		sort.Strings(in.Files)
		for _, f := range in.Files {
			w("- " + f)
		}
	}
	w("")
	w("# Tests Added or Updated")
	if len(in.Tests) == 0 {
		w("_None recorded._")
	} else {
		sort.Strings(in.Tests)
		for _, t := range in.Tests {
			w("- " + t)
		}
	}
	w("")
	w("# Sensors Run")
	if len(in.Sensors) == 0 {
		w("_None._")
	} else {
		w("| Sensor | Status | Duration |")
		w("|---|---|---|")
		for _, s := range in.Sensors {
			w(fmt.Sprintf("| %s | %s | %s |", s.Name, s.Status, s.Duration.Round(time.Millisecond)))
		}
	}
	w("")
	w("# Cost and Tokens")
	w(fmt.Sprintf("Total: $%.4f. Input tokens: %d. Output tokens: %d.",
		in.Cost.TotalUSD, in.Cost.InputTokens, in.Cost.OutputTokens))
	w("")
	w("# Risks")
	if len(in.Risks) == 0 {
		w("_None recorded._")
	} else {
		for _, r := range in.Risks {
			w("- " + r)
		}
	}
	w("")
	w("# Rollback Plan")
	if len(in.Rollback) == 0 {
		w("_Revert the implementing commit(s) and re-run `harness ci`._")
	} else {
		for _, r := range in.Rollback {
			w("- " + r)
		}
	}
	w("")
	w("# Evidence")
	if len(in.Evidence) == 0 {
		w("_None recorded._")
	} else {
		for _, e := range in.Evidence {
			w("- " + e)
		}
	}

	if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
		return "", err
	}
	return path, nil
}

// PrintLast prints the most recently produced report to out. Returns
// ErrNoReport when none exist.
var ErrNoReport = errors.New("report: no reports yet (run a workflow first)")

func PrintLast(startDir string, out io.Writer) error {
	root, err := paths.FindProjectRoot(startDir)
	if err != nil {
		return err
	}
	runsDir := filepath.Join(paths.HarnessDir(root), "runs")
	entries, err := os.ReadDir(runsDir)
	if err != nil {
		return ErrNoReport
	}
	var newest string
	var newestMod time.Time
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		path := filepath.Join(runsDir, e.Name(), "report.md")
		info, err := os.Stat(path)
		if err != nil {
			continue
		}
		if info.ModTime().After(newestMod) {
			newestMod = info.ModTime()
			newest = path
		}
	}
	if newest == "" {
		return ErrNoReport
	}
	b, err := os.ReadFile(newest)
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "# %s\n\n", filepath.Base(newest))
	_, err = out.Write(b)
	return err
}

// MostRecentRun helper for `harness report` to pull cost + sensors from DB
// when caller doesn't already have them in memory.
func MostRecentRun(ctx context.Context, root string) (sessionID, runID string, err error) {
	cfg, err := config.Load(filepath.Join(root, ".harness", "config", "harness.yaml"), root)
	if err != nil {
		return "", "", err
	}
	dbPath := config.Resolve(root, cfg.Database.Path)
	if _, err := os.Stat(dbPath); err != nil {
		return "", "", err
	}
	repo, err := sqlite.Open(dbPath)
	if err != nil {
		return "", "", err
	}
	defer repo.Close()
	sessions, err := repo.ListRecentSessions(ctx, 1)
	if err != nil || len(sessions) == 0 {
		return "", "", err
	}
	return sessions[0].ID, "", nil
}
