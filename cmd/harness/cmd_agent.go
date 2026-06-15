// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ropeixoto/harnessx/internal/app/agentcmd"
	"github.com/ropeixoto/harnessx/internal/secrets"
)

func newAgentCmd() *cobra.Command {
	c := &cobra.Command{Use: "agent", Short: "Agent adapter commands"}

	listC := &cobra.Command{
		Use:   "list",
		Short: "List registered agent adapters",
		RunE: func(cmd *cobra.Command, _ []string) error {
			dir, err := cwd()
			if err != nil {
				return err
			}
			return agentcmd.List(cmd.OutOrStdout(), dir)
		},
	}

	addC := &cobra.Command{
		Use:   "add <id>",
		Short: "Copy a bundled adapter YAML into .harness/config/agents/",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := cwd()
			if err != nil {
				return err
			}
			return agentcmd.Add(cmd.OutOrStdout(), dir, args[0])
		},
	}

	discoverC := &cobra.Command{
		Use:   "discover <binary>",
		Short: "Print a YAML scaffold for a CLI binary",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return agentcmd.Discover(cmd.OutOrStdout(), args[0])
		},
	}

	var skipRun bool
	certifyC := &cobra.Command{
		Use:   "certify <id>",
		Short: "Run the certification suite against an adapter",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := cwd()
			if err != nil {
				return err
			}
			_, err = agentcmd.Certify(cmd.Context(), cmd.OutOrStdout(), agentcmd.CertifyOptions{
				ID: args[0], StartDir: dir, SkipRun: skipRun,
			})
			return err
		},
	}
	certifyC.Flags().BoolVar(&skipRun, "skip-run", false, "skip checks that execute the CLI binary")

	c.AddCommand(listC, addC, discoverC, certifyC, newAgentLoginCmd(), newAgentInstallCmd())
	return c
}

func newAgentLoginCmd() *cobra.Command {
	var fromEnv string
	c := &cobra.Command{
		Use:   "login <id>",
		Short: "Store the API key referenced by an adapter's secret_ref",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := cwd()
			if err != nil {
				return err
			}
			reg, _, err := agentcmd.LoadAll(dir)
			if err != nil {
				return err
			}
			ad, ok := reg.Get(args[0])
			if !ok {
				return fmt.Errorf("agent %q not registered", args[0])
			}
			caps := ad.Capabilities()
			if caps.LoginCommand != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "CLI login: %s\n", caps.LoginCommand)
				if caps.AuthDocURL != "" {
					fmt.Fprintf(cmd.OutOrStdout(), "Docs:      %s\n", caps.AuthDocURL)
				}
				return nil
			}
			secretRef := guessSecretRef(args[0])
			fmt.Fprintf(cmd.OutOrStdout(), "Adapter %q has no CLI login. Storing secret %s.\n", args[0], secretRef)
			value, err := resolveSecretInput(cmd.InOrStdin(), cmd.OutOrStdout(), fromEnv, "")
			if err != nil {
				return err
			}
			backend, err := secrets.New().Set(strings.TrimPrefix(secretRef, "secret://"), value)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "\xe2\x9c\x93 stored %s in %s\n", secretRef, backend)
			return nil
		},
	}
	c.Flags().StringVar(&fromEnv, "from-env", "", "read API key from this env var")
	return c
}

func newAgentInstallCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "install <id>",
		Short: "Scaffold the YAML for a bundled adapter; alias for `agent add`",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := cwd()
			if err != nil {
				return err
			}
			return agentcmd.Add(cmd.OutOrStdout(), dir, args[0])
		},
	}
}

func guessSecretRef(id string) string {
	id = strings.TrimSuffix(id, "-api")
	id = strings.ReplaceAll(id, "-", "_")
	return "secret://" + id + "_api_key"
}
