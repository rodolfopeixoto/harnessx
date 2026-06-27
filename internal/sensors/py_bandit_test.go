package sensors

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPyBanditArgsExcludesTestsByDefault(t *testing.T) {
	got := pyBanditArgs(t.TempDir())
	joined := strings.Join(got, " ")
	if !strings.Contains(joined, ",tests") {
		t.Fatalf("default excludes should contain `tests`: %s", joined)
	}
	if strings.Contains(joined, "-c") {
		t.Fatalf("no -c expected when no .bandit present: %s", joined)
	}
	if got[len(got)-1] != "." {
		t.Fatalf("last arg should be `.`, got %s", got[len(got)-1])
	}
}

func TestPyBanditArgsPicksUpDotBandit(t *testing.T) {
	root := t.TempDir()
	cfg := filepath.Join(root, ".bandit")
	if err := os.WriteFile(cfg, []byte("[bandit]\nskips=B101\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got := pyBanditArgs(root)
	joined := strings.Join(got, " ")
	if !strings.Contains(joined, "-c "+cfg) {
		t.Fatalf("expected -c %s in args, got %s", cfg, joined)
	}
}

func TestPyBanditArgsPrefersDotBanditOverYAML(t *testing.T) {
	root := t.TempDir()
	dot := filepath.Join(root, ".bandit")
	yml := filepath.Join(root, "bandit.yaml")
	if err := os.WriteFile(dot, []byte("[bandit]\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(yml, []byte("skips: [B101]\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got := pyBanditArgs(root)
	joined := strings.Join(got, " ")
	if !strings.Contains(joined, dot) {
		t.Fatalf("expected .bandit chosen, got %s", joined)
	}
	if strings.Contains(joined, yml) {
		t.Fatalf("yaml should not be chosen when .bandit present: %s", joined)
	}
}
