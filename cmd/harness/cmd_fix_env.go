package main

import (
	"github.com/spf13/cobra"

	"github.com/ropeixoto/harnessx/internal/app/fixenvcmd"
)

func newFixEnvCmd() *cobra.Command {
	var apply bool
	c := &cobra.Command{
		Use:   "fix-env",
		Short: "Detect common environment issues and (with --apply) fix them deterministically",
		Long: `Probes the current shell + project for the env failures the harness
audit hands-on report flagged:

  - GOROOT pointing at a missing directory (breaks ` + "`go`" + ` commands)
  - PATH missing standard install dirs (/usr/local/bin, /opt/homebrew/bin)
  - Python project without .venv (sensors skip)
  - Node project without node_modules (sensors skip)
  - git user.email unset (harness drive cannot commit)
  - brew installed but not on PATH (install strategies fail)
  - Ruby project without Gemfile.lock (sensors skip)

Without --apply: prints a numbered checklist with the exact command to
run yourself. With --apply: actually executes the safe fixes
(unset GOROOT for this process, create .venv, npm install, etc.).
No LLM, no network beyond the obvious package managers.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			root, err := cwd()
			if err != nil {
				return err
			}
			_, err = fixenvcmd.Run(cmd.OutOrStdout(), fixenvcmd.Options{Root: root, Apply: apply})
			return err
		},
	}
	c.Flags().BoolVar(&apply, "apply", false, "execute the safe fixes (default: dry-run with hints)")
	return c
}
