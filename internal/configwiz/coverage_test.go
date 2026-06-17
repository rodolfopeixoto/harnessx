package configwiz

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadPropagatesParseError(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, ".harness", "config")
	_ = os.MkdirAll(p, 0o755)
	_ = os.WriteFile(filepath.Join(p, "routes.yaml"), []byte("routes: { bad: ["), 0o644)
	if _, err := Load(dir); err == nil {
		t.Fatal("want error")
	}
}

func TestSavePropagatesMkdirError(t *testing.T) {
	dir := t.TempDir()
	conflict := filepath.Join(dir, ".harness")
	_ = os.WriteFile(conflict, []byte("blocking"), 0o644)
	err := Save(dir, Snapshot{})
	if err == nil {
		t.Fatal("want mkdir error")
	}
	var pErr *os.PathError
	if !errors.As(err, &pErr) && !strings.Contains(err.Error(), "not a directory") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSetTaskPrimaryFailsWhenLoadFails(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, ".harness", "config")
	_ = os.MkdirAll(p, 0o755)
	_ = os.WriteFile(filepath.Join(p, "routes.yaml"), []byte("routes: { bad: ["), 0o644)
	if err := SetTaskPrimary(dir, "planning", "x", nil, 0, ""); err == nil {
		t.Fatal("want error")
	}
}

func TestDeleteTaskFailsWhenLoadFails(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, ".harness", "config")
	_ = os.MkdirAll(p, 0o755)
	_ = os.WriteFile(filepath.Join(p, "routes.yaml"), []byte("routes: { bad: ["), 0o644)
	if err := DeleteTask(dir, "planning"); err == nil {
		t.Fatal("want error")
	}
}
