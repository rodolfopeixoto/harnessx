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
		lookup = exec.LookPath
	}
	if _, err := lookup(s.Binary); err != nil {
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

	cmd := exec.CommandContext(ctx, s.Binary, s.Args...)
	cmd.Dir = rc.Root
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	res.Duration = time.Since(start)
	res.OutputPath = writeOutput(rc.OutputDir, s.IDValue, stdout.Bytes(), stderr.Bytes())

	exitErr := new(exec.ExitError)
	switch {
	case err == nil:
		res.Status = StatusPassed
	case errors.As(err, &exitErr):
		res.ExitCode = exitErr.ExitCode()
		res.Status = StatusFailed
		res.Detail = fmt.Sprintf("exit %d", exitErr.ExitCode())
	default:
		res.Status = StatusFailed
		res.Detail = err.Error()
	}
	return res
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
