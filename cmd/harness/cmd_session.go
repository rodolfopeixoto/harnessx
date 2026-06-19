// SPDX-License-Identifier: MIT

package main

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ropeixoto/harnessx/internal/app/sessioncmd"
	"github.com/ropeixoto/harnessx/internal/repl"
)

func newSessionCmd() *cobra.Command {
	c := &cobra.Command{Use: "session", Short: "Session commands"}
	c.AddCommand(&cobra.Command{
		Use:   "show <id>",
		Short: "Show one session's runs + sensors + cost from sqlite",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := cwd()
			if err != nil {
				return err
			}
			return sessioncmd.Show(cmd.OutOrStdout(), sessioncmd.ShowOptions{StartDir: dir, ID: args[0]})
		},
	})
	c.AddCommand(newSessionExportCmd())
	return c
}

func newSessionExportCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "export <id>",
		Short: "Emit a chat session JSONL wrapped in a sharable JSON envelope",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := cwd()
			if err != nil {
				return err
			}
			sess, err := repl.LoadSession(dir, args[0])
			if err != nil {
				return fmt.Errorf("session export %s: %w", args[0], err)
			}
			var inTok, outTok int
			var usd float64
			chatTurns := 0
			for _, t := range sess.Turns {
				if t.Action == "chat" {
					chatTurns++
				}
				inTok += t.InTokens
				outTok += t.OutTokens
				usd += t.CostUSD
			}
			env := map[string]any{
				"id":              sess.ID,
				"goal":            sess.Goal,
				"started":         sess.Started,
				"root":            sess.Root,
				"turn_count":      len(sess.Turns),
				"chat_turn_count": chatTurns,
				"in_tokens":       inTok,
				"out_tokens":      outTok,
				"cost_usd":        usd,
				"context_mark":    sess.ContextMark,
				"auto_gate":       sess.AutoGate,
				"budget_usd":      sess.BudgetUSD,
				"turns":           sess.Turns,
				"exported_by":     "harness session export",
				"schema_version":  1,
			}
			enc := json.NewEncoder(cmd.OutOrStdout())
			enc.SetIndent("", "  ")
			return enc.Encode(env)
		},
	}
}
