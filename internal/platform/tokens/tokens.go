// SPDX-License-Identifier: MIT

// Package tokens estimates LLM token counts. Phase 5 ships a 4-chars-per-token
// heuristic that's good enough for context-pack sizing and budget warnings.
// A provider-specific tokenizer can be plugged in via Estimator in later phases.
package tokens

// Estimator returns the estimated number of tokens that encoding `s` would
// produce. Implementations must be deterministic for caching.
type Estimator interface {
	Estimate(s string) int
}

type Heuristic4 struct{}

func (Heuristic4) Estimate(s string) int {
	if len(s) == 0 {
		return 0
	}
	n := len(s) / 4
	if n == 0 {
		return 1
	}
	return n
}

// EstimateBytes returns the same heuristic for byte slices, avoiding an
// allocation when the caller already has []byte.
func EstimateBytes(b []byte) int {
	if len(b) == 0 {
		return 0
	}
	n := len(b) / 4
	if n == 0 {
		return 1
	}
	return n
}
