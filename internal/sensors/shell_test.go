// SPDX-License-Identifier: MIT

package sensors

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestProjectLocalBinaryPrefersVenv(t *testing.T) {
	dir := t.TempDir()
	bin := filepath.Join(dir, ".venv", "bin")
	if err := os.MkdirAll(bin, 0o755); err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(bin, "ruff")
	if err := os.WriteFile(p, []byte("#!/bin/sh\necho ok\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	got := projectLocalBinary(dir, "ruff")
	if got != p {
		t.Errorf("want %s, got %s", p, got)
	}
}

func TestProjectLocalBinaryMissingReturnsEmpty(t *testing.T) {
	if got := projectLocalBinary(t.TempDir(), "absent"); got != "" {
		t.Errorf("expected empty for missing binary, got %q", got)
	}
}

func TestProjectLocalEnvPrependsVenvBin(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".venv", "bin"), 0o755); err != nil {
		t.Fatal(err)
	}
	env := projectLocalEnv(dir)
	var path string
	for _, e := range env {
		if strings.HasPrefix(e, "PATH=") {
			path = e[5:]
		}
	}
	want := filepath.Join(dir, ".venv", "bin")
	if !strings.HasPrefix(path, want+":") {
		t.Errorf("want PATH to start with %s:, got %s", want, path)
	}
}

func TestShellSensorRunsVenvBinaryWithoutPATH(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".venv", "bin"), 0o755); err != nil {
		t.Fatal(err)
	}
	script := "#!/bin/sh\necho hello\n"
	bin := filepath.Join(dir, ".venv", "bin", "mytool")
	if err := os.WriteFile(bin, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	s := ShellSensor{IDValue: "mytool", Binary: "mytool", Args: nil, Timeout: 5 * time.Second}
	res := s.Run(RunCtx{Ctx: context.Background(), Root: dir, OutputDir: t.TempDir()})
	if res.Status != StatusPassed {
		t.Fatalf("want passed, got %s (%s)", res.Status, res.Detail)
	}
}
