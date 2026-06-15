// SPDX-License-Identifier: MIT

package http

import (
	"net/http"

	"github.com/ropeixoto/harnessx/internal/cleanup"
	"github.com/ropeixoto/harnessx/internal/cleanup/detectors"
	"github.com/ropeixoto/harnessx/internal/hookscan"
	"github.com/ropeixoto/harnessx/internal/mcpscan"
)

func (s *Server) registerScans(mux *http.ServeMux) {
	mux.HandleFunc("/api/mcps", s.mcps)
	mux.HandleFunc("/api/hooks", s.hooks)
	mux.HandleFunc("/api/cleanup/scan", s.cleanupScan)
}

func (s *Server) mcps(w http.ResponseWriter, r *http.Request) {
	servers, err := mcpscan.Scan(s.root)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errBody(err))
		return
	}
	if servers == nil {
		servers = []mcpscan.McpServer{}
	}
	writeJSON(w, http.StatusOK, servers)
}

func (s *Server) hooks(w http.ResponseWriter, r *http.Request) {
	hooks, err := hookscan.Scan(s.root)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errBody(err))
		return
	}
	if hooks == nil {
		hooks = []hookscan.Hook{}
	}
	writeJSON(w, http.StatusOK, hooks)
}

func (s *Server) cleanupScan(w http.ResponseWriter, r *http.Request) {
	scanner := cleanup.New(
		detectors.Worktrees{},
		detectors.Caches{},
		detectors.AbandonedHarness{},
		detectors.LargeFiles{},
		detectors.VMLeftovers{},
		detectors.ClaudeLeftovers{},
	)
	findings, err := scanner.Scan(r.Context(), s.root)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errBody(err))
		return
	}
	if findings == nil {
		findings = []cleanup.Finding{}
	}
	writeJSON(w, http.StatusOK, findings)
}
