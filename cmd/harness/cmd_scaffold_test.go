// SPDX-License-Identifier: MIT

package main

import (
	"strings"
	"testing"
)

func TestScaffoldCmdHasSubcommands(t *testing.T) {
	c := newScaffoldCmd()
	wantSubs := map[string]bool{"list": false, "show": false, "apply": false}
	for _, sub := range c.Commands() {
		for name := range wantSubs {
			if strings.HasPrefix(sub.Use, name) {
				wantSubs[name] = true
			}
		}
	}
	for name, ok := range wantSubs {
		if !ok {
			t.Errorf("subcommand missing: %s", name)
		}
	}
}

func TestScaffoldApplyFlags(t *testing.T) {
	c := newScaffoldCmd()
	var apply *strings.Builder = nil
	_ = apply
	for _, sub := range c.Commands() {
		if !strings.HasPrefix(sub.Use, "apply") {
			continue
		}
		for _, want := range []string{"name", "apply", "with-git", "with-deps", "force", "git-branch"} {
			if sub.Flags().Lookup(want) == nil {
				t.Errorf("apply flag missing: %s", want)
			}
		}
		return
	}
	t.Fatal("apply subcommand not found")
}

func TestTruncStrSharedFn(t *testing.T) {
	if got := truncStr("abcde", 50); got != "abcde" {
		t.Errorf("passthrough: got %q", got)
	}
	if got := truncStr("abcde", 3); got != "ab…" {
		t.Errorf("trunc: got %q", got)
	}
}
