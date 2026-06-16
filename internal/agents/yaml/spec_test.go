// SPDX-License-Identifier: MIT

package yaml

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeSpec(t *testing.T, body string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "spec.yaml")
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestLoadValidCLISpec(t *testing.T) {
	p := writeSpec(t, `
id: claude-test
name: Claude Test
type: cli
command:
  binary: claude
  check: claude --version
strengths: [code, refactor]
`)
	s, err := Load(p)
	if err != nil {
		t.Fatal(err)
	}
	if s.ID != "claude-test" {
		t.Errorf("ID: got %q", s.ID)
	}
	if s.Command.Binary != "claude" {
		t.Errorf("binary: got %q", s.Command.Binary)
	}
	if len(s.Capabilities.Strengths) != 2 {
		t.Errorf("strengths not propagated to capabilities: got %v", s.Capabilities.Strengths)
	}
}

func TestLoadMissingIDFails(t *testing.T) {
	p := writeSpec(t, `
name: X
type: cli
command:
  binary: x
`)
	_, err := Load(p)
	if err == nil || !strings.Contains(err.Error(), "missing id") {
		t.Errorf("expected missing id error, got %v", err)
	}
}

func TestLoadMissingNameFails(t *testing.T) {
	p := writeSpec(t, `
id: x
type: cli
command:
  binary: x
`)
	_, err := Load(p)
	if err == nil || !strings.Contains(err.Error(), "missing name") {
		t.Errorf("expected missing name, got %v", err)
	}
}

func TestLoadUnknownTypeFails(t *testing.T) {
	p := writeSpec(t, `
id: x
name: X
type: bogus
`)
	_, err := Load(p)
	if err == nil || !strings.Contains(err.Error(), "unknown type") {
		t.Errorf("expected unknown type, got %v", err)
	}
}

func TestLoadCLIMissingBinaryFails(t *testing.T) {
	p := writeSpec(t, `
id: x
name: X
type: cli
`)
	_, err := Load(p)
	if err == nil || !strings.Contains(err.Error(), "missing command.binary") {
		t.Errorf("expected missing binary, got %v", err)
	}
}

func TestLoadAPIMissingEndpointFails(t *testing.T) {
	p := writeSpec(t, `
id: x
name: X
type: api
`)
	_, err := Load(p)
	if err == nil || !strings.Contains(err.Error(), "missing api.endpoint") {
		t.Errorf("expected missing endpoint, got %v", err)
	}
}

func TestLoadInteractiveMissingBinaryFails(t *testing.T) {
	p := writeSpec(t, `
id: x
name: X
type: interactive
interactive: {}
`)
	_, err := Load(p)
	if err == nil || !strings.Contains(err.Error(), "missing interactive.binary") {
		t.Errorf("expected missing interactive binary, got %v", err)
	}
}

func TestLoadInteractiveUnknownStrategyFails(t *testing.T) {
	p := writeSpec(t, `
id: x
name: X
type: interactive
interactive:
  strategy: telegrams
  binary: claude
`)
	_, err := Load(p)
	if err == nil || !strings.Contains(err.Error(), "unknown strategy") {
		t.Errorf("expected unknown strategy, got %v", err)
	}
}

func TestLoadInteractiveAcceptsKnownStrategies(t *testing.T) {
	for _, strategy := range []string{"", "pty", "tmux", "iterm2"} {
		p := writeSpec(t, `
id: x
name: X
type: interactive
interactive:
  strategyegy: `+strategy+`
  binary: claude
`)
		if _, err := Load(p); err != nil {
			t.Errorf("strategyegy %q: %v", strategy, err)
		}
	}
}

func TestLoadCopiesAuthIntoCapabilities(t *testing.T) {
	p := writeSpec(t, `
id: x
name: X
type: cli
command:
  binary: x
auth:
  login_command: x login
  doc_url: https://example.com/docs
`)
	s, err := Load(p)
	if err != nil {
		t.Fatal(err)
	}
	if s.Capabilities.LoginCommand != "x login" {
		t.Errorf("LoginCommand: got %q", s.Capabilities.LoginCommand)
	}
	if s.Capabilities.AuthDocURL != "https://example.com/docs" {
		t.Errorf("AuthDocURL: got %q", s.Capabilities.AuthDocURL)
	}
}

func TestLoadMissingFile(t *testing.T) {
	_, err := Load("/tmp/__definitely_not_a_real_spec__.yaml")
	if err == nil {
		t.Error("expected error for missing file")
	}
}
