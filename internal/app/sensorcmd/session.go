// SPDX-License-Identifier: MIT

package sensorcmd

import (
	"context"
	"os"
	"time"

	"github.com/ropeixoto/harnessx/internal/adapters/logger"
	"github.com/ropeixoto/harnessx/internal/adapters/sqlite"
	"github.com/ropeixoto/harnessx/internal/domain"
	"github.com/ropeixoto/harnessx/internal/platform/ids"
	"github.com/ropeixoto/harnessx/internal/sensors"
)

// sessionState packages a persistence bundle for a sensor Run: the SQLite
// repo, jsonl logger, and the freshly minted domain rows. hasDB=false means
// the caller must skip all writes silently (pre-init projects).
type sessionState struct {
	hasDB bool
	repo  *sqlite.Repo
	lg    *logger.JSONL
	sess  domain.Session
	run   domain.Run
}

func openSessionState(ctx context.Context, rc runtimeContext) sessionState {
	var st sessionState
	if _, err := os.Stat(rc.dbPath); err != nil {
		return st
	}
	repo, err := sqlite.Open(rc.dbPath)
	if err != nil {
		return st
	}
	st.hasDB = true
	st.repo = repo
	st.lg, _ = logger.Open(rc.logPath, rc.cfg.Logging.RotateMaxBytes)
	now := time.Now().UTC()
	st.sess = domain.Session{
		ID: ids.New(), ProjectPath: rc.root, Mode: domain.ModeAudit,
		Status: domain.StatusRunning, StartedAt: now,
	}
	st.run = domain.Run{
		ID: ids.New(), SessionID: st.sess.ID, Stage: domain.StageSensors,
		Status: domain.StatusRunning, StartedAt: now,
	}
	_ = repo.CreateSession(ctx, st.sess)
	_ = repo.CreateRun(ctx, st.run)
	return st
}

func (s *sessionState) close() {
	if s.repo != nil {
		s.repo.Close()
	}
	if s.lg != nil {
		s.lg.Close()
	}
}

func (s *sessionState) recordResult(ctx context.Context, res sensors.Result) {
	if !s.hasDB {
		return
	}
	_ = s.repo.WriteSensorResult(ctx, s.run.ID, res.ID, string(res.Status), res.Duration.Milliseconds(), res.OutputPath, time.Now().UTC())
	if s.lg != nil {
		_ = s.lg.Write("info", map[string]any{
			"stage": "sensor", "session_id": s.sess.ID, "run_id": s.run.ID,
			"sensor": res.ID, "status": string(res.Status),
			"duration_ms": res.Duration.Milliseconds(),
		})
	}
}

func (s *sessionState) finish(ctx context.Context, failed bool) {
	if !s.hasDB {
		return
	}
	end := time.Now().UTC()
	status := domain.StatusSucceeded
	exit := 0
	if failed {
		status = domain.StatusFailed
		exit = 1
	}
	_ = s.repo.FinishRun(ctx, s.run.ID, status, end, exit)
	_ = s.repo.FinishSession(ctx, s.sess.ID, status, end)
}
