// SPDX-License-Identifier: MIT

package main

import (
	"strings"
	"testing"
)

func TestNewLoopCmdRegistersFlags(t *testing.T) {
	c := newLoopCmd()
	if c.Use != "loop \"<prompt>\"" {
		t.Errorf("Use: %q", c.Use)
	}
	for _, want := range []string{"agent", "autonomy", "budget-usd", "max-attempts", "lint", "test", "apply"} {
		if c.Flags().Lookup(want) == nil {
			t.Errorf("flag missing: %s", want)
		}
	}
}

func TestLoopCmdShortIncludesLoopKeyword(t *testing.T) {
	c := newLoopCmd()
	if !strings.Contains(strings.ToLower(c.Short), "loop") {
		t.Errorf("Short should mention 'loop': %q", c.Short)
	}
}

func TestLoopCmdAcceptsZeroOrMoreArgs(t *testing.T) {
	c := newLoopCmd()
	if err := c.Args(c, []string{}); err != nil {
		t.Errorf("loop with zero args should now be allowed (F59): %v", err)
	}
	if err := c.Args(c, []string{"fix x"}); err != nil {
		t.Errorf("loop with one arg should accept: %v", err)
	}
}

func TestLoopCmdLongMentionsDeterministicLoop(t *testing.T) {
	c := newLoopCmd()
	for _, want := range []string{"lint", "test", "canonicalised", "auto-detect"} {
		if !strings.Contains(strings.ToLower(c.Long), strings.ToLower(want)) {
			t.Errorf("Long missing %q", want)
		}
	}
}

func TestLoopCmdDefaultMaxAttempts(t *testing.T) {
	c := newLoopCmd()
	f := c.Flags().Lookup("max-attempts")
	if f.DefValue != "3" {
		t.Errorf("max-attempts default: want 3, got %q", f.DefValue)
	}
}

func TestLoopCmdDefaultAutonomy(t *testing.T) {
	c := newLoopCmd()
	f := c.Flags().Lookup("autonomy")
	if f.DefValue != "safe_execute" {
		t.Errorf("autonomy default: want safe_execute, got %q", f.DefValue)
	}
}
