// SPDX-License-Identifier: MIT

package http

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ropeixoto/harnessx/internal/domain"
	"github.com/ropeixoto/harnessx/internal/platform/config"
	"github.com/ropeixoto/harnessx/internal/platform/constants"
	"github.com/ropeixoto/harnessx/internal/platform/paths"
)

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":   true,
		"root": s.root,
		"time": time.Now().UTC().Format(time.RFC3339),
	})
}

func (s *Server) listSessions(w http.ResponseWriter, r *http.Request) {
	repo, err := s.openDB()
	if err != nil {
		writeJSON(w, http.StatusOK, []domain.Session{})
		return
	}
	defer repo.Close()
	limit := atoiDefault(r.URL.Query().Get("limit"), constants.DefaultSessionListLimit)
	sessions, err := repo.ListRecentSessions(r.Context(), limit)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errBody(err))
		return
	}
	writeJSON(w, http.StatusOK, sessions)
}

func (s *Server) sessionDetail(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/sessions/")
	if id == "" {
		http.NotFound(w, r)
		return
	}
	repo, err := s.openDB()
	if err != nil {
		writeJSON(w, http.StatusNotFound, errBody(err))
		return
	}
	defer repo.Close()
	rows, err := repo.DB().QueryContext(r.Context(), `
		select id, session_id, stage, agent, status, started_at, finished_at,
		       latency_ms, input_tokens, output_tokens, estimated_cost_usd, exit_code
		from runs where session_id = ? order by started_at`, id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errBody(err))
		return
	}
	defer rows.Close()
	var runs []map[string]any
	for rows.Next() {
		var (
			runID, sessID, stage, status, startedAt string
			agent, finishedAt                       any
			latency, inTok, outTok, exit            any
			cost                                    any
		)
		if err := rows.Scan(&runID, &sessID, &stage, &agent, &status, &startedAt, &finishedAt,
			&latency, &inTok, &outTok, &cost, &exit); err != nil {
			writeJSON(w, http.StatusInternalServerError, errBody(err))
			return
		}
		runs = append(runs, map[string]any{
			"id": runID, "session_id": sessID, "stage": stage,
			"agent": agent, "status": status, "started_at": startedAt, "finished_at": finishedAt,
			"latency_ms": latency, "input_tokens": inTok, "output_tokens": outTok,
			"estimated_cost_usd": cost, "exit_code": exit,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"session_id": id, "runs": runs})
}

func (s *Server) runDetail(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/runs/")
	if id == "" {
		http.NotFound(w, r)
		return
	}
	repo, err := s.openDB()
	if err != nil {
		writeJSON(w, http.StatusNotFound, errBody(err))
		return
	}
	defer repo.Close()
	sensors, _ := repo.ListSensorResults(r.Context(), id)
	writeJSON(w, http.StatusOK, map[string]any{
		"run_id":  id,
		"sensors": sensors,
	})
}

func (s *Server) listSensorResults(w http.ResponseWriter, r *http.Request) {
	repo, err := s.openDB()
	if err != nil {
		writeJSON(w, http.StatusOK, []any{})
		return
	}
	defer repo.Close()
	rows, err := repo.DB().QueryContext(r.Context(), `
		select id, run_id, sensor, status, duration_ms, output_path, created_at
		from sensor_results order by id desc limit 500`)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errBody(err))
		return
	}
	defer rows.Close()
	var out []map[string]any
	for rows.Next() {
		var (
			id, dur                          int64
			runID, sensor, status, createdAt string
			outPath                          any
		)
		_ = rows.Scan(&id, &runID, &sensor, &status, &dur, &outPath, &createdAt)
		out = append(out, map[string]any{
			"id": id, "run_id": runID, "sensor": sensor, "status": status,
			"duration_ms": dur, "output_path": outPath, "created_at": createdAt,
		})
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) listAgents(w http.ResponseWriter, r *http.Request) {
	repo, err := s.openDB()
	if err != nil {
		writeJSON(w, http.StatusOK, []any{})
		return
	}
	defer repo.Close()
	rows, err := repo.DB().QueryContext(r.Context(), `
		select agent_id, max(score) as score, max(certified_at) as last_certified
		from agent_certifications group by agent_id`)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errBody(err))
		return
	}
	defer rows.Close()
	var out []map[string]any
	for rows.Next() {
		var id string
		var score int
		var at string
		if err := rows.Scan(&id, &score, &at); err != nil {
			continue
		}
		out = append(out, map[string]any{
			"agent_id": id, "score": score, "last_certified": at,
		})
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) listMemory(w http.ResponseWriter, r *http.Request) {
	repo, err := s.openDB()
	if err != nil {
		writeJSON(w, http.StatusOK, []any{})
		return
	}
	defer repo.Close()
	rows, err := repo.DB().QueryContext(r.Context(), `
		select id, scope, kind, content, confidence, updated_at
		from memories order by updated_at desc limit 200`)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errBody(err))
		return
	}
	defer rows.Close()
	var out []map[string]any
	for rows.Next() {
		var id, scope, kind, content, at string
		var conf float64
		_ = rows.Scan(&id, &scope, &kind, &content, &conf, &at)
		out = append(out, map[string]any{
			"id": id, "scope": scope, "kind": kind, "content": content,
			"confidence": conf, "updated_at": at,
		})
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) cost(w http.ResponseWriter, r *http.Request) {
	repo, err := s.openDB()
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{"total_usd": 0, "by_agent": []any{}})
		return
	}
	defer repo.Close()
	var total float64
	_ = repo.DB().QueryRowContext(r.Context(),
		`select coalesce(sum(estimated_cost_usd), 0) from runs`).Scan(&total)
	rows, err := repo.DB().QueryContext(r.Context(), `
		select agent, coalesce(sum(estimated_cost_usd), 0) as cost,
		       coalesce(sum(input_tokens), 0) as in_tokens,
		       coalesce(sum(output_tokens), 0) as out_tokens
		from runs where agent is not null group by agent order by cost desc`)
	var byAgent []map[string]any
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var agent string
			var cost float64
			var inTok, outTok int64
			_ = rows.Scan(&agent, &cost, &inTok, &outTok)
			byAgent = append(byAgent, map[string]any{
				"agent": agent, "cost_usd": cost,
				"input_tokens": inTok, "output_tokens": outTok,
			})
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"total_usd": total, "by_agent": byAgent,
	})
}

func (s *Server) logsTail(w http.ResponseWriter, r *http.Request) {
	cfg, _ := config.Load(filepath.Join(s.root, ".harness", "config", "harness.yaml"), s.root)
	p := config.Resolve(s.root, cfg.Logging.Path)
	b, err := os.ReadFile(p)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{"lines": []string{}})
		return
	}
	lines := strings.Split(strings.TrimRight(string(b), "\n"), "\n")
	tail := atoiDefault(r.URL.Query().Get("tail"), 100)
	if tail > 0 && len(lines) > tail {
		lines = lines[len(lines)-tail:]
	}
	writeJSON(w, http.StatusOK, map[string]any{"lines": lines})
}

func (s *Server) designManifest(w http.ResponseWriter, _ *http.Request) {
	serveProductJSON(w, s.root, "design-manifest.json")
}
func (s *Server) roadmap(w http.ResponseWriter, _ *http.Request) {
	serveProductJSON(w, s.root, "roadmap.json")
}
func (s *Server) toggles(w http.ResponseWriter, _ *http.Request) {
	serveProductJSON(w, s.root, "toggle-map.json")
}
func (s *Server) features(w http.ResponseWriter, _ *http.Request) {
	serveProductJSON(w, s.root, "feature-map.json")
}
func (s *Server) profile(w http.ResponseWriter, _ *http.Request) {
	p := filepath.Join(paths.HarnessDir(s.root), "project", "profile.json")
	serveOrEmpty(w, p, emptyProfile)
}
