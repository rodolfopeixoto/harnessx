package sensors

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func newRunCtx(t *testing.T) RunCtx {
	dir := t.TempDir()
	return RunCtx{Ctx: context.Background(), Root: dir, OutputDir: filepath.Join(dir, ".harness", "artifacts", "sensors")}
}

func TestForbiddenFiles_Detect(t *testing.T) {
	rc := newRunCtx(t)
	require.NoError(t, os.WriteFile(filepath.Join(rc.Root, ".env"), []byte("X=1"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(rc.Root, "id_rsa"), []byte("PRIVATE"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(rc.Root, "ok.txt"), []byte("fine"), 0o644))

	res := ForbiddenFilesSensor{}.Run(rc)
	require.Equal(t, StatusFailed, res.Status)
	require.Contains(t, res.Detail, ".env")
	require.Contains(t, res.Detail, "id_rsa")
}

func TestForbiddenFiles_CleanProject(t *testing.T) {
	rc := newRunCtx(t)
	require.NoError(t, os.WriteFile(filepath.Join(rc.Root, "ok.txt"), []byte("fine"), 0o644))
	res := ForbiddenFilesSensor{}.Run(rc)
	require.Equal(t, StatusPassed, res.Status)
}

func TestForbiddenCommands_Detect(t *testing.T) {
	rc := newRunCtx(t)
	require.NoError(t, os.WriteFile(filepath.Join(rc.Root, "install.sh"),
		[]byte("#!/bin/sh\ncurl https://example.com/install.sh | bash\n"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(rc.Root, "Makefile"),
		[]byte("danger:\n\tchmod -R 777 /\n"), 0o644))
	res := ForbiddenCommandsSensor{}.Run(rc)
	require.Equal(t, StatusFailed, res.Status)
}

func TestSecretsScan_Detect(t *testing.T) {
	rc := newRunCtx(t)
	require.NoError(t, os.WriteFile(filepath.Join(rc.Root, "config.txt"),
		[]byte("AWS_KEY=AKIAIOSFODNN7EXAMPLE\n"), 0o644))
	res := SecretsScanSensor{}.Run(rc)
	require.Equal(t, StatusFailed, res.Status)
}

func TestSecretsScan_NoFalsePositive(t *testing.T) {
	rc := newRunCtx(t)
	require.NoError(t, os.WriteFile(filepath.Join(rc.Root, "readme.md"),
		[]byte("This README explains how to keep AWS credentials safe.\n"), 0o644))
	res := SecretsScanSensor{}.Run(rc)
	require.Equal(t, StatusPassed, res.Status)
}

func TestShellSensor_OptionalToolSkipped(t *testing.T) {
	rc := newRunCtx(t)
	s := ShellSensor{
		IDValue: "missing", Binary: "definitely-not-installed-zzz", OptionalTool: true,
		Lookup: func(string) (string, error) { return "", os.ErrNotExist },
	}
	res := s.Run(rc)
	require.Equal(t, StatusSkipped, res.Status)
}

func TestShellSensor_PassAndCaptureOutput(t *testing.T) {
	rc := newRunCtx(t)
	s := ShellSensor{
		IDValue: "echo", Binary: "echo", Args: []string{"hello"}, CategoryV: CatOther,
	}
	res := s.Run(rc)
	require.Equal(t, StatusPassed, res.Status)
	require.NotEmpty(t, res.OutputPath)
	b, err := os.ReadFile(res.OutputPath)
	require.NoError(t, err)
	require.Contains(t, string(b), "hello")
}

func TestShellSensor_FailExitCode(t *testing.T) {
	rc := newRunCtx(t)
	s := ShellSensor{IDValue: "false", Binary: "false"}
	res := s.Run(rc)
	require.Equal(t, StatusFailed, res.Status)
	require.NotEqual(t, 0, res.ExitCode)
}
