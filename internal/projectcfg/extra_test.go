package projectcfg

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectRailsBeforeRuby(t *testing.T) {
	dir := t.TempDir()
	_ = os.MkdirAll(filepath.Join(dir, "config"), 0o755)
	_ = os.WriteFile(filepath.Join(dir, "config", "application.rb"), []byte("x"), 0o644)
	_ = os.WriteFile(filepath.Join(dir, "Gemfile"), []byte("y"), 0o644)
	got, err := Detect(dir)
	if err != nil {
		t.Fatal(err)
	}
	if got != "rails" {
		t.Errorf("want rails got %q", got)
	}
}

func TestResolveErrorsOnUnknownCommand(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module x\n"), 0o644)
	_, err := Resolve(dir, "completely-fake-command")
	if err == nil {
		t.Fatal("want error for unknown command")
	}
}

func TestLoadParsesValidYAML(t *testing.T) {
	dir := t.TempDir()
	_ = os.MkdirAll(filepath.Join(dir, ".harness", "config"), 0o755)
	_ = os.WriteFile(filepath.Join(dir, ".harness", "config", "project.yaml"),
		[]byte("stack: go\ncommands:\n  test: go test\n"), 0o644)
	p, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if p.Stack != "go" || p.Commands["test"] != "go test" {
		t.Errorf("got %+v", p)
	}
}

func TestLoadFailsOnGarbageYAML(t *testing.T) {
	dir := t.TempDir()
	_ = os.MkdirAll(filepath.Join(dir, ".harness", "config"), 0o755)
	_ = os.WriteFile(filepath.Join(dir, ".harness", "config", "project.yaml"), []byte(":::not yaml"), 0o644)
	if _, err := Load(dir); err == nil {
		t.Fatal("want parse error")
	}
}

func TestSaveCreatesDirs(t *testing.T) {
	dir := t.TempDir()
	if err := Save(dir, Project{Stack: "go"}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, ".harness", "config", "project.yaml")); err != nil {
		t.Errorf("not written: %v", err)
	}
}

func TestDefaultsAllStacksContainTestLintRun(t *testing.T) {
	for _, s := range []string{"python", "go", "rust", "ruby", "rails", "node"} {
		m := defaults(s)
		for _, key := range []string{"test", "lint", "run"} {
			if _, ok := m[key]; !ok {
				t.Errorf("stack %s missing %s", s, key)
			}
		}
	}
}
