// SPDX-License-Identifier: MIT

package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

func newCompletionCmd() *cobra.Command {
	c := &cobra.Command{
		Use:                   "completion <bash|zsh|fish|powershell>",
		Short:                 "Generate shell completion script",
		Long:                  "Output a shell-specific completion script.\n\n  source <(harness completion bash)\n  harness completion zsh > \"${fpath[1]}/_harness\"\n  harness completion install               # auto-detect shell + write to the right path",
		Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return writeCompletion(cmd, args[0], os.Stdout)
		},
	}
	c.AddCommand(newCompletionInstallCmd())
	return c
}

func writeCompletion(cmd *cobra.Command, shell string, out interface{ Write(p []byte) (int, error) }) error {
	switch shell {
	case "bash":
		return cmd.Root().GenBashCompletion(out.(*os.File))
	case "zsh":
		return cmd.Root().GenZshCompletion(out.(*os.File))
	case "fish":
		return cmd.Root().GenFishCompletion(out.(*os.File), true)
	case "powershell":
		return cmd.Root().GenPowerShellCompletionWithDesc(out.(*os.File))
	}
	return fmt.Errorf("completion: unknown shell %q", shell)
}

func newCompletionInstallCmd() *cobra.Command {
	var (
		shell  string
		dryRun bool
	)
	c := &cobra.Command{
		Use:   "install",
		Short: "Write completion script to the conventional path for your shell",
		Long: `Auto-detect SHELL (or --shell) and write the completion script to
the conventional location:

  bash (macOS, brew):   /usr/local/etc/bash_completion.d/harness
  bash (linux):         /etc/bash_completion.d/harness OR ~/.bash_completion.d/harness
  zsh:                  $ZDOTDIR/.zsh/completion/_harness OR ~/.zsh/completion/_harness
                        (add the dir to fpath then 'compinit')
  fish:                 ~/.config/fish/completions/harness.fish
  powershell:           Save-Module path; the command prints the path to source

  --dry-run                print path without writing
  --shell bash|zsh|fish    skip auto-detect`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			s := shell
			if s == "" {
				s = detectShell()
			}
			if s == "" {
				return fmt.Errorf("completion install: could not detect shell; pass --shell")
			}
			path, err := completionPath(s)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			if dryRun {
				fmt.Fprintf(out, "→ dry-run: would write %s completion to %s\n", s, path)
				return nil
			}
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				return err
			}
			var buf bytes.Buffer
			rendered, err := renderCompletion(cmd.Root(), s, &buf)
			if err != nil {
				return err
			}
			if err := os.WriteFile(path, rendered, 0o644); err != nil {
				return err
			}
			fmt.Fprintf(out, "✓ wrote %s completion to %s\n", s, path)
			fmt.Fprintln(out, hintAfterWrite(s, path))
			return nil
		},
	}
	c.Flags().StringVar(&shell, "shell", "", "bash|zsh|fish|powershell (default: auto-detect from $SHELL)")
	c.Flags().BoolVar(&dryRun, "dry-run", false, "print target path without writing")
	return c
}

func detectShell() string {
	sh := strings.ToLower(filepath.Base(os.Getenv("SHELL")))
	for _, known := range []string{"bash", "zsh", "fish"} {
		if sh == known {
			return known
		}
	}
	if runtime.GOOS == "windows" {
		return "powershell"
	}
	return ""
}

func completionPath(shell string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	switch shell {
	case "bash":
		if runtime.GOOS == "darwin" {
			if _, err := os.Stat("/usr/local/etc/bash_completion.d"); err == nil {
				return "/usr/local/etc/bash_completion.d/harness", nil
			}
			if _, err := os.Stat("/opt/homebrew/etc/bash_completion.d"); err == nil {
				return "/opt/homebrew/etc/bash_completion.d/harness", nil
			}
		}
		if _, err := os.Stat("/etc/bash_completion.d"); err == nil {
			return "/etc/bash_completion.d/harness", nil
		}
		return filepath.Join(home, ".bash_completion.d", "harness"), nil
	case "zsh":
		return filepath.Join(home, ".zsh", "completion", "_harness"), nil
	case "fish":
		return filepath.Join(home, ".config", "fish", "completions", "harness.fish"), nil
	case "powershell":
		return filepath.Join(home, "Documents", "PowerShell", "Modules", "harness", "harness.ps1"), nil
	}
	return "", fmt.Errorf("completion install: unknown shell %q", shell)
}

func renderCompletion(root *cobra.Command, shell string, buf *bytes.Buffer) ([]byte, error) {
	switch shell {
	case "bash":
		if err := root.GenBashCompletion(buf); err != nil {
			return nil, err
		}
	case "zsh":
		if err := root.GenZshCompletion(buf); err != nil {
			return nil, err
		}
	case "fish":
		if err := root.GenFishCompletion(buf, true); err != nil {
			return nil, err
		}
	case "powershell":
		if err := root.GenPowerShellCompletionWithDesc(buf); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("completion: unknown shell %q", shell)
	}
	return buf.Bytes(), nil
}

func hintAfterWrite(shell, path string) string {
	switch shell {
	case "zsh":
		dir := filepath.Dir(path)
		return fmt.Sprintf("Add to ~/.zshrc:\n  fpath+=(%s)\n  autoload -U compinit && compinit", dir)
	case "bash":
		return "Open a new shell or run: source " + path
	case "fish":
		return "fish completions pick this up automatically; open a new shell."
	case "powershell":
		return "Add to $PROFILE:\n  . " + path
	}
	return ""
}
