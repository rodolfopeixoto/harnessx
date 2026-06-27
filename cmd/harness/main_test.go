package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestRootRejectsUnknownPositional(t *testing.T) {
	root := newRoot()
	root.SetArgs([]string{"add a /orders endpoint"})
	var stdout, stderr bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&stderr)
	err := root.Execute()
	if err == nil {
		t.Fatalf("expected error for unknown positional, got nil. stdout=%q", stdout.String())
	}
	combined := strings.ToLower(err.Error() + stderr.String())
	if !strings.Contains(combined, "unknown") && !strings.Contains(combined, "unknown command") && !strings.Contains(combined, "accepts 0 arg") {
		t.Fatalf("error should mention unknown/no-args, got: %v / stderr=%s", err, stderr.String())
	}
	if strings.Contains(stdout.String()+stderr.String(), "Detected intent: feature") {
		t.Fatalf("must not fall through to feature mode: %s", stdout.String()+stderr.String())
	}
}

func TestRootBareShowsHelp(t *testing.T) {
	root := newRoot()
	root.SetArgs([]string{})
	var stdout bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&stdout)
	if err := root.Execute(); err != nil {
		t.Fatalf("bare root should not error: %v", err)
	}
	if !strings.Contains(stdout.String(), "HarnessX") && !strings.Contains(stdout.String(), "Usage:") {
		t.Fatalf("expected help output, got: %s", stdout.String())
	}
}

func TestRootKnownSubcommandWorks(t *testing.T) {
	root := newRoot()
	root.SetArgs([]string{"version"})
	var stdout bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&stdout)
	if err := root.Execute(); err != nil {
		t.Fatalf("version subcommand failed: %v", err)
	}
}
