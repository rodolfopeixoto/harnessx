package customrules

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadIgnoresSubdirectories(t *testing.T) {
	dir := t.TempDir()
	d := filepath.Join(dir, ".harness", "rules", "nested")
	_ = os.MkdirAll(d, 0o755)
	_ = os.WriteFile(filepath.Join(d, "ignored.yaml"), []byte("id: x\nforbid:\n  - y\n"), 0o644)
	rules, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(rules) != 0 {
		t.Errorf("subdir must be ignored: %v", rules)
	}
}

func TestLoadIgnoresNonYAMLExtensions(t *testing.T) {
	dir := t.TempDir()
	writeRule(t, dir, "x.txt", "anything")
	rules, _ := Load(dir)
	if len(rules) != 0 {
		t.Errorf("txt must be ignored: %v", rules)
	}
}

func TestRuleValidateRejectsMissingID(t *testing.T) {
	r := Rule{Forbid: []string{"x"}}
	if err := r.Validate(); err == nil {
		t.Fatal("want error")
	}
}

func TestRuleValidateRejectsEmptyPatterns(t *testing.T) {
	r := Rule{ID: "x"}
	if err := r.Validate(); err == nil {
		t.Fatal("want error")
	}
}

func TestRuleValidateAcceptsInfoSeverity(t *testing.T) {
	r := Rule{ID: "x", Severity: "info", Forbid: []string{"y"}}
	if err := r.Validate(); err != nil {
		t.Errorf("info should be accepted: %v", err)
	}
}

func TestAppliesToNoStacksNoGlobsAlwaysMatches(t *testing.T) {
	r := Rule{ID: "x"}
	if !r.AppliesTo("anything", "x") {
		t.Error("default rule should match")
	}
}

func TestLoadAggregatesInvalidFilesWithError(t *testing.T) {
	dir := t.TempDir()
	writeRule(t, dir, "good.yaml", "id: good\nforbid:\n  - x\n")
	writeRule(t, dir, "bad.yaml", "id: \nforbid:\n  - y\n")
	rules, err := Load(dir)
	if err == nil {
		t.Fatal("want aggregated error")
	}
	if len(rules) != 1 || rules[0].ID != "good" {
		t.Errorf("good rule should still load: %v", rules)
	}
}
