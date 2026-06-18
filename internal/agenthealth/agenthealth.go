// SPDX-License-Identifier: MIT

package agenthealth

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ropeixoto/harnessx/internal/agents"
)

type Status struct {
	AgentID    string
	OK         bool
	Detail     string
	CLIVersion string
	Err        string
	CheckedAt  time.Time
}

type Probe struct {
	id      string
	adapter agents.AgentAdapter
	tick    time.Duration

	mu      sync.RWMutex
	current Status

	running atomic.Bool
	done    chan struct{}
	cancel  context.CancelFunc
}

func New(adapter agents.AgentAdapter, interval time.Duration) *Probe {
	if interval <= 0 {
		interval = 30 * time.Second
	}
	id := ""
	if adapter != nil {
		id = adapter.ID()
	}
	return &Probe{
		id:      id,
		adapter: adapter,
		tick:    interval,
		done:    make(chan struct{}),
	}
}

func (p *Probe) Start(ctx context.Context) {
	if p == nil || p.adapter == nil {
		return
	}
	if !p.running.CompareAndSwap(false, true) {
		return
	}
	cctx, cancel := context.WithCancel(ctx)
	p.cancel = cancel
	p.runOnce(cctx)
	go p.loop(cctx)
}

func (p *Probe) Stop() {
	if p == nil {
		return
	}
	if !p.running.Load() {
		return
	}
	if p.cancel != nil {
		p.cancel()
	}
	<-p.done
}

func (p *Probe) Snapshot() Status {
	if p == nil {
		return Status{}
	}
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.current
}

func (p *Probe) loop(ctx context.Context) {
	defer close(p.done)
	defer p.running.Store(false)
	t := time.NewTicker(p.tick)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			p.runOnce(ctx)
		}
	}
}

func (p *Probe) runOnce(ctx context.Context) {
	cctx, cancel := context.WithTimeout(ctx, p.tick/2)
	defer cancel()
	res := p.adapter.Healthcheck(cctx)
	p.mu.Lock()
	p.current = Status{
		AgentID:    p.id,
		OK:         res.OK,
		Detail:     res.Detail,
		CLIVersion: res.CLIVersion,
		Err:        res.Err,
		CheckedAt:  time.Now().UTC(),
	}
	p.mu.Unlock()
}

func Badge(s Status, plain bool) string {
	if s.AgentID == "" {
		return ""
	}
	if plain {
		if s.OK {
			return "|" + s.AgentID + " ok"
		}
		return "|" + s.AgentID + " degraded"
	}
	mark := "\033[32m✓\033[0m"
	if !s.OK {
		mark = "\033[33m⚠\033[0m"
	}
	return "|" + s.AgentID + " " + mark
}
