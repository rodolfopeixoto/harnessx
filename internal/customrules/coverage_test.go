package customrules

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadPropagatesReadDirError(t *testing.T) {
	dir := t.TempDir()
	conflict := filepath.Join(dir, ".harness", "rules")
	_ = os.MkdirAll(filepath.Dir(conflict), 0o755)
	_ = os.WriteFile(conflict, []byte("blocking file"), 0o644)
	if _, err := Load(dir); err == nil {
		t.Fatal("want error when rules path is a file, not a dir")
	}
}

func TestRuleValidateAcceptsRequireOnly(t *testing.T) {
	r := Rule{ID: "req", Severity: "error", Require: []string{"x"}}
	if err := r.Validate(); err != nil {
		t.Fatalf("require-only should pass: %v", err)
	}
}

func TestAppliesToBlockedByPathGlob(t *testing.T) {
	r := Rule{ID: "x", When: When{PathGlobs: []string{"src/**"}}}
	if r.AppliesTo("python", "tests/x.py") {
		t.Error("non-matching path should not match")
	}
}

func TestLoadIgnoresFilesWithBrokenYAML(t *testing.T) {
	dir := t.TempDir()
	writeRule(t, dir, "broken.yaml", "::: invalid yaml :::")
	rules, err := Load(dir)
	if err == nil {
		t.Fatal("want error")
	}
	if len(rules) != 0 {
		t.Errorf("broken yaml should not yield rules: %v", rules)
	}
}
