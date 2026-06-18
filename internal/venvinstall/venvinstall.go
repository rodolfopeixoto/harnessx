// SPDX-License-Identifier: MIT

package venvinstall

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"strings"
)

type Strategy struct {
	Name string
	Cmds [][]string
}

type Result struct {
	Strategy string
	Steps    []StepOutcome
}

type StepOutcome struct {
	Cmd    []string
	Stdout string
	Stderr string
	Err    error
}

func (r Result) OK() bool {
	for _, s := range r.Steps {
		if s.Err != nil {
			return false
		}
	}
	return r.Strategy != ""
}

func DetectInterpreter() (string, []string) {
	for _, c := range []string{"python3.13", "python3.12", "python3.11", "python3"} {
		if _, err := exec.LookPath(c); err == nil {
			return c, nil
		}
	}
	return "", nil
}

func Install(ctx context.Context, root, requirements string, out io.Writer) (Result, error) {
	if _, err := exec.LookPath("uv"); err == nil {
		return runStrategy(ctx, root, out, Strategy{
			Name: "uv",
			Cmds: [][]string{
				{"uv", "venv", ".venv"},
				{"uv", "pip", "install", "-r", requirements, "--python", filepath.Join(".venv", "bin", "python")},
			},
		})
	}

	py, _ := DetectInterpreter()
	if py == "" {
		return Result{}, errors.New("venvinstall: no python3.11+/python3 on PATH (install python from python.org)")
	}

	strategies := []Strategy{
		{
			Name: py + "+ensurepip",
			Cmds: [][]string{
				{py, "-m", "venv", ".venv"},
				{filepath.Join(".venv", "bin", "pip"), "install", "-r", requirements},
			},
		},
		{
			Name: py + "+without-pip+get-pip",
			Cmds: [][]string{
				{py, "-m", "venv", "--without-pip", ".venv"},
				{filepath.Join(".venv", "bin", "python"), "-c", bootstrapPipScript},
				{filepath.Join(".venv", "bin", "pip"), "install", "-r", requirements},
			},
		},
		{
			Name: py + "+system-pip",
			Cmds: [][]string{
				{py, "-m", "venv", "--without-pip", ".venv"},
				{py, "-m", "pip", "install", "--prefix", ".venv", "-r", requirements},
			},
		},
	}

	var last Result
	for _, s := range strategies {
		fmt.Fprintf(out, "venv: trying strategy %s\n", s.Name)
		res, err := runStrategy(ctx, root, out, s)
		if err == nil {
			return res, nil
		}
		last = res
		fmt.Fprintf(out, "venv: strategy %s failed (%v); falling through\n", s.Name, err)
		cleanVenv(root)
	}
	return last, fmt.Errorf("venvinstall: every strategy exhausted; check host python install")
}

const bootstrapPipScript = `
import os, sys, urllib.request, tempfile, subprocess
src = "https://bootstrap.pypa.io/get-pip.py"
with tempfile.NamedTemporaryFile(suffix=".py", delete=False) as f:
    urllib.request.urlretrieve(src, f.name)
    subprocess.check_call([sys.executable, f.name])
`

func runStrategy(ctx context.Context, root string, out io.Writer, s Strategy) (Result, error) {
	res := Result{Strategy: s.Name}
	for _, c := range s.Cmds {
		fmt.Fprintf(out, "  → %s\n", strings.Join(c, " "))
		cmd := exec.CommandContext(ctx, c[0], c[1:]...)
		cmd.Dir = root
		stdoutBuf := &strings.Builder{}
		stderrBuf := &strings.Builder{}
		cmd.Stdout = io.MultiWriter(out, stdoutBuf)
		cmd.Stderr = io.MultiWriter(out, stderrBuf)
		err := cmd.Run()
		res.Steps = append(res.Steps, StepOutcome{
			Cmd: c, Stdout: stdoutBuf.String(), Stderr: stderrBuf.String(), Err: err,
		})
		if err != nil {
			return res, err
		}
	}
	return res, nil
}

func cleanVenv(root string) {
	_ = exec.Command("rm", "-rf", filepath.Join(root, ".venv")).Run()
}
