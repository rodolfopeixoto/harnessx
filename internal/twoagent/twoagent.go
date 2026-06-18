// SPDX-License-Identifier: MIT

package twoagent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/ropeixoto/harnessx/internal/ui"
)

type Problem struct {
	ID       string   `json:"id"`
	Severity string   `json:"severity"`
	Subject  string   `json:"subject"`
	Detail   string   `json:"detail"`
	Hints    []string `json:"hints,omitempty"`
	FixID    string   `json:"fix_id,omitempty"`
}

type Diagnosis struct {
	GeneratedAt time.Time `json:"generated_at"`
	Problems    []Problem `json:"problems"`
}

type Diagnoser interface {
	Diagnose(ctx context.Context, root string) ([]Problem, error)
}

func DiagnoseAll(ctx context.Context, root string, diags []Diagnoser) (Diagnosis, error) {
	d := Diagnosis{GeneratedAt: time.Now().UTC()}
	for _, dx := range diags {
		ps, err := dx.Diagnose(ctx, root)
		if err != nil {
			return d, err
		}
		d.Problems = append(d.Problems, ps...)
	}
	sort.Slice(d.Problems, func(i, j int) bool {
		return severityRank(d.Problems[i].Severity) < severityRank(d.Problems[j].Severity)
	})
	return d, nil
}

func severityRank(s string) int {
	switch s {
	case "error":
		return 0
	case "warn":
		return 1
	case "info":
		return 2
	}
	return 3
}

func SaveDiagnosis(root string, d Diagnosis) (string, error) {
	dir := filepath.Join(root, ".harness", "artifacts", "diagnoses")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	path := filepath.Join(dir, fmt.Sprintf("diag-%d.json", d.GeneratedAt.Unix()))
	body, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(path, body, 0o644); err != nil {
		return "", err
	}
	return path, nil
}

func LoadDiagnosis(path string) (Diagnosis, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return Diagnosis{}, err
	}
	var d Diagnosis
	if err := json.Unmarshal(body, &d); err != nil {
		return Diagnosis{}, err
	}
	return d, nil
}

type Fixer interface {
	ID() string
	Description() string
	Apply(ctx context.Context, root string, p Problem, out io.Writer) error
}

type FixResult struct {
	Problem Problem
	Applied bool
	Skipped bool
	Err     error
}

func ApplyAll(ctx context.Context, root string, d Diagnosis, fixers map[string]Fixer, out io.Writer) []FixResult {
	var results []FixResult
	for _, p := range d.Problems {
		if p.FixID == "" {
			results = append(results, FixResult{Problem: p, Skipped: true})
			continue
		}
		fx, ok := fixers[p.FixID]
		if !ok {
			results = append(results, FixResult{Problem: p, Skipped: true, Err: fmt.Errorf("no fixer registered for %q", p.FixID)})
			continue
		}
		fmt.Fprintf(out, "→ applying fix %s for %s\n", fx.ID(), p.Subject)
		err := fx.Apply(ctx, root, p, out)
		results = append(results, FixResult{Problem: p, Applied: err == nil, Err: err})
	}
	return results
}

type MissingToolDiagnoser struct {
	Tools []string
}

func (d MissingToolDiagnoser) Diagnose(ctx context.Context, root string) ([]Problem, error) {
	var ps []Problem
	for _, t := range d.Tools {
		if _, err := exec.LookPath(t); err == nil {
			continue
		}
		ps = append(ps, Problem{
			ID:       "tool:" + t,
			Severity: "warn",
			Subject:  t + " not installed",
			Detail:   "Required by the project; install before running the matching sensor.",
			Hints:    []string{"harness install " + t},
			FixID:    "install-tool",
		})
	}
	return ps, nil
}

type DirtyTreeDiagnoser struct{}

func (DirtyTreeDiagnoser) Diagnose(ctx context.Context, root string) ([]Problem, error) {
	cmd := exec.CommandContext(ctx, "git", "status", "--porcelain")
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		return nil, nil
	}
	if strings.TrimSpace(string(out)) == "" {
		return nil, nil
	}
	return []Problem{{
		ID:       "git:dirty",
		Severity: "warn",
		Subject:  "Working tree has uncommitted changes",
		Detail:   "harness ship refuses to run on a dirty tree.",
		Hints:    []string{"git add -A && git commit -m \"chore: snapshot\""},
		FixID:    "commit-snapshot",
	}}, nil
}

type MissingPlanDiagnoser struct{}

func (MissingPlanDiagnoser) Diagnose(ctx context.Context, root string) ([]Problem, error) {
	pin := filepath.Join(root, ".harness", "config", "plan.yaml")
	if _, err := os.Stat(pin); err == nil {
		return nil, nil
	}
	return []Problem{{
		ID:       "plan:no-pin",
		Severity: "info",
		Subject:  "No active plan pinned",
		Detail:   "Pin a plan to make the plan_scope sensor enforce scope on the next run.",
		Hints: []string{
			"harness plan write \"<intent>\" --file <path> --risk medium",
			"printf 'active_plan_id: %s\\n' \"$PLAN_ID\" > .harness/config/plan.yaml",
		},
	}}, nil
}

type InstallToolFixer struct{ HarnessBin string }

func (f InstallToolFixer) ID() string          { return "install-tool" }
func (f InstallToolFixer) Description() string { return "harness install <tool>" }
func (f InstallToolFixer) Apply(ctx context.Context, root string, p Problem, out io.Writer) error {
	tool := strings.TrimPrefix(p.ID, "tool:")
	if tool == "" {
		return errors.New("install-tool: missing tool id")
	}
	bin := f.HarnessBin
	if bin == "" {
		bin = "harness"
	}
	cmd := exec.CommandContext(ctx, bin, "install", tool)
	cmd.Dir = root
	cmd.Stdout = out
	cmd.Stderr = out
	return cmd.Run()
}

type CommitSnapshotFixer struct {
	Email   string
	Name    string
	Message string
}

func (f CommitSnapshotFixer) ID() string          { return "commit-snapshot" }
func (f CommitSnapshotFixer) Description() string { return "git add + commit" }
func (f CommitSnapshotFixer) Apply(ctx context.Context, root string, p Problem, out io.Writer) error {
	email := f.Email
	if email == "" {
		email = "you@example.com"
	}
	name := f.Name
	if name == "" {
		name = "harness fix"
	}
	msg := f.Message
	if msg == "" {
		msg = "chore: snapshot before harness fix"
	}
	if err := runGitFix(ctx, root, out, "add", "-A"); err != nil {
		return err
	}
	return runGitFix(ctx, root, out, "-c", "user.email="+email, "-c", "user.name="+name, "commit", "-m", msg)
}

func runGitFix(ctx context.Context, root string, out io.Writer, args ...string) error {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = root
	cmd.Stdout = out
	cmd.Stderr = out
	return cmd.Run()
}

func DefaultDiagnosers(stackTools []string) []Diagnoser {
	return []Diagnoser{
		MissingToolDiagnoser{Tools: stackTools},
		DirtyTreeDiagnoser{},
		MissingPlanDiagnoser{},
	}
}

func DefaultFixers(harnessBin string) map[string]Fixer {
	return map[string]Fixer{
		"install-tool":    InstallToolFixer{HarnessBin: harnessBin},
		"commit-snapshot": CommitSnapshotFixer{},
	}
}

func FormatDiagnosis(d Diagnosis) string {
	if len(d.Problems) == 0 {
		return ui.Success.Render("no problems detected ✓") + "\n"
	}
	var b strings.Builder
	for _, p := range d.Problems {
		mark := ui.MarkDot()
		switch p.Severity {
		case "error":
			mark = ui.MarkFail()
		case "warn":
			mark = ui.MarkWarn()
		case "info":
			mark = ui.MarkInfo()
		}
		fmt.Fprintf(&b, "%s %s %s\n", mark, ui.Muted.Render("["+p.ID+"]"), ui.Heading.Render(p.Subject))
		if p.Detail != "" {
			fmt.Fprintf(&b, "    %s\n", p.Detail)
		}
		for _, h := range p.Hints {
			fmt.Fprintf(&b, "    %s %s\n", ui.Muted.Render("→"), ui.Accent.Render(h))
		}
	}
	return b.String()
}
