// SPDX-License-Identifier: MIT

// Package projectcmd implements the project-registry subcommands surfaced
// under `harness project`. Logic lives in internal/workspace; this package
// is a thin renderer.
package projectcmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/ropeixoto/harnessx/internal/platform/constants"
	"github.com/ropeixoto/harnessx/internal/workspace"
)

// Options bundles flags accepted by every subcommand.
type Options struct {
	RegistryPath string
}

func openRegistry(opts Options) (*workspace.Registry, error) {
	return workspace.Open(opts.RegistryPath)
}

// Add registers a project root.
func Add(ctx context.Context, opts Options, path, displayName, slug string, out io.Writer) error {
	if path == "" {
		var err error
		path, err = os.Getwd()
		if err != nil {
			return err
		}
	}
	r, err := openRegistry(opts)
	if err != nil {
		return err
	}
	defer r.Close()
	p, err := r.Add(ctx, path, displayName, slug)
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "registered %s\n  slug: %s\n  root: %s\n  db:   %s\n", p.DisplayName, p.Slug, p.RootPath, p.DBPath)
	return nil
}

// StaleSince returns projects whose LastSeenAt is older than the
// given threshold. Archived projects are excluded.
func StaleSince(ctx context.Context, opts Options, threshold time.Time) ([]workspace.Project, error) {
	r, err := openRegistry(opts)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	projects, err := r.List(ctx, false)
	if err != nil {
		return nil, err
	}
	var stale []workspace.Project
	for _, p := range projects {
		if p.LastSeenAt == nil {
			continue
		}
		if p.LastSeenAt.Before(threshold) {
			stale = append(stale, p)
		}
	}
	return stale, nil
}

// List prints registered projects.
func List(ctx context.Context, opts Options, includeArchived bool, out io.Writer) error {
	r, err := openRegistry(opts)
	if err != nil {
		return err
	}
	defer r.Close()
	projects, err := r.List(ctx, includeArchived)
	if err != nil {
		return err
	}
	if len(projects) == 0 {
		fmt.Fprintln(out, "no projects registered yet (use 'harness project add <path>')")
		return nil
	}
	active, _ := r.Active(ctx)
	tw := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "  SLUG\tNAME\tSTATUS\tLAST SEEN\tROOT")
	for _, p := range projects {
		marker := " "
		if active.ID != "" && active.ID == p.ID {
			marker = "*"
		}
		status := "active"
		if p.ArchivedAt != nil {
			status = "archived"
		}
		seen := "—"
		if p.LastSeenAt != nil {
			seen = p.LastSeenAt.Format("2006-01-02 15:04")
		}
		fmt.Fprintf(tw, "%s %s\t%s\t%s\t%s\t%s\n", marker, p.Slug, p.DisplayName, status, seen, p.RootPath)
	}
	return tw.Flush()
}

// Switch sets the active project.
func Switch(ctx context.Context, opts Options, ref string, out io.Writer) error {
	r, err := openRegistry(opts)
	if err != nil {
		return err
	}
	defer r.Close()
	p, err := r.SetActive(ctx, ref)
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "active: %s (%s)\n", p.Slug, p.RootPath)
	return nil
}

// Current prints the resolved project.
func Current(ctx context.Context, opts Options, flag string, out io.Writer) error {
	r, err := openRegistry(opts)
	if err != nil {
		return err
	}
	defer r.Close()
	res, err := workspace.Resolve(ctx, r, workspace.ResolveOptions{Flag: flag, AutoTouch: true})
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "project: %s\n  slug:   %s\n  root:   %s\n  source: %s\n",
		res.Project.DisplayName, res.Project.Slug, res.Project.RootPath, res.Source)
	return nil
}

// Archive marks a project archived.
func Archive(ctx context.Context, opts Options, ref string, out io.Writer) error {
	r, err := openRegistry(opts)
	if err != nil {
		return err
	}
	defer r.Close()
	p, err := r.Archive(ctx, ref)
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "archived %s (%s)\n", p.Slug, p.RootPath)
	return nil
}

// Unarchive clears the archived flag.
func Unarchive(ctx context.Context, opts Options, ref string, out io.Writer) error {
	r, err := openRegistry(opts)
	if err != nil {
		return err
	}
	defer r.Close()
	p, err := r.Unarchive(ctx, ref)
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "unarchived %s (%s)\n", p.Slug, p.RootPath)
	return nil
}

// Scan walks one or more roots looking for unregistered projects (folders
// containing .harness/) and prints the candidates. Use --yes to register
// them in batch.
func Scan(ctx context.Context, opts Options, root string, registerAll bool, out io.Writer) error {
	if root == "" {
		var err error
		root, err = os.Getwd()
		if err != nil {
			return err
		}
	}
	r, err := openRegistry(opts)
	if err != nil {
		return err
	}
	defer r.Close()
	candidates, missing, err := scanFS(ctx, r, root)
	if err != nil {
		return err
	}
	if len(candidates) == 0 && len(missing) == 0 {
		fmt.Fprintln(out, "no candidates found under", root)
		return nil
	}
	if len(candidates) > 0 {
		fmt.Fprintf(out, "unregistered projects under %s:\n", root)
		for _, c := range candidates {
			fmt.Fprintf(out, "  + %s\n", c)
		}
	}
	if len(missing) > 0 {
		fmt.Fprintf(out, "registered but missing on disk:\n")
		for _, m := range missing {
			fmt.Fprintf(out, "  - %s (%s)\n", m.Slug, m.RootPath)
		}
	}
	if !registerAll {
		fmt.Fprintln(out, "re-run with --yes to register all candidates above")
		return nil
	}
	for _, c := range candidates {
		if _, err := r.Add(ctx, c, "", ""); err != nil {
			fmt.Fprintf(out, "  ! %s: %v\n", c, err)
			continue
		}
		fmt.Fprintf(out, "  + registered %s\n", c)
	}
	return nil
}

func scanFS(ctx context.Context, r *workspace.Registry, root string) ([]string, []workspace.Project, error) {
	all, err := r.List(ctx, true)
	if err != nil {
		return nil, nil, err
	}
	known := map[string]workspace.Project{}
	for _, p := range all {
		known[p.RootPath] = p
	}
	var candidates []string
	seenRoots := map[string]struct{}{}
	err = filepath.Walk(root, func(path string, info os.FileInfo, werr error) error {
		if werr != nil {
			return nil
		}
		if !info.IsDir() {
			return nil
		}
		base := filepath.Base(path)
		if base == "node_modules" || base == ".git" || base == "vendor" {
			return filepath.SkipDir
		}
		if base != constants.HarnessDir {
			return nil
		}
		projectRoot := filepath.Dir(path)
		if _, ok := seenRoots[projectRoot]; ok {
			return filepath.SkipDir
		}
		seenRoots[projectRoot] = struct{}{}
		if _, ok := known[projectRoot]; !ok {
			candidates = append(candidates, projectRoot)
		}
		return filepath.SkipDir
	})
	if err != nil {
		return nil, nil, err
	}
	var missing []workspace.Project
	for _, p := range all {
		if _, err := os.Stat(p.RootPath); errors.Is(err, os.ErrNotExist) {
			missing = append(missing, p)
		}
	}
	sort.Strings(candidates)
	sort.Slice(missing, func(i, j int) bool { return missing[i].Slug < missing[j].Slug })
	return candidates, missing, nil
}

// Forget removes the registry row for a project. Does not touch the
// project's filesystem.
func Forget(ctx context.Context, opts Options, ref string, out io.Writer) error {
	r, err := openRegistry(opts)
	if err != nil {
		return err
	}
	defer r.Close()
	if err := r.Forget(ctx, ref); err != nil {
		return err
	}
	fmt.Fprintf(out, "forgot %s (registry row removed; project files untouched)\n", ref)
	return nil
}

// ResolveRef is a tiny helper for cmd_project's Cobra arg validation: it
// trims whitespace and forwards.
func ResolveRef(args []string) string {
	if len(args) == 0 {
		return ""
	}
	return strings.TrimSpace(args[0])
}
