// SPDX-License-Identifier: MIT

package backup

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadConfigMissingReturnsDefault(t *testing.T) {
	got, err := LoadConfig(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	def := DefaultConfig()
	if got.Compression != def.Compression {
		t.Errorf("default not returned: %+v", got)
	}
}

func TestSaveAndLoadConfigRoundTrip(t *testing.T) {
	root := t.TempDir()
	c := DefaultConfig()
	c.DefaultRemote = "s3-prod"
	if err := SaveConfig(root, c); err != nil {
		t.Fatal(err)
	}
	got, err := LoadConfig(root)
	if err != nil {
		t.Fatal(err)
	}
	if got.DefaultRemote != "s3-prod" {
		t.Errorf("round-trip mismatch: %+v", got)
	}
}

func TestSaveConfigCreatesConfigDir(t *testing.T) {
	root := t.TempDir()
	if err := SaveConfig(root, DefaultConfig()); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, ".harness", "config", "backup.yaml")); err != nil {
		t.Errorf("config file missing: %v", err)
	}
}

func TestPackNameWithoutTag(t *testing.T) {
	got := PackName("")
	if !strings.HasPrefix(got, "harness-backup-") || !strings.HasSuffix(got, ".tar.gz") {
		t.Errorf("got %q", got)
	}
}

func TestPackNameWithTag(t *testing.T) {
	got := PackName("nightly")
	if !strings.Contains(got, "nightly") {
		t.Errorf("missing tag: %q", got)
	}
}

func TestPackNameSanitisesTag(t *testing.T) {
	got := PackName("v1.0 alpha+rc/1")
	for _, bad := range []string{" ", "+", "/"} {
		if strings.Contains(got, bad) {
			t.Errorf("unsanitised char %q in %q", bad, got)
		}
	}
}

func TestSafeTagPassesAllowedChars(t *testing.T) {
	cases := map[string]string{
		"abc-123":     "abc-123",
		"abc.123":     "abc-123",
		"x y z":       "x-y-z",
		"x/y":         "x-y",
		"under_score": "under_score",
	}
	for in, want := range cases {
		if got := safeTag(in); got != want {
			t.Errorf("safeTag(%q)=%q, want %q", in, got, want)
		}
	}
}

func TestSafeTagEmptyReturnsPlaceholder(t *testing.T) {
	if got := safeTag(""); got != "untagged" {
		t.Errorf("safeTag(empty)=%q", got)
	}
	if got := safeTag("@@@"); got != "---" {
		t.Errorf("safeTag(unsafe-only)=%q", got)
	}
}
