package sensors

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/ropeixoto/harnessx/internal/index"
)

func TestCoverageSensorAppliesOnlyToGo(t *testing.T) {
	s := goCoverageSensorDefault()
	if !s.AppliesTo(index.Profile{Stacks: []index.Stack{{Name: "go"}}}) {
		t.Error("must apply to go stack")
	}
	if s.AppliesTo(index.Profile{Stacks: []index.Stack{{Name: "python"}}}) {
		t.Error("must not apply to python")
	}
}

func TestCoverageSensorAppliesUniversallyWhenStacksEmpty(t *testing.T) {
	s := CoverageSensor{IDValue: "x"}
	if !s.AppliesTo(index.Profile{}) {
		t.Error("empty stacks should apply universally")
	}
}

func TestCoverageSensorIDCategoryKind(t *testing.T) {
	s := goCoverageSensorDefault()
	if s.ID() != "go_coverage_gate" {
		t.Errorf("id: %q", s.ID())
	}
	if s.Category() != CatTest {
		t.Errorf("category: %v", s.Category())
	}
	if s.Kind() != KindComputational {
		t.Errorf("kind: %v", s.Kind())
	}
}

func TestCoverageSensorPassesAtThreshold(t *testing.T) {
	out := `ok  	github.com/x/a	2.0s	coverage: 95.0% of statements
ok  	github.com/x/b	2.0s	coverage: 92.0% of statements
`
	s := CoverageSensor{
		IDValue: "x", Threshold: 0.9, Stacks: nil,
		Runner: func(ctx context.Context, root, pkg string) ([]byte, error) {
			return []byte(out), nil
		},
	}
	r := s.Run(RunCtx{Ctx: context.Background()})
	if r.Status != StatusPassed {
		t.Errorf("status: %v detail=%s", r.Status, r.Detail)
	}
	if r.Confidence < 0.9 {
		t.Errorf("confidence: %v", r.Confidence)
	}
}

func TestCoverageSensorFailsBelowThreshold(t *testing.T) {
	out := `ok  	github.com/x/a	2.0s	coverage: 70.0% of statements
`
	s := CoverageSensor{
		IDValue: "x", Threshold: 0.9,
		Runner: func(ctx context.Context, root, pkg string) ([]byte, error) {
			return []byte(out), nil
		},
	}
	r := s.Run(RunCtx{Ctx: context.Background()})
	if r.Status != StatusFailed {
		t.Errorf("expected fail, got %v", r.Status)
	}
}

func TestCoverageSensorPropagatesRunnerError(t *testing.T) {
	s := CoverageSensor{
		IDValue: "x", Threshold: 0.9,
		Runner: func(ctx context.Context, root, pkg string) ([]byte, error) {
			return []byte("boom"), errors.New("compile error")
		},
	}
	r := s.Run(RunCtx{Ctx: context.Background()})
	if r.Status != StatusFailed {
		t.Errorf("status: %v", r.Status)
	}
	if !strings.Contains(r.Detail, "compile error") {
		t.Errorf("detail missing error: %s", r.Detail)
	}
}

func TestTruncateBoundaries(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		in   string
		n    int
		want string
	}{
		{"empty", "", 4, ""},
		{"underLimit", "abc", 4, "abc"},
		{"atLimit", "abcd", 4, "abcd"},
		{"overLimit", "abcdef", 4, "abcd...[truncated]"},
	}
	for _, c := range cases {
		if got := truncate(c.in, c.n); got != c.want {
			t.Errorf("%s: got %q want %q", c.name, got, c.want)
		}
	}
}

func TestDefaultCoverageRunnerSuccessAndFailure(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell script mock not portable on windows")
	}
	shim := t.TempDir()
	okScript := "#!/bin/sh\necho 'ok  \tgithub.com/x/a\t0.1s\tcoverage: 91.0% of statements'\n"
	if err := os.WriteFile(filepath.Join(shim, "go"), []byte(okScript), 0o755); err != nil {
		t.Fatal(err)
	}
	oldPath := os.Getenv("PATH")
	t.Setenv("PATH", shim+string(os.PathListSeparator)+oldPath)
	out, err := defaultCoverageRunner(context.Background(), shim, "./...")
	if err != nil {
		t.Fatalf("expected success, got %v: %s", err, out)
	}
	if !strings.Contains(string(out), "coverage: 91.0%") {
		t.Errorf("stdout missing coverage line: %s", out)
	}

	failShim := t.TempDir()
	failScript := "#!/bin/sh\necho 'compile broken' >&2\nexit 2\n"
	if err := os.WriteFile(filepath.Join(failShim, "go"), []byte(failScript), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", failShim+string(os.PathListSeparator)+oldPath)
	out2, err2 := defaultCoverageRunner(context.Background(), failShim, "./...")
	if err2 == nil {
		t.Fatal("expected failure")
	}
	if !strings.Contains(string(out2), "compile broken") {
		t.Errorf("expected stderr captured, got %q", out2)
	}
}

func TestCatalogIncludesCoverageForGo(t *testing.T) {
	got := Catalog(index.Profile{Stacks: []index.Stack{{Name: "go"}}})
	found := false
	for _, s := range got {
		if s.ID() == "go_coverage_gate" {
			found = true
		}
	}
	if !found {
		t.Errorf("catalog missing go_coverage_gate for go stack")
	}
}

func TestCatalogOmitsCoverageWithoutGo(t *testing.T) {
	got := Catalog(index.Profile{Stacks: []index.Stack{{Name: "python"}}})
	for _, s := range got {
		if s.ID() == "go_coverage_gate" {
			t.Errorf("go_coverage_gate should not appear for python-only project")
		}
	}
}
