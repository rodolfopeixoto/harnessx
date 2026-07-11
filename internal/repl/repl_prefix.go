// SPDX-License-Identifier: MIT

package repl

import "io"

// prefixWriter wraps an io.Writer to prepend a fixed prefix to every
// line the adapter streams. The buffer stays intact across partial
// writes so a "  │ " marker never lands mid-token.
type prefixWriter struct {
	w            io.Writer
	prefix       string
	buf          []byte
	atBOL        bool
	onFirstWrite func()
	firstWritten bool
}

func (p *prefixWriter) Write(b []byte) (int, error) {
	if !p.firstWritten && len(b) > 0 {
		p.firstWritten = true
		if p.onFirstWrite != nil {
			p.onFirstWrite()
		}
	}
	if p.buf == nil {
		p.atBOL = true
	}
	for _, c := range b {
		if p.atBOL {
			p.buf = append(p.buf, []byte(p.prefix)...)
			p.atBOL = false
		}
		p.buf = append(p.buf, c)
		if c == '\n' {
			p.atBOL = true
		}
	}
	if p.atBOL {
		_, err := p.w.Write(p.buf)
		p.buf = p.buf[:0]
		if err != nil {
			return len(b), err
		}
	}
	return len(b), nil
}

func (p *prefixWriter) flush() {
	if len(p.buf) > 0 {
		_, _ = p.w.Write(p.buf)
		_, _ = p.w.Write([]byte{'\n'})
		p.buf = p.buf[:0]
	}
}
