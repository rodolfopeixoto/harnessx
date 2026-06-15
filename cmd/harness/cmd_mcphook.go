// SPDX-License-Identifier: MIT

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/ropeixoto/harnessx/internal/hookpkg"
	"github.com/ropeixoto/harnessx/internal/hookscan"
	"github.com/ropeixoto/harnessx/internal/mcppkg"
	"github.com/ropeixoto/harnessx/internal/mcpscan"
)

func newMCPCmd() *cobra.Command {
	c := &cobra.Command{Use: "mcp", Short: "MCP server discovery, listing, install"}
	c.AddCommand(mcpListCmd(), mcpScanCmd(), mcpInstallCmd(), mcpTemplatesCmd())
	return c
}

func mcpTemplatesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "templates",
		Short: "List bundled MCP server templates available to `mcp install`",
		RunE: func(cmd *cobra.Command, _ []string) error {
			names, err := mcppkg.List()
			if err != nil {
				return err
			}
			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "NAME\tTRANSPORT\tCOMMAND\tDESCRIPTION")
			for _, n := range names {
				t, err := mcppkg.Load(n)
				if err != nil {
					continue
				}
				cmdStr := t.Command
				if cmdStr == "" {
					cmdStr = t.URL
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", t.Name, t.Transport, cmdStr, t.Description)
			}
			return w.Flush()
		},
	}
}

func mcpInstallCmd() *cobra.Command {
	var (
		yes       bool
		transport string
		command   string
		url       string
	)
	c := &cobra.Command{
		Use:   "install <name>",
		Short: "Write .harness/mcp/<name>.json from a bundled template or explicit flags",
		Long: `Resolve the bundled template for <name> (see 'harness mcp templates')
and write its command + args + env into .harness/mcp/<name>.json. When
no template exists, --command / --url / --transport build the config
from scratch.

Examples:
  harness mcp install filesystem      # bundled npx @modelcontextprotocol/server-filesystem
  harness mcp install github          # bundled, needs GITHUB_PERSONAL_ACCESS_TOKEN in env
  harness mcp install my-server --command /usr/local/bin/my-mcp --transport stdio`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			root, err := cwd()
			if err != nil {
				return err
			}
			dir := filepath.Join(root, ".harness", "mcp")
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return err
			}
			target := filepath.Join(dir, name+".json")
			if _, err := os.Stat(target); err == nil && !yes {
				return fmt.Errorf("%s already exists (pass --yes to overwrite)", target)
			}
			cfg := buildMCPConfig(name, transport, command, url)
			data, err := json.MarshalIndent(cfg, "", "  ")
			if err != nil {
				return err
			}
			if err := os.WriteFile(target, data, 0o644); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Wrote %s\n", target)
			return nil
		},
	}
	c.Flags().BoolVar(&yes, "yes", false, "overwrite if exists")
	c.Flags().StringVar(&transport, "transport", "", "stdio|http (default from template)")
	c.Flags().StringVar(&command, "command", "", "binary command (overrides template)")
	c.Flags().StringVar(&url, "url", "", "URL for http transport (overrides template)")
	return c
}

func buildMCPConfig(name, transport, command, url string) map[string]any {
	cfg := map[string]any{"name": name}
	tmpl, err := mcppkg.Load(name)
	if err == nil {
		cfg["transport"] = tmpl.Transport
		if tmpl.Command != "" {
			cfg["command"] = tmpl.Command
		}
		if len(tmpl.Args) > 0 {
			cfg["args"] = tmpl.Args
		}
		if tmpl.URL != "" {
			cfg["url"] = tmpl.URL
		}
		if len(tmpl.Env) > 0 {
			cfg["env"] = tmpl.Env
		}
		if tmpl.Docs != "" {
			cfg["docs"] = tmpl.Docs
		}
	}
	if transport != "" {
		cfg["transport"] = transport
	} else if _, ok := cfg["transport"]; !ok {
		cfg["transport"] = "stdio"
	}
	if command != "" {
		cfg["command"] = command
	}
	if url != "" {
		cfg["url"] = url
	}
	return cfg
}

func newHookCmd() *cobra.Command {
	c := &cobra.Command{Use: "hook", Short: "Hook discovery, install, listing"}
	c.AddCommand(hookListCmd(), hookScanCmd(), hookInstallCmd(), hookTemplatesCmd(), hookAddCmd())
	return c
}

func hookAddCmd() *cobra.Command {
	var yes bool
	c := &cobra.Command{
		Use:   "add <event>",
		Short: "List bundled templates for an event and install one interactively",
		Long: `Discover bundled templates whose event matches <event> and prompt for
which to install. Pass --yes to install the first match without prompting.
Equivalent to listing 'harness hook templates' then 'harness hook install <name>'.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runHookAdd(cmd, args[0], yes)
		},
	}
	c.Flags().BoolVar(&yes, "yes", false, "install first matching template without prompting")
	return c
}

func runHookAdd(cmd *cobra.Command, event string, yes bool) error {
	matches, err := hookTemplatesForEvent(event)
	if err != nil {
		return err
	}
	out := cmd.OutOrStdout()
	printTemplateMenu(out, event, matches)
	pick := promptTemplatePick(cmd, len(matches), yes)
	chosen := matches[pick-1]
	root, err := cwd()
	if err != nil {
		return err
	}
	dir := filepath.Join(root, ".harness", "hooks")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	target := filepath.Join(dir, event+".sh")
	if err := os.WriteFile(target, chosen.Body, 0o755); err != nil {
		return err
	}
	fmt.Fprintf(out, "wrote %s (template: %s)\n", target, chosen.Name)
	return nil
}

func hookTemplatesForEvent(event string) ([]hookpkg.Template, error) {
	all, err := hookpkg.List()
	if err != nil {
		return nil, err
	}
	var matches []hookpkg.Template
	for _, t := range all {
		if t.Event == event {
			matches = append(matches, t)
		}
	}
	if len(matches) == 0 {
		return nil, fmt.Errorf("no bundled templates for event %q (run 'harness hook templates' to list all)", event)
	}
	return matches, nil
}

func printTemplateMenu(out io.Writer, event string, matches []hookpkg.Template) {
	fmt.Fprintf(out, "templates for event %q:\n", event)
	for i, t := range matches {
		fmt.Fprintf(out, "  [%d] %s — %s\n", i+1, t.Name, t.Description)
	}
}

func promptTemplatePick(cmd *cobra.Command, count int, yes bool) int {
	if yes {
		return 1
	}
	fmt.Fprintf(cmd.OutOrStdout(), "pick [1-%d] (default 1): ", count)
	var raw string
	_, _ = fmt.Fscanln(cmd.InOrStdin(), &raw)
	if raw == "" {
		return 1
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 1 || n > count {
		return 1
	}
	return n
}

func hookInstallCmd() *cobra.Command {
	var (
		yes      bool
		filename string
	)
	c := &cobra.Command{
		Use:   "install <name>",
		Short: "Drop a bundled hook script into .harness/hooks/",
		Long: `Install a bundled hook template into .harness/hooks/<name>.sh
and mark it executable. Templates carry the right shebang, env var
contract (HARNESS_RUN_ID, HARNESS_AGENT, HARNESS_RUN_STATUS), and
event name in a leading comment that 'harness hook scan' picks up.

  --filename other-name.sh   write under a different name
  --yes                      overwrite if exists`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := cwd()
			if err != nil {
				return err
			}
			t, err := hookpkg.Load(args[0])
			if err != nil {
				return err
			}
			dir := filepath.Join(root, ".harness", "hooks")
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return err
			}
			name := filename
			if name == "" {
				name = t.Event + ".sh"
				if t.Event == "" {
					name = args[0] + ".sh"
				}
			}
			target := filepath.Join(dir, name)
			if _, err := os.Stat(target); err == nil && !yes {
				return fmt.Errorf("%s already exists (pass --yes to overwrite)", target)
			}
			if err := os.WriteFile(target, t.Body, 0o755); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "wrote %s\n  event: %s\n  source: bundled %s\n", target, t.Event, args[0])
			return nil
		},
	}
	c.Flags().BoolVar(&yes, "yes", false, "overwrite if exists")
	c.Flags().StringVar(&filename, "filename", "", "override target filename (default: <event>.sh)")
	return c
}

func hookTemplatesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "templates",
		Short: "List bundled hook templates available to `hook install`",
		RunE: func(cmd *cobra.Command, _ []string) error {
			all, err := hookpkg.List()
			if err != nil {
				return err
			}
			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "NAME\tEVENT\tDESCRIPTION")
			for _, t := range all {
				fmt.Fprintf(w, "%s\t%s\t%s\n", t.Name, t.Event, t.Description)
			}
			return w.Flush()
		},
	}
}

func mcpScanCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "scan [root]",
		Short: "Deterministically scan filesystem for MCP server configs",
		Args:  cobra.MaximumNArgs(1),
		RunE:  mcpRun,
	}
	c.Flags().Bool("json", false, "emit JSON")
	return c
}

func mcpListCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "list [root]",
		Short: "Alias for `mcp scan`",
		Args:  cobra.MaximumNArgs(1),
		RunE:  mcpRun,
	}
	c.Flags().Bool("json", false, "emit JSON")
	return c
}

func hookScanCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "scan [root]",
		Short: "Deterministically scan filesystem for hooks",
		Args:  cobra.MaximumNArgs(1),
		RunE:  hookRun,
	}
	c.Flags().Bool("json", false, "emit JSON")
	return c
}

func hookListCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "list [root]",
		Short: "Alias for `hook scan`",
		Args:  cobra.MaximumNArgs(1),
		RunE:  hookRun,
	}
	c.Flags().Bool("json", false, "emit JSON")
	return c
}

func mcpRun(cmd *cobra.Command, args []string) error {
	root, err := rootFromArgs(args)
	if err != nil {
		return err
	}
	asJSON, _ := cmd.Flags().GetBool("json")
	servers, err := mcpscan.Scan(root)
	if err != nil {
		return err
	}
	if asJSON {
		return writeJSON(cmd, servers)
	}
	if len(servers) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "no MCP servers detected")
		return nil
	}
	tw := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "SOURCE\tNAME\tTRANSPORT\tRISK\tPATH")
	for _, s := range servers {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n", s.Source, s.Name, s.Transport, s.Risk, s.ConfigPath)
	}
	return tw.Flush()
}

func hookRun(cmd *cobra.Command, args []string) error {
	root, err := rootFromArgs(args)
	if err != nil {
		return err
	}
	asJSON, _ := cmd.Flags().GetBool("json")
	hooks, err := hookscan.Scan(root)
	if err != nil {
		return err
	}
	if asJSON {
		return writeJSON(cmd, hooks)
	}
	if len(hooks) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "no hooks detected")
		return nil
	}
	tw := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "SOURCE\tNAME\tEVENT\tSCOPE\tSTATUS\tRISK\tPATH")
	for _, h := range hooks {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n", h.Source, h.Name, h.Event, h.Scope, h.Status, h.Risk, h.ConfigPath)
	}
	return tw.Flush()
}

func writeJSON(cmd *cobra.Command, payload any) error {
	enc := json.NewEncoder(cmd.OutOrStdout())
	enc.SetIndent("", "  ")
	return enc.Encode(payload)
}
