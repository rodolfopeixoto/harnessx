// SPDX-License-Identifier: MIT

package yaml

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/ropeixoto/harnessx/internal/agents"
	"github.com/ropeixoto/harnessx/internal/agents/limits"
)

// Adapter wraps a YAML spec, executing the configured CLI binary on Run.
// All Runner injection points exist so tests can drive the adapter without
// spawning real processes.
type Adapter struct {
	Spec   Spec
	Lookup func(string) (string, error)
	Runner Runner
}

// Runner abstracts process execution. The default implementation uses
// os/exec and respects the request's timeout via context.
type Runner interface {
	Run(ctx context.Context, name string, args []string, stdin string, workingDir string) (stdout, stderr []byte, exitCode int, err error)
}

type defaultRunner struct{}

func runStreamed(ctx context.Context, name string, args []string, stdin, workingDir string, live io.Writer) ([]byte, []byte, int, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	if workingDir != "" {
		cmd.Dir = workingDir
	}
	if stdin != "" {
		cmd.Stdin = strings.NewReader(stdin)
	}
	var out, errb bytes.Buffer
	cmd.Stdout = io.MultiWriter(&out, live)
	// Drop known noisy upstream-CLI stderr lines from the live UI but
	// still capture them in errb for `harness logs` debugging.
	cmd.Stderr = io.MultiWriter(&errb, newFilteringWriter(live))
	err := cmd.Run()
	exitCode := 0
	if err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			exitCode = ee.ExitCode()
		} else {
			exitCode = -1
		}
	}
	return out.Bytes(), errb.Bytes(), exitCode, err
}

func (defaultRunner) Run(ctx context.Context, name string, args []string, stdin string, workingDir string) ([]byte, []byte, int, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	if workingDir != "" {
		cmd.Dir = workingDir
	}
	if stdin != "" {
		cmd.Stdin = strings.NewReader(stdin)
	}
	var out, errb bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errb
	err := cmd.Run()
	exitCode := 0
	if err != nil {
		exitErr := new(exec.ExitError)
		if errors.As(err, &exitErr) {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = -1
		}
	}
	return out.Bytes(), errb.Bytes(), exitCode, err
}

func New(s Spec) *Adapter {
	return &Adapter{Spec: s, Lookup: exec.LookPath, Runner: defaultRunner{}}
}

// maybeSanitiseSkills truncates SKILL.md descriptions that exceed the
// upstream CLI's hard parser limit (codex_core rejects > 1024 chars).
func maybeSanitiseSkills(adapterID string, req agents.AgentRequest) {
	cap := limits.ForAdapter(adapterID).MaxSkillDescriptionChars
	if cap <= 0 {
		return
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}
	roots := []string{filepath.Join(home, ".agents", "skills")}
	wd := req.WorkingDir
	if wd == "" {
		wd, _ = os.Getwd()
	}
	if wd == "" {
		return
	}
	outDir := filepath.Join(wd, ".harness", "runs", "_skills", adapterID)
	_, reports, _, prepErr := limits.PrepareSkills(adapterID, roots, outDir)
	if prepErr != nil {
		return
	}
	if req.LiveOut != nil && len(reports) > 0 {
		limits.WriteReport(req.LiveOut, adapterID, reports)
	}
}

func (a *Adapter) ID() string                        { return a.Spec.ID }
func (a *Adapter) Name() string                      { return a.Spec.Name }
func (a *Adapter) Capabilities() agents.Capabilities { return a.Spec.Capabilities }

func (a *Adapter) Healthcheck(ctx context.Context) agents.HealthcheckResult {
	if _, err := a.Lookup(a.Spec.Command.Binary); err != nil {
		return agents.HealthcheckResult{OK: false, Err: "binary not on PATH: " + a.Spec.Command.Binary}
	}
	check := a.Spec.Command.Check
	if check == "" {
		return agents.HealthcheckResult{OK: true, Detail: "binary present (no check command configured)"}
	}
	parts := strings.Fields(check)
	rctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	out, errb, code, err := a.Runner.Run(rctx, parts[0], parts[1:], "", "")
	if err != nil || code != 0 {
		return agents.HealthcheckResult{OK: false, Err: trimLine(string(errb)), Detail: trimLine(string(out))}
	}
	return agents.HealthcheckResult{OK: true, CLIVersion: trimLine(string(out))}
}

func (a *Adapter) Run(ctx context.Context, req agents.AgentRequest) agents.AgentResult {
	start := time.Now()
	timeout := req.Timeout
	if timeout <= 0 {
		timeout = time.Duration(maxInt(a.Spec.Execution.TimeoutSeconds, 30)) * time.Second
	}
	rctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	maybeSanitiseSkills(a.Spec.ID, req)

	args := substituteArgs(a.Spec.Run.Args, req)
	args = append(args, req.ExtraArgs...)
	stdin := ""
	if a.Spec.Execution.PromptMode == "" || a.Spec.Execution.PromptMode == "stdin" {
		stdin = req.Stdin
		if stdin == "" {
			stdin = req.Prompt
		}
	}
	wd := req.WorkingDir
	if wd == "" && a.Spec.Execution.WorkingDirectory == "project" {
		wd = "" // caller passes project root via req.WorkingDir; leave empty to inherit cwd
	}

	var stdout, stderr []byte
	var code int
	var err error
	live := req.LiveOut
	var jsonFmt *jsonStreamFormatter
	if live != nil && jsonFormat(a.Spec.Output.Format) {
		jsonFmt = newJSONStreamFormatter(live)
		live = jsonFmt
	}
	if live != nil {
		stdout, stderr, code, err = runStreamed(rctx, a.Spec.Command.Binary, args, stdin, wd, live)
		if jsonFmt != nil {
			jsonFmt.Flush()
		}
	} else {
		stdout, stderr, code, err = a.Runner.Run(rctx, a.Spec.Command.Binary, args, stdin, wd)
	}
	out := agents.AgentOutput{Stdout: stdout, Stderr: stderr}
	paths := finalMessagePaths(a.Spec.Output.FinalMessageJSONPath, a.Spec.Output.FinalMessageJSONPaths)
	for _, p := range paths {
		if v := extractJSONPath(stdout, a.Spec.Output.Format, p); v != "" {
			out.FinalMessage = v
			break
		}
	}
	if out.FinalMessage == "" {
		out.FinalMessage = trimLine(string(stdout))
	}

	failure := a.ClassifyFailure(out, code, err)
	if failure == agents.FailureNone && errors.Is(rctx.Err(), context.DeadlineExceeded) {
		failure = agents.FailureTimeout
	}

	res := agents.AgentResult{
		Output: out, ExitCode: code, Err: err,
		Duration: time.Since(start), Failure: failure,
	}
	res.Usage = a.ParseUsage(out)
	res.Usage.EstimatedCostUSD = a.estimateCost(res.Usage)
	return res
}

func (a *Adapter) ParseUsage(output agents.AgentOutput) agents.Usage {
	u := agents.Usage{Mode: "estimated"}
	usagePaths := finalMessagePaths(a.Spec.Output.UsageJSONPath, a.Spec.Output.UsageJSONPaths)
	for _, up := range usagePaths {
		raw := extractJSONPath(output.Stdout, a.Spec.Output.Format, up)
		if raw == "" {
			continue
		}
		var m map[string]any
		if err := json.Unmarshal([]byte(raw), &m); err != nil {
			continue
		}
		u.InputTokens = readInt(m, "input_tokens", "prompt_tokens", "in_tokens")
		u.OutputTokens = readInt(m, "output_tokens", "completion_tokens", "out_tokens")
		u.CachedInputTokens = readInt(m, "cached_input_tokens", "cache_read_input_tokens")
		u.ReasoningTokens = readInt(m, "reasoning_tokens")
		if u.InputTokens > 0 || u.OutputTokens > 0 {
			u.Mode = "reported"
			return u
		}
	}
	// Fallback estimate: ~4 chars per token (heuristic).
	u.InputTokens = len(output.Stdout) / 4
	u.OutputTokens = len(output.FinalMessage) / 4
	return u
}

func (a *Adapter) ClassifyFailure(output agents.AgentOutput, exitCode int, runErr error) agents.FailureType {
	if runErr == nil && exitCode == 0 {
		return agents.FailureNone
	}
	body := strings.ToLower(string(output.Stderr) + "\n" + string(output.Stdout))
	for cat, needles := range a.Spec.FailureDetection {
		for _, n := range needles {
			if strings.Contains(body, strings.ToLower(n)) {
				switch cat {
				case "rate_limit":
					return agents.FailureRateLimit
				case "context_limit":
					return agents.FailureContextLimit
				case "auth":
					return agents.FailureAuth
				case "transient":
					return agents.FailureTransient
				}
			}
		}
	}
	if errors.Is(runErr, context.DeadlineExceeded) {
		return agents.FailureTimeout
	}
	return agents.FailurePermanent
}

func (a *Adapter) estimateCost(u agents.Usage) float64 {
	in := float64(u.InputTokens) / 1_000_000.0
	out := float64(u.OutputTokens) / 1_000_000.0
	return in*a.Spec.Cost.InputTokenPricePer1M + out*a.Spec.Cost.OutputTokenPricePer1M
}

func finalMessagePaths(primary string, fallbacks []string) []string {
	out := make([]string, 0, 1+len(fallbacks))
	if primary != "" {
		out = append(out, primary)
	}
	for _, p := range fallbacks {
		if p == "" {
			continue
		}
		out = append(out, p)
	}
	return out
}

func substituteArgs(args []string, req agents.AgentRequest) []string {
	out := make([]string, len(args))
	for i, a := range args {
		s := a
		s = strings.ReplaceAll(s, "{{model}}", req.Model)
		s = strings.ReplaceAll(s, "{{prompt}}", req.Prompt)
		s = strings.ReplaceAll(s, "{{working_dir}}", req.WorkingDir)
		for k, v := range req.Extra {
			s = strings.ReplaceAll(s, "{{"+k+"}}", v)
		}
		out[i] = s
	}
	return out
}

// extractJSONPath implements a deliberately small subset of JSONPath used
// by adapter outputs: dotted paths starting with $., e.g. "$.usage.input_tokens".
// Returns the value's JSON encoding (string-encoded JSON) so callers can
// re-unmarshal as needed.
func extractJSONPath(data []byte, format, path string) string {
	if len(data) == 0 || path == "" {
		return ""
	}
	path = strings.TrimPrefix(path, "$.")
	parts := strings.Split(path, ".")

	// For JSONL output, scan lines and try each as a JSON object until one
	// resolves the path. Last writer wins (mimics the final-message convention).
	if format == "jsonl" {
		var last string
		for _, line := range bytes.Split(data, []byte("\n")) {
			if v := walkJSON(line, parts); v != "" {
				last = v
			}
		}
		return last
	}
	return walkJSON(data, parts)
}

func walkJSON(b []byte, parts []string) string {
	if len(b) == 0 {
		return ""
	}
	var cur any
	if err := json.Unmarshal(b, &cur); err != nil {
		return ""
	}
	for _, p := range parts {
		m, ok := cur.(map[string]any)
		if !ok {
			return ""
		}
		cur, ok = m[p]
		if !ok {
			return ""
		}
	}
	if s, ok := cur.(string); ok {
		return s
	}
	out, err := json.Marshal(cur)
	if err != nil {
		return ""
	}
	return string(out)
}

func readInt(m map[string]any, keys ...string) int {
	for _, k := range keys {
		if v, ok := m[k]; ok {
			switch n := v.(type) {
			case float64:
				return int(n)
			case int:
				return n
			case int64:
				return int(n)
			}
		}
	}
	return 0
}

func trimLine(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		s = s[:i]
	}
	return s
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

var _ io.Reader = (*strings.Reader)(nil)
