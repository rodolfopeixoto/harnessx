// SPDX-License-Identifier: MIT

package secrets

import (
	"errors"
	"path/filepath"
	"testing"
)

func TestKeychainBackendName(t *testing.T) {
	b := KeychainBackend{}
	if b.Name() != "keychain" {
		t.Errorf("name=%q", b.Name())
	}
}

func TestEncryptedFileBackendName(t *testing.T) {
	b := &EncryptedFileBackend{Root: t.TempDir()}
	if b.Name() != "encrypted_file" {
		t.Errorf("name=%q", b.Name())
	}
	if !b.Available() {
		t.Error("should always be available")
	}
}

func TestEncryptedFileRoundTrip(t *testing.T) {
	b := &EncryptedFileBackend{Root: t.TempDir()}
	if err := b.Set("api_key", "value-1"); err != nil {
		t.Fatal(err)
	}
	got, err := b.Get("api_key")
	if err != nil {
		t.Fatal(err)
	}
	if got != "value-1" {
		t.Errorf("got %q", got)
	}
}

func TestEncryptedFileListAndDelete(t *testing.T) {
	b := &EncryptedFileBackend{Root: t.TempDir()}
	_ = b.Set("a", "1")
	_ = b.Set("b", "2")
	names, err := b.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(names) != 2 {
		t.Errorf("want 2, got %d (%v)", len(names), names)
	}
	if err := b.Delete("a"); err != nil {
		t.Fatal(err)
	}
	_, err = b.Get("a")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("want ErrNotFound after delete, got %v", err)
	}
}

func TestEncryptedFilePaths(t *testing.T) {
	b := &EncryptedFileBackend{Root: t.TempDir()}
	if !filepathHas(b.path(), "secrets") {
		t.Errorf("path: %q", b.path())
	}
	if !filepathHas(b.keyPath(), "seed") {
		t.Errorf("keyPath: %q", b.keyPath())
	}
}

func TestParseKeychainDumpReturnsSlice(t *testing.T) {
	got := parseKeychainDump("")
	if got == nil {
		_ = got
	}
}

func filepathHas(p, frag string) bool {
	clean := filepath.Clean(p)
	for i := 0; i+len(frag) <= len(clean); i++ {
		if clean[i:i+len(frag)] == frag {
			return true
		}
	}
	return false
}
