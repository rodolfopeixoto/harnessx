# Agents

Agents are coding CLIs (Claude Code, Codex, Gemini, Kimi, future tools) plugged in via the **adapter contract**. The HarnessX core never imports vendor-specific code — every interaction goes through `internal/agents.AgentAdapter`.

## Adapter contract

```go
type AgentAdapter interface {
    ID() string
    Name() string
    Capabilities() Capabilities
    Healthcheck(ctx context.Context) HealthcheckResult
    Run(ctx context.Context, req AgentRequest) AgentResult
    ParseUsage(output AgentOutput) Usage
    ClassifyFailure(output AgentOutput, exitCode int, runErr error) FailureType
}
```

## YAML adapter

Most adapters are declared in YAML, not Go. Drop a file under `.harness/config/agents/<id>.yaml` (project override) or use a bundled template via `harness agent add <id>`.

```yaml
id: meta
name: Meta CLI
enabled: true
type: cli

command:
  binary: meta-cli
  check: meta-cli --version

capabilities:
  text: true
  vision: false
  json_output: true
  max_context_tokens: 128000

models:
  default: meta-code-large
  cheap: meta-code-small

execution:
  prompt_mode: stdin   # stdin | arg
  working_directory: project
  timeout_seconds: 1800

run:
  args: ["run", "--model", "{{model}}", "--json"]

output:
  format: jsonl
  final_message_json_path: "$.message"
  usage_json_path: "$.usage"

failure_detection:
  rate_limit: ["rate limit", "quota exceeded"]
  context_limit: ["context length"]
  auth: ["unauthorized"]

cost:
  mode: estimated
  input_token_price_per_1m: 0.0
  output_token_price_per_1m: 0.0
```

### Template substitutions

The `run.args` list runs through string substitution before exec:

| token | replaced with |
|---|---|
| `{{model}}` | resolved model name (from `--model` flag or `models.default`) |
| `{{prompt}}` | request prompt (use when `execution.prompt_mode: arg`) |
| `{{working_dir}}` | absolute project root |
| `{{k}}` for any `k` in `AgentRequest.Extra` | extra hint bag |

### Output parsing

`output.format` is either `text` or `jsonl`. With `jsonl`, the adapter scans each line as JSON and applies `final_message_json_path` / `usage_json_path` (tiny dotted JSONPath subset — `$.a.b.c`). The last writer wins, matching how Claude Code / Codex stream final assistant turns at the end of a session.

### Failure classification

The bundled YAMLs ship vocabulary for `rate_limit`, `context_limit`, `auth`, and `transient`. `IsRecoverable()` returns true for those plus `timeout`. Auth and permanent failures stop the fallback chain immediately — the next agent in line will likely hit the same wall (or worse, succeed and mask a real config problem).

## Bundled adapters

| id | binary | strengths |
|---|---|---|
| `claude` | `claude` | planning, react_design_system, security_review, design_to_product |
| `codex` | `codex` | implementation, tests, rails_api |
| `gemini` | `gemini` | prompt_refinement, cheap_review, codebase_exploration |
| `kimi` | `kimi` | codebase_exploration, dependency_audit, cheap_review |
| `fake` | `fake-cli` | test fixture (use in CI / unit tests) |

List + override: `harness agent list`. Override the bundled YAML by adding the same `id` under `.harness/config/agents/`.

## Commands

```bash
harness agent list                    # bundled + project, with cert status
harness agent add <id>                # copy bundled YAML into project
harness agent discover <binary>       # print a YAML scaffold for an unknown CLI
harness agent certify <id> [--skip-run]   # run cert suite, persist row in agent_certifications
```

## Certification

`harness agent certify <id>` runs:

1. **healthcheck** — binary on PATH + version probe succeeds
2. **capabilities_reported** — `max_context_tokens > 0`
3. **simple_prompt** — round-trip "Reply OK" (skipped when `--skip-run`)
4. **output_parseable** — adapter resolved the final message
5. **timeout_enforced** — sub-second deadline returns promptly (no hang)
6. **cancellation_honored** — cancelled context exits within 1s
7. **failure_classification** — synthetic "HTTP 429" stderr classifies as `rate_limit`

The result is summarised (`passed | partial | failed`, score 0–100) and stored in `agent_certifications`. Use `--skip-run` on machines where the binary isn't installed; the cert still validates static config + classifier vocabulary.

## Adding a new CLI

1. Run `harness agent discover <binary>` and pipe to `.harness/config/agents/<id>.yaml`.
2. Edit the scaffold: pin `models`, set `output.*` JSON paths, populate `failure_detection`.
3. Run `harness agent certify <id>` until status reaches `passed`.
4. Wire the new id into `.harness/config/routes.yaml` (or pass `--agent <id>` to commands that accept it — coming with the router exposure in a future phase).

## Router + fallback

`internal/router` selects an agent per **task type** (`implementation`, `planning`, `security_review`, etc.) from a route config:

```yaml
routes:
  implementation:
    primary: codex
    fallback: [claude, gemini, kimi]
    budget_usd: 1.0
```

`Router.Select` returns the chain + explainable reasons. `Router.Execute` walks the chain, retrying on recoverable failures and recording `FallbackEvent` per attempt. Auth/permanent failures abort the chain immediately.
