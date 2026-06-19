// SPDX-License-Identifier: MIT

package prompttpl

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidName(t *testing.T) {
	for _, v := range []string{"ok", "with-dash", "with_underscore", "a1"} {
		if !ValidName(v) {
			t.Errorf("%q should be valid", v)
		}
	}
	for _, v := range []string{"", "Upper", "with space", "x/y", "-leading-dash", "_under_first", "way-too-long-prompt-name-that-exceeds-the-forty-char-limit"} {
		if ValidName(v) {
			t.Errorf("%q should be invalid", v)
		}
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	if err := Save(dir, "hello", "world"); err != nil {
		t.Fatal(err)
	}
	body, err := Load(dir, "hello")
	if err != nil {
		t.Fatal(err)
	}
	if body != "world" {
		t.Errorf("body mismatch: %q", body)
	}
	if _, err := os.Stat(filepath.Join(dir, ".harness", "prompts", "hello.md")); err != nil {
		t.Errorf("file missing: %v", err)
	}
}

func TestSaveRejectsBadName(t *testing.T) {
	if err := Save(t.TempDir(), "Bad Name", "x"); err == nil {
		t.Fatal("want error for bad name")
	}
}

func TestSaveRejectsEmptyBody(t *testing.T) {
	if err := Save(t.TempDir(), "ok", "   "); err == nil {
		t.Fatal("want error for empty body")
	}
}

func TestListEmpty(t *testing.T) {
	got, err := List(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Errorf("want empty list, got %v", got)
	}
}

func TestListSorted(t *testing.T) {
	dir := t.TempDir()
	for _, n := range []string{"zeta", "alpha", "mu"} {
		if err := Save(dir, n, "body"); err != nil {
			t.Fatal(err)
		}
	}
	got, _ := List(dir)
	want := []string{"alpha", "mu", "zeta"}
	for i, w := range want {
		if got[i] != w {
			t.Errorf("position %d: want %s got %s", i, w, got[i])
		}
	}
}
