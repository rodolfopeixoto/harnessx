// SPDX-License-Identifier: MIT

package update

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPlatformTargetIncludesOSAndArch(t *testing.T) {
	got := PlatformTarget()
	if !strings.HasPrefix(got, "harness-") {
		t.Errorf("missing prefix: %q", got)
	}
	parts := strings.Split(got, "-")
	if len(parts) < 3 {
		t.Errorf("want harness-OS-ARCH, got %q", got)
	}
}

func TestTarballURLShape(t *testing.T) {
	got := TarballURL("owner/repo", "v1.2.3", "harness-darwin-arm64")
	want := "https://github.com/owner/repo/releases/download/v1.2.3/harness-darwin-arm64.tar.gz"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestVerifySha256MatchAndMismatch(t *testing.T) {
	dir := t.TempDir()
	tarPath := filepath.Join(dir, "x.tar.gz")
	content := []byte("hello world")
	if err := os.WriteFile(tarPath, content, 0o644); err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(content)
	hexSum := hex.EncodeToString(sum[:])
	shaPath := filepath.Join(dir, "x.tar.gz.sha256")
	if err := os.WriteFile(shaPath, []byte(hexSum+"  x.tar.gz\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := VerifySha256(tarPath, shaPath); err != nil {
		t.Errorf("matching sum should pass: %v", err)
	}
	if err := os.WriteFile(shaPath, []byte("deadbeef  x.tar.gz\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := VerifySha256(tarPath, shaPath); err == nil {
		t.Error("mismatched sum should fail")
	}
}

func TestVerifySha256MissingFiles(t *testing.T) {
	if err := VerifySha256("/tmp/_no_tar_", "/tmp/_no_sha_"); err == nil {
		t.Error("missing files should error")
	}
}

func TestVerifySha256EmptyChecksum(t *testing.T) {
	dir := t.TempDir()
	tarPath := filepath.Join(dir, "x.tar.gz")
	shaPath := filepath.Join(dir, "x.tar.gz.sha256")
	_ = os.WriteFile(tarPath, []byte("body"), 0o644)
	_ = os.WriteFile(shaPath, []byte("   \n"), 0o644)
	if err := VerifySha256(tarPath, shaPath); err == nil {
		t.Error("empty checksum should error")
	}
}

func TestDownloadFileBadURL(t *testing.T) {
	dir := t.TempDir()
	dest := filepath.Join(dir, "out")
	if err := DownloadFile("http://127.0.0.1:1/nonexistent", dest); err == nil {
		t.Error("expected error for invalid url")
	}
}
