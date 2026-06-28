package fixenvcmd

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

type Check struct {
	ID       string
	Title    string
	Severity string
	Detail   string
	FixHint  string
	FixCmd   string
}

type Options struct {
	Root        string
	Apply       bool
	Stack       string
	IncludeAll  bool
	UserEnviron func() []string
	Lookup      func(string) (string, error)
}

type Result struct {
	Checks  []Check
	Applied []string
}

func Run(out io.Writer, opts Options) (Result, error) {
	if opts.UserEnviron == nil {
		opts.UserEnviron = os.Environ
	}
	if opts.Lookup == nil {
		opts.Lookup = exec.LookPath
	}
	var res Result
	res.Checks = append(res.Checks, checkGOROOT(opts)...)
	res.Checks = append(res.Checks, checkPATH(opts)...)
	res.Checks = append(res.Checks, checkVenv(opts)...)
	res.Checks = append(res.Checks, checkNodeModules(opts)...)
	res.Checks = append(res.Checks, checkGitConfig(opts)...)
	res.Checks = append(res.Checks, checkHomebrewPATH(opts)...)
	res.Checks = append(res.Checks, checkRubyBundle(opts)...)
	for _, c := range res.Checks {
		fmt.Fprintf(out, "[%s] %s — %s\n", c.Severity, c.Title, c.Detail)
		if c.FixHint != "" {
			fmt.Fprintf(out, "  fix: %s\n", c.FixHint)
		}
		if opts.Apply && c.FixCmd != "" {
			applied, err := applyFix(c)
			if err != nil {
				fmt.Fprintf(out, "  ✗ apply failed: %v\n", err)
				continue
			}
			fmt.Fprintf(out, "  ✓ applied: %s\n", applied)
			res.Applied = append(res.Applied, c.ID)
		}
	}
	if len(res.Checks) == 0 {
		fmt.Fprintln(out, "✓ env looks healthy")
	}
	return res, nil
}

func envValue(opts Options, key string) string {
	prefix := key + "="
	for _, e := range opts.UserEnviron() {
		if strings.HasPrefix(e, prefix) {
			return strings.TrimPrefix(e, prefix)
		}
	}
	return ""
}

func checkGOROOT(opts Options) []Check {
	v := envValue(opts, "GOROOT")
	if v == "" {
		return nil
	}
	if _, err := os.Stat(v); err == nil {
		return nil
	}
	return []Check{{
		ID:       "goroot_missing",
		Title:    "GOROOT points at a missing directory",
		Severity: "high",
		Detail:   fmt.Sprintf("GOROOT=%s does not exist; `go` commands will fail with `cannot find GOROOT directory`", v),
		FixHint:  "unset GOROOT in your shell rc; modern Go installs do not need it. Run `unset GOROOT` for this shell.",
		FixCmd:   "unset:GOROOT",
	}}
}

func checkPATH(opts Options) []Check {
	path := envValue(opts, "PATH")
	if path == "" {
		return nil
	}
	wanted := []string{"/usr/local/bin", "/opt/homebrew/bin"}
	if runtime.GOOS == "linux" {
		wanted = []string{"/usr/local/bin", "/home/linuxbrew/.linuxbrew/bin"}
	}
	missing := []string{}
	parts := strings.Split(path, string(os.PathListSeparator))
	have := map[string]bool{}
	for _, p := range parts {
		have[p] = true
	}
	for _, w := range wanted {
		if _, err := os.Stat(w); err != nil {
			continue
		}
		if !have[w] {
			missing = append(missing, w)
		}
	}
	if len(missing) == 0 {
		return nil
	}
	return []Check{{
		ID:       "path_missing_standard_dirs",
		Title:    "PATH missing standard install dirs",
		Severity: "medium",
		Detail:   fmt.Sprintf("PATH does not include %s", strings.Join(missing, ", ")),
		FixHint:  fmt.Sprintf("add `export PATH=\"%s:$PATH\"` to your shell rc", strings.Join(missing, ":")),
	}}
}

func checkVenv(opts Options) []Check {
	if !hasFile(opts.Root, "requirements.txt") && !hasFile(opts.Root, "pyproject.toml") {
		return nil
	}
	venv := filepath.Join(opts.Root, ".venv")
	if st, err := os.Stat(venv); err == nil && st.IsDir() {
		return nil
	}
	return []Check{{
		ID:       "missing_venv",
		Title:    "Python project without .venv",
		Severity: "low",
		Detail:   "no `.venv` directory; sensors that expect project-local pytest/ruff will skip with binary-not-found",
		FixHint:  fmt.Sprintf("cd %s && python3 -m venv .venv && .venv/bin/pip install --upgrade pip && .venv/bin/pip install -r requirements.txt", opts.Root),
		FixCmd:   "exec:python3 -m venv " + venv + " && " + filepath.Join(venv, "bin/pip") + " install --quiet --upgrade pip && " + filepath.Join(venv, "bin/pip") + " install --quiet -r " + filepath.Join(opts.Root, "requirements.txt"),
	}}
}

func checkNodeModules(opts Options) []Check {
	if !hasFile(opts.Root, "package.json") {
		return nil
	}
	if st, err := os.Stat(filepath.Join(opts.Root, "node_modules")); err == nil && st.IsDir() {
		return nil
	}
	return []Check{{
		ID:       "missing_node_modules",
		Title:    "node project without node_modules",
		Severity: "low",
		Detail:   "no node_modules; node sensors (eslint, prettier, vitest) will skip",
		FixHint:  fmt.Sprintf("cd %s && npm install", opts.Root),
		FixCmd:   "exec:npm install --silent --no-fund --no-audit --prefix " + opts.Root,
	}}
}

func checkGitConfig(opts Options) []Check {
	if !hasFile(opts.Root, ".git") {
		return nil
	}
	user, _ := runCapture(opts.Root, "git", "config", "--get", "user.email")
	if strings.TrimSpace(user) != "" {
		return nil
	}
	return []Check{{
		ID:       "git_user_email_missing",
		Title:    "git user.email not set",
		Severity: "medium",
		Detail:   "harness will leave commits unattributed; harness drive uses `git commit`",
		FixHint:  "set with: git config --global user.email \"<you@example.com>\"",
	}}
}

func checkHomebrewPATH(opts Options) []Check {
	if runtime.GOOS != "darwin" {
		return nil
	}
	if _, err := opts.Lookup("brew"); err == nil {
		return nil
	}
	if _, err := os.Stat("/opt/homebrew/bin/brew"); err == nil {
		return []Check{{
			ID:       "homebrew_in_path",
			Title:    "brew installed but not on PATH",
			Severity: "medium",
			Detail:   "/opt/homebrew/bin/brew exists but `brew` not resolvable; `harness install` strategies that use brew will fail",
			FixHint:  "run: eval \"$(/opt/homebrew/bin/brew shellenv)\"",
		}}
	}
	return nil
}

func checkRubyBundle(opts Options) []Check {
	if !hasFile(opts.Root, "Gemfile") {
		return nil
	}
	if hasFile(opts.Root, "Gemfile.lock") {
		return nil
	}
	return []Check{{
		ID:       "missing_gemfile_lock",
		Title:    "ruby project without Gemfile.lock",
		Severity: "low",
		Detail:   "no Gemfile.lock; ruby sensors (rubocop, rspec, brakeman) will skip",
		FixHint:  fmt.Sprintf("cd %s && bundle install", opts.Root),
		FixCmd:   "exec:bundle install --quiet --gemfile=" + filepath.Join(opts.Root, "Gemfile"),
	}}
}

func hasFile(root, name string) bool {
	if root == "" {
		return false
	}
	_, err := os.Stat(filepath.Join(root, name))
	return err == nil
}

func runCapture(dir, name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	return string(out), err
}

func applyFix(c Check) (string, error) {
	switch {
	case strings.HasPrefix(c.FixCmd, "unset:"):
		key := strings.TrimPrefix(c.FixCmd, "unset:")
		_ = os.Unsetenv(key)
		return "unset " + key + " for this process (add `unset " + key + "` to your shell rc to persist)", nil
	case strings.HasPrefix(c.FixCmd, "exec:"):
		script := strings.TrimPrefix(c.FixCmd, "exec:")
		cmd := exec.Command("/bin/sh", "-c", script) //nolint:gosec // G204: user-explicit --apply path; script built from bundled FixCmd
		out, err := cmd.CombinedOutput()
		if err != nil {
			return "", fmt.Errorf("%w: %s", err, strings.TrimSpace(string(out)))
		}
		return script, nil
	}
	return "", fmt.Errorf("unsupported fix kind: %s", c.FixCmd)
}
