package sensors

import (
	"context"
	"errors"
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
