// SPDX-License-Identifier: MIT

// Package execution wires real agentic execution: prompt -> agent -> diff ->
// sensors -> autonomy gate -> apply or wait-for-approval -> report.
//
// P31 scope: build the contract, worktree manager, stdout/stderr/jsonl
// capture, sensors bridge, autonomy gate hook, and a deterministic fake
// adapter exercising the loop end-to-end. Real Claude execution rides on
// the same Executor when authenticated locally.
//
// MCP injection and hook dispatch are explicit P32 work and are surfaced
// here only as detected-but-not-active diagnostics on Result.
package execution

import (
	"context"
	"time"
)

type Mode string

const (
	ModeFeature Mode = "feature"
	ModeBugfix  Mode = "bugfix"
	ModeAsk     Mode = "ask"
	ModeReview  Mode = "review"
)

type AutonomyLevel string

const (
	AutonomyManual          AutonomyLevel = "manual"
	AutonomyPlanAndAsk      AutonomyLevel = "plan_and_ask"
	AutonomySafeExecute     AutonomyLevel = "safe_execute"
	AutonomyFullProjectLoop AutonomyLevel = "full_project_loop"
	AutonomyScheduled       AutonomyLevel = "scheduled_maintenance"
)

type Status string

const (
	StatusRunning         Status = "running"
	StatusNoChanges       Status = "no_changes"
	StatusWaitingApproval Status = "waiting_approval"
	StatusApplied         Status = "applied"
	StatusDiscarded       Status = "discarded"
	StatusSensorFailed    Status = "sensor_failed"
	StatusAgentFailed     Status = "agent_failed"
	StatusAutonomyDenied  Status = "autonomy_denied"
	// StatusConflict marks runs whose diff failed `git apply --check` —
	// the working tree was NOT modified. Audit BUG-13/14: previously such
	// runs falsely reported `applied` and dropped conflict markers in the
	// source files. The run dir keeps the original patch under
	// rejects/diff.patch so the user can review or apply manually.
	StatusConflict Status = "conflict"
)

type Request struct {
	SessionID     string
	ProjectID     string
	ProjectPath   string
	Prompt        string
	Mode          Mode
	AgentID       string
	DryRun        bool
	Apply         bool
	PlanOnly      bool
	Autonomy      AutonomyLevel
	BudgetUSD     float64
	ContextPackID string
	SpecPath      string
	PlanPath      string
	// Model picked by the router (e.g. claude-haiku-4-5 for trivial,
	// claude-opus-4-7 for complex). Empty -> adapter default.
	Model string
	// EnhancedPrompt overrides Prompt when non-empty; the executor still
	// records the original Prompt on the Result for audit.
	EnhancedPrompt string
	// Sandbox = host (default) runs the adapter on the host directly;
	// container runs it inside the project's selected runtime via the
	// internal/runtime/containers abstraction with the worktree
	// bind-mounted into /work.
	Sandbox      string
	SandboxImage string
	// PromisedFiles (audit BUG-18) is the optional list of files the
	// caller has reason to believe should change. The executor sets
	// Result.Verification.PromisedFilesUntouched to any of these that
	// don't appear in ChangedFiles after apply.
	PromisedFiles []string
}

type SensorOutcome struct {
	ID         string `json:"id"`
	Status     string `json:"status"`
	Output     string `json:"output,omitempty"`
	DurationMs int64  `json:"duration_ms"`
}

// Trajectory captures process-effort signals so operators can answer
// "was this run efficient?" (paper § 5.2.1 "evaluation beyond task
// completion"). Populated automatically by the executor.
type Trajectory struct {
	ToolCalls int   `json:"tool_calls"`
	EditCount int   `json:"edit_count"`
	WallMs    int64 `json:"wall_ms"`
}

// Verification reports oracle adequacy.
type Verification struct {
	SensorsRun    int `json:"sensors_run"`
	SensorsPassed int `json:"sensors_passed"`
	OracleCount   int `json:"oracle_count"`
	// PromisedFilesUntouched (audit BUG-18) lists files the prompt or
	// plan said would change but that don't appear in ChangedFiles.
	// Non-empty means the run was billed as `applied` but its scope
	// was narrower than the user asked for — callers should re-prompt
	// or treat the status as `incomplete`.
	PromisedFilesUntouched []string `json:"promised_files_untouched,omitempty"`
}

// Recovery reports how much the run wobbled.
type Recovery struct {
	Retries     int `json:"retries"`
	Regressions int `json:"regressions"`
}

// Replayability flags whether the run's event stream is complete
// enough to replay later.
type Replayability struct {
	EventsComplete bool `json:"events_complete"`
}

type Result struct {
	SessionID              string          `json:"session_id"`
	RunID                  string          `json:"run_id"`
	AgentID                string          `json:"agent_id"`
	Status                 Status          `json:"status"`
	StartedAt              time.Time       `json:"started_at"`
	FinishedAt             time.Time       `json:"finished_at"`
	WorktreePath           string          `json:"worktree_path,omitempty"`
	StdoutPath             string          `json:"stdout_path,omitempty"`
	StderrPath             string          `json:"stderr_path,omitempty"`
	JSONLPath              string          `json:"jsonl_path,omitempty"`
	DiffPath               string          `json:"diff_path,omitempty"`
	DiffStatPath           string          `json:"diff_stat_path,omitempty"`
	ChangedFilesPath       string          `json:"changed_files_path,omitempty"`
	ReportPath             string          `json:"report_path,omitempty"`
	ChangedFiles           []string        `json:"changed_files,omitempty"`
	Sensors                []SensorOutcome `json:"sensors,omitempty"`
	InputTokens            int             `json:"input_tokens"`
	OutputTokens           int             `json:"output_tokens"`
	EstimatedCostUSD       float64         `json:"estimated_cost_usd"`
	ExactUsageAvailable    bool            `json:"exact_usage_available"`
	MCPDetectedNotActive   []string        `json:"mcp_detected_not_active,omitempty"`
	HooksDetectedNotActive []string        `json:"hooks_detected_not_active,omitempty"`
	MCPInjected            []string        `json:"mcp_injected,omitempty"`
	MCPConfigPath          string          `json:"mcp_config_path,omitempty"`
	Hooks                  []HookOutcome   `json:"hooks,omitempty"`
	ErrorType              string          `json:"error_type,omitempty"`
	ErrorMessage           string          `json:"error_message,omitempty"`
	Trajectory             Trajectory      `json:"trajectory"`
	Verification           Verification    `json:"verification"`
	Recovery               Recovery        `json:"recovery"`
	Replayability          Replayability   `json:"replayability"`
}

type Executor interface {
	Execute(ctx context.Context, req Request) (Result, error)
}
