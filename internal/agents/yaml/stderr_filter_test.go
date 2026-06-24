// SPDX-License-Identifier: MIT

package yaml

import (
	"bytes"
	"strings"
	"testing"
)

func TestFilteringWriterDropsKnownNoise(t *testing.T) {
	var buf bytes.Buffer
	w := newFilteringWriter(&buf)
	lines := []string{
		"ERROR codex_core::session: failed to load skill /Users/x/.agents/skills/remotion-to-hyperframes/SKILL.md: invalid description: exceeds maximum length of 1024 characters\n",
		"chat: real output line\n",
		"failed to load skill /tmp/SKILL.md: somethingelse\n",
		"MCP server filesystem exited with code 1\n",
		"another legit line\n",
	}
	for _, l := range lines {
		_, _ = w.Write([]byte(l))
	}
	got := buf.String()
	for _, noise := range []string{
		"failed to load skill",
		"ERROR codex_core",
		"MCP server filesystem exited",
	} {
		if strings.Contains(got, noise) {
			t.Errorf("noise pattern %q leaked through: %s", noise, got)
		}
	}
	for _, legit := range []string{"chat: real output line", "another legit line"} {
		if !strings.Contains(got, legit) {
			t.Errorf("legit line dropped: %s", got)
		}
	}
}

func TestFilteringWriterBuffersPartialLines(t *testing.T) {
	var buf bytes.Buffer
	w := newFilteringWriter(&buf)
	_, _ = w.Write([]byte("failed to load skill /a/SKILL"))
	if buf.Len() != 0 {
		t.Errorf("partial noisy line must buffer, got %q", buf.String())
	}
	_, _ = w.Write([]byte(".md: bad\n"))
	if buf.Len() != 0 {
		t.Errorf("completed noisy line must be dropped, got %q", buf.String())
	}
	_, _ = w.Write([]byte("plain output\n"))
	if !strings.Contains(buf.String(), "plain output") {
		t.Error("subsequent legit line must pass through")
	}
}
