// SPDX-License-Identifier: MIT

package workspace

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ropeixoto/harnessx/internal/platform/constants"
	"github.com/ropeixoto/harnessx/internal/platform/paths"
)

// ResolveOptions controls how the workspace resolver picks a project.
type ResolveOptions struct {
	// Flag is the value of the global --project flag. Wins when non-empty.
	Flag string
	// Env, when set, is consulted second. If empty, HARNESS_PROJECT is read.
	Env string
	// CWD is the working directory used for the path-walk fallback. If empty,
	// os.Getwd() is used.
	CWD string
	// AutoTouch, when true, refreshes last_seen_at on the resolved project.
	AutoTouch bool
}

// ResolveSource identifies how the resolver picked the project.
type ResolveSource string

const (
	SourceFlag   ResolveSource = "flag"
	SourceEnv    ResolveSource = "env"
	SourceCWD    ResolveSource = "cwd"
	SourceActive ResolveSource = "active"
)

// Resolved bundles the chosen project with how it was picked.
type Resolved struct {
	Project Project
	Source  ResolveSource
}

// ErrNoProject is returned when no project can be resolved by any path.
var ErrNoProject = errors.New("workspace: no project resolved (use --project, HARNESS_PROJECT, run inside a project, or `harness project add`)")

// Resolve picks the active project using the precedence documented in the
// p11 spec: --project flag → HARNESS_PROJECT env → cwd walk-up → registry
// active row → ErrNoProject.
//
// The registry argument may be nil when callers only want the CWD walk-up
// behaviour (preserves v0.1.0 single-project commands when no registry exists).
func Resolve(ctx context.Context, r *Registry, opts ResolveOptions) (Resolved, error) {
	if res, ok, err := resolveFlag(ctx, r, opts); ok {
		return res, err
	}
	if res, ok, err := resolveEnv(ctx, r, opts); ok {
		return res, err
	}
	cwd, err := resolvedCWD(opts)
	if err != nil {
		return Resolved{}, err
	}
	if res, ok := resolveCWD(ctx, r, opts, cwd); ok {
		return res, nil
	}
	if r != nil {
		if p, err := r.Active(ctx); err == nil {
			return tap(r, opts.AutoTouch, Resolved{Project: p, Source: SourceActive})
		}
	}
	return Resolved{}, ErrNoProject
}

func resolveFlag(ctx context.Context, r *Registry, opts ResolveOptions) (Resolved, bool, error) {
	if opts.Flag == "" || r == nil {
		return Resolved{}, false, nil
	}
	p, err := r.Resolve(ctx, opts.Flag)
	if err != nil {
		return Resolved{}, true, fmt.Errorf("workspace: --project %q: %w", opts.Flag, err)
	}
	res, _ := tap(r, opts.AutoTouch, Resolved{Project: p, Source: SourceFlag})
	return res, true, nil
}

func resolveEnv(ctx context.Context, r *Registry, opts ResolveOptions) (Resolved, bool, error) {
	env := opts.Env
	if env == "" {
		env = os.Getenv(constants.EnvProjectOverride)
	}
	if env == "" || r == nil {
		return Resolved{}, false, nil
	}
	p, err := r.Resolve(ctx, env)
	if err != nil {
		return Resolved{}, true, fmt.Errorf("workspace: %s=%q: %w", constants.EnvProjectOverride, env, err)
	}
	res, _ := tap(r, opts.AutoTouch, Resolved{Project: p, Source: SourceEnv})
	return res, true, nil
}

func resolvedCWD(opts ResolveOptions) (string, error) {
	if opts.CWD != "" {
		return opts.CWD, nil
	}
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("workspace: getwd: %w", err)
	}
	return cwd, nil
}

func resolveCWD(ctx context.Context, r *Registry, opts ResolveOptions, cwd string) (Resolved, bool) {
	root, err := paths.FindProjectRoot(cwd)
	if err != nil || !hasMarker(root) {
		return Resolved{}, false
	}
	if r != nil {
		if p, err := r.Resolve(ctx, root); err == nil {
			res, _ := tap(r, opts.AutoTouch, Resolved{Project: p, Source: SourceCWD})
			return res, true
		}
	}
	return Resolved{Project: projectFromPath(root), Source: SourceCWD}, true
}

func projectFromPath(root string) Project {
	return Project{
		Slug:        Slugify(rootBase(root)),
		DisplayName: rootBase(root),
		RootPath:    root,
		DBPath:      defaultProjectDBPath(root),
	}
}

func rootBase(root string) string {
	// Avoid pulling filepath here; keep import set tight. The base extractor
	// in paths.FindProjectRoot already normalises separators.
	if root == "" {
		return "project"
	}
	last := 0
	for i, r := range root {
		if r == '/' || r == '\\' {
			last = i + 1
		}
	}
	if last >= len(root) {
		return "project"
	}
	return root[last:]
}

// hasMarker tests that FindProjectRoot actually identified a marker rather
// than returning its input unchanged (its documented fallback). We treat the
// returned root as valid only when at least one v0.1.0 marker exists there.
func hasMarker(root string) bool {
	for _, m := range []string{".git", ".harness", "go.mod", "package.json", "Gemfile", "Cargo.toml", "pyproject.toml", "requirements.txt"} {
		if _, err := os.Stat(filepath.Join(root, m)); err == nil {
			return true
		}
	}
	return false
}

func tap(r *Registry, doTouch bool, res Resolved) (Resolved, error) {
	if doTouch && r != nil && res.Project.ID != "" {
		_ = r.Touch(context.Background(), res.Project.ID)
	}
	return res, nil
}
