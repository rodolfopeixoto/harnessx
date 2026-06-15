// SPDX-License-Identifier: MIT

package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/ropeixoto/harnessx/internal/runtime/containers"
)

func truncateRT(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

func newRuntimeCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "runtime",
		Short: "Manage the container runtime used for sandboxed execution and tests",
		Long: `Detect, list, select, and inspect the container runtime harness uses
when it needs to run a container (sandboxed agent execution, lifecycle
tests, image pulls).

Default is auto-detect with this preference order:
  macOS: apple_container > docker > orbstack > podman > colima
  linux: docker > podman > orbstack > colima

Selection persists to .harness/config/runtime.yaml.
Override per call with HARNESS_RUNTIME=<id>.`,
	}
	c.AddCommand(newRuntimeListCmd(), newRuntimeSetCmd(), newRuntimeSelectCmd(), newRuntimeInfoCmd())
	return c
}

func newRuntimeListCmd() *cobra.Command {
	var jsonOut bool
	c := &cobra.Command{
		Use:   "list",
		Short: "List detected runtimes with versions and selection status",
		RunE: func(cmd *cobra.Command, _ []string) error {
			root, err := cwd()
			if err != nil {
				return err
			}
			ctx := cmd.Context()
			detected := containers.DetectIncluding(ctx, true)
			cfg, _ := containers.LoadConfig(root)
			out := cmd.OutOrStdout()
			if jsonOut {
				return emitRuntimesJSON(ctx, out, detected, cfg.Runtime)
			}
			w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "ID\tBINARY\tAVAILABLE\tVERSION\tSELECTED")
			for _, r := range detected {
				avail := "—"
				ver := "—"
				if r.Available(ctx) {
					avail = "✓"
					if v, err := r.Version(ctx); err == nil {
						ver = truncateRT(v, 40)
					}
				}
				selected := ""
				if cfg.Runtime == r.ID() {
					selected = "★"
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", r.ID(), r.Binary(), avail, ver, selected)
			}
			return w.Flush()
		},
	}
	c.Flags().BoolVar(&jsonOut, "json", false, "emit JSON")
	return c
}

func newRuntimeSetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set <id>",
		Short: "Pin the runtime without prompting",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := cwd()
			if err != nil {
				return err
			}
			rt, err := containers.ByID(args[0])
			if err != nil {
				return err
			}
			ctx := cmd.Context()
			cfg := containers.Config{Runtime: rt.ID()}
			if v, err := rt.Version(ctx); err == nil {
				cfg.Version = v
			}
			if err := containers.SaveConfig(root, cfg); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "runtime set: %s (%s)\n", cfg.Runtime, cfg.Version)
			return nil
		},
	}
}

func newRuntimeSelectCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "select",
		Short: "Interactive: pick from detected runtimes",
		RunE: func(cmd *cobra.Command, _ []string) error {
			root, err := cwd()
			if err != nil {
				return err
			}
			ctx := cmd.Context()
			detected := containers.Detect(ctx)
			out := cmd.OutOrStdout()
			if len(detected) == 0 {
				fmt.Fprintln(out, "No container runtime detected on this host.")
				fmt.Fprintln(out, "Install one: docker | podman | orbstack | colima | apple container (macOS 15+).")
				return nil
			}
			fmt.Fprintln(out, "Detected runtimes:")
			for i, r := range detected {
				v, _ := r.Version(ctx)
				fmt.Fprintf(out, "  [%d] %s — %s\n", i+1, r.ID(), truncateRT(v, 60))
			}
			fmt.Fprint(out, "Pick [1-", len(detected), ", or empty to skip]: ")
			reader := bufio.NewReader(os.Stdin)
			line, _ := reader.ReadString('\n')
			line = strings.TrimSpace(line)
			if line == "" {
				fmt.Fprintln(out, "skipped")
				return nil
			}
			idx, err := strconv.Atoi(line)
			if err != nil || idx < 1 || idx > len(detected) {
				return fmt.Errorf("invalid choice %q", line)
			}
			rt := detected[idx-1]
			cfg := containers.Config{Runtime: rt.ID()}
			if v, err := rt.Version(ctx); err == nil {
				cfg.Version = v
			}
			if err := containers.SaveConfig(root, cfg); err != nil {
				return err
			}
			fmt.Fprintf(out, "runtime set: %s\n", cfg.Runtime)
			return nil
		},
	}
}

func newRuntimeInfoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "info",
		Short: "Show currently selected runtime + version",
		RunE: func(cmd *cobra.Command, _ []string) error {
			root, err := cwd()
			if err != nil {
				return err
			}
			ctx := cmd.Context()
			rt, source, err := containers.Resolve(ctx, root)
			if err != nil {
				return err
			}
			v, _ := rt.Version(ctx)
			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			fmt.Fprintf(w, "runtime\t%s\n", rt.ID())
			fmt.Fprintf(w, "binary\t%s\n", rt.Binary())
			fmt.Fprintf(w, "version\t%s\n", truncateRT(v, 80))
			fmt.Fprintf(w, "source\t%s\n", source)
			return w.Flush()
		},
	}
}

func emitRuntimesJSON(ctx context.Context, out io.Writer, runtimes []containers.Runtime, selected string) error {
	type row struct {
		ID        string `json:"id"`
		Binary    string `json:"binary"`
		Available bool   `json:"available"`
		Version   string `json:"version,omitempty"`
		Selected  bool   `json:"selected"`
	}
	rows := make([]row, 0, len(runtimes))
	for _, r := range runtimes {
		ok := r.Available(ctx)
		ver := ""
		if ok {
			if v, err := r.Version(ctx); err == nil {
				ver = v
			}
		}
		rows = append(rows, row{ID: r.ID(), Binary: r.Binary(), Available: ok, Version: ver, Selected: r.ID() == selected})
	}
	return json.NewEncoder(out).Encode(rows)
}
