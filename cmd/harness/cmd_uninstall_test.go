// SPDX-License-Identifier: MIT

package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestUninstallCmdHasSubcommands(t *testing.T) {
	c := newUninstallCmd()
	want := map[string]bool{"project": false, "global": false, "all": false}
	for _, sub := range c.Commands() {
		want[sub.Use] = true
	}
	for k, ok := range want {
		if !ok {
			t.Errorf("subcommand missing: %s", k)
		}
	}
}

func TestSizeHintForDirAndFile(t *testing.T) {
	dir := t.TempDir()
	info, _ := os.Stat(dir)
	if sizeHint(info) != "directory" {
		t.Errorf("dir hint: got %q", sizeHint(info))
	}
	path := filepath.Join(dir, "f")
	_ = os.WriteFile(path, []byte("12345"), 0o644)
	info2, _ := os.Stat(path)
	if sizeHint(info2) != "5 bytes" {
		t.Errorf("file hint: got %q", sizeHint(info2))
	}
}

func TestWipePathMissing(t *testing.T) {
	var buf bytes.Buffer
	if err := wipePath(&buf, "/tmp/__nope_does_not_exist_zzz__", true); err != nil {
		t.Fatal(err)
	}
	if buf.Len() == 0 {
		t.Error("expected absent notice")
	}
}

func TestWipePathYes(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "doomed")
	_ = os.MkdirAll(target, 0o755)
	var buf bytes.Buffer
	if err := wipePath(&buf, target, true); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(target); !os.IsNotExist(err) {
		t.Error("target should be gone")
	}
}
