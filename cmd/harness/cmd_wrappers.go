// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ropeixoto/harnessx/internal/projectcfg"
	"github.com/ropeixoto/harnessx/internal/ui"
)

func newTestCmd() *cobra.Command  { return wrapperCmd("test", "Run the project test command") }
func newLintCmd() *cobra.Command  { return wrapperCmd("lint", "Run the project lint command") }
func newDevCmd() *cobra.Command   { return wrapperCmd("dev", "Run the project dev server") }
func newBenchCmd() *cobra.Command { return wrapperCmd("bench", "Run the project benchmark suite") }

func newProfileCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "profile",
		Short: "Run memory or CPU profiler against the project",
	}
	mem := wrapperCmd("profile-mem", "Memory profile")
	mem.Use = "mem"
	cpu := wrapperCmd("profile-cpu", "CPU profile")
	cpu.Use = "cpu"
	c.AddCommand(mem, cpu)
	return c
}

func wrapperCmd(name, short string) *cobra.Command {
	return &cobra.Command{
		Use:   name,
		Short: short,
		Long: fmt.Sprintf(`Reads the resolved command for %q from .harness/config/project.yaml.
Falls back to the bundled defaults for the detected stack. Pass extra
arguments after -- to append them to the underlying command.`, name),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := cwd()
			if err != nil {
				return err
			}
			line, err := projectcfg.Resolve(dir, name)
			if err != nil {
				return err
			}
			return runShellLine(cmd.Context(), cmd.OutOrStdout(), cmd.ErrOrStderr(), dir, line, args)
		},
	}
}

func runShellLine(ctx context.Context, stdout, stderr io.Writer, dir, line string, extra []string) error {
	line = strings.TrimSpace(line)
	if line == "" {
		return errors.New("wrapper: empty command")
	}
	if len(extra) > 0 {
		line = line + " " + strings.Join(extra, " ")
	}
	fmt.Fprintf(stdout, "%s %s\n", ui.Muted.Render("→"), ui.Accent.Render(line))
	cmd := exec.CommandContext(ctx, "/bin/sh", "-c", line)
	cmd.Dir = dir
	cmd.Stdin = os.Stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	cmd.Env = os.Environ()
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(stderr, "%s %s\n", ui.MarkFail(), ui.Error.Render(err.Error()))
		return err
	}
	fmt.Fprintln(stdout, ui.MarkSuccess()+" done")
	return nil
}
