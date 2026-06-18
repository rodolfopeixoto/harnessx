package activeagent

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadAbsentReturnsEmpty(t *testing.T) {
	p, err := Load(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if p.AgentID != "" {
		t.Errorf("want empty, got %q", p.AgentID)
	}
}

func TestSaveLoadRoundtrip(t *testing.T) {
	dir := t.TempDir()
	if err := Save(dir, Pin{AgentID: "claude", Model: "sonnet"}); err != nil {
		t.Fatal(err)
	}
	got, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if got.AgentID != "claude" || got.Model != "sonnet" {
		t.Errorf("roundtrip mismatch: %+v", got)
	}
}

func TestSaveRequiresAgentID(t *testing.T) {
	if err := Save(t.TempDir(), Pin{}); err == nil {
		t.Fatal("want error")
	}
}

func TestClearRemovesFile(t *testing.T) {
	dir := t.TempDir()
	_ = Save(dir, Pin{AgentID: "claude"})
	if err := Clear(dir); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, relPath)); !os.IsNotExist(err) {
		t.Errorf("file should be gone")
	}
}

func TestClearAbsentIsNoOp(t *testing.T) {
	if err := Clear(t.TempDir()); err != nil {
		t.Errorf("clear on absent file should not error: %v", err)
	}
}

func TestResolveAgentIDPrefersOverride(t *testing.T) {
	dir := t.TempDir()
	_ = Save(dir, Pin{AgentID: "kimi"})
	got := ResolveAgentID(dir, "claude")
	if got != "claude" {
		t.Errorf("override should win: %q", got)
	}
}

func TestResolveAgentIDFallsBackToPin(t *testing.T) {
	dir := t.TempDir()
	_ = Save(dir, Pin{AgentID: "kimi"})
	got := ResolveAgentID(dir, "")
	if got != "kimi" {
		t.Errorf("want kimi, got %q", got)
	}
}

func TestResolveAgentIDEmptyWhenNothing(t *testing.T) {
	if got := ResolveAgentID(t.TempDir(), ""); got != "" {
		t.Errorf("want empty, got %q", got)
	}
}
