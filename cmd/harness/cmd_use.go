// SPDX-License-Identifier: MIT

package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ropeixoto/harnessx/internal/activeagent"
	"github.com/ropeixoto/harnessx/internal/app/agentcmd"
	"github.com/ropeixoto/harnessx/internal/ui"
)

func newUseCmd() *cobra.Command {
	var (
		model string
		clear bool
	)
	c := &cobra.Command{
		Use:   "use <adapter-id>",
		Short: "Pin the active LLM adapter for the project (paper §3.5.3 governed)",
		Long: `Writes .harness/config/active.yaml. The pinned adapter is used by
'harness do', 'harness ship', and 'harness chat' unless overridden with
--agent / --adapter on the call. Run 'harness agent list' to see ids.

Examples:
  harness use claude              # pin Claude Code
  harness use kimi --model kimi-k2
  harness use --clear             # remove the pin`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := cwd()
			if err != nil {
				return err
			}
			if clear {
				if err := activeagent.Clear(dir); err != nil {
					return err
				}
				fmt.Fprintf(cmd.OutOrStdout(), "%s pin cleared\n", ui.MarkSuccess())
				return nil
			}
			if len(args) == 0 {
				p, err := activeagent.Load(dir)
				if err != nil {
					return err
				}
				if p.AgentID == "" {
					fmt.Fprintln(cmd.OutOrStdout(), "no active pin (router defaults apply)")
					return nil
				}
				fmt.Fprintf(cmd.OutOrStdout(), "active: %s model=%s\n", ui.Accent.Render(p.AgentID), p.Model)
				return nil
			}
			id := args[0]
			ids, err := agentcmd.AvailableAdapterIDs()
			if err != nil {
				return err
			}
			if !containsString(ids, id) {
				return fmt.Errorf("use: unknown adapter %q (have %v)", id, ids)
			}
			if err := activeagent.Save(dir, activeagent.Pin{AgentID: id, Model: model}); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%s pinned %s\n", ui.MarkSuccess(), ui.Accent.Render(id))
			return nil
		},
	}
	c.Flags().StringVar(&model, "model", "", "model id override (forwarded to the adapter)")
	c.Flags().BoolVar(&clear, "clear", false, "remove the pin")
	return c
}

func containsString(s []string, v string) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}
