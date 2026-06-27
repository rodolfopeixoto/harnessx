package sensors

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInferenceCandidateFlagsUntestedLargeModule(t *testing.T) {
	dir := t.TempDir()
	body := "package x\n"
	for i := 0; i < 350; i++ {
		body += "var x = 1\n"
	}
	if err := os.WriteFile(filepath.Join(dir, "big.go"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	s := InferenceCandidateSensor{IDValue: "ic", MinLinesForFlag: 300}
	res := s.Run(RunCtx{Ctx: context.Background(), Root: dir, OutputDir: t.TempDir()})
	if !strings.Contains(res.Detail, "candidate") {
		t.Fatalf("expected candidates, got %s", res.Detail)
	}
}

func TestInferenceCandidatePassesWhenTestExists(t *testing.T) {
	dir := t.TempDir()
	body := "package x\n"
	for i := 0; i < 350; i++ {
		body += "var x = 1\n"
	}
	if err := os.WriteFile(filepath.Join(dir, "big_test.go"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	s := InferenceCandidateSensor{IDValue: "ic", MinLinesForFlag: 300}
	res := s.Run(RunCtx{Ctx: context.Background(), Root: dir, OutputDir: t.TempDir()})
	if res.Detail != "no inference candidates" {
		t.Fatalf("test file should suppress flag, got %s", res.Detail)
	}
}

func TestInferenceCandidateKind(t *testing.T) {
	s := InferenceCandidateSensor{IDValue: "ic"}
	if s.Kind() != KindInferential {
		t.Fatalf("must be KindInferential, got %s", s.Kind())
	}
}
