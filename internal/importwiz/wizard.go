// SPDX-License-Identifier: MIT

package importwiz

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ropeixoto/harnessx/internal/platform/constants"
	"github.com/ropeixoto/harnessx/internal/stale"
	"github.com/ropeixoto/harnessx/internal/workspace"
)

type Step struct {
	ID     string
	Title  string
	Status string
	Detail string
}

const (
	StepDetect   = "detect"
	StepStack    = "stack"
	StepRegister = "register"
	StepIndex    = "index"
	StepDone     = "done"
)

const (
	StatusPending = "pending"
	StatusOK      = "ok"
	StatusFailed  = "failed"
)

type Options struct {
	Path        string
	DisplayName string
	Slug        string
	Confirm     bool
}

type Result struct {
	Project workspace.Project
	Steps   []Step
	Stack   []string
}

func Plan(opts Options) []Step {
	return []Step{
		{ID: StepDetect, Title: "Verify folder exists", Status: StatusPending},
		{ID: StepStack, Title: "Detect stack markers", Status: StatusPending},
		{ID: StepRegister, Title: "Register in workspace registry", Status: StatusPending, Detail: opts.Path},
		{ID: StepIndex, Title: "Capture stale fingerprints", Status: StatusPending},
		{ID: StepDone, Title: "Done", Status: StatusPending},
	}
}

func Run(ctx context.Context, registry *workspace.Registry, opts Options) (Result, error) {
	if opts.Path == "" {
		return Result{}, errors.New("importwiz: empty path")
	}
	abs, err := filepath.Abs(opts.Path)
	if err != nil {
		return Result{}, err
	}
	steps := Plan(opts)

	if _, err := os.Stat(abs); err != nil {
		steps[0].Status = StatusFailed
		return Result{Steps: steps}, fmt.Errorf("importwiz: folder %q not accessible: %w", abs, err)
	}
	steps[0].Status = StatusOK
	steps[0].Detail = abs

	stack := DetectStack(abs)
	steps[1].Status = StatusOK
	steps[1].Detail = stackDescription(stack)

	project, err := registry.Add(ctx, abs, opts.DisplayName, opts.Slug)
	if err != nil {
		steps[2].Status = StatusFailed
		return Result{Steps: steps, Stack: stack}, err
	}
	steps[2].Status = StatusOK
	steps[2].Detail = project.Slug

	if _, err := stale.Record(abs); err != nil {
		steps[3].Status = StatusFailed
		steps[3].Detail = err.Error()
	} else {
		steps[3].Status = StatusOK
		steps[3].Detail = "fingerprints saved"
	}

	steps[4].Status = StatusOK
	steps[4].Detail = "run `harness project current` to confirm"

	return Result{Project: project, Steps: steps, Stack: stack}, nil
}

func DetectStack(root string) []string {
	checks := map[string]string{
		"package.json":     "node",
		"go.mod":           "go",
		"Cargo.toml":       "rust",
		"Gemfile":          "ruby",
		"pyproject.toml":   "python",
		"requirements.txt": "python",
		"Dockerfile":       "container",
	}
	seen := map[string]bool{}
	var out []string
	for marker, label := range checks {
		if _, err := os.Stat(filepath.Join(root, marker)); err == nil {
			if !seen[label] {
				seen[label] = true
				out = append(out, label)
			}
		}
	}
	if len(out) == 0 {
		out = []string{constants.SlugFallbackName}
	}
	return out
}

func stackDescription(stack []string) string {
	if len(stack) == 0 {
		return "no markers detected"
	}
	desc := stack[0]
	for _, s := range stack[1:] {
		desc += ", " + s
	}
	return desc
}
