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

func TestPyBanditArgsDoesNotPassDashCForDotBanditINI(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, ".bandit"), []byte("[bandit]\nskips=B101\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got := pyBanditArgs(root)
	joined := strings.Join(got, " ")
	if strings.Contains(joined, "-c") {
		t.Fatalf("must not pass -c for INI .bandit (bandit auto-discovers it), got %s", joined)
	}
}

func TestPyBanditArgsPassesDashCForYAML(t *testing.T) {
	root := t.TempDir()
	yml := filepath.Join(root, "bandit.yaml")
	if err := os.WriteFile(yml, []byte("skips: [B101]\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got := pyBanditArgs(root)
	joined := strings.Join(got, " ")
	if !strings.Contains(joined, "-c "+yml) {
		t.Fatalf("expected -c %s for YAML config, got %s", yml, joined)
	}
}
