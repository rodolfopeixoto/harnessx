package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ropeixoto/harnessx/internal/install"
)

var agentUpdateRegistry = map[string]string{
	"claude":      "claude",
	"codex":       "codex",
	"antigravity": "antigravity",
	"kimi":        "kimi",
}

func newAgentUpdateCmd() *cobra.Command {
	var dryRun bool
	c := &cobra.Command{
		Use:   "update <id|all>",
		Short: "Upgrade an agent CLI to its latest version (npm / brew / pip per bundled manifest)",
		Long: `Upgrades the underlying CLI binary for a registered agent. For
example, ` + "`harness agent update gemini`" + ` runs the gemini install
manifest in upgrade mode (npm install -g @google/gemini-cli@latest on
darwin/linux). Use 'all' to upgrade every registered agent in turn.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ids := []string{args[0]}
			if args[0] == "all" {
				ids = nil
				for k := range agentUpdateRegistry {
					ids = append(ids, k)
				}
			}
			out := cmd.OutOrStdout()
			reg := install.NewRegistry()
			for _, id := range ids {
				manifestName, ok := agentUpdateRegistry[id]
				if !ok {
					fmt.Fprintf(out, "  [skip] %s: no install manifest registered\n", id)
					continue
				}
				m, err := install.LoadBundled(manifestName)
				if err != nil {
					fmt.Fprintf(out, "  [skip] %s: %v\n", id, err)
					continue
				}
				plan, err := reg.Pick(m)
				if err != nil {
					fmt.Fprintf(out, "  [skip] %s: %v\n", id, err)
					continue
				}
				fmt.Fprintf(out, "→ upgrading %s via %s\n", id, plan.Kind)
				if err := install.Execute(cmd.Context(), plan, dryRun, out, out); err != nil {
					fmt.Fprintf(out, "  [fail] %s: %v\n", id, err)
				} else {
					fmt.Fprintf(out, "  [ok]   %s\n", id)
				}
			}
			return nil
		},
	}
	c.Flags().BoolVar(&dryRun, "dry-run", false, "print commands without executing")
	return c
}
