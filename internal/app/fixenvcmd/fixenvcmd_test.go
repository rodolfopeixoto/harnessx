package fixenvcmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckGOROOTFlagsMissingDir(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "nope")
	os.RemoveAll(missing)
	opts := Options{
		Root: t.TempDir(),
		UserEnviron: func() []string {
			return []string{"GOROOT=" + missing, "PATH=/usr/bin"}
		},
	}
	checks := checkGOROOT(opts)
	if len(checks) != 1 || checks[0].ID != "goroot_missing" {
		t.Fatalf("want 1 check goroot_missing, got %+v", checks)
	}
}

func TestCheckGOROOTSilentWhenValid(t *testing.T) {
	dir := t.TempDir()
	opts := Options{
		UserEnviron: func() []string { return []string{"GOROOT=" + dir} },
	}
	if got := checkGOROOT(opts); len(got) != 0 {
		t.Fatalf("expected silent for valid GOROOT, got %+v", got)
	}
}

func TestCheckGOROOTSilentWhenUnset(t *testing.T) {
	opts := Options{UserEnviron: func() []string { return nil }}
	if got := checkGOROOT(opts); len(got) != 0 {
		t.Fatalf("expected silent when GOROOT unset, got %+v", got)
	}
}

func TestCheckVenvFlagsMissingVenvForPyProject(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "requirements.txt"), []byte("fastapi\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got := checkVenv(Options{Root: root})
	if len(got) != 1 || got[0].ID != "missing_venv" {
		t.Fatalf("want missing_venv check, got %+v", got)
	}
}

func TestCheckVenvSilentWhenVenvPresent(t *testing.T) {
	root := t.TempDir()
	_ = os.WriteFile(filepath.Join(root, "requirements.txt"), []byte("x\n"), 0o644)
	_ = os.MkdirAll(filepath.Join(root, ".venv"), 0o755)
	if got := checkVenv(Options{Root: root}); len(got) != 0 {
		t.Fatalf("want silent, got %+v", got)
	}
}

func TestCheckNodeModulesFlagsMissing(t *testing.T) {
	root := t.TempDir()
	_ = os.WriteFile(filepath.Join(root, "package.json"), []byte(`{"name":"x"}`), 0o644)
	got := checkNodeModules(Options{Root: root})
	if len(got) != 1 || got[0].ID != "missing_node_modules" {
		t.Fatalf("want missing_node_modules, got %+v", got)
	}
}

func TestRunPrintsHealthyWhenNoFindings(t *testing.T) {
	var buf bytes.Buffer
	_, err := Run(&buf, Options{
		Root:        t.TempDir(),
		UserEnviron: func() []string { return []string{"PATH=/usr/bin:/usr/local/bin:/opt/homebrew/bin"} },
		Lookup:      func(_ string) (string, error) { return "/usr/bin/brew", nil },
	})
	if err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if strings.Contains(out, "[high]") {
		t.Fatalf("unexpected high-severity findings in clean env: %s", out)
	}
}
