// SPDX-License-Identifier: MIT

package devloop

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCheckRegression(t *testing.T) {
	cases := []struct {
		baseLint, attLint, baseTest, attTest bool
		wantRegressed                        bool
		wantReason                           string
	}{
		{true, true, true, true, false, ""},
		{true, false, true, true, true, "lint regressed"},
		{true, true, true, false, true, "tests regressed"},
		{true, false, true, false, true, "lint regressed"},
		{false, false, true, true, false, ""},
	}
	for _, c := range cases {
		got, reason := checkRegression(c.baseLint, c.attLint, c.baseTest, c.attTest)
		if got != c.wantRegressed {
			t.Errorf("regressed: case=%+v want=%v got=%v", c, c.wantRegressed, got)
		}
		if c.wantReason != "" && !strings.Contains(reason, c.wantReason) {
			t.Errorf("reason: case=%+v want substring %q got %q", c, c.wantReason, reason)
		}
	}
}

func TestCanonicaliseIncludesRegressionBlockWhenSet(t *testing.T) {
	a := Attempt{Number: 2, Elapsed: 100 * time.Millisecond, Regressed: true, Regression: "lint regressed"}
	out := Canonicalise("add healthz", a)
	if !strings.Contains(out, "Regression detected") {
		t.Error("missing Regression detected block")
	}
	if !strings.Contains(out, "lint regressed") {
		t.Error("missing regression reason")
	}
	if !strings.Contains(out, "add healthz") {
		t.Error("missing original prompt")
	}
}

func TestCanonicaliseIncludesLintAndTestOutputs(t *testing.T) {
	a := Attempt{Number: 1, LintOK: false, LintOutput: "ruff: E501", TestOK: false, TestOutput: "FAILED tests/test_x.py::test_y"}
	out := Canonicalise("ping", a)
	if !strings.Contains(out, "Lint failure") || !strings.Contains(out, "ruff: E501") {
		t.Error("missing lint block")
	}
	if !strings.Contains(out, "Test failure") || !strings.Contains(out, "FAILED tests") {
		t.Error("missing test block")
	}
}

func TestTrimToLinesKeepsLastN(t *testing.T) {
	lines := strings.Repeat("line\n", 200)
	out := trimToLines(lines, 50)
	got := strings.Count(out, "line")
	if got != 49 && got != 50 {
		t.Errorf("trimToLines kept %d, want 49 or 50", got)
	}
}

func TestTrimToLinesShorterPasses(t *testing.T) {
	in := "a\nb\nc"
	if trimToLines(in, 80) != in {
		t.Error("short input should pass through")
	}
}

func TestOkOrFail(t *testing.T) {
	if okOrFail(true) != "✓" || okOrFail(false) != "✗" {
		t.Error("okOrFail mapping wrong")
	}
}

func TestRunShellSuccessAndFailure(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	ok, _ := runShell(ctx, dir, "true")
	if !ok {
		t.Error("'true' should succeed")
	}
	bad, out := runShell(ctx, dir, "false")
	if bad {
		t.Error("'false' should fail")
	}
	_ = out
}

func TestRunShellEmptyCmd(t *testing.T) {
	ok, out := runShell(context.Background(), "/tmp", "")
	if !ok || out != "" {
		t.Errorf("empty cmd should noop: ok=%v out=%q", ok, out)
	}
}

func TestResolveCommandsMissingScaffoldErrors(t *testing.T) {
	opts := Options{StartDir: t.TempDir()}
	err := resolveCommands(opts.StartDir, &opts)
	if err == nil {
		t.Fatal("expected error when no scaffold/lint/test cmd")
	}
}

func TestResolveCommandsUsesOpts(t *testing.T) {
	opts := Options{LintCmd: "echo ok", TestCmd: "echo ok"}
	if err := resolveCommands("/tmp", &opts); err != nil {
		t.Fatal(err)
	}
}

func TestDetectScaffoldMetaPython(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "requirements.txt"), []byte(""), 0o644)
	m := detectScaffoldMeta(dir)
	if m == nil || m.Language != "python" {
		t.Errorf("want python meta, got %+v", m)
	}
}

func TestDetectScaffoldMetaNone(t *testing.T) {
	m := detectScaffoldMeta(t.TempDir())
	if m != nil {
		t.Errorf("want nil for empty dir, got %+v", m)
	}
}
