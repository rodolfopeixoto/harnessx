// SPDX-License-Identifier: MIT

package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ropeixoto/harnessx/internal/agents/fake"
	"github.com/ropeixoto/harnessx/internal/specflow"
)

func fakeAdapterReturning(body string) *fake.Adapter {
	a := fake.New("fake")
	a.FinalMessage = body
	return a
}

func TestNewSpecAuthorCmdExposesFlags(t *testing.T) {
	c := newSpecAuthorCmd()
	for _, f := range []string{"adapter", "no-interactive", "skip-questions", "accept-draft"} {
		if c.Flags().Lookup(f) == nil {
			t.Errorf("flag %q missing", f)
		}
	}
}

func TestSpecAuthorCmdNonInteractiveWritesBaseline(t *testing.T) {
	root := t.TempDir()
	cwd, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(cwd) })
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	c := newSpecAuthorCmd()
	var out bytes.Buffer
	c.SetOut(&out)
	c.SetErr(&out)
	c.SetIn(strings.NewReader(""))
	c.SetArgs([]string{"--no-interactive", "add /healthz endpoint"})
	if err := c.ExecuteContext(context.Background()); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "spec saved →") {
		t.Errorf("missing save line: %s", out.String())
	}
	specs, _ := filepath.Glob(filepath.Join(root, ".harness", "artifacts", "specs", "*.md"))
	if len(specs) != 1 {
		t.Fatalf("want 1 spec file, got %d", len(specs))
	}
	body, _ := os.ReadFile(specs[0])
	for _, w := range []string{"harness-spec-id:", "## Summary", "add /healthz endpoint"} {
		if !strings.Contains(string(body), w) {
			t.Errorf("missing %q in saved spec\n%s", w, body)
		}
	}
}

func TestSpecRefineParsesSectionPrefix(t *testing.T) {
	sess := specflow.New(t.TempDir(), "x")
	sess.Body = "## Summary\nold\n"
	sess.Revisions = []specflow.Revision{{Source: "draft", Body: sess.Body}}
	a := fakeAdapterReturning("## Summary\nnew\n")
	var buf bytes.Buffer
	if err := specRefine(context.Background(), &buf, sess, a, "Summary: make it new"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(sess.Body, "new") {
		t.Errorf("body not updated: %s", sess.Body)
	}
	if sess.Revisions[1].Section != "Summary" {
		t.Errorf("section not parsed: %+v", sess.Revisions[1])
	}
}

func TestSpecRefineEmptyInstructionRejected(t *testing.T) {
	sess := specflow.New(t.TempDir(), "x")
	sess.Body = "## Summary\nx"
	sess.Revisions = []specflow.Revision{{Source: "draft", Body: sess.Body}}
	a := fakeAdapterReturning("## Summary\nx")
	if err := specRefine(context.Background(), &bytes.Buffer{}, sess, a, "Summary:   "); err == nil {
		t.Error("expected empty-instruction error")
	}
}

func TestSpecAuthorHelpListsAllCommands(t *testing.T) {
	var buf bytes.Buffer
	specAuthorHelp(&buf)
	for _, w := range []string{"/show", "/edit", "/refine", "/expand", "/shrink", "/diff", "/undo", "/save", "/cancel"} {
		if !strings.Contains(buf.String(), w) {
			t.Errorf("help missing %q\n%s", w, buf.String())
		}
	}
}
