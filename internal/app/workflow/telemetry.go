// SPDX-License-Identifier: MIT

package workflow

import (
	stdctx "context"
	"os"
	"path/filepath"
	"time"

	"github.com/ropeixoto/harnessx/internal/adapters/sqlite"
	"github.com/ropeixoto/harnessx/internal/domain"
	"github.com/ropeixoto/harnessx/internal/index"
	"github.com/ropeixoto/harnessx/internal/platform/config"
	"github.com/ropeixoto/harnessx/internal/platform/ids"
	"github.com/ropeixoto/harnessx/internal/platform/paths"
)

type runtimeCtx struct {
	root    string
	cfg     config.Config
	dbPath  string
	logPath string
	profile index.Profile
}

func newRC(startDir string) (runtimeCtx, error) {
	root, err := paths.FindProjectRoot(startDir)
	if err != nil {
		return runtimeCtx{}, err
	}
	cfg, err := config.Load(filepath.Join(root, ".harness", "config", "harness.yaml"), root)
	if err != nil {
		return runtimeCtx{}, err
	}
	rc := runtimeCtx{
		root: root, cfg: cfg,
		dbPath:  config.Resolve(root, cfg.Database.Path),
		logPath: config.Resolve(root, cfg.Logging.Path),
	}
	_ = index.ReadMap(root, index.MapProfile, &rc.profile)
	return rc, nil
}

func openTelemetry(ctx stdctx.Context, rc runtimeCtx, mode domain.Mode, stage domain.Stage) (domain.Session, domain.Run, *sqlite.Repo) {
	if _, err := os.Stat(rc.dbPath); err != nil {
		return domain.Session{}, domain.Run{}, nil
	}
	repo, err := sqlite.Open(rc.dbPath)
	if err != nil {
		return domain.Session{}, domain.Run{}, nil
	}
	now := time.Now().UTC()
	sess := domain.Session{
		ID: ids.New(), ProjectPath: rc.root, Mode: mode,
		Status: domain.StatusRunning, StartedAt: now,
	}
	run := domain.Run{
		ID: ids.New(), SessionID: sess.ID, Stage: stage,
		Status: domain.StatusRunning, StartedAt: now,
	}
	_ = repo.CreateSession(ctx, sess)
	_ = repo.CreateRun(ctx, run)
	return sess, run, repo
}

func finishTelemetry(ctx stdctx.Context, repo *sqlite.Repo, sess domain.Session, run domain.Run, status domain.Status) {
	if repo == nil || run.ID == "" {
		return
	}
	end := time.Now().UTC()
	_ = repo.FinishRun(ctx, run.ID, status, end, 0)
	_ = repo.FinishSession(ctx, sess.ID, status, end)
}

func closeRepo(r *sqlite.Repo) {
	if r != nil {
		_ = r.Close()
	}
}
