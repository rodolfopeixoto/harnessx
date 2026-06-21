// SPDX-License-Identifier: MIT

package yaml

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
)

// jsonStreamFormatter parses the Claude Code / Codex JSON-Lines
// envelope as it streams in, and emits one humanised line per
// significant event to the underlying writer:
//
//	│ ● Read <path>
//	│ ● Write <path>
//	│ ● Edit <path>
//	│ $ <bash cmd>
//	│ ⋯ thinking…
//	│ ✓ <final result text>
//
// The huge `system.init` envelope (tools[], mcp_servers[], plugins[])
// is silently swallowed — that was the wall of text users saw in the
// v0.129 walk. On a JSON parse error the raw line is passed through
// so the user never gets nothing.
type jsonStreamFormatter struct {
	w           io.Writer
	buf         bytes.Buffer
	thinkingOn  bool
	lastEmit    string
	seenSysInit bool
}

// newJSONStreamFormatter wraps the caller-supplied writer. No prefix
// is added here — repl already wraps the live writer with its own
// line-prefix decorator. If the caller does not, the events still
// land on the writer one per line, just without the chat gutter.
func newJSONStreamFormatter(w io.Writer) *jsonStreamFormatter {
	return &jsonStreamFormatter{w: w}
}

func (f *jsonStreamFormatter) Write(p []byte) (int, error) {
	f.buf.Write(p)
	for {
		line, err := f.buf.ReadBytes('\n')
		if err != nil {
			// Re-buffer the partial line and stop until more bytes arrive.
			f.buf.Reset()
			f.buf.Write(line)
			return len(p), nil
		}
		f.handleLine(bytes.TrimRight(line, "\r\n"))
	}
}

// Flush consumes any trailing bytes that did not end in a newline so a
// final assistant message without a trailing newline still renders.
func (f *jsonStreamFormatter) Flush() {
	if f.buf.Len() == 0 {
		return
	}
	line := bytes.TrimRight(f.buf.Bytes(), "\r\n")
	f.buf.Reset()
	if len(line) > 0 {
		f.handleLine(line)
	}
}

func (f *jsonStreamFormatter) handleLine(line []byte) {
	if len(line) == 0 {
		return
	}
	// The JSONL envelope sometimes wraps the whole conversation as a
	// single JSON array (we saw both forms in the wild) — handle both.
	if line[0] == '[' {
		var arr []json.RawMessage
		if err := json.Unmarshal(line, &arr); err == nil {
			for _, m := range arr {
				f.handleEvent(m)
			}
			return
		}
	}
	f.handleEvent(line)
}

func (f *jsonStreamFormatter) handleEvent(raw json.RawMessage) {
	var env struct {
		Type    string `json:"type"`
		Subtype string `json:"subtype"`
		Result  string `json:"result"`
		Message struct {
			Content []struct {
				Type     string          `json:"type"`
				Name     string          `json:"name"`
				Text     string          `json:"text"`
				Thinking string          `json:"thinking"`
				Input    json.RawMessage `json:"input"`
			} `json:"content"`
		} `json:"message"`
	}
	if err := json.Unmarshal(raw, &env); err != nil {
		// Not JSON or unexpected shape — fall back to raw passthrough,
		// truncated so a 30 KB system-init dump still does not flood
		// the screen.
		s := string(raw)
		if len(s) > 200 {
			s = s[:200] + "…"
		}
		f.emit("? " + s)
		return
	}
	switch env.Type {
	case "system":
		if !f.seenSysInit {
			f.seenSysInit = true
			f.emit("• session ready")
		}
		return
	case "user", "rate_limit_event":
		return
	case "result":
		text := strings.TrimSpace(env.Result)
		if text == "" {
			return
		}
		f.emit("✓ " + truncForChat(text, 400))
		return
	case "assistant":
		for _, c := range env.Message.Content {
			f.renderAssistantBlock(c.Type, c.Name, c.Text, c.Thinking, c.Input)
		}
		return
	}
}

func (f *jsonStreamFormatter) renderAssistantBlock(kind, toolName, text, thinking string, input json.RawMessage) {
	switch kind {
	case "text":
		t := strings.TrimSpace(text)
		if t == "" {
			return
		}
		f.emit("» " + truncForChat(t, 200))
	case "thinking":
		if f.thinkingOn {
			return
		}
		f.thinkingOn = true
		f.emit("⋯ thinking…")
	case "tool_use":
		f.thinkingOn = false
		switch toolName {
		case "Read":
			if p := jsonString(input, "file_path"); p != "" {
				f.emit("● Read " + shortenPath(p))
			}
		case "Write":
			if p := jsonString(input, "file_path"); p != "" {
				f.emit("● Write " + shortenPath(p))
			}
		case "Edit":
			if p := jsonString(input, "file_path"); p != "" {
				f.emit("● Edit " + shortenPath(p))
			}
		case "Bash":
			cmd := jsonString(input, "command")
			if cmd != "" {
				f.emit("$ " + truncForChat(cmd, 120))
			}
		case "Grep":
			pat := jsonString(input, "pattern")
			f.emit("● Grep " + truncForChat(pat, 80))
		case "Glob":
			f.emit("● Glob " + truncForChat(jsonString(input, "pattern"), 80))
		default:
			f.emit("● " + toolName)
		}
	}
}

func (f *jsonStreamFormatter) emit(body string) {
	line := body + "\n"
	if line == f.lastEmit {
		return
	}
	f.lastEmit = line
	_, _ = f.w.Write([]byte(line))
}

func jsonString(raw json.RawMessage, key string) string {
	if len(raw) == 0 {
		return ""
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(raw, &m); err != nil {
		return ""
	}
	v, ok := m[key]
	if !ok {
		return ""
	}
	var s string
	if err := json.Unmarshal(v, &s); err != nil {
		return ""
	}
	return s
}

func shortenPath(p string) string {
	// Trim a leading project root if present so the line reads as
	// "● Read app/storage.py" instead of an absolute path noise wall.
	idx := strings.LastIndex(p, "/")
	if idx < 0 || idx == len(p)-1 {
		return p
	}
	// Show last two path segments — feature folder + filename is usually
	// enough context.
	parts := strings.Split(p, "/")
	if len(parts) >= 2 {
		return parts[len(parts)-2] + "/" + parts[len(parts)-1]
	}
	return p
}

func truncForChat(s string, max int) string {
	s = strings.TrimSpace(s)
	if max <= 0 || len(s) <= max {
		return s
	}
	return s[:max] + "…"
}

// jsonFormat returns whether the spec uses a JSON-shaped output stream
// that the formatter understands. Today only the claude/codex flavour
// emits the {type,subtype,...} envelope we parse.
func jsonFormat(format string) bool {
	switch strings.ToLower(format) {
	case "json", "jsonl", "json-lines":
		return true
	}
	return false
}

// Sentinel: keep the writer interface assertion close to the type so
// future refactors that change the embed surface fail at compile time.
var _ io.Writer = (*jsonStreamFormatter)(nil)
