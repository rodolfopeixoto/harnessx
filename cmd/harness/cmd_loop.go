// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ropeixoto/harnessx/internal/devloop"
)

func newLoopCmd() *cobra.Command {
	var (
		agentID     string
		autonomy    string
		budgetUSD   float64
		maxAttempts int
		lintCmd     string
		testCmd     string
		apply       bool
	)
	c := &cobra.Command{
		Use:   "loop \"<prompt>\"",
		Short: "Deterministic dev-loop: agent → lint+test → on failure, canonicalised error fed back",
		Long: `Runs a normal feature workflow, then executes the project's lint and
test commands. If either fails, the failure output is canonicalised
into a follow-up prompt and the workflow runs again. Stops on first
pass, on --max-attempts exhaustion, or when --budget-usd is consumed.

Lint/test commands auto-detect from a scaffolded project (Python →
ruff/pytest, Go → golangci-lint/go test, etc.). Override with --lint
and --test.`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := cwd()
			if err != nil {
				return err
			}
			res, err := devloop.Run(cmd.Context(), devloop.Options{
				StartDir:    dir,
				Prompt:      strings.Join(args, " "),
				AgentID:     agentID,
				Autonomy:    autonomy,
				BudgetUSD:   budgetUSD,
				MaxAttempts: maxAttempts,
				LintCmd:     lintCmd,
				TestCmd:     testCmd,
				Apply:       apply,
			}, cmd.OutOrStdout())
			if err != nil {
				return err
			}
			status := "blocked"
			if res.Passed {
				status = "passed"
			}
			fmt.Fprintf(cmd.OutOrStdout(), "\nloop %s after %d attempt(s): %s\n", status, len(res.Attempts), res.Reason)
			return nil
		},
	}
	c.Flags().StringVar(&agentID, "agent", "claude", "agent adapter id")
	c.Flags().StringVar(&autonomy, "autonomy", "safe_execute", "autonomy level")
	c.Flags().Float64Var(&budgetUSD, "budget-usd", 1.0, "max USD across all attempts")
	c.Flags().IntVar(&maxAttempts, "max-attempts", 3, "hard cap on retry attempts (max 10)")
	c.Flags().StringVar(&lintCmd, "lint", "", "lint command (default: auto-detect from scaffold)")
	c.Flags().StringVar(&testCmd, "test", "", "test command (default: auto-detect from scaffold)")
	c.Flags().BoolVar(&apply, "apply", true, "apply diff after each agent run")
	return c
}
