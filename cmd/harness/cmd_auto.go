// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/ropeixoto/harnessx/internal/platform/ids"
	"github.com/ropeixoto/harnessx/internal/ui"
)

// autoState is the resumable checkpoint persisted under
// .harness/runs/_agent/<id>/state.json. Each phase records its
// outcome + elapsed + cost so --resume can pick up where the crash
// left off.
type autoState struct {
	RunID       string         `json:"run_id"`
	Prompt      string         `json:"prompt"`
	AgentID     string         `json:"agent_id"`
	MaxAttempts int            `json:"max_attempts"`
	BudgetUSD   float64        `json:"budget_usd"`
	Started     time.Time      `json:"started"`
	Updated     time.Time      `json:"updated"`
	Phases      map[string]any `json:"phases"`
	SpecID      string         `json:"spec_id,omitempty"`
	Attempts    int            `json:"attempts"`
	CostUSD     float64        `json:"cost_usd"`
	Done        bool           `json:"done"`
	LastError   string         `json:"last_error,omitempty"`
}

func newAutoCmd() *cobra.Command {
	var (
		maxAttempts int
		budgetUSD   float64
		agentID     string
		watch       bool
		dryRun      bool
		resumeID    string
	)
	c := &cobra.Command{
		Use:     "auto \"<prompt>\"",
		Aliases: []string{"agent-run"},
		Short:   "End-to-end agentic workflow: plan → spec → tests → impl → ci → commit (resumable)",
		Long: `Drives a feature from a single prompt through the full pipeline:
  1. plan        — deterministic JSON plan
  2. spec        — harness feature → .harness/artifacts/specs/<id>.md
  3. tests       — failing tests via harness drive (red phase)
  4. impl        — harness do --autonomy safe_execute
  5. ci          — harness ci; on failure feed canonicalised error
                   back to step 4 up to --max-attempts
  6. commit      — conventional commit on green

State persists under .harness/runs/_agent/<run-id>/state.json so
--resume <run-id> picks up after crash.`,
		Args: cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := cwd()
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			in := cmd.InOrStdin()

			if !dryRun {
				agentID, err = resolveDefaultAgent(agentID, root, out, in)
				if err != nil {
					return err
				}
			}

			var state *autoState
			if resumeID != "" {
				state, err = loadAutoState(root, resumeID)
				if err != nil {
					return fmt.Errorf("auto resume %s: %w", resumeID, err)
				}
				fmt.Fprintf(out, "auto: resuming run %s (%d attempts so far, $%.4f)\n", state.RunID, state.Attempts, state.CostUSD)
			} else {
				prompt := strings.Join(args, " ")
				if strings.TrimSpace(prompt) == "" {
					return fmt.Errorf("auto: prompt required")
				}
				state = &autoState{
					RunID:       strings.ToLower(ids.New()),
					Prompt:      prompt,
					AgentID:     agentID,
					MaxAttempts: maxAttempts,
					BudgetUSD:   budgetUSD,
					Started:     time.Now().UTC(),
					Phases:      map[string]any{},
				}
				if err := saveAutoState(root, state); err != nil {
					return err
				}
			}

			runAutoPipeline(cmd.Context(), root, state, out, dryRun, watch)
			return saveAutoState(root, state)
		},
	}
	c.Flags().IntVar(&maxAttempts, "max-attempts", 5, "max impl→ci retries before giving up")
	c.Flags().Float64Var(&budgetUSD, "budget-usd", 2.0, "hard cap across the whole pipeline")
	c.Flags().StringVar(&agentID, "agent", "", "agent adapter id (defaults via active.yaml / env / TTY picker)")
	c.Flags().BoolVar(&watch, "watch", false, "stream live phase output (default: per-phase summary)")
	c.Flags().BoolVar(&dryRun, "dry-run", false, "print planned phases without executing")
	c.Flags().StringVar(&resumeID, "resume", "", "resume an interrupted run by id")
	return c
}

func runAutoPipeline(ctx context.Context, root string, st *autoState, out io.Writer, dryRun, watch bool) {
	phases := []struct {
		name string
		args []string
	}{
		{"plan", []string{"plan", st.Prompt}},
		{"spec", []string{"feature", st.Prompt, "--yes"}},
		{"tests", []string{"drive", "--features", filepath.Join(".harness", "artifacts", "specs"), st.Prompt, "--max-attempts", "1"}},
		{"impl", []string{"do", st.Prompt, "--autonomy", "safe_execute", "--agent", st.AgentID}},
		{"ci", []string{"ci"}},
	}
	for _, p := range phases {
		if _, done := st.Phases[p.name]; done {
			fmt.Fprintf(out, "  %s %s — already completed (skip)\n", ui.MarkSuccess(), p.name)
			continue
		}
		if dryRun {
			fmt.Fprintf(out, "  - %-6s harness %s\n", p.name, strings.Join(p.args, " "))
			continue
		}
		start := time.Now()
		err := runHarnessSubcommand(ctx, out, root, p.args...)
		elapsed := time.Since(start)
		st.Phases[p.name] = map[string]any{
			"elapsed_ms": elapsed.Milliseconds(),
			"ok":         err == nil,
			"err":        errString(err),
		}
		st.Updated = time.Now().UTC()
		_ = saveAutoState(root, st)
		if err != nil {
			st.LastError = err.Error()
			fmt.Fprintf(out, "  %s %s — %v\n", ui.MarkFail(), p.name, err)
			return
		}
		fmt.Fprintf(out, "  %s %s (%s)\n", ui.MarkSuccess(), p.name, elapsed.Round(time.Millisecond))
	}
	st.Done = true
}

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func autoStatePath(root, runID string) string {
	return filepath.Join(root, ".harness", "runs", "_agent", runID, "state.json")
}

func saveAutoState(root string, st *autoState) error {
	p := autoStatePath(root, st.RunID)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	body, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, body, 0o644)
}

func loadAutoState(root, runID string) (*autoState, error) {
	body, err := os.ReadFile(autoStatePath(root, runID))
	if err != nil {
		return nil, err
	}
	var st autoState
	if err := json.Unmarshal(body, &st); err != nil {
		return nil, err
	}
	if st.Phases == nil {
		st.Phases = map[string]any{}
	}
	return &st, nil
}
