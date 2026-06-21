// SPDX-License-Identifier: MIT

package repl

import "io"

const (
	bracketedPasteStart      = "\x1b[200~"
	bracketedPasteEnd        = "\x1b[201~"
	bracketedPasteEnableSeq  = "\x1b[?2004h"
	bracketedPasteDisableSeq = "\x1b[?2004l"
)

type pasteCoalescingReader struct {
	src      io.Reader
	inPaste  bool
	buf      []byte
	carry    []byte
	startSeq []byte
	endSeq   []byte
}

func newPasteCoalescingReader(src io.Reader) *pasteCoalescingReader {
	return &pasteCoalescingReader{
		src:      src,
		startSeq: []byte(bracketedPasteStart),
		endSeq:   []byte(bracketedPasteEnd),
	}
}

func (p *pasteCoalescingReader) Close() error {
	if c, ok := p.src.(io.Closer); ok {
		return c.Close()
	}
	return nil
}

func (p *pasteCoalescingReader) Read(out []byte) (int, error) {
	if len(p.buf) > 0 {
		n := copy(out, p.buf)
		p.buf = p.buf[n:]
		return n, nil
	}
	tmp := make([]byte, len(out))
	n, err := p.src.Read(tmp)
	if n == 0 {
		return 0, err
	}
	data := append(p.carry, tmp[:n]...)
	p.carry = nil
	processed := p.process(data)
	if len(processed) == 0 && err == nil {
		return p.Read(out)
	}
	written := copy(out, processed)
	if written < len(processed) {
		p.buf = append(p.buf, processed[written:]...)
	}
	return written, err
}

func (p *pasteCoalescingReader) process(data []byte) []byte {
	out := make([]byte, 0, len(data))
	i := 0
	for i < len(data) {
		if !p.inPaste {
			if idx := matchPrefix(data[i:], p.startSeq); idx >= 0 {
				p.inPaste = true
				i += idx
				continue
			}
			if isPartialPrefix(data[i:], p.startSeq) {
				p.carry = append(p.carry, data[i:]...)
				return out
			}
			out = append(out, data[i])
			i++
			continue
		}
		if idx := matchPrefix(data[i:], p.endSeq); idx >= 0 {
			p.inPaste = false
			i += idx
			continue
		}
		if isPartialPrefix(data[i:], p.endSeq) {
			p.carry = append(p.carry, data[i:]...)
			return out
		}
		switch data[i] {
		case '\r', '\n':
			out = append(out, '\\', data[i])
		default:
			out = append(out, data[i])
		}
		i++
	}
	return out
}

func matchPrefix(haystack, needle []byte) int {
	if len(haystack) < len(needle) {
		return -1
	}
	for i, b := range needle {
		if haystack[i] != b {
			return -1
		}
	}
	return len(needle)
}

func isPartialPrefix(tail, needle []byte) bool {
	if len(tail) == 0 || len(tail) >= len(needle) {
		return false
	}
	for i, b := range tail {
		if needle[i] != b {
			return false
		}
	}
	return true
}
