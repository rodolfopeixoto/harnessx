// SPDX-License-Identifier: MIT

package router

import (
	"context"
	"database/sql"
	"sort"
)

// AgentStats summarises historical performance for a single adapter.
type AgentStats struct {
	AgentID     string
	Successes   int
	Failures    int
	TotalRuns   int
	TotalCost   float64
	AvgLatency  int64
	SuccessRate float64
}

// LoadStats reads aggregate per-agent metrics from a sqlite DB. Empty
// result is normal pre-Phase 6 (no executions yet).
func LoadStats(ctx context.Context, db *sql.DB) (map[string]AgentStats, error) {
	if db == nil {
		return nil, nil
	}
	rows, err := db.QueryContext(ctx, `
		select agent,
		       sum(case when status = 'succeeded' then 1 else 0 end) as ok,
		       sum(case when status = 'failed' then 1 else 0 end) as fail,
		       count(*) as total,
		       coalesce(sum(estimated_cost_usd), 0) as cost,
		       coalesce(avg(latency_ms), 0) as latency
		from runs
		where agent is not null and agent != ''
		group by agent`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]AgentStats{}
	for rows.Next() {
		var s AgentStats
		var avgLatency float64
		if err := rows.Scan(&s.AgentID, &s.Successes, &s.Failures, &s.TotalRuns, &s.TotalCost, &avgLatency); err != nil {
			continue
		}
		s.AvgLatency = int64(avgLatency)
		if s.TotalRuns > 0 {
			s.SuccessRate = float64(s.Successes) / float64(s.TotalRuns)
		}
		out[s.AgentID] = s
	}
	return out, rows.Err()
}

// ApplyStats re-orders a RouteConfig's fallback chain by historical
// success rate (desc), keeping the configured primary in place. Agents
// with zero history sort last; ties broken by lower avg latency.
//
// The configured primary is never demoted — operators want predictability
// on the happy path. Fallback ordering is where stats kick in.
func ApplyStats(cfg RouteConfig, stats map[string]AgentStats) RouteConfig {
	if len(stats) == 0 || len(cfg.Fallback) < 2 {
		return cfg
	}
	chain := append([]string{}, cfg.Fallback...)
	sort.SliceStable(chain, func(i, j int) bool {
		si, oki := stats[chain[i]]
		sj, okj := stats[chain[j]]
		// No history: send to the back.
		if !oki && okj {
			return false
		}
		if oki && !okj {
			return true
		}
		if oki && okj {
			if si.SuccessRate != sj.SuccessRate {
				return si.SuccessRate > sj.SuccessRate
			}
			if si.AvgLatency != sj.AvgLatency {
				return si.AvgLatency < sj.AvgLatency
			}
		}
		return chain[i] < chain[j]
	})
	out := cfg
	out.Fallback = chain
	return out
}
