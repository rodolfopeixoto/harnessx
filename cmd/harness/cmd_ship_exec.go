// SPDX-License-Identifier: MIT

package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

// runHarness shells out to the harness binary with plain-mode env forced so
// captured stdout/stderr does not contain ANSI escapes — ship loops need to
// pattern-match on rate-limit markers.
func runHarness(ctx context.Context, bin, root string, out io.Writer, args []string) error {
	cmd := exec.CommandContext(ctx, bin, args...)
	cmd.Dir = root
	cmd.Env = append(os.Environ(), "HARNESS_PLAIN=1", "NO_COLOR=1")
	var stderr bytes.Buffer
	cmd.Stdout = out
	cmd.Stderr = io.MultiWriter(out, &stderr)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%w: %s", err, stderr.String())
	}
	return nil
}

func runGit(ctx context.Context, root string, args ...string) error {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git %s: %v: %s", strings.Join(args, " "), err, string(out))
	}
	return nil
}

func gitDirty(ctx context.Context, root string) (bool, error) {
	cmd := exec.CommandContext(ctx, "git", "status", "--porcelain")
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("git status: %w", err)
	}
	return len(bytes.TrimSpace(out)) > 0, nil
}

var rateLimitRe = regexp.MustCompile(`(?i)(429|rate[\s_-]?limit|too many requests|quota.*exceed)`)

func isRateLimit(s string) bool { return rateLimitRe.MatchString(s) }
