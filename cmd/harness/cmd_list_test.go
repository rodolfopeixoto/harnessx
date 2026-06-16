// SPDX-License-Identifier: MIT

package main

import (
	"strings"
	"testing"
)

func TestListCmdUseAndArgs(t *testing.T) {
	c := newListCmd()
	if c.Use != "list" {
		t.Errorf("Use: %q", c.Use)
	}
	if err := c.Args(c, []string{"extra"}); err == nil {
		t.Error("list with extra args should error")
	}
	if err := c.Args(c, []string{}); err != nil {
		t.Errorf("list with no args should be ok: %v", err)
	}
}

func TestListCmdShortMentionsComposite(t *testing.T) {
	c := newListCmd()
	if !strings.Contains(strings.ToLower(c.Short), "composite") {
		t.Errorf("Short should mention composite: %q", c.Short)
	}
}
