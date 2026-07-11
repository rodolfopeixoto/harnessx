// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/ropeixoto/harnessx/internal/hookpkg"
	"github.com/ropeixoto/harnessx/internal/hookscan"
)

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
