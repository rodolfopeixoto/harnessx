// SPDX-License-Identifier: MIT

package main

import "testing"

func TestWrapperCommandsRegistered(t *testing.T) {
	for _, n := range []string{"test", "lint", "dev", "bench"} {
		c := wrapperCmd(n, "")
		if c.Use != n {
			t.Errorf("Use: want %q got %q", n, c.Use)
		}
		if c.RunE == nil {
			t.Errorf("%s: RunE missing", n)
		}
	}
}

func TestProfileCmdHasMemAndCpu(t *testing.T) {
	c := newProfileCmd()
	subs := map[string]bool{"mem": false, "cpu": false}
	for _, sc := range c.Commands() {
		subs[sc.Use] = true
	}
	for k, ok := range subs {
		if !ok {
			t.Errorf("profile subcmd missing: %s", k)
		}
	}
}
