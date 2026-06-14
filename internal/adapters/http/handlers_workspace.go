// SPDX-License-Identifier: MIT

package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/ropeixoto/harnessx/internal/workspace"
)

// workspaceHandlers wires registry endpoints. The registry handle is opened
// on first call and reused; nil on failure so endpoints degrade gracefully
// to empty responses (matches the v0.1.0 pattern in handlers.go).
type workspaceHandlers struct{ srv *Server }

func (s *Server) registerWorkspace(mux *http.ServeMux) {
	wh := &workspaceHandlers{srv: s}
	mux.HandleFunc("/api/workspace/projects", wh.projects)
	mux.HandleFunc("/api/workspace/switch", wh.switchActive)
	mux.HandleFunc("/api/workspace/current", wh.current)
}

func (h *workspaceHandlers) openRegistry() (*workspace.Registry, error) {
	return workspace.Open("")
}

func (h *workspaceHandlers) projects(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.listProjects(w, r)
	case http.MethodPost:
		h.addProject(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *workspaceHandlers) listProjects(w http.ResponseWriter, r *http.Request) {
	reg, err := h.openRegistry()
	if err != nil {
		writeJSON(w, http.StatusOK, []workspace.Project{})
		return
	}
	defer reg.Close()
	includeArchived := r.URL.Query().Get("archived") == "1"
	projects, err := reg.List(r.Context(), includeArchived)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errBody(err))
		return
	}
	writeJSON(w, http.StatusOK, projects)
}

func (h *workspaceHandlers) addProject(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Path string `json:"path"`
		Name string `json:"name"`
		Slug string `json:"slug"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || strings.TrimSpace(body.Path) == "" {
		writeJSON(w, http.StatusBadRequest, errBody(errors.New("payload requires {path}")))
		return
	}
	reg, err := h.openRegistry()
	if err != nil {
		writeJSON(w, http.StatusServiceUnavailable, errBody(err))
		return
	}
	defer reg.Close()
	p, err := reg.Add(r.Context(), body.Path, body.Name, body.Slug)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errBody(err))
		return
	}
	writeJSON(w, http.StatusCreated, p)
}

func (h *workspaceHandlers) switchActive(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		Ref string `json:"ref"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, errBody(err))
		return
	}
	reg, err := h.openRegistry()
	if err != nil {
		writeJSON(w, http.StatusServiceUnavailable, errBody(err))
		return
	}
	defer reg.Close()
	p, err := reg.SetActive(r.Context(), body.Ref)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errBody(err))
		return
	}
	writeJSON(w, http.StatusOK, p)
}

func (h *workspaceHandlers) current(w http.ResponseWriter, r *http.Request) {
	reg, err := h.openRegistry()
	if err != nil {
		writeJSON(w, http.StatusServiceUnavailable, errBody(err))
		return
	}
	defer reg.Close()
	res, err := workspace.Resolve(r.Context(), reg, workspace.ResolveOptions{
		Flag: r.URL.Query().Get("project"),
	})
	if err != nil {
		writeJSON(w, http.StatusNotFound, errBody(err))
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"project": res.Project,
		"source":  res.Source,
	})
}
