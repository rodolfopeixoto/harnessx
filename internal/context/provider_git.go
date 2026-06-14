// SPDX-License-Identifier: MIT

package context

import (
	"bytes"
	"context"
	"os/exec"
	"strings"
)

type GitProvider struct{}

func (GitProvider) Name() string { return "git" }

func (GitProvider) Apply(ctx context.Context, root string, pack *Pack) error {
	if !hasBinary("git") {
		pack.Stats.ProvidersSkipped++
		return nil
	}
	pack.GitStatus = strings.TrimSpace(runGit(ctx, root, "status", "--porcelain"))
	pack.GitDiff = strings.TrimSpace(runGit(ctx, root, "diff", "HEAD"))

	// Promote changed paths into RelevantFiles for downstream providers.
	for _, line := range strings.Split(pack.GitStatus, "\n") {
		if len(line) < 4 {
			continue
		}
		path := strings.TrimSpace(line[3:])
		if path == "" {
			continue
		}
		pack.RelevantFiles = appendFile(pack.RelevantFiles, FileEntry{
			Path: path, Reason: "git_status",
		})
	}
	pack.Stats.ProvidersRan++
	return nil
}

func runGit(ctx context.Context, dir string, args ...string) string {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	var out, errb bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errb
	if err := cmd.Run(); err != nil {
		return ""
	}
	return out.String()
}

func appendFile(in []FileEntry, e FileEntry) []FileEntry {
	for _, existing := range in {
		if existing.Path == e.Path {
			return in
		}
	}
	return append(in, e)
}
