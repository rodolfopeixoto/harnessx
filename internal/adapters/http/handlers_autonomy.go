// SPDX-License-Identifier: MIT

package http

import (
	"net/http"

	"github.com/ropeixoto/harnessx/internal/autonomy"
	"github.com/ropeixoto/harnessx/internal/health"
)

func (s *Server) registerAutonomy(mux *http.ServeMux) {
	mux.HandleFunc("/api/autonomy", func(w http.ResponseWriter, _ *http.Request) {
		ops := []autonomy.Operation{autonomy.OpRead, autonomy.OpPlan, autonomy.OpExecuteLowRisk, autonomy.OpExecuteHighRisk, autonomy.OpClean, autonomy.OpSchedule}
		levels := autonomy.AllLevels()
		out := map[string]map[string]autonomy.Decision{}
		for _, lvl := range levels {
			out[string(lvl)] = map[string]autonomy.Decision{}
			for _, op := range ops {
				dec, _ := autonomy.Gate(lvl, op)
				out[string(lvl)][string(op)] = dec
			}
		}
		writeJSON(w, http.StatusOK, out)
	})
	mux.HandleFunc("/api/health/score", func(w http.ResponseWriter, _ *http.Request) {
		score := health.Inputs{
			TestsPassPct: 100, SensorsPassPct: 100, SecurityFindings: 0,
			PerfBudgetExceeded: false, OutdatedDeps: 1, DocsCoverage: 70,
			DesignParityPct: 80, RoadmapClearPct: 60, MemoryFreshDays: 10, InvalidConfigs: 0,
		}.Compute()
		writeJSON(w, http.StatusOK, score)
	})
}
