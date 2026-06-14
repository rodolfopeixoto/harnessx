// SPDX-License-Identifier: MIT

package sensors

import (
	"context"
	"sort"
)

// Runner executes a slice of sensors in deterministic order:
//  1. computational sensors first (alphabetical by ID)
//  2. inferential sensors next (alphabetical by ID)
//
// A failed sensor never aborts the run — every sensor produces a Result
// so reports can show the full picture. Callers decide what to do with
// failures (e.g. `harness ci` exits non-zero on any failed status).
type Runner struct {
	OnResult func(Result) // optional: per-sensor callback (for streaming UIs)
}

func (r *Runner) Run(ctx context.Context, sensors []Sensor, rc RunCtx) []Result {
	rc.Ctx = ctx
	ordered := orderForExecution(sensors)
	out := make([]Result, 0, len(ordered))
	for _, s := range ordered {
		res := s.Run(rc)
		if res.Category == "" {
			res.Category = s.Category()
		}
		if res.Kind == "" {
			res.Kind = s.Kind()
		}
		if r.OnResult != nil {
			r.OnResult(res)
		}
		out = append(out, res)
		if ctx.Err() != nil {
			break
		}
	}
	return out
}

func orderForExecution(in []Sensor) []Sensor {
	out := make([]Sensor, len(in))
	copy(out, in)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Kind() != out[j].Kind() {
			return out[i].Kind() == KindComputational
		}
		return out[i].ID() < out[j].ID()
	})
	return out
}

// Summary aggregates outcomes for one Run.
type Summary struct {
	Total   int
	Passed  int
	Failed  int
	Skipped int
}

func Summarize(rs []Result) Summary {
	s := Summary{Total: len(rs)}
	for _, r := range rs {
		switch r.Status {
		case StatusPassed:
			s.Passed++
		case StatusFailed:
			s.Failed++
		case StatusSkipped:
			s.Skipped++
		}
	}
	return s
}
