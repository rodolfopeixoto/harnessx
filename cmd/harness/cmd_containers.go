// SPDX-License-Identifier: MIT

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/ropeixoto/harnessx/internal/runtime/containers"
)

const containersConfirmEnv = "HARNESS_CONTAINERS_I_UNDERSTAND"

func newContainersCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "containers",
		Short: "List, kill, and prune containers via the selected runtime",
		Long: `Cross-runtime container operations. Uses the runtime resolved by
'harness runtime info' (env > project config > auto-detect).

Two-key safety on prune:
  - interactive y/N prompt, OR
  - export HARNESS_CONTAINERS_I_UNDERSTAND=1 for non-interactive flows`,
	}
	c.AddCommand(newContainersListCmd(), newContainersKillCmd(), newContainersPruneCmd())
	return c
}

func newContainersListCmd() *cobra.Command {
	var (
		all     bool
		jsonOut bool
	)
	c := &cobra.Command{
		Use:   "list",
		Short: "List containers (running by default; --all includes stopped)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			root, err := cwd()
			if err != nil {
				return err
			}
			ctx := cmd.Context()
			rt, _, err := containers.Resolve(ctx, root)
			if err != nil {
				return err
			}
			list, err := rt.List(ctx, containers.ListOptions{All: all})
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			if jsonOut {
				return json.NewEncoder(out).Encode(list)
			}
			w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "ID\tNAME\tIMAGE\tSTATUS")
			for _, c := range list {
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", short(c.ID, 12), c.Name, truncateRT(c.Image, 40), truncateRT(c.Status, 40))
			}
			return w.Flush()
		},
	}
	c.Flags().BoolVar(&all, "all", false, "include stopped containers")
	c.Flags().BoolVar(&jsonOut, "json", false, "emit JSON")
	return c
}

func newContainersKillCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "kill <id> [<id>...]",
		Short: "Stop + remove one or more containers",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := cwd()
			if err != nil {
				return err
			}
			ctx := cmd.Context()
			rt, _, err := containers.Resolve(ctx, root)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			for _, id := range args {
				if err := rt.Kill(ctx, id); err != nil {
					fmt.Fprintf(out, "✗ %s: %v\n", id, err)
					continue
				}
				fmt.Fprintf(out, "✓ killed %s\n", id)
			}
			return nil
		},
	}
}

func newContainersPruneCmd() *cobra.Command {
	var (
		stopped   bool
		all       bool
		olderThan time.Duration
		jsonOut   bool
	)
	c := &cobra.Command{
		Use:   "prune",
		Short: "Remove stopped containers (two-key safety: prompt OR env var)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			root, err := cwd()
			if err != nil {
				return err
			}
			ctx := cmd.Context()
			rt, _, err := containers.Resolve(ctx, root)
			if err != nil {
				return err
			}
			opts := containers.PruneOptions{Stopped: stopped, All: all, OlderThan: olderThan}
			opts.IUnderstand = os.Getenv(containersConfirmEnv) == "1"
			if !opts.IUnderstand {
				if !confirmInteractivePrune(cmd.OutOrStdout(), opts, rt.ID()) {
					return fmt.Errorf("prune aborted (set %s=1 or confirm interactively)", containersConfirmEnv)
				}
				opts.IUnderstand = true
			}
			res, err := rt.Prune(ctx, opts)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			if jsonOut {
				return json.NewEncoder(out).Encode(res)
			}
			fmt.Fprintf(out, "pruned: %d\n", len(res.Pruned))
			for _, id := range res.Pruned {
				fmt.Fprintf(out, "  ✓ %s\n", short(id, 12))
			}
			if len(res.Skipped) > 0 {
				fmt.Fprintf(out, "skipped: %d\n", len(res.Skipped))
			}
			return nil
		},
	}
	c.Flags().BoolVar(&stopped, "stopped", true, "only prune stopped containers (default)")
	c.Flags().BoolVar(&all, "all", false, "include running containers (DANGEROUS)")
	c.Flags().DurationVar(&olderThan, "older-than", 0, "only prune containers older than this duration (e.g. 720h)")
	c.Flags().BoolVar(&jsonOut, "json", false, "emit JSON")
	return c
}

func confirmInteractivePrune(out interface{ Write([]byte) (int, error) }, opts containers.PruneOptions, runtimeID string) bool {
	mode := "stopped"
	if opts.All {
		mode = "ALL (incl. running)"
	}
	cutoff := "any age"
	if opts.OlderThan > 0 {
		cutoff = "older than " + opts.OlderThan.String()
	}
	fmt.Fprintf(out, "About to prune %s containers via %s (%s).\n", mode, runtimeID, cutoff)
	fmt.Fprint(out, "Type 'yes' to confirm: ")
	reader := bufio.NewReader(os.Stdin)
	line, _ := reader.ReadString('\n')
	return strings.TrimSpace(line) == "yes"
}

func short(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
