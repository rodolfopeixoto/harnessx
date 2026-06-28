package learncmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ropeixoto/harnessx/internal/execution"
)

func writeRun(t *testing.T, root, id, body string) {
	t.Helper()
	dir := filepath.Join(root, ".harness", "runs", id)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "meta.json"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestRunEmptyShowsZero(t *testing.T) {
	var buf bytes.Buffer
	_, err := Run(&buf, Options{Root: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "0 runs analyzed") {
		t.Fatalf("expected zero-run summary, got: %s", buf.String())
	}
}

func TestRunSurfacesDominantAdapter(t *testing.T) {
	root := t.TempDir()
	for _, id := range []string{"r1", "r2", "r3"} {
		writeRun(t, root, id, `{"run_id":"`+id+`","agent_id":"claude","status":"applied","input_tokens":1000,"output_tokens":500,"estimated_cost_usd":0.05}`)
	}
	writeRun(t, root, "r4", `{"run_id":"r4","agent_id":"codex","status":"applied","input_tokens":900,"output_tokens":400,"estimated_cost_usd":0.04}`)
	var buf bytes.Buffer
	res, err := Run(&buf, Options{Root: root})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "Most-used adapter") {
		t.Fatalf("expected dominant adapter pattern, got: %s", buf.String())
	}
	if res.RunsAnalyzed != 4 {
		t.Errorf("want 4 runs, got %d", res.RunsAnalyzed)
	}
}

func TestRunSurfacesRecurringError(t *testing.T) {
	root := t.TempDir()
	for _, id := range []string{"r1", "r2"} {
		writeRun(t, root, id, `{"run_id":"`+id+`","agent_id":"claude","status":"agent_failed","error_type":"budget_exceeded"}`)
	}
	var buf bytes.Buffer
	if _, err := Run(&buf, Options{Root: root}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "budget_exceeded") {
		t.Fatalf("expected recurring failure flag, got: %s", buf.String())
	}
	if !strings.Contains(buf.String(), "Raise --budget-usd") {
		t.Fatalf("expected actionable fix, got: %s", buf.String())
	}
}

func TestApplySetsAutonomyOnWaitingApproval(t *testing.T) {
	root := t.TempDir()
	for i := 0; i < 4; i++ {
		writeRun(t, root, "r"+string(rune('a'+i)), `{"run_id":"r","agent_id":"claude","status":"waiting_approval","changed_files":["x"]}`)
	}
	var buf bytes.Buffer
	res, err := Run(&buf, Options{Root: root, Apply: true})
	if err != nil {
		t.Fatal(err)
	}
	if !containsLearn(res.Applied, "autonomy_safe_execute") {
		t.Fatalf("expected autonomy fix applied, got %+v\noutput:\n%s", res.Applied, buf.String())
	}
	body, err := os.ReadFile(filepath.Join(root, ".harness", "config", "autonomy"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(string(body)) != "safe_execute" {
		t.Fatalf("expected safe_execute, got %s", string(body))
	}
}

func containsLearn(ss []string, want string) bool {
	for _, s := range ss {
		if s == want {
			return true
		}
	}
	return false
}

func TestUpdateIncrementalAppendsAndDedupes(t *testing.T) {
	root := t.TempDir()
	run1 := makeExecRes("r1", "claude", "applied", 1000, 500, 0.05, "")
	inc1, path1, err := UpdateIncremental(root, run1)
	if err != nil {
		t.Fatal(err)
	}
	if path1 == "" {
		t.Fatal("expected written path")
	}
	if inc1.RunsSeen != 1 {
		t.Errorf("want 1 run seen, got %d", inc1.RunsSeen)
	}
	if inc1.TokensTotal != 1500 {
		t.Errorf("want 1500 tokens, got %d", inc1.TokensTotal)
	}
	if _, _, err := UpdateIncremental(root, run1); err != nil {
		t.Fatal(err)
	}
	inc2, err := LoadIncremental(root)
	if err != nil {
		t.Fatal(err)
	}
	if inc2.RunsSeen != 1 {
		t.Errorf("idempotent on same run id; want still 1, got %d", inc2.RunsSeen)
	}
	run2 := makeExecRes("r2", "codex", "applied", 2000, 800, 0.10, "")
	inc3, _, err := UpdateIncremental(root, run2)
	if err != nil {
		t.Fatal(err)
	}
	if inc3.RunsSeen != 2 {
		t.Errorf("want 2 runs seen, got %d", inc3.RunsSeen)
	}
	if inc3.ByAdapter["codex"] != 1 || inc3.ByAdapter["claude"] != 1 {
		t.Errorf("ByAdapter wrong: %+v", inc3.ByAdapter)
	}
}

func makeExecRes(id, adapter, status string, in, out int, cost float64, errType string) execution.Result {
	return execution.Result{
		RunID:            id,
		AgentID:          adapter,
		Status:           execution.Status(status),
		InputTokens:      in,
		OutputTokens:     out,
		EstimatedCostUSD: cost,
		ErrorType:        errType,
	}
}

func TestRunWriteFilePersists(t *testing.T) {
	root := t.TempDir()
	writeRun(t, root, "r1", `{"run_id":"r1","agent_id":"x","status":"applied"}`)
	var buf bytes.Buffer
	res, err := Run(&buf, Options{Root: root, WriteFile: true})
	if err != nil {
		t.Fatal(err)
	}
	if res.WrittenPath == "" {
		t.Fatalf("expected written path, got empty")
	}
	if _, err := os.Stat(res.WrittenPath); err != nil {
		t.Fatalf("file not written: %v", err)
	}
}
