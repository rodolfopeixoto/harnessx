// SPDX-License-Identifier: MIT

// fake-agent is a deterministic stand-in for Claude/Codex/Gemini used by
// HarnessX execution tests and the e2e smoke. It reads the prompt from
// stdin, parses a single instruction of the form
//
//	create <path> with content: <content>
//
// (case-insensitive, repeatable on separate lines), writes those files
// inside the current working directory, then emits a JSON object on
// stdout matching the claude/codex output contract:
//
//	{"result": "wrote N files", "usage": {"input_tokens": 12, "output_tokens": 8}}
//
// When the prompt contains "fail" the agent exits non-zero so the
// executor's agent_failed path is exercised.
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var createPattern = regexp.MustCompile(`(?i)^create\s+(\S+)\s+with\s+content:\s*(.*)$`)

func main() {
	data, err := io.ReadAll(bufio.NewReader(os.Stdin))
	if err != nil {
		fmt.Fprintf(os.Stderr, "fake-agent: read stdin: %v\n", err)
		os.Exit(2)
	}
	prompt := string(data)
	if strings.Contains(strings.ToLower(prompt), "fail") {
		fmt.Fprintln(os.Stderr, "fake-agent: failure requested")
		os.Exit(1)
	}
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "fake-agent: cwd: %v\n", err)
		os.Exit(2)
	}
	count := 0
	for _, line := range strings.Split(prompt, "\n") {
		m := createPattern.FindStringSubmatch(strings.TrimSpace(line))
		if m == nil {
			continue
		}
		target := filepath.Join(cwd, m[1])
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			fmt.Fprintf(os.Stderr, "fake-agent: mkdir: %v\n", err)
			os.Exit(2)
		}
		if err := os.WriteFile(target, []byte(m[2]+"\n"), 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "fake-agent: write: %v\n", err)
			os.Exit(2)
		}
		count++
	}
	out := map[string]any{
		"result": fmt.Sprintf("wrote %d files", count),
		"usage":  map[string]int{"input_tokens": 12, "output_tokens": 8},
	}
	enc := json.NewEncoder(os.Stdout)
	if err := enc.Encode(out); err != nil {
		fmt.Fprintf(os.Stderr, "fake-agent: encode: %v\n", err)
		os.Exit(2)
	}
}
