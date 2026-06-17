package planscope

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writePlan(t *testing.T, root, id, body string) {
	t.Helper()
	d := filepath.Join(root, ".harness", "artifacts", "plans")
	_ = os.MkdirAll(d, 0o755)
	_ = os.WriteFile(filepath.Join(d, "PLAN-"+id+".md"), []byte(body), 0o644)
}

const plan = `# PLAN-x

## Intent

add /healthz

## Files in scope

- ` + "`app/main.py`" + `

## Invariants

- ok

## Validation

` + "```sh\nharness ci\n```" + `

## Risk tier

low
`

func TestCheckErrorsWithoutPlanID(t *testing.T) {
	if _, err := Check(context.Background(), Options{Root: "x"}); err == nil {
		t.Fatal("want error")
	}
}

func TestCheckErrorsWithoutRoot(t *testing.T) {
	if _, err := Check(context.Background(), Options{PlanID: "x"}); err == nil {
		t.Fatal("want error")
	}
}

func TestCheckAllowsInScopeFiles(t *testing.T) {
	dir := t.TempDir()
	writePlan(t, dir, "x", plan)
	r, err := Check(context.Background(), Options{
		Root: dir, PlanID: "x",
		DiffSource: func(ctx context.Context, root string) ([]string, error) {
			return []string{"app/main.py"}, nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !r.Pass() {
		t.Errorf("want pass, got %+v", r)
	}
}

func TestCheckFlagsOutOfScope(t *testing.T) {
	dir := t.TempDir()
	writePlan(t, dir, "x", plan)
	r, err := Check(context.Background(), Options{
		Root: dir, PlanID: "x",
		DiffSource: func(ctx context.Context, root string) ([]string, error) {
			return []string{"app/main.py", "config/secrets.py"}, nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if r.Pass() {
		t.Errorf("want fail")
	}
	if len(r.Violations) != 1 || r.Violations[0].Path != "config/secrets.py" {
		t.Errorf("violations: %+v", r.Violations)
	}
}

func TestCheckPropagatesDiffSourceError(t *testing.T) {
	dir := t.TempDir()
	writePlan(t, dir, "x", plan)
	_, err := Check(context.Background(), Options{
		Root: dir, PlanID: "x",
		DiffSource: func(ctx context.Context, root string) ([]string, error) {
			return nil, errors.New("git boom")
		},
	})
	if err == nil {
		t.Fatal("want error")
	}
}

func TestCheckPropagatesPlanLoadError(t *testing.T) {
	_, err := Check(context.Background(), Options{Root: t.TempDir(), PlanID: "ghost"})
	if err == nil {
		t.Fatal("want error")
	}
}

func TestFormatResultPass(t *testing.T) {
	got := FormatResult(Result{PlanID: "abc"})
	if !strings.Contains(got, "all changed files in scope") {
		t.Errorf("got %q", got)
	}
}

func TestFormatResultFail(t *testing.T) {
	got := FormatResult(Result{PlanID: "abc", Violations: []Violation{{Path: "x.py", Reason: "out"}}})
	if !strings.Contains(got, "✗ x.py") {
		t.Errorf("got %q", got)
	}
}
