// SPDX-License-Identifier: MIT

// Package initcmd bootstraps a project's .harness/ directory: writes the
// default config, creates the SQLite database, applies migrations, and
// records a bootstrap session + run.
package initcmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/ropeixoto/harnessx/internal/adapters/logger"
	"github.com/ropeixoto/harnessx/internal/adapters/sqlite"
	"github.com/ropeixoto/harnessx/internal/domain"
	"github.com/ropeixoto/harnessx/internal/platform/config"
	"github.com/ropeixoto/harnessx/internal/platform/ids"
	"github.com/ropeixoto/harnessx/internal/platform/paths"
)

// configTemplate mirrors templates/.harness/config/harness.yaml. Embedding
// it here keeps `harness init` self-contained (no template lookup at runtime).
const configTemplate = `# HarnessX project configuration. Created by ` + "`harness init`" + `.
# See docs/configuration.md for the full schema.

version: 1

project:
  name: {{.Name}}
  root: {{.Root}}

database:
  path: .harness/db/harness.sqlite

logging:
  path: .harness/logs/events.jsonl
  rotate_max_bytes: 10485760

agents: {}
routes: {}
sensors: {}
context: {}
`

const hookTemplate = `#!/bin/sh
# .harness/hooks/pre-tool-use.sh — runs before every adapter invocation.
# Exit 0 to allow the run. Non-zero blocks unless autonomy=full_project_loop.
#
# Available bundled templates (install via 'harness hook add pre-tool-use'):
#   lint     — go vet + golangci-lint on staged files
#   secrets  — refuse runs when .env exposes a key/token
#   noforce  — refuse force-push prompts
exit 0
`

const harnessGitignore = `db/
logs/
cache/
artifacts/
`

type Result struct {
	Root        string
	HarnessDir  string
	ConfigPath  string
	DBPath      string
	LogPath     string
	SessionID   string
	RunID       string
	AlreadyInit bool
}

type Options struct {
	StartDir string // working dir from which to resolve the project root
	Force    bool   // overwrite existing config (db is preserved)
}

func Run(ctx context.Context, opts Options, out io.Writer) (Result, error) {
	root, err := paths.FindProjectRoot(opts.StartDir)
	if err != nil {
		return Result{}, fmt.Errorf("init: locate root: %w", err)
	}
	hd := paths.HarnessDir(root)
	res := Result{
		Root:       root,
		HarnessDir: hd,
		ConfigPath: filepath.Join(hd, "config", "harness.yaml"),
		DBPath:     filepath.Join(hd, "db", "harness.sqlite"),
		LogPath:    filepath.Join(hd, "logs", "events.jsonl"),
	}

	if _, err := os.Stat(res.ConfigPath); err == nil {
		res.AlreadyInit = true
		if !opts.Force {
			fmt.Fprintf(out, "harness: already initialised at %s\n", hd)
		}
	}

	// Create directory layout.
	for _, sub := range []string{"config", "db", "logs", "cache", "artifacts", "product", "project", "hooks"} {
		if err := os.MkdirAll(filepath.Join(hd, sub), 0o755); err != nil {
			return res, fmt.Errorf("init: mkdir %s: %w", sub, err)
		}
	}

	hookPath := filepath.Join(hd, "hooks", "pre-tool-use.sh")
	if err := writeIfMissing(hookPath, []byte(hookTemplate), opts.Force); err != nil {
		return res, err
	}
	if err := os.Chmod(hookPath, 0o755); err != nil {
		return res, err
	}

	// Write .harness/.gitignore.
	if err := writeIfMissing(filepath.Join(hd, ".gitignore"), []byte(harnessGitignore), opts.Force); err != nil {
		return res, err
	}

	// Ensure the project's root .gitignore excludes .harness/worktrees/ — git
	// worktrees created during agent runs would otherwise show up as embedded
	// repos when the user runs `git add .` (audit BUG-16).
	if err := ensureRootGitignoreLine(root, ".harness/worktrees/"); err != nil {
		return res, err
	}

	// Write config from template.
	if !res.AlreadyInit || opts.Force {
		tpl, err := template.New("cfg").Parse(configTemplate)
		if err != nil {
			return res, fmt.Errorf("init: parse template: %w", err)
		}
		f, err := os.Create(res.ConfigPath)
		if err != nil {
			return res, fmt.Errorf("init: create config: %w", err)
		}
		err = tpl.Execute(f, struct{ Name, Root string }{
			Name: filepath.Base(root),
			Root: root,
		})
		_ = f.Close()
		if err != nil {
			return res, fmt.Errorf("init: write config: %w", err)
		}
	}

	// Load (possibly merged) config to learn DB + log paths.
	cfg, err := config.Load(res.ConfigPath, root)
	if err != nil {
		return res, err
	}
	res.DBPath = config.Resolve(root, cfg.Database.Path)
	res.LogPath = config.Resolve(root, cfg.Logging.Path)

	// Open DB (applies migrations) and record bootstrap session/run.
	repo, err := sqlite.Open(res.DBPath)
	if err != nil {
		return res, err
	}
	defer repo.Close()

	lg, err := logger.Open(res.LogPath, cfg.Logging.RotateMaxBytes)
	if err != nil {
		return res, err
	}
	defer lg.Close()

	now := time.Now().UTC()
	sess := domain.Session{
		ID: ids.New(), ProjectPath: root,
		Mode: domain.ModeBootstrap, Status: domain.StatusRunning, StartedAt: now,
	}
	run := domain.Run{
		ID: ids.New(), SessionID: sess.ID,
		Stage: domain.StageInit, Status: domain.StatusRunning, StartedAt: now,
	}
	if err := repo.CreateSession(ctx, sess); err != nil {
		return res, fmt.Errorf("init: record session: %w", err)
	}
	if err := repo.CreateRun(ctx, run); err != nil {
		return res, fmt.Errorf("init: record run: %w", err)
	}
	_ = lg.Write("info", map[string]any{
		"stage": "init", "session_id": sess.ID, "run_id": run.ID, "root": root,
	})

	end := time.Now().UTC()
	if err := repo.FinishRun(ctx, run.ID, domain.StatusSucceeded, end, 0); err != nil {
		return res, err
	}
	if err := repo.FinishSession(ctx, sess.ID, domain.StatusSucceeded, end); err != nil {
		return res, err
	}

	res.SessionID = sess.ID
	res.RunID = run.ID

	fmt.Fprintf(out, "harness: initialised %s\n", hd)
	fmt.Fprintf(out, "  config:  %s\n", res.ConfigPath)
	fmt.Fprintf(out, "  db:      %s\n", res.DBPath)
	fmt.Fprintf(out, "  log:     %s\n", res.LogPath)
	return res, nil
}

func writeIfMissing(path string, data []byte, force bool) error {
	_, err := os.Stat(path)
	if err == nil && !force {
		return nil
	}
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// ensureRootGitignoreLine appends line to root/.gitignore if not already
// present; the file is created when missing. Idempotent.
func ensureRootGitignoreLine(root, line string) error {
	path := filepath.Join(root, ".gitignore")
	existing, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	for _, candidate := range strings.Split(string(existing), "\n") {
		if strings.TrimSpace(candidate) == line {
			return nil
		}
	}
	var buf strings.Builder
	buf.Write(existing)
	if len(existing) > 0 && !strings.HasSuffix(string(existing), "\n") {
		buf.WriteByte('\n')
	}
	buf.WriteString("# HarnessX worktrees (do not commit)\n")
	buf.WriteString(line)
	buf.WriteByte('\n')
	return os.WriteFile(path, []byte(buf.String()), 0o644)
}
