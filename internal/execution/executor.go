// SPDX-License-Identifier: MIT

package execution

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/oklog/ulid/v2"

	"github.com/ropeixoto/harnessx/internal/agents"
	"github.com/ropeixoto/harnessx/internal/autonomy"
	"github.com/ropeixoto/harnessx/internal/hookscan"
	"github.com/ropeixoto/harnessx/internal/index"
	"github.com/ropeixoto/harnessx/internal/mcpscan"
	"github.com/ropeixoto/harnessx/internal/sensors"
)

// DefaultExecutor wires worktree -> agent -> diff -> sensors -> autonomy
// gate -> report. It does not know about Claude/Codex/Gemini specifically;
// the caller supplies an agents.AgentAdapter (built from a YAML spec).
type DefaultExecutor struct {
	ProjectRoot string
	Manager     *Manager
	Adapter     agents.AgentAdapter
	Sensors     []sensors.Sensor
	Profile     index.Profile
	Clock       func() time.Time
	IDGen       func() string
	// Status receives "calling <adapter>" / "<adapter> returned in <dur>"
	// notices around the adapter.Run call. nil = silent. Workflow wires
	// this to a stderr-writing closure so the operator sees a heartbeat
	// while the LLM is working.
	Status func(string)
	// LiveOut, when non-nil, is teed into the agent subprocess stdout/
	// stderr so the operator sees the CLI output in real time.
	LiveOut io.Writer
}

func NewDefaultExecutor(root string, adapter agents.AgentAdapter, ss []sensors.Sensor, p index.Profile) *DefaultExecutor {
	return &DefaultExecutor{
		ProjectRoot: root,
		Manager:     NewManager(root),
		Adapter:     adapter,
		Sensors:     ss,
		Profile:     p,
		Clock:       time.Now,
		IDGen:       newRunID,
	}
}

func newRunID() string {
	return "run_" + ulid.Make().String()
}

// Execute runs the loop once. PlanOnly skips agent invocation but still
// writes a report. DryRun keeps the diff in the worktree (no apply).
// Apply attempts to merge the worktree into the project root after gate
// allow.
func (e *DefaultExecutor) Execute(ctx context.Context, req Request) (Result, error) {
	if e.Adapter == nil {
		return Result{}, errors.New("execution: nil adapter")
	}
	if e.ProjectRoot == "" {
		return Result{}, errors.New("execution: empty project root")
	}
	res := Result{
		SessionID: req.SessionID,
		RunID:     e.IDGen(),
		AgentID:   e.Adapter.ID(),
		Status:    StatusRunning,
		StartedAt: e.Clock(),
	}
	runDir := filepath.Join(e.ProjectRoot, ".harness", "runs", res.RunID)
	if err := os.MkdirAll(runDir, 0o755); err != nil {
		return res, fmt.Errorf("execution: mkdir runDir: %w", err)
	}
	res.MCPDetectedNotActive, res.HooksDetectedNotActive = detectDeferrals(e.ProjectRoot)

	if req.PlanOnly {
		res.Status = StatusNoChanges
		res.FinishedAt = e.Clock()
		res.ReportPath = writeReport(runDir, req, res, "plan-only mode, agent not invoked")
		writeMeta(runDir, res)
		return res, nil
	}

	wt, err := e.Manager.Prepare(ctx, res.RunID)
	if err != nil {
		res.Status = StatusAgentFailed
		res.ErrorType = "worktree_prepare"
		res.ErrorMessage = err.Error()
		res.FinishedAt = e.Clock()
		return res, err
	}
	res.WorktreePath = wt.Path

	stdoutPath := filepath.Join(runDir, "stdout.log")
	stderrPath := filepath.Join(runDir, "stderr.log")
	res.StdoutPath = stdoutPath
	res.StderrPath = stderrPath

	agentReq := e.buildAgentRequest(req, wt, runDir, &res)

	if blocked, err := e.dispatchPreHooks(ctx, req, wt, runDir, &res); blocked {
		return res, err
	}

	agentRes := e.invokeAdapter(ctx, req, wt, agentReq)
	if err := os.WriteFile(stdoutPath, agentRes.Output.Stdout, 0o644); err != nil {
		return res, fmt.Errorf("execution: write stdout: %w", err)
	}
	if err := os.WriteFile(stderrPath, agentRes.Output.Stderr, 0o644); err != nil {
		return res, fmt.Errorf("execution: write stderr: %w", err)
	}
	res.InputTokens = agentRes.Usage.InputTokens
	res.OutputTokens = agentRes.Usage.OutputTokens
	res.EstimatedCostUSD = agentRes.Usage.EstimatedCostUSD
	res.ExactUsageAvailable = agentRes.Usage.Mode == "reported"

	if agentRes.Failure != agents.FailureNone || agentRes.Err != nil {
		e.finalizeAgentFailure(ctx, req, wt, runDir, &res, agentRes)
		return res, nil
	}

	changed, err := e.captureAndRecordDiff(ctx, wt, runDir, &res)
	if err != nil {
		return res, err
	}

	if len(changed) == 0 {
		return e.finalizeNoChanges(ctx, req, wt, runDir, &res)
	}

	res.Sensors = RunSensors(ctx, e.Sensors, e.Profile, wt.Path, runDir)
	postHooks, _ := DispatchHooks(ctx, e.ProjectRoot, HookPostToolUse, []string{"HARNESS_RUN_ID=" + res.RunID, "HARNESS_AGENT=" + e.Adapter.ID()})
	res.Hooks = append(res.Hooks, postHooks...)
	risk := ClassifyRisk(changed)
	policy, _ := autonomy.LoadPolicy(e.ProjectRoot)
	dec, reason := GateApplyWithPolicy(req.Autonomy, risk, res.Sensors, policy, changed)
	e.applyGate(ctx, req, wt, runDir, &res, dec, reason)
	if hasFailedSensor(res.Sensors) && res.Status != StatusAutonomyDenied {
		res.Status = StatusSensorFailed
	}

	res.FinishedAt = e.Clock()
	res.Verification.PromisedFilesUntouched = untouchedPromisedFiles(req.PromisedFiles, res.ChangedFiles)
	populateMetrics(&res)
	summary := fmt.Sprintf("status=%s files=%d risk=%s decision=%s", res.Status, len(changed), risk, dec)
	res.ReportPath = writeReport(runDir, req, res, summary)
	writeMeta(runDir, res)
	return res, nil
}

// untouchedPromisedFiles returns promised entries absent from changed.
// Matching is exact-path; callers normalise before passing.
func untouchedPromisedFiles(promised, changed []string) []string {
	if len(promised) == 0 {
		return nil
	}
	hit := make(map[string]struct{}, len(changed))
	for _, c := range changed {
		hit[c] = struct{}{}
	}
	var miss []string
	for _, p := range promised {
		if _, ok := hit[p]; !ok {
			miss = append(miss, p)
		}
	}
	return miss
}

// populateMetrics derives Trajectory / Verification / Recovery /
// Replayability from data the executor already collected. Pure
// function over res so unit tests can call it directly.
func populateMetrics(res *Result) {
	res.Trajectory.ToolCalls = len(res.Hooks)
	res.Trajectory.EditCount = len(res.ChangedFiles)
	if !res.FinishedAt.IsZero() && !res.StartedAt.IsZero() {
		res.Trajectory.WallMs = res.FinishedAt.Sub(res.StartedAt).Milliseconds()
	}
	res.Verification.SensorsRun = len(res.Sensors)
	for _, s := range res.Sensors {
		if s.Status == "passed" {
			res.Verification.SensorsPassed++
		}
	}
	res.Verification.OracleCount = len(res.Sensors)
	res.Replayability.EventsComplete = res.JSONLPath != "" || res.StdoutPath != "" || res.StderrPath != ""
}

func (e *DefaultExecutor) captureAndRecordDiff(ctx context.Context, wt Worktree, runDir string, res *Result) ([]string, error) {
	changed, err := CaptureDiff(ctx, wt, runDir)
	if err != nil {
		res.Status = StatusAgentFailed
		res.ErrorType = "diff_capture"
		res.ErrorMessage = err.Error()
		res.FinishedAt = e.Clock()
		return nil, err
	}
	res.ChangedFiles = changed
	if wt.Kind == "git_worktree" {
		res.DiffPath = filepath.Join(runDir, "diff.patch")
		res.DiffStatPath = filepath.Join(runDir, "diff-stat.txt")
	}
	res.ChangedFilesPath = filepath.Join(runDir, "changed-files.json")
	return changed, nil
}

func (e *DefaultExecutor) applyGate(ctx context.Context, req Request, wt Worktree, runDir string, res *Result, dec autonomy.Decision, reason string) {
	switch {
	case dec == autonomy.DecisionDeny:
		res.Status = StatusAutonomyDenied
		res.ErrorMessage = reason
		_ = e.Manager.Cleanup(ctx, wt)
		res.WorktreePath = ""
	case dec == autonomy.DecisionApproval, !req.Apply:
		res.Status = StatusWaitingApproval
	case req.Apply:
		if err := ApplyWorktreeDiff(ctx, e.ProjectRoot, wt, runDir); err != nil {
			if errors.Is(err, ErrApplyConflict) {
				res.Status = StatusConflict
				res.ErrorType = "apply_conflict"
				res.ErrorMessage = err.Error()
				// Keep worktree on disk so the user can rerun apply or
				// pull the rejected hunks manually.
			} else {
				res.Status = StatusAgentFailed
				res.ErrorType = "apply_failed"
				res.ErrorMessage = err.Error()
			}
		} else {
			res.Status = StatusApplied
			_ = e.Manager.Cleanup(ctx, wt)
			res.WorktreePath = ""
		}
	}
}

func hasFailedSensor(ss []SensorOutcome) bool {
	for _, s := range ss {
		if s.Status == "failed" {
			return true
		}
	}
	return false
}

func (e *DefaultExecutor) finalizeAgentFailure(ctx context.Context, req Request, wt Worktree, runDir string, res *Result, agentRes agents.AgentResult) {
	res.Status = StatusAgentFailed
	res.ErrorType = string(agentRes.Failure)
	if agentRes.Err != nil {
		res.ErrorMessage = agentRes.Err.Error()
	}
	res.FinishedAt = e.Clock()
	_ = e.Manager.Cleanup(ctx, wt)
	res.WorktreePath = ""
	res.ReportPath = writeReport(runDir, req, *res, "agent failed")
	writeMeta(runDir, *res)
}

func (e *DefaultExecutor) finalizeNoChanges(ctx context.Context, req Request, wt Worktree, runDir string, res *Result) (Result, error) {
	res.Status = StatusNoChanges
	res.FinishedAt = e.Clock()
	_ = e.Manager.Cleanup(ctx, wt)
	res.WorktreePath = ""
	res.ReportPath = writeReport(runDir, req, *res, "agent produced no changes")
	writeMeta(runDir, *res)
	if req.Mode == ModeFeature || req.Mode == ModeBugfix {
		return *res, fmt.Errorf("agent produced no changes for %s mode", req.Mode)
	}
	return *res, nil
}

func (e *DefaultExecutor) invokeAdapter(ctx context.Context, req Request, wt Worktree, agentReq agents.AgentRequest) agents.AgentResult {
	if e.Status != nil {
		e.Status("calling " + e.Adapter.ID() + "...")
	}
	if e.LiveOut != nil {
		agentReq.LiveOut = e.LiveOut
	}
	start := time.Now()
	var res agents.AgentResult
	if SandboxMode(req.Sandbox) != SandboxContainer {
		res = e.Adapter.Run(ctx, agentReq)
	} else {
		binary := e.Adapter.ID()
		sb := SandboxSpec{Mode: SandboxContainer, Image: req.SandboxImage}
		r, err := runInContainer(ctx, e.ProjectRoot, sb, wt, agentReq, binary)
		res = r
		if err != nil {
			res.Err = err
			res.Failure = agents.FailureTransient
		}
	}
	if e.Status != nil {
		e.Status(fmt.Sprintf("%s returned in %s", e.Adapter.ID(), time.Since(start).Round(time.Millisecond)))
	}
	return res
}

func (e *DefaultExecutor) buildAgentRequest(req Request, wt Worktree, runDir string, res *Result) agents.AgentRequest {
	prompt := req.Prompt
	if req.EnhancedPrompt != "" {
		prompt = req.EnhancedPrompt
	}
	r := agents.AgentRequest{
		Prompt:     prompt,
		Model:      req.Model,
		WorkingDir: wt.Path,
		Timeout:    5 * time.Minute,
		Extra:      map[string]string{},
	}
	if e.Adapter.Capabilities().MCP {
		if path, names, err := BuildMCPConfig(e.ProjectRoot, runDir); err == nil && path != "" {
			res.MCPConfigPath = path
			res.MCPInjected = names
			r.Extra["mcp_config"] = path
			r.ExtraArgs = append(r.ExtraArgs, "--mcp-config", path)
		}
	}
	return r
}

func (e *DefaultExecutor) dispatchPreHooks(ctx context.Context, req Request, wt Worktree, runDir string, res *Result) (bool, error) {
	preHooks, _ := DispatchHooks(ctx, e.ProjectRoot, HookPreToolUse,
		[]string{"HARNESS_RUN_ID=" + res.RunID, "HARNESS_AGENT=" + e.Adapter.ID()})
	res.Hooks = append(res.Hooks, preHooks...)
	failures := FormatHookFailures(preHooks)
	if failures == "" || req.Autonomy == AutonomyFullProjectLoop {
		return false, nil
	}
	res.Status = StatusAutonomyDenied
	res.ErrorType = "pre_hook_blocked"
	res.ErrorMessage = failures
	res.FinishedAt = e.Clock()
	_ = e.Manager.Cleanup(ctx, wt)
	res.WorktreePath = ""
	res.ReportPath = writeReport(runDir, req, *res, "pre-tool-use hook blocked execution")
	writeMeta(runDir, *res)
	return true, fmt.Errorf("pre-tool-use hook blocked: %s", failures)
}

func writeMeta(runDir string, res Result) {
	path := filepath.Join(runDir, "meta.json")
	data, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return
	}
	_ = os.WriteFile(path, data, 0o644)
}

var ErrRunIncomplete = errors.New("execution: run incomplete (meta.json missing)")

func ListRuns(projectRoot string) ([]Result, error) {
	dir := filepath.Join(projectRoot, ".harness", "runs")
	entries, err := os.ReadDir(dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	out := make([]Result, 0, len(entries))
	for i := len(entries) - 1; i >= 0; i-- {
		e := entries[i]
		if !e.IsDir() {
			continue
		}
		r, err := LoadRun(projectRoot, e.Name())
		if err != nil {
			if errors.Is(err, ErrRunIncomplete) {
				out = append(out, degradedRun(projectRoot, e.Name()))
			}
			continue
		}
		out = append(out, r)
	}
	return out, nil
}

func degradedRun(projectRoot, runID string) Result {
	r := Result{RunID: runID, Status: StatusIncomplete}
	runDir := filepath.Join(projectRoot, ".harness", "runs", runID)
	if st, err := os.Stat(runDir); err == nil {
		r.StartedAt = st.ModTime()
	}
	reportPath := filepath.Join(runDir, "report.md")
	if _, err := os.Stat(reportPath); err == nil {
		r.ReportPath = reportPath
	}
	return r
}

func LoadRun(projectRoot, runID string) (Result, error) {
	runDir := filepath.Join(projectRoot, ".harness", "runs", runID)
	if _, err := os.Stat(runDir); err != nil {
		return Result{}, err
	}
	path := filepath.Join(runDir, "meta.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Result{}, fmt.Errorf("%w: %s", ErrRunIncomplete, runID)
		}
		return Result{}, err
	}
	var r Result
	if err := json.Unmarshal(data, &r); err != nil {
		return Result{}, err
	}
	return r, nil
}

func detectDeferrals(root string) ([]string, []string) {
	var mcps, hooks []string
	if list, err := mcpscan.Scan(root); err == nil {
		for _, m := range list {
			mcps = append(mcps, m.Source+"/"+m.Name)
		}
	}
	if list, err := hookscan.Scan(root); err == nil {
		for _, h := range list {
			hooks = append(hooks, h.Source+"/"+h.Name)
		}
	}
	return mcps, hooks
}

func writeReport(runDir string, req Request, res Result, summary string) string {
	path := filepath.Join(runDir, "report.md")
	var b []byte
	b = append(b, []byte(fmt.Sprintf("# Run Report — %s\n\n", res.RunID))...)
	b = append(b, []byte(fmt.Sprintf("## Summary\n\n%s\n\n", summary))...)
	b = append(b, []byte(fmt.Sprintf("## Prompt\n\n```\n%s\n```\n\n", req.Prompt))...)
	b = append(b, []byte(fmt.Sprintf("## Mode\n\n%s\n\n", req.Mode))...)
	b = append(b, []byte(fmt.Sprintf("## Agent\n\n%s\n\n", res.AgentID))...)
	b = append(b, []byte(fmt.Sprintf("## Worktree\n\n%s\n\n", res.WorktreePath))...)
	if len(res.ChangedFiles) > 0 {
		b = append(b, []byte("## Changed Files\n\n")...)
		for _, c := range res.ChangedFiles {
			b = append(b, []byte("- "+c+"\n")...)
		}
		b = append(b, '\n')
	}
	if res.DiffPath != "" {
		b = append(b, []byte("## Diff\n\n`"+res.DiffPath+"`\n\n")...)
	}
	if len(res.Sensors) > 0 {
		b = append(b, []byte("## Sensors\n\n| ID | Status | Duration (ms) |\n|---|---|---|\n")...)
		for _, s := range res.Sensors {
			b = append(b, []byte(fmt.Sprintf("| %s | %s | %d |\n", s.ID, s.Status, s.DurationMs))...)
		}
		b = append(b, '\n')
	}
	b = append(b, []byte(fmt.Sprintf("## Cost and Tokens\n\nInput: %d | Output: %d | Estimated: $%.4f | Exact: %t\n\n",
		res.InputTokens, res.OutputTokens, res.EstimatedCostUSD, res.ExactUsageAvailable))...)
	if len(res.MCPInjected) > 0 {
		b = append(b, []byte(fmt.Sprintf("## MCP\n\nInjected %d MCP server(s) via %s\n\n", len(res.MCPInjected), res.MCPConfigPath))...)
		j, _ := json.MarshalIndent(res.MCPInjected, "", "  ")
		b = append(b, j...)
		b = append(b, '\n', '\n')
	} else if len(res.MCPDetectedNotActive) > 0 {
		b = append(b, []byte("## MCP\n\nDetected but not injected (adapter capability mcp=false)\n\n")...)
		j, _ := json.MarshalIndent(res.MCPDetectedNotActive, "", "  ")
		b = append(b, j...)
		b = append(b, '\n', '\n')
	}
	if len(res.Hooks) > 0 {
		b = append(b, []byte("## Hooks\n\n| Name | Event | Exit | Duration (ms) | Skipped |\n|---|---|---|---|---|\n")...)
		for _, h := range res.Hooks {
			b = append(b, []byte(fmt.Sprintf("| %s | %s | %d | %d | %t |\n", h.Name, h.Event, h.ExitCode, h.DurationMs, h.Skipped))...)
		}
		b = append(b, '\n')
	} else if len(res.HooksDetectedNotActive) > 0 {
		b = append(b, []byte("## Hooks\n\nDetected but no pre/post-tool-use scripts executable\n\n")...)
		j, _ := json.MarshalIndent(res.HooksDetectedNotActive, "", "  ")
		b = append(b, j...)
		b = append(b, '\n', '\n')
	}
	b = append(b, []byte(fmt.Sprintf("## Status\n\n%s\n\n", res.Status))...)
	if res.ErrorMessage != "" {
		b = append(b, []byte(fmt.Sprintf("## Error\n\n```\n%s: %s\n```\n", res.ErrorType, res.ErrorMessage))...)
	}
	_ = os.WriteFile(path, b, 0o644)
	return path
}
