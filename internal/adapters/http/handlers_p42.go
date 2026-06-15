// SPDX-License-Identifier: MIT

package http

import (
	"net/http"

	"github.com/ropeixoto/harnessx/internal/install"
	"github.com/ropeixoto/harnessx/internal/runtime/containers"
	"github.com/ropeixoto/harnessx/internal/secrets"
)

func (s *Server) registerP42(mux *http.ServeMux) {
	mux.HandleFunc("/api/runtime", s.runtimeInfo)
	mux.HandleFunc("/api/runtimes", s.runtimesList)
	mux.HandleFunc("/api/containers", s.containersList)
	mux.HandleFunc("/api/images", s.imagesList)
	mux.HandleFunc("/api/install", s.installList)
	mux.HandleFunc("/api/secrets/names", s.secretsNames)
}

func (s *Server) runtimeInfo(w http.ResponseWriter, r *http.Request) {
	rt, source, err := containers.Resolve(r.Context(), s.root)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{"runtime": nil, "source": "none", "error": err.Error()})
		return
	}
	v, _ := rt.Version(r.Context())
	writeJSON(w, http.StatusOK, map[string]any{
		"runtime": rt.ID(), "binary": rt.Binary(), "version": v, "source": source,
	})
}

func (s *Server) runtimesList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	all := containers.DetectIncluding(ctx, true)
	cfg, _ := containers.LoadConfig(s.root)
	type row struct {
		ID        string `json:"id"`
		Binary    string `json:"binary"`
		Available bool   `json:"available"`
		Version   string `json:"version,omitempty"`
		Selected  bool   `json:"selected"`
	}
	rows := make([]row, 0, len(all))
	for _, rt := range all {
		ok := rt.Available(ctx)
		ver := ""
		if ok {
			if v, err := rt.Version(ctx); err == nil {
				ver = v
			}
		}
		rows = append(rows, row{ID: rt.ID(), Binary: rt.Binary(), Available: ok, Version: ver, Selected: rt.ID() == cfg.Runtime})
	}
	writeJSON(w, http.StatusOK, rows)
}

func (s *Server) containersList(w http.ResponseWriter, r *http.Request) {
	rt, _, err := containers.Resolve(r.Context(), s.root)
	if err != nil {
		writeJSON(w, http.StatusOK, []any{})
		return
	}
	all := r.URL.Query().Get("all") == "true"
	list, err := rt.List(r.Context(), containers.ListOptions{All: all})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errBody(err))
		return
	}
	if list == nil {
		list = []containers.Container{}
	}
	writeJSON(w, http.StatusOK, list)
}

func (s *Server) imagesList(w http.ResponseWriter, r *http.Request) {
	rt, _, err := containers.Resolve(r.Context(), s.root)
	if err != nil {
		writeJSON(w, http.StatusOK, []any{})
		return
	}
	images, err := rt.ListImages(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errBody(err))
		return
	}
	if images == nil {
		images = []containers.Image{}
	}
	writeJSON(w, http.StatusOK, images)
}

func (s *Server) installList(w http.ResponseWriter, r *http.Request) {
	names, err := install.ListBundled()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errBody(err))
		return
	}
	type row struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Category    string `json:"category"`
		Binary      string `json:"binary"`
		Installed   bool   `json:"installed"`
	}
	rows := make([]row, 0, len(names))
	for _, n := range names {
		m, err := install.LoadBundled(n)
		if err != nil {
			continue
		}
		rows = append(rows, row{
			Name:        m.Name,
			Description: m.Description,
			Category:    m.Category,
			Binary:      m.Probe.Binary,
			Installed:   binaryOnPath(m.Probe.Binary),
		})
	}
	writeJSON(w, http.StatusOK, rows)
}

func (s *Server) secretsNames(w http.ResponseWriter, r *http.Request) {
	store := secrets.New()
	perBackend, err := store.List()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errBody(err))
		return
	}
	if perBackend == nil {
		perBackend = map[string][]string{}
	}
	for k, v := range perBackend {
		if v == nil {
			perBackend[k] = []string{}
		}
	}
	for _, b := range store.Backends() {
		if _, ok := perBackend[b.Name()]; !ok {
			perBackend[b.Name()] = []string{}
		}
	}
	writeJSON(w, http.StatusOK, perBackend)
}
