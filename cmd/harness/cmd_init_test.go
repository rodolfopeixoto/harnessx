// SPDX-License-Identifier: MIT

package main

import (
	"strings"
	"testing"
)

func TestInitCmdHasGitFlags(t *testing.T) {
	c := newInitCmd()
	for _, want := range []string{"force", "git", "all", "git-branch", "slug"} {
		if c.Flags().Lookup(want) == nil {
			t.Errorf("init flag missing: %s", want)
		}
	}
}

func TestInitCmdLongMentionsGitAndAll(t *testing.T) {
	c := newInitCmd()
	low := strings.ToLower(c.Long)
	for _, want := range []string{"git init", "registry"} {
		if !strings.Contains(low, want) {
			t.Errorf("Long missing %q", want)
		}
	}
}

func TestInitCmdGitBranchDefault(t *testing.T) {
	c := newInitCmd()
	f := c.Flags().Lookup("git-branch")
	if f == nil || f.DefValue != "main" {
		t.Errorf("git-branch default: want main, got %v", f)
	}
}
