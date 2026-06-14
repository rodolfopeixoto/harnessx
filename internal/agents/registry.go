// SPDX-License-Identifier: MIT

package agents

import (
	"errors"
	"sort"
	"sync"
)

// Registry holds AgentAdapter instances by ID. Safe for concurrent reads;
// writes are protected by mutex.
type Registry struct {
	mu sync.RWMutex
	m  map[string]AgentAdapter
}

func NewRegistry() *Registry {
	return &Registry{m: map[string]AgentAdapter{}}
}

func (r *Registry) Register(a AgentAdapter) error {
	if a == nil || a.ID() == "" {
		return errors.New("registry: nil or unidentified adapter")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.m[a.ID()]; exists {
		return errors.New("registry: duplicate adapter id: " + a.ID())
	}
	r.m[a.ID()] = a
	return nil
}

func (r *Registry) Get(id string) (AgentAdapter, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	a, ok := r.m[id]
	return a, ok
}

func (r *Registry) IDs() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	ids := make([]string, 0, len(r.m))
	for id := range r.m {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

func (r *Registry) All() []AgentAdapter {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]AgentAdapter, 0, len(r.m))
	for _, a := range r.m {
		out = append(out, a)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID() < out[j].ID() })
	return out
}
