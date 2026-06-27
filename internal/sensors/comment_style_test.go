package sensors

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCommentStylePassesEmptyProject(t *testing.T) {
	dir := t.TempDir()
	s := CommentStyleSensor{IDValue: "comment_style"}
	res := s.Run(RunCtx{Ctx: context.Background(), Root: dir, OutputDir: t.TempDir()})
	if res.Status != StatusPassed {
		t.Fatalf("want passed, got %s — %s", res.Status, res.Detail)
	}
}

func TestCommentStyleFlagsTODOWithoutRef(t *testing.T) {
	dir := t.TempDir()
	body := "package x\n// TODO finish later\n"
	if err := os.WriteFile(filepath.Join(dir, "a.go"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	s := CommentStyleSensor{IDValue: "comment_style"}
	res := s.Run(RunCtx{Ctx: context.Background(), Root: dir, OutputDir: t.TempDir()})
	if res.Status != StatusFailed {
		t.Fatalf("want failed, got %s", res.Status)
	}
}

func TestCommentStyleAllowsTODOWithRef(t *testing.T) {
	dir := t.TempDir()
	body := "package x\n// TODO FOO-123 finish later\n"
	if err := os.WriteFile(filepath.Join(dir, "a.go"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	s := CommentStyleSensor{IDValue: "comment_style"}
	res := s.Run(RunCtx{Ctx: context.Background(), Root: dir, OutputDir: t.TempDir()})
	if res.Status != StatusPassed {
		t.Fatalf("want passed, got %s — %s", res.Status, res.Detail)
	}
}

func TestCommentStyleFlagsWhatComment(t *testing.T) {
	dir := t.TempDir()
	body := "package x\n// increment the counter\nvar i = 0\n"
	if err := os.WriteFile(filepath.Join(dir, "a.go"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	s := CommentStyleSensor{IDValue: "comment_style"}
	res := s.Run(RunCtx{Ctx: context.Background(), Root: dir, OutputDir: t.TempDir()})
	if res.Status != StatusFailed {
		t.Fatalf("want failed, got %s", res.Status)
	}
}

func TestCommentStyleDisabledViaConfig(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, ".harness", "config")
	if err := os.MkdirAll(cfg, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cfg, "comment_style.yaml"), []byte("disabled: true\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	body := "package x\n// TODO finish later\n"
	if err := os.WriteFile(filepath.Join(dir, "a.go"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	s := CommentStyleSensor{IDValue: "comment_style"}
	res := s.Run(RunCtx{Ctx: context.Background(), Root: dir, OutputDir: t.TempDir()})
	if res.Status != StatusSkipped {
		t.Fatalf("want skipped, got %s — %s", res.Status, res.Detail)
	}
}

func TestCommentStyleConfigBannedPattern(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, ".harness", "config")
	if err := os.MkdirAll(cfg, 0o755); err != nil {
		t.Fatal(err)
	}
	yaml := `banned_patterns:
  - "^// XXX"
require_todo_ref: false
forbid_what_comments: false
`
	if err := os.WriteFile(filepath.Join(cfg, "comment_style.yaml"), []byte(yaml), 0o644); err != nil {
		t.Fatal(err)
	}
	body := "package x\n// XXX legacy\n"
	if err := os.WriteFile(filepath.Join(dir, "a.go"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	s := CommentStyleSensor{IDValue: "comment_style"}
	res := s.Run(RunCtx{Ctx: context.Background(), Root: dir, OutputDir: t.TempDir()})
	if res.Status != StatusFailed {
		t.Fatalf("want failed, got %s", res.Status)
	}
	if !strings.Contains(res.Detail, "1 comment issue") {
		t.Errorf("detail mismatch: %s", res.Detail)
	}
}
