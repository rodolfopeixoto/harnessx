// SPDX-License-Identifier: MIT

// Package indexcmd wires the `harness project index|inspect` use cases on
// top of internal/index, persisting a session+run per invocation.
package indexcmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/ropeixoto/harnessx/internal/adapters/logger"
	"github.com/ropeixoto/harnessx/internal/adapters/sqlite"
	"github.com/ropeixoto/harnessx/internal/domain"
	"github.com/ropeixoto/harnessx/internal/index"
	"github.com/ropeixoto/harnessx/internal/platform/config"
	"github.com/ropeixoto/harnessx/internal/platform/ids"
	"github.com/ropeixoto/harnessx/internal/platform/paths"
)

type IndexOptions struct {
	StartDir string
	Force    bool
}

func RunIndex(ctx context.Context, opts IndexOptions, out io.Writer) (index.Result, error) {
	root, err := paths.FindProjectRoot(opts.StartDir)
	if err != nil {
		return index.Result{}, err
	}
	cfgPath := filepath.Join(root, ".harness", "config", "harness.yaml")
	cfg, err := config.Load(cfgPath, root)
	if err != nil {
		return index.Result{}, err
	}

	dbPath := config.Resolve(root, cfg.Database.Path)
	logPath := config.Resolve(root, cfg.Logging.Path)

	// Best-effort: record a session/run when DB exists, but do not block the
	// index pass if the project hasn't run `harness init` yet.
	var repo *sqlite.Repo
	var lg *logger.JSONL
	var sess domain.Session
	var run domain.Run
	hasDB := false
	if _, err := os.Stat(dbPath); err == nil {
		repo, err = sqlite.Open(dbPath)
		if err == nil {
			hasDB = true
			defer repo.Close()
			lg, _ = logger.Open(logPath, cfg.Logging.RotateMaxBytes)
			if lg != nil {
				defer lg.Close()
			}
			now := time.Now().UTC()
			sess = domain.Session{
				ID: ids.New(), ProjectPath: root, Mode: domain.ModeSetup,
				Status: domain.StatusRunning, StartedAt: now,
			}
			run = domain.Run{
				ID: ids.New(), SessionID: sess.ID, Stage: domain.Stage("project_index"),
				Status: domain.StatusRunning, StartedAt: now,
			}
			_ = repo.CreateSession(ctx, sess)
			_ = repo.CreateRun(ctx, run)
			if lg != nil {
				_ = lg.Write("info", map[string]any{
					"stage": "project_index", "session_id": sess.ID, "run_id": run.ID, "root": root, "force": opts.Force,
				})
			}
		}
	}

	res, buildErr := index.Build(index.Options{Root: root, Force: opts.Force})

	if hasDB {
		status := domain.StatusSucceeded
		if buildErr != nil {
			status = domain.StatusFailed
		}
		end := time.Now().UTC()
		_ = repo.FinishRun(ctx, run.ID, status, end, exitFromErr(buildErr))
		_ = repo.FinishSession(ctx, sess.ID, status, end)
	}
	if buildErr != nil {
		return res, buildErr
	}

	fmt.Fprintf(out, "harness: indexed %s\n", res.OutputDir)
	if len(res.Updated) > 0 {
		fmt.Fprintln(out, "  updated:")
		for _, m := range res.Updated {
			fmt.Fprintf(out, "    - %s\n", m)
		}
	}
	if len(res.Skipped) > 0 {
		fmt.Fprintln(out, "  skipped (inputs unchanged):")
		for _, m := range res.Skipped {
			fmt.Fprintf(out, "    - %s\n", m)
		}
	}
	if !hasDB {
		fmt.Fprintln(out, "note: .harness/ not initialised — telemetry skipped. Run `harness init` first.")
	}
	return res, nil
}

func exitFromErr(err error) int {
	if err == nil {
		return 0
	}
	return 1
}

type InspectOptions struct {
	StartDir string
	Map      string // empty = list all maps
}

func RunInspect(opts InspectOptions, out io.Writer) error {
	root, err := paths.FindProjectRoot(opts.StartDir)
	if err != nil {
		return err
	}
	dir := filepath.Join(root, ".harness", "project")

	if opts.Map == "" {
		fmt.Fprintf(out, "project maps under %s:\n", dir)
		for _, m := range index.AllMaps() {
			info, err := os.Stat(filepath.Join(dir, string(m)))
			if err != nil {
				fmt.Fprintf(out, "  - %s (missing)\n", m)
				continue
			}
			fmt.Fprintf(out, "  - %s (%d bytes, %s)\n", m, info.Size(), info.ModTime().Format(time.RFC3339))
		}
		return nil
	}

	target := filepath.Join(dir, opts.Map)
	if filepath.Ext(opts.Map) == "" {
		target = filepath.Join(dir, opts.Map+".json")
	}
	b, err := os.ReadFile(target)
	if err != nil {
		return fmt.Errorf("inspect: %w", err)
	}
	var v any
	if err := json.Unmarshal(b, &v); err != nil {
		// fall back to raw bytes
		out.Write(b)
		return nil
	}
	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
