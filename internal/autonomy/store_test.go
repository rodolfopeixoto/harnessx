package autonomy

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveLoadRoundTrip(t *testing.T) {
	root := t.TempDir()
	if err := Save(root, SafeExecute); err != nil {
		t.Fatalf("Save: %v", err)
	}
	got, err := Load(root)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got != SafeExecute {
		t.Fatalf("want %s, got %s", SafeExecute, got)
	}
	data, err := os.ReadFile(filepath.Join(root, ActiveFileRel))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != string(SafeExecute)+"\n" {
		t.Errorf("file content: want %q, got %q", string(SafeExecute)+"\n", string(data))
	}
}

func TestLoadDefaultsToManualWhenMissing(t *testing.T) {
	root := t.TempDir()
	got, err := Load(root)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got != Manual {
		t.Fatalf("want %s default, got %s", Manual, got)
	}
}

func TestSaveRejectsInvalidLevel(t *testing.T) {
	root := t.TempDir()
	if err := Save(root, Level("nope")); err == nil {
		t.Fatal("expected error for invalid level")
	}
}
