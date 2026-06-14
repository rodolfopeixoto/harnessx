// SPDX-License-Identifier: MIT

package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/ropeixoto/harnessx/internal/importwiz"
	"github.com/ropeixoto/harnessx/internal/stale"
	"github.com/ropeixoto/harnessx/internal/workspace"
)

type importHandlers struct{ srv *Server }

func (s *Server) registerImport(mux *http.ServeMux) {
	ih := &importHandlers{srv: s}
	mux.HandleFunc("/api/workspace/import", ih.runImport)
	mux.HandleFunc("/api/workspace/stale", ih.staleCurrent)
	mux.HandleFunc("/api/workspace/stale/", ih.staleBySlug)
}

func (h *importHandlers) runImport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		Path string `json:"path"`
		Name string `json:"name"`
		Slug string `json:"slug"`
		Yes  bool   `json:"yes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || strings.TrimSpace(body.Path) == "" {
		writeJSON(w, http.StatusBadRequest, errBody(errors.New("payload requires {path}")))
		return
	}
	reg, err := workspace.Open("")
	if err != nil {
		writeJSON(w, http.StatusServiceUnavailable, errBody(err))
		return
	}
	defer reg.Close()
	res, err := importwiz.Run(r.Context(), reg, importwiz.Options{Path: body.Path, DisplayName: body.Name, Slug: body.Slug, Confirm: body.Yes})
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errBody(err))
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func (h *importHandlers) staleCurrent(w http.ResponseWriter, r *http.Request) {
	entries, err := stale.Detect(h.srv.root)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errBody(err))
		return
	}
	writeJSON(w, http.StatusOK, entries)
}

func (h *importHandlers) staleBySlug(w http.ResponseWriter, r *http.Request) {
	slug := strings.TrimPrefix(r.URL.Path, "/api/workspace/stale/")
	if slug == "" {
		http.NotFound(w, r)
		return
	}
	reg, err := workspace.Open("")
	if err != nil {
		writeJSON(w, http.StatusServiceUnavailable, errBody(err))
		return
	}
	defer reg.Close()
	project, err := reg.Resolve(r.Context(), slug)
	if err != nil {
		writeJSON(w, http.StatusNotFound, errBody(err))
		return
	}
	entries, err := stale.Detect(project.RootPath)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errBody(err))
		return
	}
	writeJSON(w, http.StatusOK, entries)
}
