// SPDX-License-Identifier: MIT

package repl

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

func TestPasteCoalescingPassthroughOutsidePaste(t *testing.T) {
	r := newPasteCoalescingReader(strings.NewReader("hello\nworld\n"))
	got, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "hello\nworld\n" {
		t.Errorf("want unchanged, got %q", got)
	}
}

func TestPasteCoalescingEscapesNewlinesInsidePaste(t *testing.T) {
	in := bracketedPasteStart + "line one\nline two\nline three" + bracketedPasteEnd + "\n"
	r := newPasteCoalescingReader(strings.NewReader(in))
	got, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	want := `line one\` + "\n" + `line two\` + "\n" + "line three\n"
	if string(got) != want {
		t.Errorf("paste escape wrong\n got: %q\nwant: %q", got, want)
	}
}

func TestPasteCoalescingHandlesSplitMarkers(t *testing.T) {
	full := bracketedPasteStart + "abc\ndef" + bracketedPasteEnd
	r := newPasteCoalescingReader(&chunkReader{chunks: [][]byte{
		[]byte(full[:3]),
		[]byte(full[3:7]),
		[]byte(full[7:]),
	}})
	var buf bytes.Buffer
	tmp := make([]byte, 4)
	for {
		n, err := r.Read(tmp)
		if n > 0 {
			buf.Write(tmp[:n])
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
	}
	if !strings.Contains(buf.String(), `abc\`+"\n"+"def") {
		t.Errorf("paste across chunks lost newline escape: %q", buf.String())
	}
	if strings.Contains(buf.String(), bracketedPasteStart) {
		t.Errorf("start marker leaked: %q", buf.String())
	}
}

func TestIsPartialPrefix(t *testing.T) {
	if !isPartialPrefix([]byte("\x1b["), []byte(bracketedPasteStart)) {
		t.Error("ESC[ should be partial prefix of paste start")
	}
	if isPartialPrefix([]byte("foo"), []byte(bracketedPasteStart)) {
		t.Error("foo is not a partial prefix")
	}
	if isPartialPrefix([]byte(bracketedPasteStart), []byte(bracketedPasteStart)) {
		t.Error("full match should not count as partial prefix")
	}
}

func TestPasteCoalescingClose(t *testing.T) {
	r := newPasteCoalescingReader(io.NopCloser(strings.NewReader("")))
	if err := r.Close(); err != nil {
		t.Errorf("close: %v", err)
	}
}

type chunkReader struct {
	chunks [][]byte
	idx    int
}

func (c *chunkReader) Read(p []byte) (int, error) {
	if c.idx >= len(c.chunks) {
		return 0, io.EOF
	}
	n := copy(p, c.chunks[c.idx])
	c.chunks[c.idx] = c.chunks[c.idx][n:]
	if len(c.chunks[c.idx]) == 0 {
		c.idx++
	}
	return n, nil
}
