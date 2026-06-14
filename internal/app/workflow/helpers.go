// SPDX-License-Identifier: MIT

package workflow

import (
	"fmt"
	"io"
	"os"
	"strings"

	hxcontext "github.com/ropeixoto/harnessx/internal/context"
	"github.com/ropeixoto/harnessx/internal/domain"
	"github.com/ropeixoto/harnessx/internal/index"
	"github.com/ropeixoto/harnessx/internal/platform/paths"
)

// confirmInteractive prompts on stdin for a y/N answer. Returns false
// when stdin isn't a terminal (CI, redirected) so non-interactive
// callers can't be tricked into auto-approving — the spec mandates
// explicit human approval for risky changes.
func confirmInteractive(out io.Writer, prompt string) bool {
	if !isTerminal(os.Stdin) {
		return false
	}
	fmt.Fprint(out, prompt)
	var resp string
	if _, err := fmt.Fscanln(os.Stdin, &resp); err != nil {
		return false
	}
	resp = strings.ToLower(strings.TrimSpace(resp))
	return resp == "y" || resp == "yes"
}

func isTerminal(f *os.File) bool {
	if f == nil {
		return false
	}
	info, err := f.Stat()
	if err != nil {
		return false
	}
	return (info.Mode() & os.ModeCharDevice) != 0
}

func filesOf(pack *hxcontext.Pack) []string {
	if pack == nil {
		return nil
	}
	out := make([]string, 0, len(pack.RelevantFiles))
	for _, f := range pack.RelevantFiles {
		out = append(out, f.Path)
	}
	return out
}

func stackNames(p index.Profile) []string {
	out := make([]string, 0, len(p.Stacks))
	for _, s := range p.Stacks {
		out = append(out, s.Name)
	}
	return out
}

// estimateCost approximates planning cost as 10 % of the per-run budget.
// Real adapter runs override the row with reported usage.
func estimateCost(budgetUSD float64) float64 {
	if budgetUSD <= 0 {
		return 0
	}
	return budgetUSD * 0.10
}

func riskHints(mode domain.Mode) []string {
	switch mode {
	case domain.ModeBugfix:
		return []string{"regression coverage may be incomplete", "scope creep beyond root cause"}
	case domain.ModeFeature:
		return []string{"new public API may need contract test", "feature toggle missing for partial rollout"}
	case domain.ModeOptimization:
		return []string{"removing tools that look unused may break ops", "missing perf baseline"}
	}
	return nil
}

// taskFor maps a workflow mode to a route key understood by routes.yaml.
// Falls back to "implementation" so unrouted modes still pick something
// sensible without crashing.
func taskFor(mode domain.Mode) string {
	switch mode {
	case domain.ModeQuestion:
		return "codebase_exploration"
	case domain.ModeBugfix:
		return "implementation"
	case domain.ModeOptimization:
		return "resource_optimization"
	case domain.ModeAudit:
		return "dependency_audit"
	case domain.ModeReview:
		return "security_review"
	case domain.ModeDesignToProduct:
		return "design_to_product"
	default:
		return "implementation"
	}
}

// EnsureRoot is exported so callers can prefer the workflow's resolver
// when CLI cwd differs from the project root.
func EnsureRoot(startDir string) (string, error) {
	return paths.FindProjectRoot(startDir)
}

// PromptOrError returns nil when prompt is non-empty, an error suitable
// for cobra's RunE otherwise.
func PromptOrError(prompt string) error {
	if strings.TrimSpace(prompt) == "" {
		return fmt.Errorf("prompt is empty")
	}
	return nil
}
