# P31 — Real Agentic Execution Loop

## 1. Current state (trace, June 15 2026, develop tip + feature/p31-real-agentic-execution)

Infrastructure layer is complete. Agentic execution loop is structurally wired but does not produce real diff/report/sensor artifacts.

### 1.1 Command paths today

| Cmd | Entrypoint | App | Behaviour |
|---|---|---|---|
| `harness feature "<p>"` | `cmd/harness/cmd_workflow.go:91` → `newWorkflowCmd("feature", ..., workflow.Feature)` | `internal/app/workflow/workflow.go:102` `Feature()` | calls `planThenMaybeExecute()` |
| `harness bugfix "<p>"` | same factory, mode=`ModeBugfix` | same | same |
| `harness ask "<p>"` | `cmd/harness/cmd_workflow.go::newAskCmd` | `workflow.go::Ask` (line 51) | builds context pack, writes report, **never calls agent** |
| `harness run "<p>"` | same factory, no mode hint | same | classifier picks intent |

### 1.2 `planThenMaybeExecute` (workflow.go:113-211)

1. `intent.Classify(prompt)` (119)
2. `spec.NewFromPrompt` + `sp.Write` (137-138)
3. `hxcontext.Build` (146-152)
4. `plan.New` + `pl.Write` (154-169) — hardcodes `pl.SensorsToRun = [forbidden_files, forbidden_commands, secrets_scan, "stack rule pack"]`
5. Confirmation gate (182-188)
6. `executeAgents(ctx, rc, mode, prompt, budget, pack, out)` (192) — calls router.Select + adapter chain
7. `writeReport` (195-207) → `reportcmd.Build` writes `.harness/artifacts/reports/<run_id>.md`
8. `finishTelemetry` (209-210)

### 1.3 `executeAgents` (internal/app/workflow/execute.go:28-67)

Loads adapters via `agentcmd.LoadAll`, builds default routes, calls `r.Select(task)` for chain, then loops invoking `adapter.Run(ctx, AgentRequest)`. **Result is captured but**:
- no diff artifact is created from agent output
- no sensors are run after the agent returns
- no autonomy gate guards apply
- cost/tokens are recorded only if `usage_json_path` parses (real Claude needs auth → no JSON usage block → 0)

### 1.4 Agent adapter (internal/agents/yaml/adapter.go:85-128)

`Adapter.Run(ctx, req)`:
- substitutes `{{prompt}} {{model}} {{working_dir}}` into args
- spawns binary, optional stdin prompt mode
- extracts `$.result` from stdout as FinalMessage
- extracts `$.usage` for tokens
- classifies failure via regex from spec.FailureDetection

This is the right abstraction. Missing: it has no concept of "execution workspace", "produced diff", "tool calls".

### 1.5 Sensors (internal/sensors/)

`Sensor` interface + Runner exist. `ChangedFilesSensor` calls `git diff --name-only HEAD`. **Runner.Run is never called from workflow.** Sensors are listed in `pl.SensorsToRun` only as strings.

### 1.6 Autonomy (internal/autonomy/autonomy.go:11-106)

`Gate(level, op) Decision{Allow|RequireApproval|Deny}` defined. **Never called from workflow.**

### 1.7 Persistence (internal/adapters/sqlite/repo.go)

Run table cols: id, session_id, stage, agent, model, status, prompt_hash, context_hash, started_at, finished_at, latency_ms, input_tokens, cached_input_tokens, output_tokens, reasoning_tokens, estimated_cost_usd, exit_code, fallback_from, error_type.
Methods: CreateRun, FinishRun, UpdateRunCostAndTokens, WriteSensorResult, ListSensorResults. Sufficient for P31 with one schema addition (worktree_path, diff_path, report_path).

### 1.8 Worktree

Only ref: `internal/cleanup/detectors/worktrees.go` (cleanup detector). **No active worktree manager.** Must add.

### 1.9 Diff handling

`ChangedFilesSensor` captures `git diff --name-only`. **No unified diff patch artifact.** Must add.

### 1.10 Report

`internal/app/reportcmd/report.go:55-165` writes a markdown report. Sections already match what P31 needs. Wire it from new Executor.

### 1.11 MCP / hooks

Discovered via `internal/mcpscan` + `internal/hookscan` and exposed via HTTP + CLI. **Not passed into adapter.Run** and **not executed pre/post**. P31 leaves both as detected-but-not-active and records that fact in the report. P32 wires them.

## 2. Contract

```go
package execution

type Mode string
const (
    ModeFeature Mode = "feature"
    ModeBugfix  Mode = "bugfix"
    ModeAsk     Mode = "ask"
    ModeReview  Mode = "review"
)

type AutonomyLevel string

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
}

type SensorOutcome struct {
    ID       string
    Status   string // passed|failed|skipped|blocked|not_configured
    Output   string
    DurationMs int64
}

type Result struct {
    SessionID           string
    RunID               string
    AgentID             string
    Status              string // running|no_changes|waiting_approval|applied|sensor_failed|agent_failed|denied
    StartedAt           time.Time
    FinishedAt          time.Time
    WorktreePath        string
    StdoutPath          string
    StderrPath          string
    JSONLPath           string
    DiffPath            string
    DiffStatPath        string
    ChangedFilesPath    string
    ReportPath          string
    ChangedFiles        []string
    Sensors             []SensorOutcome
    InputTokens         int
    OutputTokens        int
    EstimatedCostUSD    float64
    ExactUsageAvailable bool
    MCPDetectedNotActive  []string
    HooksDetectedNotActive []string
    ErrorType           string
    ErrorMessage        string
}

type Executor interface {
    Execute(ctx context.Context, req Request) (Result, error)
}
```

## 3. Required CLI

```
harness feature "<p>" [--agent claude|fake|codex|gemini] [--dry-run|--apply|--plan-only]
                      [--autonomy manual|plan_and_ask|safe_execute|full_project_loop]
                      [--budget-usd 1.00] [--json]
harness bugfix  "<p>"  (same flags)
harness ask     "<p>"  (--agent + --json, plan-only forced)
harness run list
harness run inspect <run-id>
harness run approve <run-id>
harness run discard <run-id>
harness run sensors <run-id>
harness run report  <run-id>
```

`--dry-run` (default for `feature` and `bugfix` unless `--apply`): worktree created, agent invoked, diff captured, sensors run, but no merge to project root.

`--apply`: after diff + sensors pass + autonomy.Gate=Allow → apply patch to project root.

`--plan-only`: skip agent invocation; just spec + plan + report.

## 4. New packages / files

```
internal/execution/
  types.go          // Request, Result, SensorOutcome, Executor
  worktree.go       // git worktree add/remove under .harness/worktrees/<run-id>
  capture.go        // tee stdout/stderr/jsonl to files under .harness/runs/<run-id>
  diff.go           // git diff > diff.patch, git diff --stat, changed-files.json
  sensors.go        // bridge sensors.Runner -> []SensorOutcome
  gate.go           // call autonomy.Gate based on changed-file risk classifier
  executor.go       // orchestrate the loop
  apply.go          // git apply / commit / merge to project root

cmd/harness/cmd_run.go        // run list|inspect|approve|discard|sensors|report
internal/app/runcmd/runcmd.go // app layer for run subcommands

templates/agents/fake.yaml    // points at bundled fake binary
cmd/fake-agent/main.go        // deterministic test agent: reads stdin prompt,
                              // if prompt matches /create.*fixture/ writes a file
internal/adapters/sqlite/migrations/0003_run_artifacts.sql
                              // add columns: worktree_path, diff_path, report_path, autonomy_decision
```

## 5. Test plan

Unit (`internal/execution/*_test.go`):
- Request validation (mode + agent + autonomy)
- worktree create/remove idempotent
- diff detection on staged + unstaged changes
- sensor outcome aggregation
- gate decision matrix (level × op × risk)

Integration with fake agent (no real Claude):
- fixture repo, fake agent produces diff → status=waiting_approval (dry-run) or applied (--apply + safe_execute)
- fake agent produces no diff → status=no_changes, exit non-zero for feature mode
- fake agent fails → status=agent_failed
- sensors fail → status=sensor_failed, gate blocks apply
- autonomy=manual → never auto-apply

E2E shell smoke `scripts/e2e-phase31.sh`:
```
./bin/harness feature "create tmp/fixture.md saying hello" --agent fake --apply
./bin/harness run list      # 1 row
./bin/harness run inspect <run-id>
./bin/harness run report <run-id>
test -f tmp/fixture.md
```

Real Claude behind env flag: `HARNESS_REAL_CLAUDE=1 go test ./internal/execution -run RealClaude` (local only).

## 6. Risks + mitigations

| Risk | Mitigation |
|---|---|
| Worktree leaks if process killed | Worktree dir naming includes run-id; `harness run discard` removes; cleanup detector already catches strays |
| Agent prints diff to stdout but doesn't write files | Two-phase: capture stdout; if no file-system diff but stdout contains unified-diff markers, try `git apply` from stdout |
| `git apply` partial failure | Stage to worktree first; reject patch with non-zero exit; never auto-apply to project root |
| Sensor false-negative passes destructive change | Risk classifier marks Dockerfile/migrations/secrets/deps as high-risk → forced approval regardless of autonomy level |
| Real Claude requires auth interactively | Adapter healthcheck during certify; if `signal: killed` on simple_prompt, executor surfaces `error_type=auth_required` with actionable next step |
| Cost telemetry inaccurate | Mark `ExactUsageAvailable=false` when JSON usage not parsed; report shows "estimated" |

## 7. Rollback

P31 lives behind no flag (feature flag would mean two code paths to maintain). If reverted, drop branch + revert migration `0003_run_artifacts.sql`. Run rows from P31 stay readable by older binary (new columns nullable).

## 8. MCP / hooks (explicit deferral)

P31 detects MCPs and hooks during executor setup and records them in `Result.MCPDetectedNotActive` / `HooksDetectedNotActive`. The report includes the line:

> MCP configs detected but not injected yet. P32 required.
> Hooks detected but not executed yet. P32 required.

No claim is made that they are wired.

## 9. Out of scope

- Codex / Gemini / Kimi adapter polish (contract works, parsing tuning deferred)
- Inspector tabs (P33)
- ActionService backend persistence (P33)
- Streaming SSE for Active Run page (P33)
- Memory promotion UI surface (P33)

## 10. Acceptance

P31 done when every criterion in the original prompt's "Acceptance criteria" block passes locally:

1. feature command invokes adapter for real
2. fake-agent E2E produces real diff
3. claude adapter invokable when authenticated
4. stdout/stderr/jsonl captured
5. no_changes detected, non-zero exit for feature/bugfix
6. diff.patch generated
7. sensors run after diff
8. autonomy gate blocks unsafe apply
9. report generated
10. run inspect shows paths + status
11. terminal output actionable
12. dashboard surfaces run/report/sensor
13. tests cover success/failure/no-change
14. MCP/hooks reported as detected-but-not-active
