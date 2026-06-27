// SPDX-License-Identifier: MIT

package sensors

import (
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/ropeixoto/harnessx/internal/index"
)

func hasNodeModules(root string) bool {
	if root == "" {
		return false
	}
	st, err := os.Stat(filepath.Join(root, "node_modules"))
	return err == nil && st.IsDir()
}

func hasCargoLock(root string) bool {
	if root == "" {
		return false
	}
	_, err := os.Stat(filepath.Join(root, "Cargo.lock"))
	return err == nil
}

func hasBundleInstalled(root string) bool {
	if root == "" {
		return false
	}
	if _, err := os.Stat(filepath.Join(root, "Gemfile.lock")); err != nil {
		return false
	}
	return true
}

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
		if hasNodeModules(p.Root) {
			all = append(all,
				ShellSensor{IDValue: "node_eslint", CategoryV: CatLint, Binary: "npx", Args: []string{"--no-install", "eslint", "."}, Stacks: []string{"react", "nextjs", "vite", "node"}, OptionalTool: true},
				ShellSensor{IDValue: "node_prettier_check", CategoryV: CatFormat, Binary: "npx", Args: []string{"--no-install", "prettier", "--check", "."}, Stacks: []string{"react", "nextjs", "vite", "node"}, OptionalTool: true},
				ShellSensor{IDValue: "node_typecheck", CategoryV: CatTypecheck, Binary: "npx", Args: []string{"--no-install", "tsc", "--noEmit"}, Stacks: []string{"react", "nextjs", "vite", "node"}, OptionalTool: true},
				ShellSensor{IDValue: "node_test", CategoryV: CatTest, Binary: "npm", Args: []string{"test", "--", "--run"}, Stacks: []string{"react", "nextjs", "vite", "node"}, OptionalTool: true, Timeout: 10 * time.Minute},
			)
		}
	}

	if hasStack("rails") || hasStack("ruby") {
		if hasBundleInstalled(p.Root) {
			all = append(all,
				ShellSensor{IDValue: "ruby_rubocop", CategoryV: CatLint, Binary: "rubocop", Args: nil, Stacks: []string{"rails", "ruby"}, OptionalTool: true},
				ShellSensor{IDValue: "ruby_rspec", CategoryV: CatTest, Binary: "rspec", Args: nil, Stacks: []string{"rails", "ruby"}, OptionalTool: true, Timeout: 10 * time.Minute},
				ShellSensor{IDValue: "ruby_brakeman", CategoryV: CatSecurity, Binary: "brakeman", Args: []string{"-q"}, Stacks: []string{"rails"}, OptionalTool: true},
				ShellSensor{IDValue: "ruby_bundle_audit", CategoryV: CatDeps, Binary: "bundle-audit", Args: []string{"check", "--update"}, Stacks: []string{"rails", "ruby"}, OptionalTool: true},
			)
		}
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
		)
		if hasCargoLock(p.Root) {
			all = append(all,
				ShellSensor{IDValue: "rust_clippy", CategoryV: CatLint, Binary: "cargo", Args: []string{"clippy", "--all-targets", "--all-features", "--offline", "--", "-D", "warnings"}, Stacks: []string{"rust"}, OptionalTool: true},
				ShellSensor{IDValue: "rust_test", CategoryV: CatTest, Binary: "cargo", Args: []string{"test", "--all", "--offline"}, Stacks: []string{"rust"}, OptionalTool: true, Timeout: 15 * time.Minute},
				ShellSensor{IDValue: "rust_audit", CategoryV: CatDeps, Binary: "cargo-audit", Args: []string{"audit"}, Stacks: []string{"rust"}, OptionalTool: true},
			)
		}
	}

	if hasStack("docker") {
		all = append(all,
			ShellSensor{IDValue: "docker_hadolint", CategoryV: CatImage, Binary: "hadolint", Args: []string{"Dockerfile"}, Stacks: []string{"docker"}, OptionalTool: true},
		)
	}

	if hasStack("java") {
		all = append(all,
			ShellSensor{IDValue: "java_checkstyle", CategoryV: CatLint, Binary: "checkstyle", Args: []string{"-c", "/google_checks.xml", "src"}, Stacks: []string{"java"}, OptionalTool: true},
			ShellSensor{IDValue: "java_spotbugs", CategoryV: CatLint, Binary: "spotbugs", Args: []string{"-textui", "."}, Stacks: []string{"java"}, OptionalTool: true},
			ShellSensor{IDValue: "java_mvn_test", CategoryV: CatTest, Binary: "mvn", Args: []string{"-q", "test"}, Stacks: []string{"java"}, OptionalTool: true, Timeout: 15 * time.Minute},
			ShellSensor{IDValue: "java_gradle_test", CategoryV: CatTest, Binary: "gradle", Args: []string{"-q", "test"}, Stacks: []string{"java"}, OptionalTool: true, Timeout: 15 * time.Minute},
		)
	}

	if hasStack("kotlin") {
		all = append(all,
			ShellSensor{IDValue: "kotlin_ktlint", CategoryV: CatLint, Binary: "ktlint", Args: []string{"--reporter=plain", "src/**/*.kt"}, Stacks: []string{"kotlin"}, OptionalTool: true},
			ShellSensor{IDValue: "kotlin_detekt", CategoryV: CatLint, Binary: "detekt", Args: nil, Stacks: []string{"kotlin"}, OptionalTool: true},
			ShellSensor{IDValue: "kotlin_gradle_test", CategoryV: CatTest, Binary: "gradle", Args: []string{"-q", "test"}, Stacks: []string{"kotlin"}, OptionalTool: true, Timeout: 15 * time.Minute},
		)
	}

	if hasStack("swift") {
		all = append(all,
			ShellSensor{IDValue: "swift_format", CategoryV: CatFormat, Binary: "swift-format", Args: []string{"lint", "-r", "Sources"}, Stacks: []string{"swift"}, OptionalTool: true},
			ShellSensor{IDValue: "swift_lint", CategoryV: CatLint, Binary: "swiftlint", Args: []string{"--strict"}, Stacks: []string{"swift"}, OptionalTool: true},
			ShellSensor{IDValue: "swift_test", CategoryV: CatTest, Binary: "swift", Args: []string{"test"}, Stacks: []string{"swift"}, OptionalTool: true, Timeout: 15 * time.Minute},
		)
	}

	if hasStack("elixir") {
		all = append(all,
			ShellSensor{IDValue: "elixir_format", CategoryV: CatFormat, Binary: "mix", Args: []string{"format", "--check-formatted"}, Stacks: []string{"elixir"}, OptionalTool: true},
			ShellSensor{IDValue: "elixir_credo", CategoryV: CatLint, Binary: "mix", Args: []string{"credo", "--strict"}, Stacks: []string{"elixir"}, OptionalTool: true},
			ShellSensor{IDValue: "elixir_dialyzer", CategoryV: CatTypecheck, Binary: "mix", Args: []string{"dialyzer"}, Stacks: []string{"elixir"}, OptionalTool: true, Timeout: 15 * time.Minute},
			ShellSensor{IDValue: "elixir_test", CategoryV: CatTest, Binary: "mix", Args: []string{"test"}, Stacks: []string{"elixir"}, OptionalTool: true, Timeout: 10 * time.Minute},
			ShellSensor{IDValue: "elixir_sobelow", CategoryV: CatSecurity, Binary: "mix", Args: []string{"sobelow", "--exit"}, Stacks: []string{"elixir"}, OptionalTool: true},
		)
	}

	if hasStack("php") || hasStack("laravel") || hasStack("symfony") {
		all = append(all,
			ShellSensor{IDValue: "php_cs_fixer", CategoryV: CatFormat, Binary: "php-cs-fixer", Args: []string{"fix", "--dry-run", "--diff"}, Stacks: []string{"php", "laravel", "symfony"}, OptionalTool: true},
			ShellSensor{IDValue: "php_stan", CategoryV: CatTypecheck, Binary: "phpstan", Args: []string{"analyse", "--no-progress"}, Stacks: []string{"php", "laravel", "symfony"}, OptionalTool: true},
			ShellSensor{IDValue: "php_psalm", CategoryV: CatTypecheck, Binary: "psalm", Args: []string{"--no-progress"}, Stacks: []string{"php", "laravel", "symfony"}, OptionalTool: true},
			ShellSensor{IDValue: "php_unit", CategoryV: CatTest, Binary: "phpunit", Args: nil, Stacks: []string{"php", "laravel", "symfony"}, OptionalTool: true, Timeout: 10 * time.Minute},
			ShellSensor{IDValue: "php_pest", CategoryV: CatTest, Binary: "pest", Args: nil, Stacks: []string{"php", "laravel", "symfony"}, OptionalTool: true, Timeout: 10 * time.Minute},
		)
	}

	if hasStack("dotnet") {
		all = append(all,
			ShellSensor{IDValue: "dotnet_format", CategoryV: CatFormat, Binary: "dotnet", Args: []string{"format", "--verify-no-changes"}, Stacks: []string{"dotnet"}, OptionalTool: true},
			ShellSensor{IDValue: "dotnet_build", CategoryV: CatLint, Binary: "dotnet", Args: []string{"build", "-warnaserror"}, Stacks: []string{"dotnet"}, OptionalTool: true, Timeout: 10 * time.Minute},
			ShellSensor{IDValue: "dotnet_test", CategoryV: CatTest, Binary: "dotnet", Args: []string{"test", "--nologo", "--verbosity", "quiet"}, Stacks: []string{"dotnet"}, OptionalTool: true, Timeout: 15 * time.Minute},
		)
	}

	if hasStack("dart") || hasStack("flutter") {
		all = append(all,
			ShellSensor{IDValue: "dart_format", CategoryV: CatFormat, Binary: "dart", Args: []string{"format", "--set-exit-if-changed", "."}, Stacks: []string{"dart", "flutter"}, OptionalTool: true},
			ShellSensor{IDValue: "dart_analyze", CategoryV: CatLint, Binary: "dart", Args: []string{"analyze", "--fatal-infos"}, Stacks: []string{"dart", "flutter"}, OptionalTool: true},
			ShellSensor{IDValue: "dart_test", CategoryV: CatTest, Binary: "dart", Args: []string{"test"}, Stacks: []string{"dart"}, OptionalTool: true, Timeout: 10 * time.Minute},
			ShellSensor{IDValue: "flutter_test", CategoryV: CatTest, Binary: "flutter", Args: []string{"test"}, Stacks: []string{"flutter"}, OptionalTool: true, Timeout: 15 * time.Minute},
		)
	}

	out := make([]Sensor, 0, len(all))
	for _, s := range all {
		out = append(out, Wrap(s))
	}
	for _, s := range p.Stacks {
		out = append(out, smellSensorsFor(s.Name)...)
	}
	out = append(out, CommentStyleSensor{IDValue: "comment_style"})
	sort.Slice(out, func(i, j int) bool { return out[i].ID() < out[j].ID() })
	return out
}
