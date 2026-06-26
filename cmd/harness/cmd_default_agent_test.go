// SPDX-License-Identifier: MIT

package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveDefaultAgentExplicitWins(t *testing.T) {
	t.Setenv("HARNESS_DEFAULT_AGENT", "kimi")
	got, err := resolveDefaultAgent("codex", t.TempDir(), &bytes.Buffer{}, strings.NewReader(""))
	if err != nil {
		t.Fatal(err)
	}
	if got != "codex" {
		t.Errorf("explicit must win, got %q", got)
	}
}

func TestResolveDefaultAgentReadsActiveYAML(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".harness", "config"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".harness", "config", "active.yaml"), []byte("agent_id: gemini\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := resolveDefaultAgent("", root, &bytes.Buffer{}, strings.NewReader(""))
	if err != nil {
		t.Fatal(err)
	}
	if got != "gemini" {
		t.Errorf("want gemini from active.yaml, got %q", got)
	}
}

func TestResolveDefaultAgentReadsEnvWhenNoPin(t *testing.T) {
	t.Setenv("HARNESS_DEFAULT_AGENT", "ollama")
	got, err := resolveDefaultAgent("", t.TempDir(), &bytes.Buffer{}, strings.NewReader(""))
	if err != nil {
		t.Fatal(err)
	}
	if got != "ollama" {
		t.Errorf("want ollama from env, got %q", got)
	}
}

func TestResolveDefaultAgentNonTTYNoPinErrors(t *testing.T) {
	t.Setenv("HARNESS_DEFAULT_AGENT", "")
	_, err := resolveDefaultAgent("", t.TempDir(), &bytes.Buffer{}, strings.NewReader(""))
	if err == nil {
		t.Error("no pin + non-TTY must hard-error")
	}
	if !strings.Contains(err.Error(), "harness use") {
		t.Errorf("error must mention `harness use`, got %v", err)
	}
}

func TestIsTerminalReaderRejectsStringsReader(t *testing.T) {
	if isTerminalReader(strings.NewReader("")) {
		t.Error("strings.Reader is not a TTY")
	}
}
