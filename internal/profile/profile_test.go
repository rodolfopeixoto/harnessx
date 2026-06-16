// SPDX-License-Identifier: MIT

package profile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSnapshotPopulated(t *testing.T) {
	h := Snapshot()
	if h.Sys == 0 {
		t.Fatal("sys must be populated")
	}
	if h.CapturedAt.IsZero() {
		t.Fatal("captured at must be set")
	}
}

func TestSnapshotString(t *testing.T) {
	h := Snapshot()
	s := h.String()
	if !strings.Contains(s, "alloc=") || !strings.Contains(s, "sys=") {
		t.Fatalf("string missing fields: %s", s)
	}
}

func TestWriteHeap(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "heap.pprof")
	if err := WriteHeap(p); err != nil {
		t.Fatal(err)
	}
	st, err := os.Stat(p)
	if err != nil {
		t.Fatal(err)
	}
	if st.Size() == 0 {
		t.Fatal("heap profile empty")
	}
}

func TestStartCPU(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "cpu.pprof")
	stop, err := StartCPU(p)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 1_000_000; i++ {
		_ = i * i
	}
	if err := stop(); err != nil {
		t.Fatal(err)
	}
	st, err := os.Stat(p)
	if err != nil {
		t.Fatal(err)
	}
	if st.Size() == 0 {
		t.Fatal("cpu profile empty")
	}
}

func TestStartCPUBadPath(t *testing.T) {
	if _, err := StartCPU("/nonexistent-dir-xyz/cpu.pprof"); err == nil {
		t.Fatal("expected error on bad path")
	}
}

func TestWriteHeapBadPath(t *testing.T) {
	if err := WriteHeap("/nonexistent-dir-xyz/heap.pprof"); err == nil {
		t.Fatal("expected error")
	}
}

func TestDiffPct(t *testing.T) {
	if got := DiffPct(100, 110); got != 10 {
		t.Fatalf("want 10, got %v", got)
	}
	if got := DiffPct(0, 100); got != 0 {
		t.Fatalf("want 0 for zero base, got %v", got)
	}
	if got := DiffPct(100, 50); got != -50 {
		t.Fatalf("want -50, got %v", got)
	}
}
