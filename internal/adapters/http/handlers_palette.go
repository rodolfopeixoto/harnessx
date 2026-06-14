// SPDX-License-Identifier: MIT

package http

import (
	"net/http"

	"github.com/ropeixoto/harnessx/internal/app/catalogcmd"
	"github.com/ropeixoto/harnessx/internal/palette"
	"github.com/ropeixoto/harnessx/internal/workspace"
)

func (s *Server) registerPalette(mux *http.ServeMux) {
	mux.HandleFunc("/api/palette", s.paletteSearch)
}

func (s *Server) paletteSearch(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	reg, _ := workspace.Open("")
	if reg != nil {
		defer reg.Close()
	}
	cat := catalogcmd.New()
	p := palette.New(
		palette.ProjectsSource{Registry: reg},
		palette.CapabilitiesSource{Catalog: cat, Root: s.root},
		palette.CommandsSource{Commands: palette.BuiltinCommands},
	)
	hits, err := p.Search(r.Context(), q)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errBody(err))
		return
	}
	writeJSON(w, http.StatusOK, hits)
}
