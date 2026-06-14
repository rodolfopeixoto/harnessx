// SPDX-License-Identifier: MIT

// Package speccmd wires `harness spec init` — scaffolds a fresh
// spec-driven-development document under .harness/artifacts/specs/.
// Manual companion to the auto-generated specs that `harness feature`
// produces.
package speccmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ropeixoto/harnessx/internal/domain"
	"github.com/ropeixoto/harnessx/internal/platform/paths"
	"github.com/ropeixoto/harnessx/internal/spec"
)

type InitOptions struct {
	StartDir string
	Name     string
	Prompt   string
	Mode     domain.Mode
}

func Init(out io.Writer, opts InitOptions) error {
	if opts.Name == "" {
		return errors.New("spec init: --name is required")
	}
	root, err := paths.FindProjectRoot(opts.StartDir)
	if err != nil {
		return err
	}
	mode := opts.Mode
	if mode == "" {
		mode = domain.ModeFeature
	}
	prompt := opts.Prompt
	if prompt == "" {
		prompt = "Manual spec for " + opts.Name
	}
	sp := spec.NewFromPrompt(prompt, mode)
	sp.Title = opts.Name

	dir := filepath.Join(paths.HarnessDir(root), "artifacts", "specs")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	path := filepath.Join(dir, slug(opts.Name)+".md")
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("spec init: %s already exists", path)
	}
	written, err := sp.Write(root)
	if err != nil {
		return err
	}
	if err := os.Rename(written, path); err != nil {
		return fmt.Errorf("spec init: rename %s → %s: %w", written, path, err)
	}
	fmt.Fprintf(out, "wrote spec: %s\n", path)
	fmt.Fprintln(out, "edit the markdown, then either commit it or feed it back via `harness feature`.")
	return nil
}

func slug(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == ' ' || r == '_' || r == '-' || r == '/':
			b.WriteByte('-')
		}
	}
	out := strings.Trim(b.String(), "-")
	if out == "" {
		out = "spec-" + time.Now().UTC().Format("20060102T150405")
	}
	return out
}
