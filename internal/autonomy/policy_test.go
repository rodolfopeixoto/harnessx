// SPDX-License-Identifier: MIT

package autonomy

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMatchPath_DenyWins(t *testing.T) {
	p := Policy{
		AllowPaths: []string{"src/**"},
		DenyPaths:  []string{"src/secrets/**"},
	}
	if got := p.MatchPath("src/secrets/api.key"); got != DecisionDeny {
		t.Fatalf("want deny, got %v", got)
	}
	if got := p.MatchPath("src/main.go"); got != DecisionAllow {
		t.Fatalf("want allow, got %v", got)
	}
}

func TestMatchPath_NoAllowListAllowsByDefault(t *testing.T) {
	p := Policy{DenyPaths: []string{"secrets/**"}}
	if got := p.MatchPath("README.md"); got != DecisionAllow {
		t.Fatalf("want allow, got %v", got)
	}
	if got := p.MatchPath("secrets/api.key"); got != DecisionDeny {
		t.Fatalf("want deny, got %v", got)
	}
}

func TestMatchPath_WithAllowListNonMatchNeedsApproval(t *testing.T) {
	p := Policy{AllowPaths: []string{"docs/**"}}
	if got := p.MatchPath("docs/intro.md"); got != DecisionAllow {
		t.Fatalf("want allow, got %v", got)
	}
	if got := p.MatchPath("src/main.go"); got != DecisionApproval {
		t.Fatalf("want approval, got %v", got)
	}
}

func TestMatchPath_BasenameGlob(t *testing.T) {
	p := Policy{DenyPaths: []string{".env*"}}
	if got := p.MatchPath("config/.env.prod"); got != DecisionDeny {
		t.Fatalf("want deny, got %v", got)
	}
}

func TestMatchCommand_PrefixMatch(t *testing.T) {
	p := Policy{
		AllowCommands: []string{"go test", "npm test"},
		DenyCommands:  []string{"rm -rf /", "git push --force"},
	}
	if got := p.MatchCommand("go test ./..."); got != DecisionAllow {
		t.Fatalf("want allow, got %v", got)
	}
	if got := p.MatchCommand("git push --force --tags"); got != DecisionDeny {
		t.Fatalf("want deny, got %v", got)
	}
	if got := p.MatchCommand("python build.py"); got != DecisionApproval {
		t.Fatalf("want approval, got %v", got)
	}
}

func TestLoadPolicy_MissingFileNoError(t *testing.T) {
	dir := t.TempDir()
	p, err := LoadPolicy(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(p.DenyPaths) != 0 {
		t.Fatalf("expected empty policy, got %+v", p)
	}
}

func TestLoadPolicy_ParsesYAML(t *testing.T) {
	dir := t.TempDir()
	cfgDir := filepath.Join(dir, ".harness", "config")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	body := `level: safe_execute
allow_paths:
  - "src/**"
deny_paths:
  - "secrets/**"
allow_commands:
  - "go test"
deny_commands:
  - "git push --force"
`
	if err := os.WriteFile(filepath.Join(cfgDir, "autonomy.yaml"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	p, err := LoadPolicy(dir)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if p.Level != "safe_execute" {
		t.Fatalf("level: %s", p.Level)
	}
	if len(p.AllowPaths) != 1 || p.AllowPaths[0] != "src/**" {
		t.Fatalf("allow_paths: %+v", p.AllowPaths)
	}
}
