// SPDX-License-Identifier: MIT

package backup

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

type Rclone struct {
	Bin string
}

func NewRclone() (*Rclone, error) {
	bin, err := exec.LookPath("rclone")
	if err != nil {
		return nil, errors.New("backup: rclone not on PATH (run: harness install rclone)")
	}
	return &Rclone{Bin: bin}, nil
}

func (r *Rclone) Listremotes(ctx context.Context) ([]string, error) {
	out, err := exec.CommandContext(ctx, r.Bin, "listremotes").Output()
	if err != nil {
		return nil, err
	}
	var names []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		names = append(names, strings.TrimSuffix(line, ":"))
	}
	return names, nil
}

func (r *Rclone) Copy(ctx context.Context, src, dst string) error {
	return run(ctx, r.Bin, "copy", src, dst, "--progress")
}

func (r *Rclone) Ls(ctx context.Context, remotePath string) ([]string, error) {
	out, err := exec.CommandContext(ctx, r.Bin, "lsf", remotePath).Output()
	if err != nil {
		return nil, err
	}
	var lines []string
	for _, l := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		l = strings.TrimSpace(l)
		if l == "" {
			continue
		}
		lines = append(lines, l)
	}
	return lines, nil
}

func (r *Rclone) Sync(ctx context.Context, src, dst string, dryRun bool) error {
	args := []string{"sync", src, dst, "--progress"}
	if dryRun {
		args = append(args, "--dry-run")
	}
	return run(ctx, r.Bin, args...)
}

func (r *Rclone) ConfigCreate(ctx context.Context, name, provider string, extraArgs ...string) error {
	args := append([]string{"config", "create", name, provider}, extraArgs...)
	return run(ctx, r.Bin, args...)
}

func (r *Rclone) ConfigInteractive(ctx context.Context, name, provider string) error {
	cmd := exec.CommandContext(ctx, r.Bin, "config", "create", name, provider)
	return cmd.Run()
}

func run(ctx context.Context, bin string, args ...string) error {
	cmd := exec.CommandContext(ctx, bin, args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("rclone %v: %w: %s", args, err, strings.TrimSpace(stderr.String()))
	}
	return nil
}
