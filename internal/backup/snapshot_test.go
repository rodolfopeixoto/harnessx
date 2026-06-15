package backup

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPackUnpack_RoundTrip(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()
	configDir := filepath.Join(src, ".harness", "config")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(configDir, "harness.yaml"), []byte("hello\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, ".harness", "secrets.enc"), []byte("DO NOT COPY"), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg := DefaultConfig()
	tar := filepath.Join(t.TempDir(), "snap.tar.gz")
	m, err := Pack(src, cfg, "test-tag", tar, false)
	if err != nil {
		t.Fatalf("pack: %v", err)
	}
	if len(m.IncludedFiles) != 1 {
		t.Fatalf("expected 1 included file (harness.yaml), got %d: %+v", len(m.IncludedFiles), m.IncludedFiles)
	}
	if _, err := Unpack(tar, dst, true); err != nil {
		t.Fatalf("unpack: %v", err)
	}
	got, err := os.ReadFile(filepath.Join(dst, ".harness", "config", "harness.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "hello\n" {
		t.Fatalf("round-trip mismatch: %q", got)
	}
	if _, err := os.Stat(filepath.Join(dst, ".harness", "secrets.enc")); err == nil {
		t.Fatal("secrets.enc leaked into snapshot")
	}
}

func TestPackIncludesSecretsWhenAllowed(t *testing.T) {
	src := t.TempDir()
	if err := os.MkdirAll(filepath.Join(src, ".harness"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, ".harness", "secrets.enc"), []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg := DefaultConfig()
	cfg.Include = append(cfg.Include, ".harness/secrets.enc")
	tar := filepath.Join(t.TempDir(), "snap.tar.gz")
	m, err := Pack(src, cfg, "", tar, true)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, f := range m.IncludedFiles {
		if f.Path == ".harness/secrets.enc" {
			found = true
		}
	}
	if !found {
		t.Fatal("expected secrets.enc included when allowed")
	}
}

func TestPolicy_DefaultExcludesSecrets(t *testing.T) {
	p := newPolicy(DefaultConfig(), false)
	if p.Allow(".harness/secrets.enc") {
		t.Fatal("secrets.enc should be denied by default")
	}
	if p.Allow(".harness/secret-seed") {
		t.Fatal("secret-seed should be denied by default")
	}
	if !p.Allow(".harness/config/harness.yaml") {
		t.Fatal("config file should be allowed")
	}
}

func TestPolicy_IncludeSecretsOptIn(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Include = append(cfg.Include, ".harness/secrets.enc")
	p := newPolicy(cfg, true)
	if !p.Allow(".harness/secrets.enc") {
		t.Fatal("opt-in should allow secrets.enc")
	}
}
