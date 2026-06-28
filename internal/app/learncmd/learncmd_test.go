package learncmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
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
