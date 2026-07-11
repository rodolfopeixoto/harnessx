// SPDX-License-Identifier: MIT

package execution

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/oklog/ulid/v2"

	"github.com/ropeixoto/harnessx/internal/agents"
	"github.com/ropeixoto/harnessx/internal/autonomy"
	"github.com/ropeixoto/harnessx/internal/index"
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
func (e *DefaultExecutor) Execute(ctx context.Context, req Request) (Result, error) { //nolint:gocyclo // critical execution path; readability suffers when split arbitrarily
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

	if budgetErr := e.enforceBudget(ctx, req, wt, runDir, &res); budgetErr != nil {
		return res, budgetErr
	}

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

func (e *DefaultExecutor) enforceBudget(ctx context.Context, req Request, wt Worktree, runDir string, res *Result) error {
	if req.BudgetUSD <= 0 || res.EstimatedCostUSD <= req.BudgetUSD {
		return nil
	}
	res.Status = StatusAgentFailed
	res.ErrorType = "budget_exceeded"
	res.ErrorMessage = fmt.Sprintf("estimated cost $%.4f exceeded --budget-usd $%.4f", res.EstimatedCostUSD, req.BudgetUSD)
	res.FinishedAt = e.Clock()
	_ = e.Manager.Cleanup(ctx, wt)
	res.WorktreePath = ""
	res.ReportPath = writeReport(runDir, req, *res, res.ErrorMessage)
	writeMeta(runDir, *res)
	return fmt.Errorf("execution: %s", res.ErrorMessage)
}
