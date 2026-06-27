package execution

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestListRunsIncludesIncompleteDirs(t *testing.T) {
	root := t.TempDir()
	runs := filepath.Join(root, ".harness", "runs")
	if err := os.MkdirAll(filepath.Join(runs, "run_complete"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(runs, "run_orphan"), 0o755); err != nil {
		t.Fatal(err)
	}
	meta := `{"run_id":"run_complete","agent_id":"fake","status":"applied"}`
	if err := os.WriteFile(filepath.Join(runs, "run_complete", "meta.json"), []byte(meta), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(runs, "run_orphan", "report.md"), []byte("# orphan"), 0o644); err != nil {
		t.Fatal(err)
	}

	out, err := ListRuns(root)
	if err != nil {
		t.Fatalf("ListRuns: %v", err)
	}
	if len(out) != 2 {
		t.Fatalf("expected 2 runs (complete + orphan), got %d: %+v", len(out), out)
	}
	var seenIncomplete, seenComplete bool
	for _, r := range out {
		switch r.RunID {
		case "run_orphan":
			seenIncomplete = true
			if r.Status != StatusIncomplete {
				t.Errorf("orphan run status: want %s, got %s", StatusIncomplete, r.Status)
			}
			if r.ReportPath == "" {
				t.Errorf("orphan run should expose report path")
			}
		case "run_complete":
			seenComplete = true
			if r.Status != "applied" {
				t.Errorf("complete run status: want applied, got %s", r.Status)
			}
		}
	}
	if !seenIncomplete || !seenComplete {
		t.Fatalf("missing entries: incomplete=%v complete=%v", seenIncomplete, seenComplete)
	}
}

func TestLoadRunReturnsTypedErrorWhenMetaMissing(t *testing.T) {
	root := t.TempDir()
	runDir := filepath.Join(root, ".harness", "runs", "run_x")
	if err := os.MkdirAll(runDir, 0o755); err != nil {
		t.Fatal(err)
	}
	_, err := LoadRun(root, "run_x")
	if err == nil {
		t.Fatal("expected error when meta.json missing")
	}
	if !errors.Is(err, ErrRunIncomplete) {
		t.Fatalf("expected ErrRunIncomplete, got %v", err)
	}
}

func TestListRunsEmptyDirNoError(t *testing.T) {
	root := t.TempDir()
	out, err := ListRuns(root)
	if err != nil {
		t.Fatalf("ListRuns: %v", err)
	}
	if len(out) != 0 {
		t.Fatalf("expected empty, got %d", len(out))
	}
}
