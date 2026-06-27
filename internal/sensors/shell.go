// SPDX-License-Identifier: MIT

package sensors

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// ShellSensor wraps a shell command. When OptionalTool is true and the
// binary is missing from PATH, Run returns StatusSkipped rather than
// StatusFailed — this keeps optional rule-pack tools from making CI red.
type ShellSensor struct {
	IDValue      string
	CategoryV    Category
	KindV        Kind
	Binary       string
	Args         []string
	Stacks       []string // empty = applies whenever Binary is present
	OptionalTool bool
	Timeout      time.Duration
	Lookup       func(string) (string, error) // injectable for tests
}

func (s ShellSensor) ID() string         { return s.IDValue }
func (s ShellSensor) Category() Category { return s.CategoryV }
func (s ShellSensor) Kind() Kind {
	if s.KindV != "" {
		return s.KindV
	}
	return KindComputational
}

func (s ShellSensor) AppliesTo(p stackHaver) bool {
	if len(s.Stacks) == 0 {
		return true
	}
	have := map[string]bool{}
	for _, st := range p.stackList() {
		have[st] = true
	}
	for _, want := range s.Stacks {
		if have[want] {
			return true
		}
	}
	return false
}

// stackHaver lets tests pass a tiny shim instead of a full index.Profile.
type stackHaver interface{ stackList() []string }

// profileShim adapts an index.Profile to stackHaver via an internal helper
// (kept private to avoid leaking the index dependency outside this package).
type profileShim struct{ names []string }

func (p profileShim) stackList() []string { return p.names }

func (s ShellSensor) Run(rc RunCtx) Result {
	start := time.Now()
	res := Result{ID: s.IDValue, Category: s.CategoryV, Kind: s.Kind()}
	lookup := s.Lookup
	if lookup == nil {
		lookup = func(name string) (string, error) {
			if p := projectLocalBinary(rc.Root, name); p != "" {
				return p, nil
			}
			return exec.LookPath(name)
		}
	}
	resolved, err := lookup(s.Binary)
	if err != nil {
		res.Status = ifElse(s.OptionalTool, StatusSkipped, StatusFailed)
		res.Detail = "binary not on PATH: " + s.Binary
		res.Duration = time.Since(start)
		return res
	}

	timeout := s.Timeout
	if timeout <= 0 {
		timeout = 5 * time.Minute
	}
	ctx, cancel := context.WithTimeout(rc.Ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, resolved, s.Args...)
	cmd.Dir = rc.Root
	cmd.Env = projectLocalEnv(rc.Root)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	res.Duration = time.Since(start)
	res.OutputPath = writeOutput(rc.OutputDir, s.IDValue, stdout.Bytes(), stderr.Bytes())

	exitErr := new(exec.ExitError)
	switch {
	case err == nil:
		res.Status = StatusPassed
	case errors.As(err, &exitErr):
		res.ExitCode = exitErr.ExitCode()
		res.Status = StatusFailed
		hint := firstNonEmptyLine(stderr.Bytes())
		if hint == "" {
			hint = firstNonEmptyLine(stdout.Bytes())
		}
		if hint != "" {
			res.Detail = fmt.Sprintf("exit %d — %s", exitErr.ExitCode(), hint)
		} else {
			res.Detail = fmt.Sprintf("exit %d", exitErr.ExitCode())
		}
	default:
		res.Status = StatusFailed
		res.Detail = err.Error()
	}
	return res
}

func firstNonEmptyLine(b []byte) string {
	const max = 200
	start := 0
	for i := 0; i < len(b); i++ {
		if b[i] == '\n' {
			line := bytes.TrimSpace(b[start:i])
			if len(line) > 0 {
				if len(line) > max {
					line = line[:max]
				}
				return string(line)
			}
			start = i + 1
		}
	}
	tail := bytes.TrimSpace(b[start:])
	if len(tail) > max {
		tail = tail[:max]
	}
	return string(tail)
}

// projectLocalBinary returns an absolute path to a project-scoped binary
// (`.venv/bin/<name>`, `venv/bin/<name>`, `node_modules/.bin/<name>`)
// when one exists and is executable. Lets python/node CI run without the
// user having to `source .venv/bin/activate` before every `harness ci`.
func projectLocalBinary(root, name string) string {
	if root == "" || name == "" {
		return ""
	}
	candidates := []string{
		filepath.Join(root, ".venv", "bin", name),
		filepath.Join(root, "venv", "bin", name),
		filepath.Join(root, "node_modules", ".bin", name),
	}
	for _, p := range candidates {
		if st, err := os.Stat(p); err == nil && !st.IsDir() && st.Mode()&0o111 != 0 {
			return p
		}
	}
	return ""
}

// projectLocalEnv prepends project-local bin dirs to PATH so children of
// the wrapped binary (e.g. ruff invoking python via a shebang) also
// resolve against the project venv before falling back to system tools.
func projectLocalEnv(root string) []string {
	if root == "" {
		return os.Environ()
	}
	env := os.Environ()
	prepend := []string{
		filepath.Join(root, ".venv", "bin"),
		filepath.Join(root, "venv", "bin"),
		filepath.Join(root, "node_modules", ".bin"),
	}
	var keep []string
	for _, p := range prepend {
		if st, err := os.Stat(p); err == nil && st.IsDir() {
			keep = append(keep, p)
		}
	}
	if len(keep) == 0 {
		return env
	}
	for i, e := range env {
		if len(e) >= 5 && e[:5] == "PATH=" {
			env[i] = "PATH=" + joinWithColons(append(keep, e[5:])...)
			return env
		}
	}
	return append(env, "PATH="+joinWithColons(append(keep, os.Getenv("PATH"))...))
}

func joinWithColons(parts ...string) string {
	out := ""
	for i, p := range parts {
		if i > 0 {
			out += ":"
		}
		out += p
	}
	return out
}

func writeOutput(dir, id string, stdout, stderr []byte) string {
	if dir == "" {
		return ""
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return ""
	}
	p := filepath.Join(dir, id+".log")
	var buf bytes.Buffer
	buf.WriteString("--- stdout ---\n")
	buf.Write(stdout)
	if len(stdout) > 0 && stdout[len(stdout)-1] != '\n' {
		buf.WriteByte('\n')
	}
	buf.WriteString("--- stderr ---\n")
	buf.Write(stderr)
	if len(stderr) > 0 && stderr[len(stderr)-1] != '\n' {
		buf.WriteByte('\n')
	}
	if err := os.WriteFile(p, buf.Bytes(), 0o644); err != nil {
		return ""
	}
	return p
}

func ifElse[T any](cond bool, a, b T) T {
	if cond {
		return a
	}
	return b
}
