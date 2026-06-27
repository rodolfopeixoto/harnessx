package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func withHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	return home
}

func TestConfigSourcesSetGetReset(t *testing.T) {
	home := withHome(t)
	set := newConfigSourcesCmd()
	set.SetArgs([]string{"set", "update_repo", "myfork/harnessx"})
	var buf bytes.Buffer
	set.SetOut(&buf)
	set.SetErr(&buf)
	if err := set.Execute(); err != nil {
		t.Fatalf("set: %v", err)
	}
	body, err := os.ReadFile(filepath.Join(home, ".config", "harness", "sources.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), "update_repo: myfork/harnessx") {
		t.Fatalf("file content: %s", string(body))
	}

	get := newConfigSourcesCmd()
	get.SetArgs([]string{"get", "update_repo"})
	var out bytes.Buffer
	get.SetOut(&out)
	get.SetErr(&out)
	if err := get.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "myfork/harnessx") {
		t.Fatalf("get output: %s", out.String())
	}

	reset := newConfigSourcesCmd()
	reset.SetArgs([]string{"reset"})
	if err := reset.Execute(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(home, ".config", "harness", "sources.yaml")); !os.IsNotExist(err) {
		t.Fatalf("sources.yaml should be deleted, stat err=%v", err)
	}
}

func TestConfigSourcesRejectsUnknownKey(t *testing.T) {
	withHome(t)
	cmd := newConfigSourcesCmd()
	cmd.SetArgs([]string{"set", "nope", "value"})
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	if err := cmd.Execute(); err == nil {
		t.Fatalf("expected error for unknown key, got nil. out=%s", buf.String())
	}
}

func TestResolveRepoReadsSourcesFile(t *testing.T) {
	home := withHome(t)
	dir := filepath.Join(home, ".config", "harness")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "sources.yaml"), []byte("update_repo: myfork/harnessx\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got := resolveRepo()
	if got != "myfork/harnessx" {
		t.Fatalf("want myfork/harnessx, got %s", got)
	}
}

func TestResolveRepoEnvOverridesFile(t *testing.T) {
	home := withHome(t)
	dir := filepath.Join(home, ".config", "harness")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "sources.yaml"), []byte("update_repo: ignored/from-file\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HARNESS_UPDATE_REPO", "wins/from-env")
	if got := resolveRepo(); got != "wins/from-env" {
		t.Fatalf("env should win: got %s", got)
	}
}

func TestResolveRepoDefaultsWhenUnset(t *testing.T) {
	withHome(t)
	t.Setenv("HARNESS_UPDATE_REPO", "")
	if got := resolveRepo(); got != defaultRepo {
		t.Fatalf("default: want %s, got %s", defaultRepo, got)
	}
}
