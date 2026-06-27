package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEnsurePytestCovDetectsPlugin(t *testing.T) {
	dir := t.TempDir()
	bin := filepath.Join(dir, "pytest")
	body := "#!/bin/sh\necho '  --cov=PATH      coverage report'\n"
	if err := os.WriteFile(bin, []byte(body), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := ensurePytestCov(bin); err != nil {
		t.Fatalf("want nil, got %v", err)
	}
}

func TestEnsurePytestCovReportsMissingPlugin(t *testing.T) {
	dir := t.TempDir()
	bin := filepath.Join(dir, "pytest")
	body := "#!/bin/sh\necho '  -q  quiet'\n"
	if err := os.WriteFile(bin, []byte(body), 0o755); err != nil {
		t.Fatal(err)
	}
	err := ensurePytestCov(bin)
	if err == nil {
		t.Fatal("want error when --cov missing")
	}
	if !strings.Contains(err.Error(), "pytest-cov") {
		t.Fatalf("error should mention pytest-cov: %v", err)
	}
}
