// SPDX-License-Identifier: MIT

package repl

import (
	"bufio"
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/chzyer/readline"
)

func TestBufioPromptReaderReadsSingleLine(t *testing.T) {
	out := &bytes.Buffer{}
	r := &bufioPromptReader{r: bufioReader("hello\n"), w: out}
	got, err := r.ReadInput(">> ", "...")
	if err != nil {
		t.Fatal(err)
	}
	if got != "hello" {
		t.Errorf("want hello, got %q", got)
	}
	if !strings.HasPrefix(out.String(), ">> ") {
		t.Errorf("prompt missing: %q", out.String())
	}
}

func TestBufioPromptReaderHandlesBackslashContinuation(t *testing.T) {
	out := &bytes.Buffer{}
	r := &bufioPromptReader{r: bufioReader("hello \\\nworld\n"), w: out}
	got, err := r.ReadInput(">> ", "→ ")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "hello") || !strings.Contains(got, "world") {
		t.Errorf("backslash join failed: %q", got)
	}
	if !strings.Contains(out.String(), "→ ") {
		t.Errorf("continuation marker missing: %q", out.String())
	}
}

func TestBufioPromptReaderHandlesTripleQuoteHeredoc(t *testing.T) {
	out := &bytes.Buffer{}
	r := &bufioPromptReader{r: bufioReader("\"\"\"\nline one\nline two\nthird\n\"\"\"\n"), w: out}
	got, err := r.ReadInput(">> ", "→ ")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"line one", "line two", "third"} {
		if !strings.Contains(got, want) {
			t.Errorf("heredoc missing %q in %q", want, got)
		}
	}
	if strings.Contains(got, hereDocMarker) {
		t.Errorf("heredoc marker leaked into payload: %q", got)
	}
}

func TestBufioPromptReaderEOF(t *testing.T) {
	r := &bufioPromptReader{r: bufioReader(""), w: io.Discard}
	if _, err := r.ReadInput(">> ", "..."); err == nil {
		t.Fatal("want EOF error")
	}
}

func TestBufioPromptReaderClose(t *testing.T) {
	r := &bufioPromptReader{r: bufioReader(""), w: io.Discard}
	if err := r.Close(); err != nil {
		t.Errorf("close: %v", err)
	}
}

func TestNewPromptReaderFallsBackToBufioForNonTTY(t *testing.T) {
	p := newPromptReader(strings.NewReader("x\n"), io.Discard, "", nil)
	if _, ok := p.(*bufioPromptReader); !ok {
		t.Errorf("expected bufioPromptReader, got %T", p)
	}
	_ = p.Close()
}

func TestIsTerminalFalseForStrings(t *testing.T) {
	if isTerminal(strings.NewReader("x")) {
		t.Error("strings.Reader is not a terminal")
	}
}

func TestChatCompleterIncludesStaticAndDynamic(t *testing.T) {
	c := chatCompleter([]string{"claude", "codex"}, []string{"my-feature"})
	tree, ok := c.(*readline.PrefixCompleter)
	if !ok || tree == nil {
		t.Fatal("expected PrefixCompleter")
	}
	if tree.Tree("") == "" {
		t.Fatal("completer tree empty")
	}
}

func TestToPcItemsKeepsOrder(t *testing.T) {
	got := toPcItems([]string{"a", "b", "c"})
	if len(got) != 3 {
		t.Fatalf("want 3, got %d", len(got))
	}
}

func TestAsReadCloserWrapsPlainReader(t *testing.T) {
	rc := asReadCloser(strings.NewReader("x"))
	if err := rc.Close(); err != nil {
		t.Errorf("close: %v", err)
	}
}

func bufioReader(s string) *bufio.Reader {
	return bufio.NewReader(strings.NewReader(s))
}
