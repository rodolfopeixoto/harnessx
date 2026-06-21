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

func (s *slashSuggester) render(line string) {
	if s.out == nil {
		return
	}
	matches := s.matches(line)
	s.clear()
	if len(matches) == 0 {
		return
	}
	s.lastShown = len(matches)
	fmt.Fprint(s.out, "\x1b[s")
	for i, m := range matches {
		fmt.Fprintf(s.out, "\n\x1b[2K  %s", padRight(m, slashPopupColWidth))
		if i == 0 {
			fmt.Fprint(s.out, "  ← TAB to complete")
		}
	}
	fmt.Fprint(s.out, "\x1b[u")
}

func (s *slashSuggester) clear() {
	if s.lastShown == 0 {
		return
	}
	fmt.Fprint(s.out, "\x1b[s")
	for i := 0; i < s.lastShown; i++ {
		fmt.Fprint(s.out, "\n\x1b[2K")
	}
	fmt.Fprint(s.out, "\x1b[u")
	s.lastShown = 0
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
