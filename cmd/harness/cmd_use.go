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
		tier  string
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
			if tier != "" {
				if model != "" {
					return fmt.Errorf("use: --tier and --model are mutually exclusive")
				}
				resolved, err := resolveTierModel(dir, id, tier)
				if err != nil {
					return err
				}
				model = resolved
			}
			if err := activeagent.Save(dir, activeagent.Pin{AgentID: id, Model: model}); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%s pinned %s\n", ui.MarkSuccess(), ui.Accent.Render(id))
			return nil
		},
	}
	c.Flags().StringVar(&model, "model", "", "model id override (forwarded to the adapter)")
	c.Flags().StringVar(&tier, "tier", "", "resolve a model tier (cheap|default|deep) from the adapter YAML")
	c.Flags().BoolVar(&clear, "clear", false, "remove the pin")
	return c
}

// resolveTierModel maps a tier label (cheap|default|deep|...) to a concrete
// model id declared in the adapter's YAML (capabilities.models). The empty
// string is returned when the tier is absent — callers should surface the
// available tiers in the error message.
func resolveTierModel(root, adapterID, tier string) (string, error) {
	reg, _, err := agentcmd.LoadAll(root)
	if err != nil {
		return "", err
	}
	a, ok := reg.Get(adapterID)
	if !ok {
		return "", fmt.Errorf("use: adapter %q not registered", adapterID)
	}
	models := a.Capabilities().Models
	if len(models) == 0 {
		return "", fmt.Errorf("use: adapter %q declares no model tiers", adapterID)
	}
	id, ok := models[tier]
	if !ok || id == "" {
		available := make([]string, 0, len(models))
		for t := range models {
			available = append(available, t)
		}
		return "", fmt.Errorf("use: tier %q not defined for %s (available: %v)", tier, adapterID, available)
	}
	return id, nil
}

func containsString(s []string, v string) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}
