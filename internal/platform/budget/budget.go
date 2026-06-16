// SPDX-License-Identifier: MIT

// Package budget enforces per-task USD ceilings. Callers ask the Guard
// whether a planned cost is allowed; on excess the Guard records the
// breach and returns a sentinel error.
package budget

import (
	"errors"
	"fmt"
	"sync"
)

var ErrBudgetExceeded = errors.New("budget: ceiling exceeded")

type Guard struct {
	mu      sync.Mutex
	cap     float64
	spent   float64
	entries []Entry
}

// Entry records one charge against the budget. Lets callers render a
// breakdown table at the end of a run instead of the single mysterious
// remaining-line.
type Entry struct {
	Label string // e.g. "claude-sonnet-4-6", "plan", "loop attempt #2"
	USD   float64
	Tag   string // e.g. "reported", "estimated"
	Note  string // free-form (tokens in/out, mode, etc.)
}

func New(usd float64) *Guard {
	if usd <= 0 {
		usd = 1.0
	}
	return &Guard{cap: usd}
}

func (g *Guard) Remaining() float64 {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.cap - g.spent
}

func (g *Guard) Spent() float64 {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.spent
}

// Charge records spend and returns ErrBudgetExceeded if the total would
// exceed the cap. The spend is recorded regardless so reports show the
// actual breach amount.
func (g *Guard) Charge(usd float64) error {
	return g.ChargeWith(Entry{Label: "uncategorised", USD: usd})
}

func (g *Guard) ChargeWith(e Entry) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.spent += e.USD
	g.entries = append(g.entries, e)
	if g.spent > g.cap {
		return fmt.Errorf("%w: spent $%.4f of $%.4f", ErrBudgetExceeded, g.spent, g.cap)
	}
	return nil
}

func (g *Guard) Entries() []Entry {
	g.mu.Lock()
	defer g.mu.Unlock()
	out := make([]Entry, len(g.entries))
	copy(out, g.entries)
	return out
}

func (g *Guard) Cap() float64 {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.cap
}
