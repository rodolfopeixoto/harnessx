// SPDX-License-Identifier: MIT

package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"

	"github.com/ropeixoto/harnessx/internal/adapters/sqlite"
	"github.com/ropeixoto/harnessx/internal/domain"
	"github.com/ropeixoto/harnessx/internal/platform/config"
	"github.com/ropeixoto/harnessx/internal/platform/paths"
)

func (s *Server) openDB() (*sqlite.Repo, error) {
	cfg, err := config.Load(filepath.Join(s.root, ".harness", "config", "harness.yaml"), s.root)
	if err != nil {
		return nil, err
	}
	dbPath := config.Resolve(s.root, cfg.Database.Path)
	if _, err := os.Stat(dbPath); err != nil {
		return nil, fmt.Errorf("dashboard: db missing at %s", dbPath)
	}
	return sqlite.Open(dbPath)
}

func serveProductJSON(w http.ResponseWriter, root, name string) {
	serveFileJSON(w, filepath.Join(paths.HarnessDir(root), "product", name))
}

func serveFileJSON(w http.ResponseWriter, path string) {
	b, err := os.ReadFile(path)
	if err != nil {
		writeJSON(w, http.StatusNotFound, errBody(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(b)
}

// serveOrEmpty serves the JSON file at path, or — when the file is
// absent — writes a 200 with the supplied empty-envelope marshaller.
// Used by endpoints whose artifact is produced by a later optional step
// (e.g. `harness project index`) so the dashboard can render the empty
// state instead of an error.
func serveOrEmpty(w http.ResponseWriter, path string, empty func() any) {
	if _, err := os.Stat(path); err != nil {
		writeJSON(w, http.StatusOK, empty())
		return
	}
	serveFileJSON(w, path)
}

func emptyProfile() any {
	return map[string]any{
		"stacks":       []any{},
		"languages":    []any{},
		"markers":      []any{},
		"generated_at": nil,
	}
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func errBody(err error) map[string]any {
	return map[string]any{"error": err.Error()}
}

func atoiDefault(s string, def int) int {
	if s == "" {
		return def
	}
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return def
		}
		n = n*10 + int(c-'0')
	}
	return n
}

func logRequests(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
	})
}

// SortedSessions sorts sessions newest-first; exported for callers that
// need stable ordering outside the API surface.
func SortedSessions(in []domain.Session) []domain.Session {
	out := make([]domain.Session, len(in))
	copy(out, in)
	sort.Slice(out, func(i, j int) bool { return out[i].StartedAt.After(out[j].StartedAt) })
	return out
}
