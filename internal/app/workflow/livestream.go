// SPDX-License-Identifier: MIT

package workflow

import (
	"bytes"
	"io"
)

type linePrefixWriter struct {
	w      io.Writer
	prefix []byte
	buf    bytes.Buffer
}

func newAgentLivePrefix(w io.Writer) io.Writer {
	return &linePrefixWriter{w: w, prefix: []byte("  │ ")}
}

func (l *linePrefixWriter) Write(p []byte) (int, error) {
	n := len(p)
	l.buf.Write(p)
	for {
		idx := bytes.IndexByte(l.buf.Bytes(), '\n')
		if idx < 0 {
			break
		}
		line := l.buf.Next(idx + 1)
		if _, err := l.w.Write(l.prefix); err != nil {
			return n, err
		}
		if _, err := l.w.Write(line); err != nil {
			return n, err
		}
	}
	return n, nil
}
