// SPDX-License-Identifier: MIT

package main

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/ropeixoto/harnessx/internal/ui"
)

func askYesNo(in io.Reader, out io.Writer, prompt string, defaultYes bool) bool {
	def := "Y/n"
	if !defaultYes {
		def = "y/N"
	}
	fmt.Fprintf(out, "  %s [%s]: ", prompt, def)
	answer := trimAndLower(readPromptLine(in))
	if answer == "" {
		return defaultYes
	}
	switch answer {
	case "y", "yes", "1", "true", "ok":
		return true
	case "n", "no", "0", "false":
		return false
	}
	return defaultYes
}

func askChoice(in io.Reader, out io.Writer, prompt string, options []string, defaultIdx int) (int, error) {
	if len(options) == 0 {
		return 0, fmt.Errorf("askChoice: no options")
	}
	fmt.Fprintln(out, prompt)
	for i, o := range options {
		marker := "  "
		if i == defaultIdx {
			marker = "→ "
		}
		fmt.Fprintf(out, "  %s%d) %s\n", marker, i+1, o)
	}
	fmt.Fprintf(out, "  pick [1-%d, default %d]: ", len(options), defaultIdx+1)
	answer := trimAndLower(readPromptLine(in))
	if answer == "" {
		return defaultIdx, nil
	}
	n, err := strconv.Atoi(answer)
	if err != nil || n < 1 || n > len(options) {
		fmt.Fprintf(out, "  %s invalid choice %q — using default %d\n", ui.MarkWarn(), answer, defaultIdx+1)
		return defaultIdx, nil
	}
	return n - 1, nil
}

func askLine(in io.Reader, out io.Writer, prompt, defaultVal string) (string, bool) {
	fmt.Fprintf(out, "  %s [%s]: ", prompt, defaultVal)
	answer := strings.TrimSpace(readPromptLine(in))
	if answer == "" {
		return defaultVal, true
	}
	return answer, true
}

func readPromptLine(in io.Reader) string {
	if br, ok := in.(*bufio.Reader); ok {
		line, _ := br.ReadString('\n')
		return strings.TrimRight(line, "\r\n")
	}
	buf := make([]byte, 256)
	n, _ := in.Read(buf)
	answer := ""
	for i := 0; i < n; i++ {
		if buf[i] == '\n' || buf[i] == '\r' {
			break
		}
		answer += string(buf[i])
	}
	return answer
}

func trimAndLower(s string) string {
	out := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == ' ' || c == '\t' {
			continue
		}
		if c >= 'A' && c <= 'Z' {
			c += 32
		}
		out = append(out, c)
	}
	return string(out)
}
