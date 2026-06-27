package sensors

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSmellSensorPassesOnCleanFile(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "small.go")
	if err := os.WriteFile(src, []byte("package x\nfunc Add(a,b int) int { return a+b }\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	s := SmellSensor{IDValue: "go_smell", Extensions: []string{".go"}, MaxFileLines: 600, MaxFuncLines: 60, MaxNestDepth: 6, MaxMethodsCls: 25}
	res := s.Run(RunCtx{Ctx: context.Background(), Root: dir, OutputDir: t.TempDir()})
	if res.Status != StatusPassed {
		t.Fatalf("want passed, got %s — %s", res.Status, res.Detail)
	}
}

func TestSmellSensorFlagsLongFile(t *testing.T) {
	dir := t.TempDir()
	body := "package x\n"
	for i := 0; i < 700; i++ {
		body += "// line\n"
	}
	if err := os.WriteFile(filepath.Join(dir, "big.go"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	s := SmellSensor{IDValue: "go_smell", Extensions: []string{".go"}, MaxFileLines: 600}
	res := s.Run(RunCtx{Ctx: context.Background(), Root: dir, OutputDir: t.TempDir()})
	if res.Status != StatusFailed {
		t.Fatalf("want failed, got %s", res.Status)
	}
	if !strings.Contains(res.Detail, "smell") {
		t.Errorf("detail should mention smell, got %s", res.Detail)
	}
}

func TestSmellSensorFlagsLongMethod(t *testing.T) {
	dir := t.TempDir()
	var b strings.Builder
	b.WriteString("package x\nfunc Big() {\n")
	for i := 0; i < 80; i++ {
		b.WriteString("    x := 1\n")
	}
	b.WriteString("}\n")
	if err := os.WriteFile(filepath.Join(dir, "long.go"), []byte(b.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	s := SmellSensor{IDValue: "go_smell", Extensions: []string{".go"}, MaxFuncLines: 50}
	res := s.Run(RunCtx{Ctx: context.Background(), Root: dir, OutputDir: t.TempDir()})
	if res.Status != StatusFailed {
		t.Fatalf("want failed, got %s", res.Status)
	}
}

func TestSmellSensorSkipsNodeModules(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "node_modules", "foo"), 0o755); err != nil {
		t.Fatal(err)
	}
	body := "function huge(){\n" + strings.Repeat("  return 1;\n", 200) + "}\n"
	if err := os.WriteFile(filepath.Join(dir, "node_modules", "foo", "index.js"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	s := SmellSensor{IDValue: "node_smell", Extensions: []string{".js"}, MaxFuncLines: 50}
	res := s.Run(RunCtx{Ctx: context.Background(), Root: dir, OutputDir: t.TempDir()})
	if res.Status != StatusPassed {
		t.Fatalf("must skip node_modules, got %s", res.Status)
	}
}

func TestDefaultExtensionsKnownStacks(t *testing.T) {
	for _, s := range []string{"go", "python", "java", "kotlin", "swift", "elixir", "php", "dotnet", "dart"} {
		if got := defaultExtensions(s); len(got) == 0 {
			t.Errorf("stack %s: expected extensions, got none", s)
		}
	}
	if got := defaultExtensions("unknown"); got != nil {
		t.Errorf("unknown stack should return nil, got %v", got)
	}
}
