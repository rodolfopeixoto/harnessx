// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

// topics indexes short in-CLI tutorials for the most-asked subjects.
// Each entry is a self-contained snippet the user can paste into their
// terminal to make progress without reading the full manual.
var topics = map[string]string{
	"quickstart": `Quickstart — 5 commands

  1. harness doctor                  # check tools + agent CLIs
  2. cd your-project && harness init # writes .harness/
  3. harness project add . --slug myapp
  4. harness feature "create HELLO.md with content: hi" \
         --agent fake-real --apply --autonomy safe_execute
  5. harness runs list

Then open the dashboard:
  harness dashboard --addr :7373
  open http://localhost:7373
`,
	"agents": `Agents — install + login + use

Adapters live in:
  templates/agents/{claude,codex,gemini,kimi,fake-real}.yaml

List + certify:
  harness agent list
  harness agent certify claude

Login per CLI (harness does NOT wrap login):
  claude login                          # Anthropic OAuth
  codex auth login                      # OpenAI OAuth
  gemini auth login                     # Google OAuth
  kimi login --provider moonshot        # Moonshot

Run with an agent:
  harness feature "<prompt>" --agent claude --apply --autonomy safe_execute
`,
	"sensors": `Sensors — what they are + how to run

Sensors are deterministic checks the executor runs after the agent
returns. A failed sensor blocks 'apply'.

  harness sensor list                  # bundled sensors
  harness sensor run --root .          # ad-hoc run
  harness runs sensors <run-id>        # per-run results

Common sensors: forbidden_files, forbidden_commands, secrets_scan,
changed_files, performance_budget, plus stack-specific packs.
`,
	"hooks": `Hooks — pre/post tool-use

A hook is an executable script under .harness/hooks/. The Executor
runs it before adapter.Run (pre-tool-use) and after sensors
(post-tool-use). Non-zero pre-hook blocks the run unless autonomy =
full_project_loop.

  mkdir -p .harness/hooks
  cat > .harness/hooks/pre-tool-use.sh <<'SH'
  #!/bin/bash
  echo "running for $HARNESS_RUN_ID"
  exit 0
  SH
  chmod +x .harness/hooks/pre-tool-use.sh
  harness hook scan
`,
	"autonomy": `Autonomy — five levels + per-path policy

Levels:
  manual                — every action denies or asks
  plan_and_ask          — plan ok, every execution asks
  safe_execute          — low-risk auto-apply; high-risk asks (default)
  full_project_loop     — low-risk auto; bypasses pre-hook block
  scheduled_maintenance — low-risk asks; high-risk denies

Per-project policy file .harness/config/autonomy.yaml:
  level: safe_execute
  allow_paths: ["src/**", "docs/**"]
  deny_paths:  ["secrets/**", ".env*", "infra/prod/**"]
  allow_commands: ["go test", "npm test"]
  deny_commands:  ["rm -rf /", "git push --force"]
`,
	"mcp": `MCP — install + scan + injection

  harness mcp install filesystem --command npx --yes
  harness mcp scan

When an adapter declares capabilities.mcp=true, the Executor merges
discovered MCPs into runs/<id>/mcp-config.json and appends
--mcp-config <path> to the agent invocation.
`,
	"update": `Update — channels + self-update

Channels:
  stable   — newest non-prerelease (default)
  beta     — includes pre-releases (vX.Y.Z-beta*, -rc*)
  develop  — clone develop branch + build from source (git + go)

  harness update                    # latest stable
  harness update --channel beta
  harness update --tag v0.4.0
  harness update --dry-run
  harness update status             # is something newer?
  harness update channels           # list available releases
`,
	"input": `Multi-input prompts

  harness feature "summarize" --prompt-file brief.md
  harness feature "summarize" --pdf brief.pdf      # needs pdftotext
  harness feature "redo this layout" --image mockup.png \
      --agent claude                                # vision-capable only
`,
	"tracker": `Tracker — metrics + audit

  harness metrics --since 1d --json
  harness metrics --since 7d
  harness audit --limit 20 --kind sensor

Aggregates per-run state from .harness/runs/*/meta.json
and the audit JSONL at .harness/audit/events.jsonl.
`,
	"billing": `Billing — Anthropic streams and which adapter hits which

Anthropic splits spending into three buckets (post 2026-06-15):

  Subscription usage        interactive claude in terminal/IDE, claude.ai chat
  Agent SDK monthly credit  claude -p / Agent SDK / GitHub Action
  Pay-as-you-go API         calls with an Anthropic API key (x-api-key)

How HarnessX adapters map:

  --agent claude            uses 'claude --print --output-format json' -> Agent SDK credit
                            ($20-$200/month depending on plan)
  --agent anthropic-api     direct API key -> pay-as-you-go, no monthly cap

Pick by workload:

  automation-heavy          use anthropic-api with --budget-usd per run
  exploration + a bit       use claude CLI, opt in Agent SDK credit at
                            https://console.anthropic.com -> plan settings

Set a per-run cap with --budget-usd 0.50. Watch harness metrics --since 1d.

Full breakdown: docs/anthropic-billing.md
`,
}

func newHelpCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "help [topic]",
		Short: "Show in-CLI tutorial for a topic (run with no args to list)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			if len(args) == 0 {
				keys := make([]string, 0, len(topics))
				for k := range topics {
					keys = append(keys, k)
				}
				sort.Strings(keys)
				fmt.Fprintln(out, "Topics:")
				for _, k := range keys {
					summary := firstLine(topics[k])
					fmt.Fprintf(out, "  %-12s %s\n", k, summary)
				}
				fmt.Fprintln(out, "\nRun: harness help <topic>")
				return nil
			}
			body, ok := topics[args[0]]
			if !ok {
				return fmt.Errorf("unknown topic %q; run 'harness help' to list", args[0])
			}
			fmt.Fprintln(out, body)
			return nil
		},
	}
	return c
}

func firstLine(s string) string {
	if i := strings.Index(s, "\n"); i > 0 {
		return s[:i]
	}
	return s
}
