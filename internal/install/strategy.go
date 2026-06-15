// SPDX-License-Identifier: MIT

package install

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
)

type Strategy interface {
	Kind() string
	Plan(m Manifest, args map[string]string) (Plan, error)
}

func strategyAvailable(kind string) bool {
	bin := strategyBinary(kind)
	if bin == "" {
		return true
	}
	_, err := exec.LookPath(bin)
	return err == nil
}

type Plan struct {
	Kind        string
	Command     []string
	Env         []string
	Description string
}

func (p Plan) String() string {
	if len(p.Command) == 0 {
		return p.Description
	}
	return fmt.Sprintf("%s: %v", p.Kind, p.Command)
}

type StrategyRegistry struct {
	strategies map[string]Strategy
}

func NewRegistry() *StrategyRegistry {
	r := &StrategyRegistry{strategies: map[string]Strategy{}}
	r.Register(&BrewStrategy{})
	r.Register(&AptStrategy{})
	r.Register(&DnfStrategy{})
	r.Register(&PacmanStrategy{})
	r.Register(&GoInstallStrategy{})
	r.Register(&NpmGlobalStrategy{})
	r.Register(&CargoInstallStrategy{})
	r.Register(&PipUserStrategy{})
	return r
}

func (r *StrategyRegistry) Register(s Strategy) { r.strategies[s.Kind()] = s }

func (r *StrategyRegistry) Get(kind string) (Strategy, bool) {
	s, ok := r.strategies[kind]
	return s, ok
}

func (r *StrategyRegistry) Pick(m Manifest) (Plan, error) {
	for _, sm := range m.Strategies {
		if !sm.Platform.Matches(runtime.GOOS, runtime.GOARCH) {
			continue
		}
		s, ok := r.Get(sm.Kind)
		if !ok {
			continue
		}
		if !strategyAvailable(sm.Kind) {
			continue
		}
		return s.Plan(m, sm.Args)
	}
	return Plan{}, fmt.Errorf("install: no viable strategy for %s on %s/%s", m.Name, runtime.GOOS, runtime.GOARCH)
}

func strategyBinary(kind string) string {
	switch kind {
	case "brew":
		return "brew"
	case "apt":
		return "apt-get"
	case "dnf":
		return "dnf"
	case "pacman":
		return "pacman"
	case "go_install":
		return "go"
	case "npm_global":
		return "npm"
	case "cargo_install":
		return "cargo"
	case "pip_user":
		return "pip3"
	}
	return ""
}

func Execute(ctx context.Context, p Plan, dryRun bool, stdout, stderr Writer) error {
	if dryRun {
		fmt.Fprintf(stdout, "→ dry-run: %s\n", p.String())
		return nil
	}
	if len(p.Command) == 0 {
		return fmt.Errorf("install: empty plan command")
	}
	fmt.Fprintf(stdout, "→ %s\n", p.String())
	cmd := exec.CommandContext(ctx, p.Command[0], p.Command[1:]...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	if len(p.Env) > 0 {
		cmd.Env = p.Env
	}
	return cmd.Run()
}

type Writer interface {
	Write(p []byte) (int, error)
}

type BrewStrategy struct{}

func (BrewStrategy) Kind() string { return ("brew") }
func (BrewStrategy) Plan(m Manifest, args map[string]string) (Plan, error) {
	pkg := args["package"]
	if pkg == "" {
		pkg = m.Name
	}
	flags := []string{"install"}
	if v := args["version"]; v != "" {
		pkg = pkg + "@" + v
	}
	return Plan{Kind: "brew", Command: append([]string{"brew"}, append(flags, pkg)...)}, nil
}

type AptStrategy struct{}

func (AptStrategy) Kind() string { return ("apt") }
func (AptStrategy) Plan(m Manifest, args map[string]string) (Plan, error) {
	pkg := args["package"]
	if pkg == "" {
		pkg = m.Name
	}
	return Plan{Kind: "apt", Command: []string{"sudo", "apt-get", "install", "-y", pkg}}, nil
}

type DnfStrategy struct{}

func (DnfStrategy) Kind() string { return ("dnf") }
func (DnfStrategy) Plan(m Manifest, args map[string]string) (Plan, error) {
	pkg := args["package"]
	if pkg == "" {
		pkg = m.Name
	}
	return Plan{Kind: "dnf", Command: []string{"sudo", "dnf", "install", "-y", pkg}}, nil
}

type PacmanStrategy struct{}

func (PacmanStrategy) Kind() string { return ("pacman") }
func (PacmanStrategy) Plan(m Manifest, args map[string]string) (Plan, error) {
	pkg := args["package"]
	if pkg == "" {
		pkg = m.Name
	}
	return Plan{Kind: "pacman", Command: []string{"sudo", "pacman", "-S", "--noconfirm", pkg}}, nil
}

type GoInstallStrategy struct{}

func (GoInstallStrategy) Kind() string { return ("go_install") }
func (GoInstallStrategy) Plan(m Manifest, args map[string]string) (Plan, error) {
	pkg := args["package"]
	if pkg == "" {
		return Plan{}, fmt.Errorf("install: go_install needs args.package")
	}
	ver := args["version"]
	if ver == "" {
		ver = "latest"
	}
	return Plan{Kind: "go_install", Command: []string{"go", "install", pkg + "@" + ver}}, nil
}

type NpmGlobalStrategy struct{}

func (NpmGlobalStrategy) Kind() string { return ("npm_global") }
func (NpmGlobalStrategy) Plan(m Manifest, args map[string]string) (Plan, error) {
	pkg := args["package"]
	if pkg == "" {
		pkg = m.Name
	}
	if v := args["version"]; v != "" {
		pkg = pkg + "@" + v
	}
	return Plan{Kind: "npm_global", Command: []string{"npm", "install", "-g", pkg}}, nil
}

type CargoInstallStrategy struct{}

func (CargoInstallStrategy) Kind() string { return ("cargo_install") }
func (CargoInstallStrategy) Plan(m Manifest, args map[string]string) (Plan, error) {
	pkg := args["package"]
	if pkg == "" {
		pkg = m.Name
	}
	return Plan{Kind: "cargo_install", Command: []string{"cargo", "install", pkg}}, nil
}

type PipUserStrategy struct{}

func (PipUserStrategy) Kind() string { return ("pip_user") }
func (PipUserStrategy) Plan(m Manifest, args map[string]string) (Plan, error) {
	pkg := args["package"]
	if pkg == "" {
		pkg = m.Name
	}
	return Plan{Kind: "pip_user", Command: []string{"pip3", "install", "--user", pkg}}, nil
}
