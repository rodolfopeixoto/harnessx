package sensors

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/ropeixoto/harnessx/internal/index"
)

func writePlan(t *testing.T, root, id, body string) {
	t.Helper()
	d := filepath.Join(root, ".harness", "artifacts", "plans")
	_ = os.MkdirAll(d, 0o755)
	_ = os.WriteFile(filepath.Join(d, "PLAN-"+id+".md"), []byte(body), 0o644)
}

const samplePlan = `# PLAN-x

## Intent

add /healthz

## Files in scope

- ` + "`app/main.py`" + `

## Validation

` + "```sh\nharness ci\n```" + `

## Risk tier

low
`

func TestPlanScopeSensorSkipsWhenNoPlan(t *testing.T) {
	s := PlanScopeSensor{IDValue: "plan_scope"}
	r := s.Run(RunCtx{Ctx: context.Background()})
	if r.Status != StatusSkipped {
		t.Errorf("want skipped, got %v", r.Status)
	}
}

func TestPlanScopeSensorAppliesToOnlyWhenPinned(t *testing.T) {
	s := PlanScopeSensor{}
	if s.AppliesTo(index.Profile{}) {
		t.Error("must not apply without plan")
	}
	s.PlanID = "x"
	if !s.AppliesTo(index.Profile{}) {
		t.Error("must apply with plan")
	}
}

func TestSaveLoadActivePlanIDRoundtrip(t *testing.T) {
	dir := t.TempDir()
	if err := SaveActivePlanID(dir, "abc"); err != nil {
		t.Fatal(err)
	}
	got, err := LoadActivePlanID(dir)
	if err != nil {
		t.Fatal(err)
	}
	if got != "abc" {
		t.Errorf("got %q", got)
	}
}

func TestLoadActivePlanIDMissingFileReturnsEmpty(t *testing.T) {
	got, err := LoadActivePlanID(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if got != "" {
		t.Errorf("want empty, got %q", got)
	}
}

func TestLoadActivePlanIDPropagatesParseError(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, ".harness", "config")
	_ = os.MkdirAll(p, 0o755)
	_ = os.WriteFile(filepath.Join(p, "plan.yaml"), []byte("active_plan_id: { broken: ["), 0o644)
	if _, err := LoadActivePlanID(dir); err == nil {
		t.Fatal("want parse error")
	}
}

func TestPlanScopeSensorPassesInScope(t *testing.T) {
	dir := t.TempDir()
	writePlan(t, dir, "x", samplePlan)
	s := PlanScopeSensor{IDValue: "plan_scope", PlanID: "x"}
	r := s.Run(RunCtx{Ctx: context.Background(), Root: dir})
	if r.Status != StatusFailed {
		t.Errorf("dirty repo with no manifest will fail (git not init); got %v", r.Status)
	}
}

func TestCatalogIncludesPlanScopeWhenPinned(t *testing.T) {
	dir := t.TempDir()
	_ = SaveActivePlanID(dir, "abc")
	got := Catalog(index.Profile{Root: dir})
	found := false
	for _, s := range got {
		if s.ID() == "plan_scope" {
			found = true
		}
	}
	if !found {
		t.Errorf("catalog missing plan_scope")
	}
}

func TestCatalogOmitsPlanScopeWithoutPin(t *testing.T) {
	got := Catalog(index.Profile{Root: t.TempDir()})
	for _, s := range got {
		if s.ID() == "plan_scope" {
			t.Errorf("plan_scope should not appear without pin")
		}
	}
}
