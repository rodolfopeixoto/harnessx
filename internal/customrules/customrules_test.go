package customrules

import (
	"os"
	"path/filepath"
	"testing"
)

func writeRule(t *testing.T, root, name, body string) {
	t.Helper()
	d := filepath.Join(root, ".harness", "rules")
	if err := os.MkdirAll(d, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(d, name), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestLoadEmptyDirReturnsNil(t *testing.T) {
	dir := t.TempDir()
	rules, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if rules != nil {
		t.Errorf("want nil rules, got %v", rules)
	}
}

func TestLoadParsesValidRule(t *testing.T) {
	dir := t.TempDir()
	writeRule(t, dir, "no-print.yaml", `
id: no-print
description: forbid stray print() calls
severity: warn
when:
  stacks: [python]
  path_globs: ["**/*.py"]
forbid:
  - "print("
`)
	rules, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(rules) != 1 || rules[0].ID != "no-print" {
		t.Fatalf("unexpected rules: %+v", rules)
	}
}

func TestLoadRejectsInvalidSeverity(t *testing.T) {
	dir := t.TempDir()
	writeRule(t, dir, "bad.yaml", "id: bad\nseverity: catastrophic\nforbid:\n  - x\n")
	rules, err := Load(dir)
	if err == nil {
		t.Fatal("expected validation error")
	}
	if len(rules) != 0 {
		t.Errorf("invalid rule should not load: %+v", rules)
	}
}

func TestAppliesToStackFilter(t *testing.T) {
	r := Rule{ID: "x", When: When{Stacks: []string{"python"}}}
	if !r.AppliesTo("python", "x.py") {
		t.Error("python stack should match")
	}
	if r.AppliesTo("go", "x.go") {
		t.Error("go stack should not match")
	}
}

func TestAppliesToWildcardStack(t *testing.T) {
	r := Rule{ID: "x", When: When{Stacks: []string{"*"}}}
	if !r.AppliesTo("anything", "x") {
		t.Error("wildcard should match")
	}
}

func TestAppliesToPathGlob(t *testing.T) {
	r := Rule{ID: "x", When: When{PathGlobs: []string{"*.py"}}}
	if !r.AppliesTo("python", "a.py") {
		t.Error("py should match")
	}
	if r.AppliesTo("python", "a.go") {
		t.Error("go should not match")
	}
}
