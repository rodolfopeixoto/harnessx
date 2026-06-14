// SPDX-License-Identifier: MIT

package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/ropeixoto/harnessx/internal/app/catalogcmd"
	"github.com/ropeixoto/harnessx/internal/catalog"
	"github.com/ropeixoto/harnessx/internal/domain"
)

type catalogHandlers struct{ srv *Server }

func (s *Server) registerCatalog(mux *http.ServeMux) {
	ch := &catalogHandlers{srv: s}
	mux.HandleFunc("/api/catalog/kinds", ch.kinds)
	mux.HandleFunc("/api/catalog/items", ch.items)
	mux.HandleFunc("/api/catalog/plan", ch.plan)
}

func (ch *catalogHandlers) kinds(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, domain.AllCapabilityKinds())
}

func (ch *catalogHandlers) items(w http.ResponseWriter, r *http.Request) {
	kind := domain.CapabilityKind(r.URL.Query().Get("kind"))
	c := catalogcmd.New()
	var (
		caps []domain.Capability
		err  error
	)
	if kind != "" {
		caps, err = c.DiscoverKind(r.Context(), ch.srv.root, kind)
	} else {
		caps, err = c.Discover(r.Context(), ch.srv.root)
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errBody(err))
		return
	}
	writeJSON(w, http.StatusOK, caps)
}

func (ch *catalogHandlers) plan(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		Kind domain.CapabilityKind `json:"kind"`
		Name string                `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Kind == "" || body.Name == "" {
		writeJSON(w, http.StatusBadRequest, errBody(errors.New("payload requires {kind,name}")))
		return
	}
	c := catalogcmd.New()
	ops, err := c.Plan(r.Context(), ch.srv.root, body.Kind, body.Name)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errBody(err))
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"ops":  ops,
		"hash": catalog.HashOps(ops),
	})
}
