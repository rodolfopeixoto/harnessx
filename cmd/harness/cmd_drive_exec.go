// SPDX-License-Identifier: MIT

package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// runHarnessChild shells out to the compiled harness binary. `bin` comes
// from os.Executable() in the caller, so paths never come from user input.
func runHarnessChild(ctx context.Context, bin, dir string, out io.Writer, args []string) error {
	c := exec.CommandContext(ctx, bin, args...) //nolint:gosec // bin is the harness binary path from os.Executable(), args are harness subcommands
	c.Dir = dir
	c.Env = append(os.Environ(), "HARNESS_PLAIN=1", "NO_COLOR=1")
	c.Stdout = out
	c.Stderr = out
	if err := c.Run(); err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			return fmt.Errorf("exit %d", ee.ExitCode())
		}
		return err
	}
	return nil
}

func runGitInDir(ctx context.Context, dir string, args ...string) error {
	c := exec.CommandContext(ctx, "git", args...)
	c.Dir = dir
	var buf bytes.Buffer
	c.Stdout = &buf
	c.Stderr = &buf
	if err := c.Run(); err != nil {
		return fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, buf.String())
	}
	return nil
}

func gitTreeSnapshot(ctx context.Context, root string) string {
	c := exec.CommandContext(ctx, "git", "status", "--porcelain")
	c.Dir = root
	out, err := c.Output()
	if err != nil {
		return ""
	}
	return string(out)
}
