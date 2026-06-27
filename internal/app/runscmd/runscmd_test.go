package runscmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupRun(t *testing.T, root, id string, meta string) {
	t.Helper()
	dir := filepath.Join(root, ".harness", "runs", id)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if meta != "" {
		if err := os.WriteFile(filepath.Join(dir, "meta.json"), []byte(meta), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func TestListEmpty(t *testing.T) {
	root := t.TempDir()
	var buf bytes.Buffer
	if err := List(&buf, Options{Root: root}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "no runs") {
		t.Fatalf("want 'no runs', got: %s", buf.String())
	}
}

func TestListIncludesIncomplete(t *testing.T) {
	root := t.TempDir()
	setupRun(t, root, "r1", `{"run_id":"r1","status":"applied"}`)
	setupRun(t, root, "r2", "")
	if err := os.WriteFile(filepath.Join(root, ".harness", "runs", "r2", "report.md"), []byte("# orphan"), 0o644); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := List(&buf, Options{Root: root}); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"r1", "r2", "incomplete"} {
		if !strings.Contains(buf.String(), want) {
			t.Errorf("missing %q in output:\n%s", want, buf.String())
		}
	}
}

func TestInspectIncompleteMessage(t *testing.T) {
	root := t.TempDir()
	setupRun(t, root, "r_orphan", "")
	var buf bytes.Buffer
	err := Inspect(&buf, Options{Root: root, RunID: "r_orphan"})
	if err == nil {
		t.Fatal("expected error for orphan run")
	}
	if !strings.Contains(err.Error(), "meta.json") {
		t.Fatalf("error should mention meta.json: %v", err)
	}
}

func TestReportFallsBackToOrphanReportMD(t *testing.T) {
	root := t.TempDir()
	setupRun(t, root, "r_orphan", "")
	body := "# stub report"
	if err := os.WriteFile(filepath.Join(root, ".harness", "runs", "r_orphan", "report.md"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := Report(&buf, Options{Root: root, RunID: "r_orphan"}); err != nil {
		t.Fatal(err)
	}
	if buf.String() != body {
		t.Fatalf("want %q, got %q", body, buf.String())
	}
}
