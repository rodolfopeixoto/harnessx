// SPDX-License-Identifier: MIT

// Package certify runs a deterministic adapter certification suite and
// returns a structured result + score.
package certify

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/ropeixoto/harnessx/internal/agents"
)

type CheckStatus string

const (
	StatusPassed  CheckStatus = "passed"
	StatusFailed  CheckStatus = "failed"
	StatusSkipped CheckStatus = "skipped"
)

type Check struct {
	Name   string      `json:"name"`
	Status CheckStatus `json:"status"`
	Detail string      `json:"detail,omitempty"`
}

type Result struct {
	AgentID    string  `json:"agent_id"`
	CLIVersion string  `json:"cli_version,omitempty"`
	Score      int     `json:"score"`
	Status     string  `json:"status"` // passed | partial | failed
	Checks     []Check `json:"checks"`
}

// DetailsJSON returns a compact JSON encoding suitable for the SQLite
// agent_certifications.details_json column.
func (r Result) DetailsJSON() string {
	b, err := json.Marshal(r)
	if err != nil {
		return "{}"
	}
	return string(b)
}

type Options struct {
	// SimplePrompt drives the "simple prompt works" check. Empty value
	// skips the check.
	SimplePrompt string
	// SkipRun skips Run-dependent checks. Useful when binaries are not
	// installed (e.g. cert preflight, CI without API keys).
	SkipRun bool
	// PerCheckTimeout bounds each network-bearing check.
	PerCheckTimeout time.Duration
}

// Run executes the certification suite against adapter a.
func Run(ctx context.Context, a agents.AgentAdapter, opts Options) Result {
	if opts.PerCheckTimeout <= 0 {
		opts.PerCheckTimeout = 10 * time.Second
	}
	r := Result{AgentID: a.ID()}

	// 1. Healthcheck — binary exists + authentication probe.
	hc := a.Healthcheck(ctx)
	r.CLIVersion = hc.CLIVersion
	if hc.OK {
		r.Checks = append(r.Checks, Check{Name: "healthcheck", Status: StatusPassed, Detail: hc.CLIVersion})
	} else {
		r.Checks = append(r.Checks, Check{Name: "healthcheck", Status: StatusFailed, Detail: combine(hc.Err, hc.Detail)})
	}

	// 2. Capabilities self-report sanity.
	caps := a.Capabilities()
	if caps.MaxContextTokens > 0 {
		r.Checks = append(r.Checks, Check{Name: "capabilities_reported", Status: StatusPassed,
			Detail: "max_context_tokens=" + intToStr(caps.MaxContextTokens)})
	} else {
		r.Checks = append(r.Checks, Check{Name: "capabilities_reported", Status: StatusFailed,
			Detail: "max_context_tokens is zero"})
	}

	if opts.SkipRun || !hc.OK {
		r.Checks = append(r.Checks, Check{Name: "simple_prompt", Status: StatusSkipped, Detail: "skipped: healthcheck failed or SkipRun set"})
		r.Checks = append(r.Checks, Check{Name: "output_parseable", Status: StatusSkipped})
		r.Checks = append(r.Checks, Check{Name: "timeout_enforced", Status: StatusSkipped})
		r.Checks = append(r.Checks, Check{Name: "cancellation_honored", Status: StatusSkipped})
	} else {
		// 3. Simple prompt works.
		rctx, cancel := context.WithTimeout(ctx, opts.PerCheckTimeout)
		prompt := opts.SimplePrompt
		if prompt == "" {
			prompt = "Reply with the single word OK."
		}
		res := a.Run(rctx, agents.AgentRequest{Prompt: prompt, Timeout: opts.PerCheckTimeout})
		cancel()
		if res.Err == nil && res.ExitCode == 0 {
			r.Checks = append(r.Checks, Check{Name: "simple_prompt", Status: StatusPassed, Detail: truncate(res.Output.FinalMessage, 80)})
			r.Checks = append(r.Checks, Check{Name: "output_parseable", Status: StatusPassed})
		} else {
			r.Checks = append(r.Checks, Check{Name: "simple_prompt", Status: StatusFailed, Detail: combine(errString(res.Err), truncate(string(res.Output.Stderr), 200))})
			r.Checks = append(r.Checks, Check{Name: "output_parseable", Status: StatusSkipped})
		}

		// 4. Timeout enforcement — request with a sub-second timeout against
		// a healthy adapter should still return promptly (not hang).
		tctx, tcancel := context.WithTimeout(ctx, 2*time.Second)
		tres := a.Run(tctx, agents.AgentRequest{Prompt: "noop", Timeout: 200 * time.Millisecond})
		tcancel()
		if tres.Duration < 2*time.Second {
			r.Checks = append(r.Checks, Check{Name: "timeout_enforced", Status: StatusPassed, Detail: tres.Duration.String()})
		} else {
			r.Checks = append(r.Checks, Check{Name: "timeout_enforced", Status: StatusFailed, Detail: "run exceeded 2s ceiling"})
		}

		// 5. Cancellation — cancelling immediately must return quickly.
		cctx, ccancel := context.WithCancel(ctx)
		ccancel()
		cres := a.Run(cctx, agents.AgentRequest{Prompt: "noop"})
		if cres.Duration < time.Second {
			r.Checks = append(r.Checks, Check{Name: "cancellation_honored", Status: StatusPassed, Detail: cres.Duration.String()})
		} else {
			r.Checks = append(r.Checks, Check{Name: "cancellation_honored", Status: StatusFailed, Detail: "cancelled run took ≥ 1s"})
		}
	}

	// 6. Failure classification self-test — feed synthetic stderr.
	fakeOut := agents.AgentOutput{Stderr: []byte("HTTP 429: rate limit exceeded")}
	cls := a.ClassifyFailure(fakeOut, 1, nil)
	if cls != agents.FailureNone {
		r.Checks = append(r.Checks, Check{Name: "failure_classification", Status: StatusPassed, Detail: string(cls)})
	} else {
		r.Checks = append(r.Checks, Check{Name: "failure_classification", Status: StatusFailed, Detail: "did not classify a rate-limit stderr"})
	}

	r.Score, r.Status = score(r.Checks)
	return r
}

func score(checks []Check) (int, string) {
	if len(checks) == 0 {
		return 0, "failed"
	}
	passed := 0
	failed := 0
	for _, c := range checks {
		switch c.Status {
		case StatusPassed:
			passed++
		case StatusFailed:
			failed++
		}
	}
	pct := passed * 100 / len(checks)
	switch {
	case failed == 0 && pct >= 100:
		return pct, "passed"
	case failed == 0:
		return pct, "passed"
	case pct >= 50:
		return pct, "partial"
	default:
		return pct, "failed"
	}
}

func combine(a, b string) string {
	switch {
	case a != "" && b != "":
		return a + " | " + b
	case a != "":
		return a
	default:
		return b
	}
}

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func intToStr(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

func truncate(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

func looksLikeAuth(s string) bool {
	low := strings.ToLower(s)
	for _, hint := range []string{"unauthorized", "not logged in", "401", "invalid api key", "authentication failed", "auth required"} {
		if strings.Contains(low, hint) {
			return true
		}
	}
	return false
}
