// SPDX-License-Identifier: MIT

package agentcmd

import (
	"fmt"
	"io"
	"strings"

	"github.com/ropeixoto/harnessx/internal/agents"
	"github.com/ropeixoto/harnessx/internal/agents/certify"
)

func renderCertification(out io.Writer, a agents.AgentAdapter, res certify.Result) {
	fmt.Fprintf(out, "Certification for %s — status: %s, score: %d/100\n", a.ID(), res.Status, res.Score)
	fmt.Fprintln(out)
	for _, c := range res.Checks {
		fmt.Fprintf(out, "  [%s] %s\n", iconFor(c.Status), c.Name)
		if desc := checkDescriptions[c.Name]; desc != "" {
			fmt.Fprintf(out, "       what: %s\n", desc)
		}
		if c.Detail != "" {
			fmt.Fprintf(out, "       detail: %s\n", c.Detail)
		}
		if c.Status == certify.StatusFailed {
			if hint := checkRemediation(c.Name, c.Detail, a); hint != "" {
				fmt.Fprintf(out, "       fix: %s\n", hint)
			}
		}
	}
	fmt.Fprintln(out)
	fmt.Fprintln(out, summariseCertification(a, res))
}

func iconFor(s certify.CheckStatus) string {
	switch s {
	case certify.StatusPassed:
		return "✓"
	case certify.StatusFailed:
		return "✗"
	default:
		return "·"
	}
}

var checkDescriptions = map[string]string{
	"healthcheck":            "binary on PATH + responds to --version",
	"capabilities_reported":  "adapter YAML reported its capabilities (context window, vision, tool-use)",
	"simple_prompt":          "round-trip with a one-line prompt; proves the CLI can answer",
	"output_parseable":       "JSON / text output matched the spec's final_message + usage paths",
	"timeout_enforced":       "adapter respects per-call timeout (cancels long runs)",
	"cancellation_honored":   "adapter aborts on ctx.Cancel within 1s",
	"failure_classification": "stderr regex maps to rate_limit / auth / context_limit / transient",
}

func checkRemediation(name, detail string, a agents.AgentAdapter) string {
	if r, ok := remediationByCheck[name]; ok {
		return r(detail, a)
	}
	return ""
}

var remediationByCheck = map[string]func(string, agents.AgentAdapter) string{
	"healthcheck":          remediateHealthcheck,
	"simple_prompt":        remediateSimplePrompt,
	"output_parseable":     func(string, agents.AgentAdapter) string { return remediateOutputParseable() },
	"timeout_enforced":     func(string, agents.AgentAdapter) string { return remediateTimeoutEnforced() },
	"cancellation_honored": func(string, agents.AgentAdapter) string { return remediateCancellationHonored() },
}

func remediateHealthcheck(_ string, a agents.AgentAdapter) string {
	if login := a.Capabilities().LoginCommand; login != "" {
		return "binary missing or unhealthy — run: " + login + " (docs: " + a.Capabilities().AuthDocURL + ")"
	}
	return "binary missing or unhealthy — install the CLI then retry: harness install " + a.ID()
}

func remediateSimplePrompt(detail string, a agents.AgentAdapter) string {
	low := strings.ToLower(detail)
	manualSmoke := "echo \"ping\" | " + a.ID() + " --print --output-format json"
	switch {
	case strings.Contains(low, "signal: killed"), strings.Contains(low, "context deadline"):
		return "CLI exceeded the simple_prompt timeout — bump it with --simple-timeout 180s.\n            manual smoke: " + manualSmoke
	case strings.Contains(low, "unauthorized"), strings.Contains(low, "401"), strings.Contains(low, "invalid api key"):
		return "auth failure — run: " + suggestLogin(a)
	case strings.Contains(low, "command not found"):
		return "CLI missing — run: harness install " + a.ID()
	case strings.Contains(low, "timeout"):
		return "CLI did not respond in time — bump --simple-timeout, then: " + manualSmoke
	default:
		return "smoke prompt failed — manual check: " + manualSmoke
	}
}

func remediateOutputParseable() string {
	return "CLI output did not match the YAML JSONPath — verify the adapter spec under .harness/config/agents/"
}

func remediateTimeoutEnforced() string {
	return "agent does not honour per-call timeout — open an upstream issue or pin a newer CLI version"
}

func remediateCancellationHonored() string {
	return "agent does not cancel on context cancel — pin a newer CLI version"
}

func suggestLogin(a agents.AgentAdapter) string {
	if login := a.Capabilities().LoginCommand; login != "" {
		return login
	}
	return a.ID() + " login"
}

func summariseCertification(a agents.AgentAdapter, res certify.Result) string {
	failed, skipped := countCheckStatuses(res)
	switch {
	case failed == 0 && skipped == 0:
		return "✓ ready — " + a.ID() + " is usable end-to-end."
	case failed == 0:
		return "✓ usable — " + a.ID() + " works; some optional checks were skipped."
	case failed == 1:
		return "partial — " + a.ID() + " usable for runs that do not need the failing check.\n" +
			"            Use harness feature ... --agent " + a.ID() + " --apply to try a real run.\n" +
			"            Re-run harness agent certify " + a.ID() + " after fixing the failure above."
	default:
		return "blocked — fix the failures above before using " + a.ID() + "."
	}
}

func countCheckStatuses(res certify.Result) (failed, skipped int) {
	for _, c := range res.Checks {
		switch c.Status {
		case certify.StatusFailed:
			failed++
		case certify.StatusSkipped:
			skipped++
		}
	}
	return failed, skipped
}
