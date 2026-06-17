// SPDX-License-Identifier: MIT

package orchestrate

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ropeixoto/harnessx/internal/agents"
)

func NewAdapterRunner(reg *agents.Registry, root string, timeout time.Duration) AdapterRunner {
	return func(ctx context.Context, step Step, prev []BlackboardEntry) (string, error) {
		if step.Adapter == "" {
			return "", errors.New("orchestrate: empty adapter id")
		}
		a, ok := reg.Get(step.Adapter)
		if !ok {
			return "", fmt.Errorf("orchestrate: adapter %q not registered", step.Adapter)
		}
		prompt := buildRolePrompt(step, prev)
		req := agents.AgentRequest{
			Prompt:     prompt,
			WorkingDir: root,
			Timeout:    timeout,
			Extra:      map[string]string{"role": string(step.Role)},
		}
		res := a.Run(ctx, req)
		if res.Err != nil {
			return res.Output.FinalMessage, res.Err
		}
		final := res.Output.FinalMessage
		if strings.TrimSpace(final) == "" {
			final = string(res.Output.Stdout)
		}
		return final, nil
	}
}

func buildRolePrompt(step Step, prev []BlackboardEntry) string {
	var b strings.Builder
	fmt.Fprintf(&b, "You are the %s role in a HarnessX orchestration (paper §4.1.1).\n", step.Role)
	if step.Prompt != "" {
		fmt.Fprintf(&b, "\nTask:\n%s\n", step.Prompt)
	}
	if len(prev) > 0 {
		b.WriteString("\nPrevious blackboard entries (most recent first):\n")
		for i := len(prev) - 1; i >= 0 && len(prev)-i <= 3; i-- {
			e := prev[i]
			fmt.Fprintf(&b, "- step %d (%s): %s\n  %s\n", e.Step, e.Role, e.Status, trim(e.Stdout, 400))
		}
	}
	b.WriteString("\nRespond with the artifact your role owns; keep output focused.\n")
	return b.String()
}

func trim(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
