package workflow

import (
	"bytes"
	"strings"
	"testing"
)

func TestLinePrefixWriterAddsPrefixPerLine(t *testing.T) {
	var buf bytes.Buffer
	w := newAgentLivePrefix(&buf)
	if _, err := w.Write([]byte("first\nsecond\nthird")); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	for _, want := range []string{"first", "second"} {
		if !strings.Contains(got, "│ "+want) {
			t.Errorf("missing prefixed %q: %q", want, got)
		}
	}
	if strings.Contains(got, "│ third") {
		t.Errorf("third has no newline, should be buffered: %q", got)
	}
}

func TestLinePrefixWriterStreamsAcrossWrites(t *testing.T) {
	var buf bytes.Buffer
	w := newAgentLivePrefix(&buf)
	_, _ = w.Write([]byte("alpha "))
	_, _ = w.Write([]byte("beta\n"))
	if !strings.Contains(buf.String(), "│ alpha beta") {
		t.Errorf("multi-write join failed: %q", buf.String())
	}
}
