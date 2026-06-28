package install

import (
	"runtime"
	"strings"
	"testing"
)

func TestLoadBundled_KnownTools(t *testing.T) {
	for _, name := range []string{"gopls", "ripgrep", "antigravity", "claude"} {
		m, err := LoadBundled(name)
		if err != nil {
			t.Fatalf("load %s: %v", name, err)
		}
		if m.Name != name {
			t.Errorf("manifest %s name=%q", name, m.Name)
		}
		if len(m.Strategies) == 0 {
			t.Errorf("manifest %s has no strategies", name)
		}
	}
}

func TestLoadBundled_Unknown(t *testing.T) {
	if _, err := LoadBundled("definitely-not-bundled"); err == nil {
		t.Fatal("expected error")
	}
}

func TestListBundled_NonEmpty(t *testing.T) {
	names, err := ListBundled()
	if err != nil {
		t.Fatal(err)
	}
	if len(names) < 5 {
		t.Fatalf("expected at least 5 bundled manifests, got %d", len(names))
	}
}

func TestPlatformMatch(t *testing.T) {
	all := PlatformMatch{}
	if !all.Matches("darwin", "amd64") {
		t.Fatal("empty match should accept all")
	}
	darwinOnly := PlatformMatch{OS: []string{"darwin"}}
	if darwinOnly.Matches("linux", "amd64") {
		t.Fatal("darwin-only matched linux")
	}
	if !darwinOnly.Matches("darwin", runtime.GOARCH) {
		t.Fatal("darwin-only rejected darwin")
	}
}

func TestRegistry_PicksFirstAvailable(t *testing.T) {
	r := NewRegistry()
	m := Manifest{
		Name: "fake",
		Strategies: []StrategyManifest{
			{Kind: "nonexistent_kind"},
			{Kind: "go_install", Args: map[string]string{"package": "example.com/x"}},
		},
	}
	if !strings.Contains(plan(t, r, m).String(), "go install") {
		t.Fatal("expected go install plan")
	}
}

func plan(t *testing.T, r *StrategyRegistry, m Manifest) Plan {
	t.Helper()
	p, err := r.Pick(m)
	if err != nil {
		t.Fatalf("pick: %v", err)
	}
	return p
}
