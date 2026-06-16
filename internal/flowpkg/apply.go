// SPDX-License-Identifier: MIT

package flowpkg

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"time"
)

type ApplyOptions struct {
	Root string
	Dry  bool
}

type PhaseResult struct {
	Name    string
	Skipped bool
	Output  string
	Err     error
}

func Apply(ctx context.Context, f Flow, opts ApplyOptions, out io.Writer) ([]PhaseResult, error) {
	if err := f.Validate(); err != nil {
		return nil, err
	}
	var results []PhaseResult
	for _, p := range f.Phases {
		fmt.Fprintf(out, "→ phase %s (kind=%s)\n", p.Name, p.Kind)
		if opts.Dry {
			results = append(results, PhaseResult{Name: p.Name, Skipped: true})
			continue
		}
		results = append(results, runPhase(ctx, p, opts.Root))
	}
	return results, nil
}

func runPhase(ctx context.Context, p Phase, root string) PhaseResult {
	switch p.Kind {
	case PhaseDeterministic:
		return runShell(ctx, p, root)
	case PhaseLLM:
		return PhaseResult{Name: p.Name, Skipped: true, Output: "llm phase wiring deferred to v0.94"}
	case PhaseSensor:
		return PhaseResult{Name: p.Name, Skipped: true, Output: "sensor phase wiring deferred to v0.94"}
	}
	return PhaseResult{Name: p.Name, Err: fmt.Errorf("unknown kind %q", p.Kind)}
}

func runShell(ctx context.Context, p Phase, root string) PhaseResult {
	if len(p.Cmd) == 0 {
		return PhaseResult{Name: p.Name, Err: fmt.Errorf("deterministic phase %q has empty cmd", p.Name)}
	}
	timeout := time.Duration(p.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 60 * time.Second
	}
	cctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	cmd := exec.CommandContext(cctx, p.Cmd[0], p.Cmd[1:]...)
	cmd.Dir = root
	body, err := cmd.CombinedOutput()
	return PhaseResult{Name: p.Name, Output: string(body), Err: err}
}
