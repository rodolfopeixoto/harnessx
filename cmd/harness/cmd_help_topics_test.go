// SPDX-License-Identifier: MIT

package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestTopicsHasCoreTopics(t *testing.T) {
	want := []string{"quickstart", "agents", "sensors", "hooks", "autonomy", "mcp", "update", "input", "tracker", "billing", "do", "loop", "scaffold"}
	for _, w := range want {
		if _, ok := topics[w]; !ok {
			t.Errorf("topic missing: %s", w)
		}
	}
}

func TestTopicsBodiesAreNonEmpty(t *testing.T) {
	for name, body := range topics {
		if strings.TrimSpace(body) == "" {
			t.Errorf("topic %s body is empty", name)
		}
	}
}

func TestFirstLineExtractsHeadline(t *testing.T) {
	if got := firstLine("hello\nworld\n"); got != "hello" {
		t.Errorf("firstLine: got %q", got)
	}
	if got := firstLine("oneliner"); got != "oneliner" {
		t.Errorf("firstLine without newline: got %q", got)
	}
}

func TestHelpCmdListsTopics(t *testing.T) {
	c := newHelpCmd()
	var buf bytes.Buffer
	c.SetOut(&buf)
	if err := c.RunE(c, []string{}); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, want := range []string{"Topics:", "quickstart", "do", "loop", "scaffold"} {
		if !strings.Contains(out, want) {
			t.Errorf("help no-args missing %q", want)
		}
	}
}

func TestHelpCmdPrintsTopic(t *testing.T) {
	c := newHelpCmd()
	var buf bytes.Buffer
	c.SetOut(&buf)
	if err := c.RunE(c, []string{"do"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "harness do") {
		t.Errorf("help do should print body: %s", buf.String())
	}
}

func TestHelpCmdUnknownTopic(t *testing.T) {
	c := newHelpCmd()
	c.SetOut(new(bytes.Buffer))
	err := c.RunE(c, []string{"definitely-not-real"})
	if err == nil {
		t.Error("unknown topic should error")
	}
}
