// SPDX-License-Identifier: MIT

// Package artifactcmd implements `harness artifact ls` + `harness artifact
// cat` for the on-disk artifact tree under .harness/artifacts/.
package artifactcmd

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"

	"github.com/ropeixoto/harnessx/internal/platform/paths"
)

type ListOptions struct {
	StartDir string
	Kind     string // optional filter (specs|plans|reports|sensors|perf)
}

func List(out io.Writer, opts ListOptions) error {
	root, err := paths.FindProjectRoot(opts.StartDir)
	if err != nil {
		return err
	}
	base := filepath.Join(root, ".harness", "artifacts")
	if opts.Kind != "" {
		base = filepath.Join(base, opts.Kind)
	}
	if _, err := os.Stat(base); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			fmt.Fprintln(out, "(no artifacts under .harness/artifacts)")
			return nil
		}
		return fmt.Errorf("artifact ls: %w", err)
	}
	type row struct {
		rel   string
		size  int64
		mtime string
	}
	var rows []row
	_ = filepath.WalkDir(base, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		info, _ := d.Info()
		rel, _ := filepath.Rel(filepath.Join(root, ".harness", "artifacts"), p)
		rows = append(rows, row{rel: rel, size: info.Size(), mtime: info.ModTime().Format("2006-01-02T15:04:05Z07:00")})
		return nil
	})
	sort.Slice(rows, func(i, j int) bool { return rows[i].mtime > rows[j].mtime })
	if len(rows) == 0 {
		fmt.Fprintln(out, "(no artifacts under .harness/artifacts)")
		return nil
	}
	fmt.Fprintf(out, "%-25s %10s  %s\n", "MTIME", "BYTES", "PATH")
	for _, r := range rows {
		fmt.Fprintf(out, "%-25s %10d  %s\n", r.mtime, r.size, r.rel)
	}
	return nil
}

type CatOptions struct {
	StartDir string
	Path     string // relative under .harness/artifacts
}

func Cat(out io.Writer, opts CatOptions) error {
	if opts.Path == "" {
		return errors.New("artifact cat: missing path")
	}
	root, err := paths.FindProjectRoot(opts.StartDir)
	if err != nil {
		return err
	}
	full := filepath.Join(root, ".harness", "artifacts", filepath.Clean(opts.Path))
	abs, err := filepath.Abs(full)
	if err != nil {
		return err
	}
	guard, _ := filepath.Abs(filepath.Join(root, ".harness", "artifacts"))
	if !startsWith(abs, guard) {
		return errors.New("artifact cat: refusing path outside .harness/artifacts")
	}
	f, err := os.Open(abs)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(out, f)
	return err
}

func startsWith(p, prefix string) bool {
	if p == prefix {
		return true
	}
	return len(p) > len(prefix) && p[:len(prefix)] == prefix && (p[len(prefix)] == filepath.Separator)
}
