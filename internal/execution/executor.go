// SPDX-License-Identifier: MIT

package execution

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

	prompt := req.Prompt
	if req.EnhancedPrompt != "" {
		prompt = req.EnhancedPrompt
	}
	agentReq := agents.AgentRequest{
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
			agentReq.Extra["mcp_config"] = path
			agentReq.ExtraArgs = append(agentReq.ExtraArgs, "--mcp-config", path)
		}
	}

	preHooks, _ := DispatchHooks(ctx, e.ProjectRoot, HookPreToolUse, []string{"HARNESS_RUN_ID=" + res.RunID, "HARNESS_AGENT=" + e.Adapter.ID()})
	res.Hooks = append(res.Hooks, preHooks...)
	if failures := FormatHookFailures(preHooks); failures != "" && req.Autonomy != AutonomyFullProjectLoop {
		res.Status = StatusAutonomyDenied
		res.ErrorType = "pre_hook_blocked"
		res.ErrorMessage = failures
		res.FinishedAt = e.Clock()
		_ = e.Manager.Cleanup(ctx, wt)
		res.WorktreePath = ""
		res.ReportPath = writeReport(runDir, req, res, "pre-tool-use hook blocked execution")
		writeMeta(runDir, res)
		return res, fmt.Errorf("pre-tool-use hook blocked: %s", failures)
	}

	agentRes := e.Adapter.Run(ctx, agentReq)
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
		res.Status = StatusAgentFailed
		res.ErrorType = string(agentRes.Failure)
		if agentRes.Err != nil {
			res.ErrorMessage = agentRes.Err.Error()
		}
		res.FinishedAt = e.Clock()
		_ = e.Manager.Cleanup(ctx, wt)
		res.WorktreePath = ""
		res.ReportPath = writeReport(runDir, req, res, "agent failed")
		writeMeta(runDir, res)
		return res, nil
	}

	changed, err := CaptureDiff(ctx, wt, runDir)
	if err != nil {
		res.Status = StatusAgentFailed
		res.ErrorType = "diff_capture"
		res.ErrorMessage = err.Error()
		res.FinishedAt = e.Clock()
		return res, err
	}
	res.ChangedFiles = changed
	if wt.Kind == "git_worktree" {
		res.DiffPath = filepath.Join(runDir, "diff.patch")
		res.DiffStatPath = filepath.Join(runDir, "diff-stat.txt")
	}
	res.ChangedFilesPath = filepath.Join(runDir, "changed-files.json")

	if len(changed) == 0 {
		res.Status = StatusNoChanges
		res.FinishedAt = e.Clock()
		_ = e.Manager.Cleanup(ctx, wt)
		res.WorktreePath = ""
		res.ReportPath = writeReport(runDir, req, res, "agent produced no changes")
		writeMeta(runDir, res)
		if req.Mode == ModeFeature || req.Mode == ModeBugfix {
			return res, fmt.Errorf("agent produced no changes for %s mode", req.Mode)
		}
		return res, nil
	}

	res.Sensors = RunSensors(ctx, e.Sensors, e.Profile, wt.Path, runDir)
	postHooks, _ := DispatchHooks(ctx, e.ProjectRoot, HookPostToolUse, []string{"HARNESS_RUN_ID=" + res.RunID, "HARNESS_AGENT=" + e.Adapter.ID()})
	res.Hooks = append(res.Hooks, postHooks...)
	risk := ClassifyRisk(changed)
	policy, _ := autonomy.LoadPolicy(e.ProjectRoot)
	dec, reason := GateApplyWithPolicy(req.Autonomy, risk, res.Sensors, policy, changed)

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
			res.Status = StatusAgentFailed
			res.ErrorType = "apply_failed"
			res.ErrorMessage = err.Error()
		} else {
			res.Status = StatusApplied
			_ = e.Manager.Cleanup(ctx, wt)
			res.WorktreePath = ""
		}
	}

	hasFailedSensor := false
	for _, s := range res.Sensors {
		if s.Status == "failed" {
			hasFailedSensor = true
			break
		}
	}
	if hasFailedSensor && res.Status != StatusAutonomyDenied {
		res.Status = StatusSensorFailed
	}

	res.FinishedAt = e.Clock()
	summary := fmt.Sprintf("status=%s files=%d risk=%s decision=%s", res.Status, len(changed), risk, dec)
	res.ReportPath = writeReport(runDir, req, res, summary)
	writeMeta(runDir, res)
	return res, nil
}

func writeMeta(runDir string, res Result) {
	path := filepath.Join(runDir, "meta.json")
	data, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return
	}
	_ = os.WriteFile(path, data, 0o644)
}

// ListRuns reads the .harness/runs directory and returns one Result per
// run, newest first by RunID.
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
			continue
		}
		out = append(out, r)
	}
	return out, nil
}

// LoadRun rehydrates the persisted Result for a single run id.
func LoadRun(projectRoot, runID string) (Result, error) {
	path := filepath.Join(projectRoot, ".harness", "runs", runID, "meta.json")
	data, err := os.ReadFile(path)
	if err != nil {
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
