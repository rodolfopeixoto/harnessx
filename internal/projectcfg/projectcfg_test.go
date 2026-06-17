// SPDX-License-Identifier: MIT

package projectcfg

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveLoadRoundtrip(t *testing.T) {
	dir := t.TempDir()
	p := Project{Stack: "python", Commands: map[string]string{"test": "pytest"}}
	if err := Save(dir, p); err != nil {
		t.Fatal(err)
	}
	got, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if got.Stack != "python" || got.Commands["test"] != "pytest" {
		t.Fatalf("roundtrip mismatch: %+v", got)
	}
}

func TestResolveUsesProjectYamlFirst(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module x\n"), 0o644)
	_ = Save(dir, Project{Stack: "python", Commands: map[string]string{"test": "custom"}})
	got, err := Resolve(dir, "test")
	if err != nil {
		t.Fatal(err)
	}
	if got != "custom" {
		t.Fatalf("want custom got %q", got)
	}
}

func TestResolveFallsBackToStackDefaults(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module x\n"), 0o644)
	got, err := Resolve(dir, "test")
	if err != nil {
		t.Fatal(err)
	}
	if got != "go test -race ./..." {
		t.Fatalf("want go default got %q", got)
	}
}

func TestResolveErrorsWhenStackUndetectable(t *testing.T) {
	dir := t.TempDir()
	if _, err := Resolve(dir, "test"); err == nil {
		t.Fatal("expected error")
	}
}

func TestDetectMatrix(t *testing.T) {
	for probe, want := range map[string]string{
		"go.mod":         "go",
		"Cargo.toml":     "rust",
		"pyproject.toml": "python",
		"Gemfile":        "ruby",
		"package.json":   "node",
	} {
		dir := t.TempDir()
		_ = os.WriteFile(filepath.Join(dir, probe), []byte("x"), 0o644)
		got, err := Detect(dir)
		if err != nil {
			t.Errorf("%s: %v", probe, err)
			continue
		}
		if got != want {
			t.Errorf("%s: want %q got %q", probe, want, got)
		}
	}
}

func TestKnownCommandsContainsCore(t *testing.T) {
	keys := KnownCommands()
	must := map[string]bool{"test": false, "lint": false, "dev": false, "bench": false}
	for _, k := range keys {
		if _, ok := must[k]; ok {
			must[k] = true
		}
	}
	for k, ok := range must {
		if !ok {
			t.Errorf("missing %q in KnownCommands()", k)
		}
	}
}

func TestFromMetaStripsEmpty(t *testing.T) {
	p := FromMeta("go", map[string]string{"test": "go test", "lint": "  ", "run": ""})
	if _, ok := p.Commands["lint"]; ok {
		t.Errorf("lint should be stripped")
	}
	if _, ok := p.Commands["run"]; ok {
		t.Errorf("run should be stripped")
	}
	if p.Commands["test"] != "go test" {
		t.Errorf("test preserved incorrectly: %q", p.Commands["test"])
	}
}
