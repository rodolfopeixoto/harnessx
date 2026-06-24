// SPDX-License-Identifier: MIT

package yaml

import (
	"bytes"
	"io"
	"regexp"
)

// stderrNoisePatterns are upstream-CLI stderr lines that are not
// actionable for the user and just clutter the live REPL output.
// Add patterns conservatively — anything matching here is silently
// dropped from the live writer (it still lands in the captured
// stderr buffer for debugging via `harness logs`).
var stderrNoisePatterns = []*regexp.Regexp{
	// codex: skill loader complains about each long SKILL.md file
	regexp.MustCompile(`failed to load skill .*SKILL\.md`),
	regexp.MustCompile(`invalid description: exceeds maximum length`),
	// codex_core::session noise that fires for every session start
	regexp.MustCompile(`ERROR codex_core::session: failed to load skill`),
	// claude: bootstrap noise about MCP server discovery
	regexp.MustCompile(`MCP server [^ ]+ exited with`),
}

// filteringWriter wraps a writer and drops complete lines that match
// any of the configured noise patterns. Partial lines are buffered
// until a newline arrives so a noisy line is never half-printed.
type filteringWriter struct {
	out      io.Writer
	patterns []*regexp.Regexp
	buf      []byte
}

func newFilteringWriter(out io.Writer) *filteringWriter {
	return &filteringWriter{out: out, patterns: stderrNoisePatterns}
}

func (f *filteringWriter) Write(p []byte) (int, error) {
	n := len(p)
	f.buf = append(f.buf, p...)
	for {
		idx := bytes.IndexByte(f.buf, '\n')
		if idx < 0 {
			return n, nil
		}
		line := f.buf[:idx+1]
		f.buf = f.buf[idx+1:]
		if !f.matchesNoise(line) {
			if _, err := f.out.Write(line); err != nil {
				return n, err
			}
		}
	}
}

func (f *filteringWriter) matchesNoise(line []byte) bool {
	for _, re := range f.patterns {
		if re.Match(line) {
			return true
		}
	}
	return false
}
