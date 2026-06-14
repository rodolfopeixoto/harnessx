// SPDX-License-Identifier: MIT

// Package http exposes a small read-only REST surface for the dashboard.
// Every endpoint reads from SQLite or the on-disk artifact tree — there
// is no separate cache, so what the dashboard shows is what HarnessX
// actually recorded.
//
// File split:
//   - server.go    Server + New + Addr + Handler + Start lifecycle.
//   - handlers.go  Every /api/* handler.
//   - static.go    Static SPA serving + built-in HTML fallback.
//   - helpers.go   openDB + writeJSON + atoiDefault + sort helpers.
package http

import (
	"context"
	"errors"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/ropeixoto/harnessx/internal/platform/constants"
)

type Server struct {
	root string
	mu   sync.RWMutex
	addr string
	srv  *http.Server
	dist string
}

// Options bundles caller-tunable knobs.
type Options struct {
	Root string
	Addr string
	Dist string
}

func New(opts Options) (*Server, error) {
	if opts.Root == "" {
		return nil, errors.New("http: missing root")
	}
	if opts.Addr == "" {
		opts.Addr = constants.DefaultDashboardAddr
	}
	return &Server{root: opts.Root, addr: opts.Addr, dist: opts.Dist}, nil
}

// Addr returns the server's listen address. Useful after Start() resolves
// "" (any free port) to a concrete addr.
func (s *Server) Addr() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.addr
}

// Handler builds the multiplexer without binding to a socket. Tests use
// this to drive the API via httptest.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", s.health)
	mux.HandleFunc("/api/sessions", s.listSessions)
	mux.HandleFunc("/api/sessions/", s.sessionDetail)
	mux.HandleFunc("/api/runs/", s.runDetail)
	mux.HandleFunc("/api/sensors", s.listSensorResults)
	mux.HandleFunc("/api/agents", s.listAgents)
	mux.HandleFunc("/api/memory", s.listMemory)
	mux.HandleFunc("/api/cost", s.cost)
	mux.HandleFunc("/api/logs", s.logsTail)
	mux.HandleFunc("/api/design", s.designManifest)
	mux.HandleFunc("/api/roadmap", s.roadmap)
	mux.HandleFunc("/api/toggles", s.toggles)
	mux.HandleFunc("/api/features", s.features)
	mux.HandleFunc("/api/profile", s.profile)
	s.registerWorkspace(mux)
	s.registerCatalog(mux)
	s.registerImport(mux)
	s.registerPalette(mux)
	mux.HandleFunc("/", s.staticOrFallback)
	return logRequests(mux)
}

func (s *Server) Start(ctx context.Context) error {
	s.mu.RLock()
	addr := s.addr
	s.mu.RUnlock()
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	s.mu.Lock()
	s.addr = lis.Addr().String()
	s.srv = &http.Server{
		Handler:           s.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}
	srv := s.srv
	s.mu.Unlock()
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), constants.DefaultDashboardShutdownTimeout)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()
	return srv.Serve(lis)
}
