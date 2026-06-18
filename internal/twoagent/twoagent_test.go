package twoagent

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type stubDiag struct {
	problems []Problem
	err      error
}

func (s stubDiag) Diagnose(ctx context.Context, root string) ([]Problem, error) {
	return s.problems, s.err
}

func TestDiagnoseAllAggregatesAndSortsBySeverity(t *testing.T) {
	d, err := DiagnoseAll(context.Background(), t.TempDir(), []Diagnoser{
		stubDiag{problems: []Problem{{ID: "a", Severity: "info"}}},
		stubDiag{problems: []Problem{{ID: "b", Severity: "error"}}},
		stubDiag{problems: []Problem{{ID: "c", Severity: "warn"}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"b", "c", "a"}
	for i, p := range d.Problems {
		if p.ID != want[i] {
			t.Errorf("problem[%d] want %q got %q", i, want[i], p.ID)
		}
	}
}

func TestDiagnoseAllPropagatesError(t *testing.T) {
	_, err := DiagnoseAll(context.Background(), t.TempDir(), []Diagnoser{stubDiag{err: errors.New("boom")}})
	if err == nil {
		t.Fatal("want error")
	}
}

func TestSaveLoadDiagnosisRoundtrip(t *testing.T) {
	dir := t.TempDir()
	d := Diagnosis{Problems: []Problem{{ID: "x", Severity: "warn", Subject: "subject"}}}
	path, err := SaveDiagnosis(dir, d)
	if err != nil {
		t.Fatal(err)
	}
	got, err := LoadDiagnosis(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Problems) != 1 || got.Problems[0].ID != "x" {
		t.Errorf("roundtrip mismatch: %+v", got)
	}
}

func TestLoadDiagnosisBadJSON(t *testing.T) {
	p := filepath.Join(t.TempDir(), "x.json")
	_ = os.WriteFile(p, []byte("not json"), 0o644)
	if _, err := LoadDiagnosis(p); err == nil {
		t.Fatal("want error")
	}
}

type stubFixer struct {
	id  string
	err error
}

func (s stubFixer) ID() string          { return s.id }
func (s stubFixer) Description() string { return s.id }
func (s stubFixer) Apply(ctx context.Context, root string, p Problem, out io.Writer) error {
	return s.err
}

func TestApplyAllRoutesByFixID(t *testing.T) {
	d := Diagnosis{Problems: []Problem{
		{ID: "p1", FixID: "f1"},
		{ID: "p2", FixID: "f-missing"},
		{ID: "p3"},
	}}
	fixers := map[string]Fixer{"f1": fakeFixer{}}
	var buf bytes.Buffer
	res := ApplyAll(context.Background(), t.TempDir(), d, fixers, &buf)
	if len(res) != 3 {
		t.Fatalf("want 3 results, got %d", len(res))
	}
	if !res[0].Applied {
		t.Errorf("p1 should be applied")
	}
	if res[1].Applied || res[1].Skipped == false {
		t.Errorf("p2 should be skipped (no fixer)")
	}
	if res[2].Applied || res[2].Skipped == false {
		t.Errorf("p3 should be skipped (no fix id)")
	}
}

type fakeFixer struct{}

func (f fakeFixer) ID() string          { return "f1" }
func (f fakeFixer) Description() string { return "" }
func (f fakeFixer) Apply(ctx context.Context, root string, p Problem, out io.Writer) error {
	return nil
}

func TestApplyAllUsesIOWriter(t *testing.T) {
	d := Diagnosis{Problems: []Problem{{ID: "p", FixID: "f1"}}}
	fixers := map[string]Fixer{"f1": fakeFixer{}}
	var buf bytes.Buffer
	_ = ApplyAll(context.Background(), t.TempDir(), d, fixers, &buf)
	if !strings.Contains(buf.String(), "applying fix") {
		t.Errorf("missing log line: %s", buf.String())
	}
}

func TestSeverityRankOrder(t *testing.T) {
	for _, c := range []struct {
		a, b string
	}{
		{"error", "warn"},
		{"warn", "info"},
		{"info", "other"},
	} {
		if severityRank(c.a) >= severityRank(c.b) {
			t.Errorf("%s should rank lower than %s", c.a, c.b)
		}
	}
}

func TestFormatDiagnosisEmpty(t *testing.T) {
	got := FormatDiagnosis(Diagnosis{})
	if !strings.Contains(got, "no problems") {
		t.Errorf("missing happy path: %q", got)
	}
}

func TestFormatDiagnosisWithProblems(t *testing.T) {
	got := FormatDiagnosis(Diagnosis{Problems: []Problem{
		{ID: "x", Severity: "error", Subject: "boom", Hints: []string{"do thing"}},
	}})
	for _, want := range []string{"✗", "[x]", "boom", "do thing"} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in %q", want, got)
		}
	}
}

func TestMissingToolDiagnoser(t *testing.T) {
	d := MissingToolDiagnoser{Tools: []string{"definitely-not-on-path-xyz"}}
	ps, err := d.Diagnose(context.Background(), t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if len(ps) != 1 || ps[0].FixID != "install-tool" {
		t.Errorf("unexpected: %+v", ps)
	}
}

func TestMissingPlanDiagnoser(t *testing.T) {
	d := MissingPlanDiagnoser{}
	ps, _ := d.Diagnose(context.Background(), t.TempDir())
	if len(ps) != 1 {
		t.Errorf("want 1 problem, got %d", len(ps))
	}
}

func TestMissingPlanSkippedWhenPinned(t *testing.T) {
	dir := t.TempDir()
	_ = os.MkdirAll(filepath.Join(dir, ".harness", "config"), 0o755)
	_ = os.WriteFile(filepath.Join(dir, ".harness", "config", "plan.yaml"), []byte("active_plan_id: x"), 0o644)
	ps, _ := MissingPlanDiagnoser{}.Diagnose(context.Background(), dir)
	if len(ps) != 0 {
		t.Errorf("want 0 problems, got %+v", ps)
	}
}

func TestDefaultDiagnosersReturnsThree(t *testing.T) {
	got := DefaultDiagnosers([]string{"x"})
	if len(got) != 3 {
		t.Errorf("want 3 default diagnosers, got %d", len(got))
	}
}

func TestDefaultFixersHasInstallAndCommit(t *testing.T) {
	got := DefaultFixers("/usr/local/bin/harness")
	for _, k := range []string{"install-tool", "commit-snapshot"} {
		if _, ok := got[k]; !ok {
			t.Errorf("missing fixer %q", k)
		}
	}
}

func TestSaveDiagnosisMarshalsJSON(t *testing.T) {
	dir := t.TempDir()
	d := Diagnosis{Problems: []Problem{{ID: "x"}}}
	path, err := SaveDiagnosis(dir, d)
	if err != nil {
		t.Fatal(err)
	}
	body, _ := os.ReadFile(path)
	var back Diagnosis
	if err := json.Unmarshal(body, &back); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
}
