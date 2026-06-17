// SPDX-License-Identifier: MIT

package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstallPrePushHook_FreshRepo(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	path, err := InstallPrePushHook(dir, false)
	if err != nil {
		t.Fatalf("install: %v", err)
	}
	want := filepath.Join(dir, ".git", "hooks", "pre-push")
	if path != want {
		t.Fatalf("path: want %s got %s", want, path)
	}
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(body, []byte("harness ci")) {
		t.Fatal("hook missing 'harness ci'")
	}
	if !bytes.Contains(body, []byte(prePushHookMarker)) {
		t.Fatal("hook missing managed marker")
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != hookFileMode {
		t.Fatalf("perm: want %o got %o", hookFileMode, info.Mode().Perm())
	}
}

func TestInstallPrePushHook_NotGitRepo(t *testing.T) {
	dir := t.TempDir()
	if _, err := InstallPrePushHook(dir, false); err == nil {
		t.Fatal("expected error for missing .git")
	}
}

func TestInstallPrePushHook_Idempotent(t *testing.T) {
	dir := t.TempDir()
	_ = os.Mkdir(filepath.Join(dir, ".git"), 0o755)
	if _, err := InstallPrePushHook(dir, false); err != nil {
		t.Fatal(err)
	}
	if _, err := InstallPrePushHook(dir, false); err != nil {
		t.Fatalf("second install must succeed: %v", err)
	}
}

func TestInstallPrePushHook_RefusesForeignHook(t *testing.T) {
	dir := t.TempDir()
	hooks := filepath.Join(dir, ".git", "hooks")
	if err := os.MkdirAll(hooks, 0o755); err != nil {
		t.Fatal(err)
	}
	foreign := []byte("#!/bin/sh\necho custom hook\n")
	if err := os.WriteFile(filepath.Join(hooks, "pre-push"), foreign, 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := InstallPrePushHook(dir, false); err == nil {
		t.Fatal("expected refusal on foreign hook")
	} else if !strings.Contains(err.Error(), "--force") {
		t.Fatalf("error must mention --force: %v", err)
	}
	if _, err := InstallPrePushHook(dir, true); err != nil {
		t.Fatalf("force install failed: %v", err)
	}
	body, _ := os.ReadFile(filepath.Join(hooks, "pre-push"))
	if !bytes.Contains(body, []byte(prePushHookMarker)) {
		t.Fatal("force did not replace foreign hook")
	}
}

func TestInstallGitHooksCmd_HasForceFlag(t *testing.T) {
	c := newInstallGitHooksCmd()
	if c.Flags().Lookup("force") == nil {
		t.Fatal("--force flag missing")
	}
}

func TestIsHarnessManagedHook(t *testing.T) {
	if !isHarnessManagedHook([]byte("# Managed by HarnessX\n")) {
		t.Fatal("should detect marker")
	}
	if isHarnessManagedHook([]byte("#!/bin/sh\necho hi\n")) {
		t.Fatal("false positive")
	}
}
