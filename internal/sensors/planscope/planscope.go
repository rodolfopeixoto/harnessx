// SPDX-License-Identifier: MIT

package planscope

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ropeixoto/harnessx/internal/plancontract"
)

type DiffSource func(ctx context.Context, root string) ([]string, error)

type Violation struct {
	Path   string
	Reason string
}

type Result struct {
	PlanID     string
	Violations []Violation
}

func (r Result) Pass() bool { return len(r.Violations) == 0 }

type Options struct {
	Root       string
	PlanID     string
	DiffSource DiffSource
}

func Check(ctx context.Context, opts Options) (Result, error) {
	if opts.PlanID == "" {
		return Result{}, errors.New("planscope: missing plan id")
	}
	if opts.Root == "" {
		return Result{}, errors.New("planscope: missing root")
	}
	contract, err := plancontract.Load(opts.Root, opts.PlanID)
	if err != nil {
		return Result{}, err
	}
	source := opts.DiffSource
	if source == nil {
		source = GitChangedFiles
	}
	files, err := source(ctx, opts.Root)
	if err != nil {
		return Result{}, fmt.Errorf("planscope: enumerate diff: %w", err)
	}
	res := Result{PlanID: contract.ID}
	for _, f := range files {
		if !contract.InScope(f) {
			res.Violations = append(res.Violations, Violation{
				Path:   f,
				Reason: "outside PLAN file scope",
			})
		}
	}
	return res, nil
}

func GitChangedFiles(ctx context.Context, root string) ([]string, error) {
	cmd := exec.CommandContext(ctx, "git", "status", "--porcelain")
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	var paths []string
	for _, line := range bytes.Split(out, []byte("\n")) {
		s := strings.TrimSpace(string(line))
		if s == "" {
			continue
		}
		parts := strings.Fields(s)
		if len(parts) < 2 {
			continue
		}
		paths = append(paths, filepath.ToSlash(parts[len(parts)-1]))
	}
	return paths, nil
}

func FormatResult(r Result) string {
	var b strings.Builder
	if r.Pass() {
		fmt.Fprintf(&b, "planscope: %s — all changed files in scope\n", r.PlanID)
		return b.String()
	}
	fmt.Fprintf(&b, "planscope: %s — %d violation(s)\n", r.PlanID, len(r.Violations))
	for _, v := range r.Violations {
		fmt.Fprintf(&b, "  ✗ %s — %s\n", v.Path, v.Reason)
	}
	return b.String()
}
