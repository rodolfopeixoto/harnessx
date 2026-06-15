package secrets

import (
	"path/filepath"
	"testing"
)

func TestEnvBackend(t *testing.T) {
	t.Setenv("HARNESS_SECRET_TEST_KEY", "hello")
	v, err := EnvBackend{}.Get("test_key")
	if err != nil {
		t.Fatal(err)
	}
	if v != "hello" {
		t.Fatalf("expected hello, got %q", v)
	}
}

func TestEncryptedFile_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	b := &EncryptedFileBackend{Root: dir}
	if err := b.Set("token", "supersecret"); err != nil {
		t.Fatal(err)
	}
	got, err := b.Get("token")
	if err != nil {
		t.Fatal(err)
	}
	if got != "supersecret" {
		t.Fatalf("got %q", got)
	}
	if err := b.Delete("token"); err != nil {
		t.Fatal(err)
	}
	if _, err := b.Get("token"); err != ErrNotFound {
		t.Fatalf("expected ErrNotFound after delete, got %v", err)
	}
	if _, err := filepath.Abs(b.path()); err != nil {
		t.Fatal(err)
	}
}

func TestStore_PrefersFirstBackendWithValue(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HARNESS_SECRET_FOO", "from-env")
	enc := &EncryptedFileBackend{Root: dir}
	_ = enc.Set("foo", "from-file")
	s := NewWith(EnvBackend{}, enc)
	v, err := s.Get("foo")
	if err != nil {
		t.Fatal(err)
	}
	if v != "from-env" {
		t.Fatalf("expected env to win, got %q", v)
	}
}

func TestStore_Resolve(t *testing.T) {
	t.Setenv("HARNESS_SECRET_BAR", "envtoken")
	s := New()
	v, err := s.Resolve("secret://bar")
	if err != nil {
		t.Fatal(err)
	}
	if v != "envtoken" {
		t.Fatalf("got %q", v)
	}
	v, err = s.Resolve("plain-value")
	if err != nil {
		t.Fatal(err)
	}
	if v != "plain-value" {
		t.Fatalf("plain pass-through failed: %q", v)
	}
}
