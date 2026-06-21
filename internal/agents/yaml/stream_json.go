// SPDX-License-Identifier: MIT

package yaml

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"

	"github.com/ropeixoto/harnessx/internal/platform/constants"
)

type jsonEvent struct {
	Type    string `json:"type"`
	Subtype string `json:"subtype"`
	Result  string `json:"result"`
	Message struct {
		Content []jsonContentBlock `json:"content"`
	} `json:"message"`
}

type jsonContentBlock struct {
	Type     string          `json:"type"`
	Name     string          `json:"name"`
	Text     string          `json:"text"`
	Thinking string          `json:"thinking"`
	Input    json.RawMessage `json:"input"`
}

type jsonStreamFormatter struct {
	w           io.Writer
	buf         bytes.Buffer
	thinkingOn  bool
	lastEmit    string
	seenSysInit bool
}

func newJSONStreamFormatter(w io.Writer) *jsonStreamFormatter {
	return &jsonStreamFormatter{w: w}
}

func (f *jsonStreamFormatter) Write(p []byte) (int, error) {
	f.buf.Write(p)
	for {
		line, err := f.buf.ReadBytes('\n')
		if err != nil {
			f.buf.Reset()
			f.buf.Write(line)
			return len(p), nil
		}
		f.handleLine(bytes.TrimRight(line, "\r\n"))
	}
}

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
	var env jsonEvent
	if err := json.Unmarshal(raw, &env); err != nil {
		f.emit("? " + truncForChat(string(raw), constants.ChatRawFallbackMax))
		return
	}
	switch env.Type {
	case "system":
		if !f.seenSysInit {
			f.seenSysInit = true
			f.emit("• session ready")
		}
	case "user", "rate_limit_event":
		return
	case "result":
		if text := strings.TrimSpace(env.Result); text != "" {
			f.emit("✓ " + truncForChat(text, constants.ChatResultMax))
		}
	case "assistant":
		for _, c := range env.Message.Content {
			f.renderAssistantBlock(c)
		}
	}
}

func (f *jsonStreamFormatter) renderAssistantBlock(c jsonContentBlock) {
	switch c.Type {
	case "text":
		if t := strings.TrimSpace(c.Text); t != "" {
			f.emit("» " + truncForChat(t, constants.ChatPromptTextMax))
		}
	case "thinking":
		if f.thinkingOn {
			return
		}
		f.thinkingOn = true
		f.emit("⋯ thinking…")
	case "tool_use":
		f.thinkingOn = false
		f.renderToolUse(c.Name, c.Input)
	}
}

func (f *jsonStreamFormatter) renderToolUse(name string, input json.RawMessage) {
	switch name {
	case "Read":
		f.emitIfPath("● Read ", input)
	case "Write":
		f.emitIfPath("● Write ", input)
	case "Edit":
		f.emitIfPath("● Edit ", input)
	case "Bash":
		if cmd := jsonString(input, "command"); cmd != "" {
			f.emit("$ " + truncForChat(cmd, constants.ChatBashCommandMax))
		}
	case "Grep":
		f.emit("● Grep " + truncForChat(jsonString(input, "pattern"), constants.ChatGrepGlobMax))
	case "Glob":
		f.emit("● Glob " + truncForChat(jsonString(input, "pattern"), constants.ChatGrepGlobMax))
	default:
		f.emit("● " + name)
	}
}

func (f *jsonStreamFormatter) emitIfPath(prefix string, input json.RawMessage) {
	if p := jsonString(input, "file_path"); p != "" {
		f.emit(prefix + shortenPath(p))
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
	idx := strings.LastIndex(p, "/")
	if idx < 0 || idx == len(p)-1 {
		return p
	}
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

func jsonFormat(format string) bool {
	switch strings.ToLower(format) {
	case "json", "jsonl", "json-lines":
		return true
	}
	return false
}

var _ io.Writer = (*jsonStreamFormatter)(nil)
