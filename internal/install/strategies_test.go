// SPDX-License-Identifier: MIT

package install

import (
	"strings"
	"testing"
)

func TestBrewStrategyPlan(t *testing.T) {
	p, err := BrewStrategy{}.Plan(Manifest{Name: "ripgrep"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !sliceHasFn(p.Command, "brew") || !sliceHasFn(p.Command, "install") || !sliceHasFn(p.Command, "ripgrep") {
		t.Errorf("brew plan: %v", p.Command)
	}
}

func TestBrewStrategyHonorsPackageOverride(t *testing.T) {
	p, _ := BrewStrategy{}.Plan(Manifest{Name: "x"}, map[string]string{"package": "real"})
	if !sliceHasFn(p.Command, "real") {
		t.Errorf("override ignored: %v", p.Command)
	}
}

func TestBrewStrategyHonorsVersion(t *testing.T) {
	p, _ := BrewStrategy{}.Plan(Manifest{Name: "go"}, map[string]string{"version": "1.21"})
	joined := strings.Join(p.Command, " ")
	if !strings.Contains(joined, "@1.21") {
		t.Errorf("version not pinned: %v", p.Command)
	}
}

func TestAptStrategyPlan(t *testing.T) {
	p, err := AptStrategy{}.Plan(Manifest{Name: "ripgrep"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !sliceHasFn(p.Command, "apt-get") {
		t.Errorf("apt plan: %v", p.Command)
	}
}

func TestDnfStrategyPlan(t *testing.T) {
	p, err := DnfStrategy{}.Plan(Manifest{Name: "ripgrep"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !sliceHasFn(p.Command, "dnf") {
		t.Errorf("dnf plan: %v", p.Command)
	}
}

func TestPacmanStrategyPlan(t *testing.T) {
	p, err := PacmanStrategy{}.Plan(Manifest{Name: "ripgrep"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !sliceHasFn(p.Command, "pacman") {
		t.Errorf("pacman plan: %v", p.Command)
	}
}

func TestGoInstallStrategyPlan(t *testing.T) {
	p, err := GoInstallStrategy{}.Plan(Manifest{Name: "gopls"}, map[string]string{"package": "golang.org/x/tools/gopls@latest"})
	if err != nil {
		t.Fatal(err)
	}
	if !sliceHasFn(p.Command, "go") || !sliceHasFn(p.Command, "install") {
		t.Errorf("go install plan: %v", p.Command)
	}
}

func TestNpmGlobalStrategyPlan(t *testing.T) {
	p, err := NpmGlobalStrategy{}.Plan(Manifest{Name: "@anthropic-ai/claude-code"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !sliceHasFn(p.Command, "npm") {
		t.Errorf("npm plan: %v", p.Command)
	}
}

func TestCargoInstallStrategyPlan(t *testing.T) {
	p, err := CargoInstallStrategy{}.Plan(Manifest{Name: "ripgrep"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !sliceHasFn(p.Command, "cargo") {
		t.Errorf("cargo plan: %v", p.Command)
	}
}

func TestPipUserStrategyPlan(t *testing.T) {
	p, err := PipUserStrategy{}.Plan(Manifest{Name: "ruff"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !sliceHasFn(p.Command, "pip3") && !sliceHasFn(p.Command, "pip") {
		t.Errorf("pip plan: %v", p.Command)
	}
}

func TestPlanString(t *testing.T) {
	p := Plan{Kind: "brew", Command: []string{"brew", "install", "x"}}
	if !strings.Contains(p.String(), "brew") {
		t.Errorf("string: %s", p.String())
	}
	empty := Plan{Description: "noop"}
	if empty.String() != "noop" {
		t.Errorf("empty cmd should use description: %s", empty.String())
	}
}

func sliceHasFn(haystack []string, needle string) bool {
	for _, h := range haystack {
		if h == needle {
			return true
		}
	}
	return false
}
