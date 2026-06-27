// SPDX-License-Identifier: MIT

package sensors

import (
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/ropeixoto/harnessx/internal/index"
)

func pyPipAuditArgs(root string) []string {
	_ = root
	return []string{"-l"}
}

func pyBanditArgs(root string) []string {
	args := []string{"-r", "--exclude", pyBanditExcludesCSV}
	for _, candidate := range []string{"bandit.yaml", "bandit.yml"} {
		path := filepath.Join(root, candidate)
		if st, err := os.Stat(path); err == nil && !st.IsDir() {
			args = append(args, "-c", path)
			break
		}
	}
	args = append(args, ".")
	return args
}

const pyExcludesCSV = ".venv,venv,.git,__pycache__,.pytest_cache,.mypy_cache,.ruff_cache,node_modules,dist,build"
const pyExcludesRE = `(^|/)\.?venv/|(^|/)__pycache__/|(^|/)\.git/|(^|/)\.pytest_cache/|(^|/)\.mypy_cache/|(^|/)\.ruff_cache/|(^|/)node_modules/|(^|/)dist/|(^|/)build/`
const pyBanditExcludesCSV = pyExcludesCSV + ",tests"

// Catalog returns every sensor known to HarnessX, filtered to those that
// apply to the given project profile. Universal sensors always apply;
// stack-specific shell sensors require the matching stack to be detected.
// Skip-if-tool-missing behaviour lives in ShellSensor itself.
func Catalog(p index.Profile) []Sensor {
	out := []Sensor{
		ForbiddenFilesSensor{},
		ForbiddenCommandsSensor{},
		SecretsScanSensor{},
		ChangedFilesSensor{},
		PerformanceBudgetSensor{},
	}
	out = append(out, stackSensors(p)...)
	for _, s := range p.Stacks {
		if s.Name == "go" {
			out = append(out, goCoverageSensorDefault())
			break
		}
	}
	if plan, _ := LoadActivePlanID(p.Root); plan != "" {
		out = append(out, PlanScopeSensor{IDValue: "plan_scope", PlanID: plan})
	}
	return out
}

// stackSensors emits the per-stack rule pack. Sensors target a single tool
// each. OptionalTool=true so missing binaries skip rather than fail.
func stackSensors(p index.Profile) []Sensor {
	var all []ShellSensor

	hasStack := func(name string) bool {
		for _, s := range p.Stacks {
			if s.Name == name {
				return true
			}
		}
		return false
	}

	if hasStack("go") {
		all = append(all,
			ShellSensor{IDValue: "go_format", CategoryV: CatFormat, Binary: "gofmt", Args: []string{"-l", "."}, Stacks: []string{"go"}},
			ShellSensor{IDValue: "go_vet", CategoryV: CatLint, Binary: "go", Args: []string{"vet", "./..."}, Stacks: []string{"go"}},
			ShellSensor{IDValue: "go_test", CategoryV: CatTest, Binary: "go", Args: []string{"test", "./..."}, Stacks: []string{"go"}, Timeout: 10 * time.Minute},
			ShellSensor{IDValue: "go_staticcheck", CategoryV: CatLint, Binary: "staticcheck", Args: []string{"./..."}, Stacks: []string{"go"}, OptionalTool: true},
			ShellSensor{IDValue: "go_golangci_lint", CategoryV: CatLint, Binary: "golangci-lint", Args: []string{"run"}, Stacks: []string{"go"}, OptionalTool: true},
			ShellSensor{IDValue: "go_vuln", CategoryV: CatSecurity, Binary: "govulncheck", Args: []string{"./..."}, Stacks: []string{"go"}, OptionalTool: true},
		)
	}

	if hasStack("react") || hasStack("nextjs") || hasStack("vite") || hasStack("node") {
		all = append(all,
			ShellSensor{IDValue: "node_eslint", CategoryV: CatLint, Binary: "npx", Args: []string{"--no-install", "eslint", "."}, Stacks: []string{"react", "nextjs", "vite", "node"}, OptionalTool: true},
			ShellSensor{IDValue: "node_prettier_check", CategoryV: CatFormat, Binary: "npx", Args: []string{"--no-install", "prettier", "--check", "."}, Stacks: []string{"react", "nextjs", "vite", "node"}, OptionalTool: true},
			ShellSensor{IDValue: "node_typecheck", CategoryV: CatTypecheck, Binary: "npx", Args: []string{"--no-install", "tsc", "--noEmit"}, Stacks: []string{"react", "nextjs", "vite", "node"}, OptionalTool: true},
			ShellSensor{IDValue: "node_test", CategoryV: CatTest, Binary: "npm", Args: []string{"test", "--", "--run"}, Stacks: []string{"react", "nextjs", "vite", "node"}, OptionalTool: true, Timeout: 10 * time.Minute},
		)
	}

	if hasStack("rails") || hasStack("ruby") {
		all = append(all,
			ShellSensor{IDValue: "ruby_rubocop", CategoryV: CatLint, Binary: "rubocop", Args: nil, Stacks: []string{"rails", "ruby"}, OptionalTool: true},
			ShellSensor{IDValue: "ruby_rspec", CategoryV: CatTest, Binary: "rspec", Args: nil, Stacks: []string{"rails", "ruby"}, OptionalTool: true, Timeout: 10 * time.Minute},
			ShellSensor{IDValue: "ruby_brakeman", CategoryV: CatSecurity, Binary: "brakeman", Args: []string{"-q"}, Stacks: []string{"rails"}, OptionalTool: true},
			ShellSensor{IDValue: "ruby_bundle_audit", CategoryV: CatDeps, Binary: "bundle-audit", Args: []string{"check", "--update"}, Stacks: []string{"rails", "ruby"}, OptionalTool: true},
		)
	}

	if hasStack("python") {
		all = append(all,
			ShellSensor{IDValue: "py_ruff", CategoryV: CatLint, Binary: "ruff", Args: []string{"check", "--exclude", pyExcludesCSV, "."}, Stacks: []string{"python"}, OptionalTool: true},
			ShellSensor{IDValue: "py_ruff_format", CategoryV: CatFormat, Binary: "ruff", Args: []string{"format", "--check", "--exclude", pyExcludesCSV, "."}, Stacks: []string{"python"}, OptionalTool: true},
			ShellSensor{IDValue: "py_mypy", CategoryV: CatTypecheck, Binary: "mypy", Args: []string{"--exclude", pyExcludesRE, "."}, Stacks: []string{"python"}, OptionalTool: true},
			ShellSensor{IDValue: "py_pytest", CategoryV: CatTest, Binary: "pytest", Args: nil, Stacks: []string{"python"}, OptionalTool: true, Timeout: 10 * time.Minute},
			ShellSensor{IDValue: "py_bandit", CategoryV: CatSecurity, Binary: "bandit", Args: pyBanditArgs(p.Root), Stacks: []string{"python"}, OptionalTool: true},
			ShellSensor{IDValue: "py_pip_audit", CategoryV: CatDeps, Binary: "pip-audit", Args: pyPipAuditArgs(p.Root), Stacks: []string{"python"}, OptionalTool: true},
		)
	}

	if hasStack("rust") {
		all = append(all,
			ShellSensor{IDValue: "rust_fmt", CategoryV: CatFormat, Binary: "cargo", Args: []string{"fmt", "--check"}, Stacks: []string{"rust"}, OptionalTool: true},
			ShellSensor{IDValue: "rust_clippy", CategoryV: CatLint, Binary: "cargo", Args: []string{"clippy", "--all-targets", "--all-features", "--", "-D", "warnings"}, Stacks: []string{"rust"}, OptionalTool: true},
			ShellSensor{IDValue: "rust_test", CategoryV: CatTest, Binary: "cargo", Args: []string{"test", "--all"}, Stacks: []string{"rust"}, OptionalTool: true, Timeout: 15 * time.Minute},
			ShellSensor{IDValue: "rust_audit", CategoryV: CatDeps, Binary: "cargo", Args: []string{"audit"}, Stacks: []string{"rust"}, OptionalTool: true},
		)
	}

	if hasStack("docker") {
		all = append(all,
			ShellSensor{IDValue: "docker_hadolint", CategoryV: CatImage, Binary: "hadolint", Args: []string{"Dockerfile"}, Stacks: []string{"docker"}, OptionalTool: true},
		)
	}

	out := make([]Sensor, 0, len(all))
	for _, s := range all {
		out = append(out, Wrap(s))
	}
	// Stable order: by ID. The runner re-sorts to put computational first;
	// within each kind, alphabetical keeps output deterministic.
	sort.Slice(out, func(i, j int) bool { return out[i].ID() < out[j].ID() })
	return out
}
