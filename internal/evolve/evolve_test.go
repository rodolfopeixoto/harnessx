package evolve

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeEvents(t *testing.T, dir string, lines []string) {
	t.Helper()
	p := filepath.Join(dir, ".harness", "logs")
	_ = os.MkdirAll(p, 0o755)
	_ = os.WriteFile(filepath.Join(p, "events.jsonl"), []byte(strings.Join(lines, "\n")+"\n"), 0o644)
}

func TestDiagnoseHandlesMissingLog(t *testing.T) {
	d, err := Diagnose(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if d.Events != 0 || len(d.Clusters) != 0 {
		t.Errorf("unexpected diagnosis: %+v", d)
	}
}

func TestDiagnoseClustersFailures(t *testing.T) {
	dir := t.TempDir()
	writeEvents(t, dir, []string{
		`{"time":"t1","level":"info","fields":{"stage":"init","status":"ok"}}`,
		`{"time":"t2","level":"error","fields":{"stage":"do","status":"failed","sensor":"py_pytest"}}`,
		`{"time":"t3","level":"error","fields":{"stage":"do","status":"failed","sensor":"py_pytest"}}`,
		`{"time":"t4","level":"error","fields":{"stage":"do","status":"failed","sensor":"go_test"}}`,
	})
	d, err := Diagnose(dir)
	if err != nil {
		t.Fatal(err)
	}
	if d.Events != 4 || d.Failures != 3 {
		t.Fatalf("counts: %+v", d)
	}
	if len(d.Clusters) != 2 {
		t.Fatalf("clusters: %+v", d.Clusters)
	}
	if d.Clusters[0].Count != 2 {
		t.Errorf("top cluster should have count 2, got %d", d.Clusters[0].Count)
	}
}

func TestPromoteRequiresHITL(t *testing.T) {
	if err := Promote(t.TempDir(), PromoteOptions{HITL: false}); err == nil {
		t.Fatal("expected error without --hitl")
	} else if !strings.Contains(err.Error(), "hitl") {
		t.Errorf("error must mention hitl: %v", err)
	}
}

func TestProposeAppendsMutationsLog(t *testing.T) {
	dir := t.TempDir()
	id, err := Propose(dir, Mutation{Component: "router", Description: "lower retry"})
	if err != nil {
		t.Fatal(err)
	}
	if id == "" {
		t.Fatal("missing id")
	}
	body, err := os.ReadFile(filepath.Join(dir, ".harness", "logs", "mutations.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), id) {
		t.Errorf("log does not contain id %s: %s", id, body)
	}
}

func TestReplayCountsSignatureMatches(t *testing.T) {
	dir := t.TempDir()
	writeEvents(t, dir, []string{
		`{"level":"error","fields":{"stage":"do","status":"failed","sensor":"py_pytest"}}`,
	})
	d, _ := Diagnose(dir)

	tracePath := filepath.Join(dir, "trace.jsonl")
	_ = os.WriteFile(tracePath, []byte(strings.Join([]string{
		`{"level":"error","fields":{"stage":"do","status":"failed","sensor":"py_pytest"}}`,
		`{"level":"error","fields":{"stage":"do","status":"failed","sensor":"go_test"}}`,
	}, "\n")+"\n"), 0o644)

	res, err := Replay(dir, tracePath, d)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		t.Fatal(err)
	}
	if res.Replayed != 2 || res.Matched != 1 {
		t.Errorf("res: %+v", res)
	}
}
