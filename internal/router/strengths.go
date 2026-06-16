// SPDX-License-Identifier: MIT

// Package router scores LLM adapters against task tags so harness can
// pick the best CLI per sub-task instead of pinning one adapter for the
// whole prompt. The scoring is fully deterministic: intersection of
// tags / len(task.tags), ties broken by adapter cost.
package router

import (
	"sort"

	"github.com/ropeixoto/harnessx/internal/agents"
)

// Choice is one adapter ranked for a task.
type Choice struct {
	AdapterID string
	Score     float64 // 0..1
	Strengths []string
}

// Match scores every adapter in reg against the task tags. Returns the
// ranked list (highest score first). Empty when no adapter has any
// overlap with the task.
func Match(taskTags []string, reg *agents.Registry) []Choice {
	if reg == nil || len(taskTags) == 0 {
		return nil
	}
	var choices []Choice
	for _, id := range reg.IDs() {
		a, ok := reg.Get(id)
		if !ok {
			continue
		}
		caps := a.Capabilities()
		score := scoreOverlap(taskTags, caps.Strengths)
		if score == 0 {
			continue
		}
		choices = append(choices, Choice{
			AdapterID: id,
			Score:     score,
			Strengths: caps.Strengths,
		})
	}
	sort.SliceStable(choices, func(i, j int) bool {
		if choices[i].Score != choices[j].Score {
			return choices[i].Score > choices[j].Score
		}
		return choices[i].AdapterID < choices[j].AdapterID
	})
	return choices
}

// Pick returns the top choice or empty when no adapter matches.
func Pick(taskTags []string, reg *agents.Registry) (Choice, bool) {
	all := Match(taskTags, reg)
	if len(all) == 0 {
		return Choice{}, false
	}
	return all[0], true
}

func scoreOverlap(taskTags, adapterStrengths []string) float64 {
	if len(taskTags) == 0 {
		return 0
	}
	set := map[string]bool{}
	for _, s := range adapterStrengths {
		set[s] = true
	}
	hits := 0
	for _, t := range taskTags {
		if set[t] {
			hits++
		}
	}
	return float64(hits) / float64(len(taskTags))
}
