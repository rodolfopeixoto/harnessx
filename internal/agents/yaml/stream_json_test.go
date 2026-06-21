// SPDX-License-Identifier: MIT

package yaml

import (
	"bytes"
	"strings"
	"testing"
)

// A trimmed fixture of the Claude Code JSON-Lines envelope users saw
// flooding the chat in v0.129. The full payload also has a 30 KB
// `tools[]` + `mcp_servers[]` + `slash_commands[]` block that the
// formatter should swallow silently.
const claudeJSONLFixture = `[{"type":"system","subtype":"init","cwd":"/p","tools":["Read","Write","Edit","Bash"],"mcp_servers":[{"name":"x","status":"connected"}],"model":"claude-opus-4-7","session_id":"abc"},{"type":"assistant","message":{"content":[{"type":"thinking","thinking":"…"}]}},{"type":"assistant","message":{"content":[{"type":"tool_use","name":"Read","input":{"file_path":"/p/app/storage.py"}}]}},{"type":"assistant","message":{"content":[{"type":"tool_use","name":"Write","input":{"file_path":"/p/tests/test_products.py"}}]}},{"type":"assistant","message":{"content":[{"type":"tool_use","name":"Bash","input":{"command":"pytest -q"}}]}},{"type":"assistant","message":{"content":[{"type":"text","text":"Done. 9/9 pass."}]}},{"type":"result","subtype":"success","is_error":false,"result":"Done. 9/9 pass."}]
`

func TestJSONStreamFormatterEmitsHumanisedLines(t *testing.T) {
	var buf bytes.Buffer
	f := newJSONStreamFormatter(&buf)
	if _, err := f.Write([]byte(claudeJSONLFixture)); err != nil {
		t.Fatal(err)
	}
	f.Flush()
	got := buf.String()
	want := []string{
		"• session ready",
		"⋯ thinking…",
		"● Read app/storage.py",
		"● Write tests/test_products.py",
		"$ pytest -q",
		"» Done. 9/9 pass.",
		"✓ Done. 9/9 pass.",
	}
	for _, w := range want {
		if !strings.Contains(got, w) {
			t.Errorf("missing %q in output:\n%s", w, got)
		}
	}
	if strings.Contains(got, "mcp_servers") || strings.Contains(got, "slash_commands") {
		t.Errorf("system.init metadata leaked into stream:\n%s", got)
	}
}

func TestJSONStreamFormatterFallsBackOnGarbage(t *testing.T) {
	var buf bytes.Buffer
	f := newJSONStreamFormatter(&buf)
	_, _ = f.Write([]byte("not a json line\n"))
	if !strings.Contains(buf.String(), "? not a json line") {
		t.Errorf("expected raw fallback, got %q", buf.String())
	}
}

func TestJSONStreamFormatterDeduplicatesThinking(t *testing.T) {
	var buf bytes.Buffer
	f := newJSONStreamFormatter(&buf)
	thinkLine := `{"type":"assistant","message":{"content":[{"type":"thinking","thinking":"…"}]}}` + "\n"
	for i := 0; i < 5; i++ {
		_, _ = f.Write([]byte(thinkLine))
	}
	count := strings.Count(buf.String(), "thinking…")
	if count != 1 {
		t.Errorf("expected 1 thinking line, got %d (output=%q)", count, buf.String())
	}
}

func TestShortenPathTrimsToLastTwoSegments(t *testing.T) {
	cases := map[string]string{
		"/Users/x/dev/p/app/storage.py": "app/storage.py",
		"app/storage.py":                "app/storage.py",
		"storage.py":                    "storage.py",
	}
	for in, want := range cases {
		if got := shortenPath(in); got != want {
			t.Errorf("shortenPath(%q)=%q want %q", in, got, want)
		}
	}
}
