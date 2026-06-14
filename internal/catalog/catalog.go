// SPDX-License-Identifier: MIT

// Package catalog is the unified manager for HarnessX capabilities: Agents,
// MCPs, Hooks, Sensors, Skills, Context Providers, Resource Providers and
// Plugins. Discovery is deterministic (filesystem + manifest scan, never
// LLM-driven); installs run via a Plan→approval→Apply pipeline so every
// mutation has a unified diff and an audit trail.
package catalog

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"

	"github.com/ropeixoto/harnessx/internal/domain"
)

// Kind is the per-capability plug-in contract. Each implementation lives in
// internal/catalog/kinds/<kind>.go.
type Kind interface {
	Kind() domain.CapabilityKind
	Discover(ctx context.Context, root string) ([]domain.Capability, error)
	Plan(ctx context.Context, root, name string) ([]domain.FileOp, error)
}

// Catalog wires a set of Kind implementations.
type Catalog struct {
	mu    sync.RWMutex
	kinds map[domain.CapabilityKind]Kind
}

// New returns a Catalog with no kinds registered. Callers register the kinds
// they care about (or use NewDefault to register every bundled one).
func New() *Catalog {
	return &Catalog{kinds: map[domain.CapabilityKind]Kind{}}
}

// Register adds a Kind to the Catalog. Duplicate kinds replace silently —
// tests can swap in fakes without rebuilding the Catalog.
func (c *Catalog) Register(k Kind) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.kinds[k.Kind()] = k
}

// Kinds returns registered kinds in deterministic (enum) order.
func (c *Catalog) Kinds() []domain.CapabilityKind {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := []domain.CapabilityKind{}
	for _, k := range domain.AllCapabilityKinds() {
		if _, ok := c.kinds[k]; ok {
			out = append(out, k)
		}
	}
	return out
}

// Discover walks every registered kind and returns the merged + sorted
// capability list. A kind's Discover failure aborts the call.
func (c *Catalog) Discover(ctx context.Context, root string) ([]domain.Capability, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	var out []domain.Capability
	for _, k := range domain.AllCapabilityKinds() {
		impl, ok := c.kinds[k]
		if !ok {
			continue
		}
		caps, err := impl.Discover(ctx, root)
		if err != nil {
			return nil, fmt.Errorf("catalog: discover %s: %w", k, err)
		}
		out = append(out, caps...)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Kind != out[j].Kind {
			return out[i].Kind < out[j].Kind
		}
		return out[i].Name < out[j].Name
	})
	return out, nil
}

// DiscoverKind returns capabilities for a single kind, useful for /api/catalog/items.
func (c *Catalog) DiscoverKind(ctx context.Context, root string, kind domain.CapabilityKind) ([]domain.Capability, error) {
	c.mu.RLock()
	impl, ok := c.kinds[kind]
	c.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("catalog: unknown kind %q", kind)
	}
	return impl.Discover(ctx, root)
}

// Show looks up a capability by kind + name and returns the populated record.
// Returns ErrUnknownCapability when no Kind matches.
func (c *Catalog) Show(ctx context.Context, root string, kind domain.CapabilityKind, name string) (domain.Capability, error) {
	caps, err := c.DiscoverKind(ctx, root, kind)
	if err != nil {
		return domain.Capability{}, err
	}
	for _, cap := range caps {
		if cap.Name == name {
			return cap, nil
		}
	}
	return domain.Capability{}, ErrUnknownCapability
}

// Plan delegates to the matching Kind. Returned ops are absolute paths,
// suitable for direct execution by Apply.
func (c *Catalog) Plan(ctx context.Context, root string, kind domain.CapabilityKind, name string) ([]domain.FileOp, error) {
	c.mu.RLock()
	impl, ok := c.kinds[kind]
	c.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("catalog: unknown kind %q", kind)
	}
	return impl.Plan(ctx, root, name)
}

// ErrUnknownCapability is returned when Show cannot find the requested item.
var ErrUnknownCapability = errors.New("catalog: capability not found")
