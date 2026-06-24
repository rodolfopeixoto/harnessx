// SPDX-License-Identifier: MIT

package repl

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/chzyer/readline"
)

const (
	slashPopupMaxRows  = 6
	slashPopupColWidth = 20
)

type slashSuggester struct {
	out       io.Writer
	allSlash  []string
	lastShown int
	lastLine  string
}

func newSlashSuggester(out io.Writer) *slashSuggester {
	cmds := make([]string, len(staticSlashCommands))
	copy(cmds, staticSlashCommands)
	sort.Strings(cmds)
	return &slashSuggester{out: out, allSlash: cmds}
}

func (s *slashSuggester) Listener() readline.Listener {
	return readline.FuncListener(func(line []rune, pos int, key rune) ([]rune, int, bool) {
		s.render(string(line))
		return nil, 0, false
	})
}

// render keeps the popup correctly drawn without ever emitting a raw
// '\n'. Earlier versions used \n + \x1b[s/\x1b[u to draw rows below
// the prompt; when the cursor sat near the bottom of the terminal the
// newlines forced the buffer to scroll, breaking the saved cursor and
// leaving ghost "← TAB to complete" rows on every keystroke. We now
// use \x1b[B (cursor down — never scrolls) + \x1b[2K (erase entire
// line) + \x1b[1G (column 1), and only redraw when the matches set
// actually changed.
func (s *slashSuggester) render(line string) {
	if s.out == nil {
		return
	}
	matches := s.matches(line)
	key := joinKey(matches)
	if key == s.lastLine {
		return
	}
	s.eraseLast()
	s.lastLine = key
	if len(matches) == 0 {
		return
	}
	s.lastShown = len(matches)
	fmt.Fprint(s.out, "\x1b[s")
	for i, m := range matches {
		fmt.Fprintf(s.out, "\x1b[1B\x1b[2K\x1b[1G  %s", padRight(m, slashPopupColWidth))
		if i == 0 {
			fmt.Fprint(s.out, "  ← TAB to complete")
		}
	}
	fmt.Fprint(s.out, "\x1b[u")
}

func (s *slashSuggester) eraseLast() {
	if s.lastShown == 0 {
		return
	}
	fmt.Fprint(s.out, "\x1b[s")
	for i := 0; i < s.lastShown; i++ {
		fmt.Fprint(s.out, "\x1b[1B\x1b[2K\x1b[1G")
	}
	fmt.Fprint(s.out, "\x1b[u")
	s.lastShown = 0
}

func joinKey(matches []string) string {
	return strings.Join(matches, "|")
}

func (s *slashSuggester) matches(line string) []string {
	trimmed := strings.TrimLeft(line, " \t")
	if !strings.HasPrefix(trimmed, "/") {
		return nil
	}
	prefix := trimmed
	if idx := strings.IndexAny(trimmed, " \t"); idx > 0 {
		prefix = trimmed[:idx]
	}
	out := make([]string, 0, slashPopupMaxRows)
	for _, c := range s.allSlash {
		if c == prefix && strings.IndexAny(trimmed, " \t") < 0 {
			return nil
		}
		if strings.HasPrefix(c, prefix) {
			out = append(out, c)
			if len(out) == slashPopupMaxRows {
				break
			}
		}
	}
	return out
}
