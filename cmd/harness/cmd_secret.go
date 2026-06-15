// SPDX-License-Identifier: MIT

package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/ropeixoto/harnessx/internal/secrets"
)

func newSecretCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "secret",
		Short: "Manage API keys and tokens used by HarnessX adapters",
		Long: `Cross-platform secret store. Backends tried in order:

  env             — process env (HARNESS_SECRET_<UPPER> or <UPPER>)
  keychain        — macOS Keychain via /usr/bin/security
  secret_service  — Linux libsecret via secret-tool
  encrypted_file  — AES-GCM file at ~/.harness/secrets.enc (last resort)

Secrets are referenced from adapter YAML as secret://<name>.
Values are never logged; harness secret get redacts by default.`,
	}
	c.AddCommand(newSecretListCmd(), newSecretSetCmd(), newSecretGetCmd(), newSecretUnsetCmd(), newSecretInfoCmd())
	return c
}

func newSecretListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List secret names per backend (values redacted)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			store := secrets.New()
			perBackend, err := store.List()
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "BACKEND\tNAME")
			for _, b := range store.Backends() {
				names := perBackend[b.Name()]
				if len(names) == 0 {
					continue
				}
				for _, n := range names {
					fmt.Fprintf(w, "%s\t%s\n", b.Name(), n)
				}
			}
			return w.Flush()
		},
	}
}

func newSecretSetCmd() *cobra.Command {
	var (
		fromEnv string
		from    string
	)
	c := &cobra.Command{
		Use:   "set <name>",
		Short: "Set a secret (prompts on stdin; --from-env or --from-file accepted)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			value, err := resolveSecretInput(cmd.InOrStdin(), cmd.OutOrStdout(), fromEnv, from)
			if err != nil {
				return err
			}
			backend, err := secrets.New().Set(name, value)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "✓ stored %q in %s\n", name, backend)
			return nil
		},
	}
	c.Flags().StringVar(&fromEnv, "from-env", "", "read value from this env var")
	c.Flags().StringVar(&from, "from-file", "", "read value from this file")
	return c
}

func resolveSecretInput(in io.Reader, out io.Writer, fromEnv, fromFile string) (string, error) {
	if fromEnv != "" {
		v := os.Getenv(fromEnv)
		if v == "" {
			return "", fmt.Errorf("env %s is empty", fromEnv)
		}
		return v, nil
	}
	if fromFile != "" {
		data, err := os.ReadFile(fromFile)
		if err != nil {
			return "", err
		}
		return strings.TrimRight(string(data), "\n"), nil
	}
	if f, ok := in.(*os.File); ok && term.IsTerminal(int(f.Fd())) {
		fmt.Fprint(out, "value (hidden): ")
		raw, err := term.ReadPassword(int(f.Fd()))
		if err != nil {
			return "", err
		}
		fmt.Fprintln(out)
		return string(raw), nil
	}
	reader := bufio.NewReader(in)
	v, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", err
	}
	return strings.TrimRight(v, "\n"), nil
}

func newSecretGetCmd() *cobra.Command {
	var reveal bool
	c := &cobra.Command{
		Use:   "get <name>",
		Short: "Get a secret (redacted unless --reveal)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			v, err := secrets.New().Get(args[0])
			if err != nil {
				return err
			}
			if reveal {
				fmt.Fprintln(cmd.OutOrStdout(), v)
			} else {
				fmt.Fprintln(cmd.OutOrStdout(), redact(v))
			}
			return nil
		},
	}
	c.Flags().BoolVar(&reveal, "reveal", false, "print plaintext (dangerous)")
	return c
}

func newSecretUnsetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "unset <name>",
		Short: "Remove a secret from every writable backend",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := secrets.New().Delete(args[0]); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "✓ removed %q\n", args[0])
			return nil
		},
	}
}

func newSecretInfoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "info",
		Short: "Show which backends are active on this host",
		RunE: func(cmd *cobra.Command, _ []string) error {
			store := secrets.New()
			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "BACKEND\tAVAILABLE\tWRITABLE")
			for _, b := range store.Backends() {
				avail := "—"
				if b.Available() {
					avail = "✓"
				}
				writeStr := "✓"
				if b.Name() == "env" {
					writeStr = "—"
				}
				fmt.Fprintf(w, "%s\t%s\t%s\n", b.Name(), avail, writeStr)
			}
			return w.Flush()
		},
	}
}

func redact(s string) string {
	if len(s) <= 4 {
		return "****"
	}
	return s[:2] + strings.Repeat("*", len(s)-4) + s[len(s)-2:]
}
