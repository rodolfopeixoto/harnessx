package agenthealth

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ropeixoto/harnessx/internal/agents"
)

type fakeAdapter struct {
	id    string
	calls atomic.Int64
	mu    sync.Mutex
	state agents.HealthcheckResult
}

func (f *fakeAdapter) ID() string                        { return f.id }
func (f *fakeAdapter) Name() string                      { return f.id }
func (f *fakeAdapter) Capabilities() agents.Capabilities { return agents.Capabilities{} }
func (f *fakeAdapter) Healthcheck(ctx context.Context) agents.HealthcheckResult {
	f.calls.Add(1)
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.state
}
func (f *fakeAdapter) Run(ctx context.Context, req agents.AgentRequest) agents.AgentResult {
	return agents.AgentResult{}
}
func (f *fakeAdapter) ParseUsage(o agents.AgentOutput) agents.Usage { return agents.Usage{} }
func (f *fakeAdapter) ClassifyFailure(o agents.AgentOutput, c int, e error) agents.FailureType {
	return agents.FailureNone
}

func (f *fakeAdapter) Flip(ok bool, detail string) {
	f.mu.Lock()
	f.state = agents.HealthcheckResult{OK: ok, Detail: detail}
	f.mu.Unlock()
}

func TestProbeStartCallsHealthAndStoresStatus(t *testing.T) {
	a := &fakeAdapter{id: "claude"}
	a.Flip(true, "ready")
	p := New(a, 50*time.Millisecond)
	p.Start(context.Background())
	defer p.Stop()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if a.calls.Load() >= 2 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if a.calls.Load() < 2 {
		t.Fatalf("expected >=2 healthcheck calls, got %d", a.calls.Load())
	}
	s := p.Snapshot()
	if !s.OK || s.AgentID != "claude" {
		t.Errorf("unexpected snapshot: %+v", s)
	}
}

func TestProbeStopBlocksUntilLoopEnds(t *testing.T) {
	a := &fakeAdapter{id: "x"}
	a.Flip(true, "")
	p := New(a, 20*time.Millisecond)
	p.Start(context.Background())
	time.Sleep(40 * time.Millisecond)
	p.Stop()
	if p.running.Load() {
		t.Error("running flag should be cleared after Stop")
	}
}

func TestProbeStartIsIdempotent(t *testing.T) {
	a := &fakeAdapter{id: "x"}
	a.Flip(true, "")
	p := New(a, 20*time.Millisecond)
	p.Start(context.Background())
	p.Start(context.Background())
	defer p.Stop()
	time.Sleep(40 * time.Millisecond)
	if !p.running.Load() {
		t.Error("probe should still be running")
	}
}

func TestProbeStopOnUnstartedIsNoOp(t *testing.T) {
	a := &fakeAdapter{id: "x"}
	p := New(a, 20*time.Millisecond)
	p.Stop()
}

func TestProbeSnapshotConcurrentRace(t *testing.T) {
	a := &fakeAdapter{id: "x"}
	a.Flip(true, "")
	p := New(a, 5*time.Millisecond)
	p.Start(context.Background())
	defer p.Stop()
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			deadline := time.Now().Add(200 * time.Millisecond)
			for time.Now().Before(deadline) {
				_ = p.Snapshot()
			}
		}()
	}
	go func() {
		deadline := time.Now().Add(200 * time.Millisecond)
		for time.Now().Before(deadline) {
			a.Flip(true, "x")
			time.Sleep(time.Millisecond)
			a.Flip(false, "y")
		}
	}()
	wg.Wait()
}

func TestNilProbeMethodsSafe(t *testing.T) {
	var p *Probe
	p.Start(context.Background())
	p.Stop()
	if p.Snapshot().AgentID != "" {
		t.Error("nil snapshot should be empty")
	}
}

func TestProbeSwapRetargetsAdapter(t *testing.T) {
	a := &fakeAdapter{id: "ollama"}
	a.Flip(true, "")
	p := New(a, 20*time.Millisecond)
	p.Start(context.Background())
	defer p.Stop()
	time.Sleep(40 * time.Millisecond)
	if id := p.Snapshot().AgentID; id != "ollama" {
		t.Fatalf("pre-swap snapshot id = %q want ollama", id)
	}
	b := &fakeAdapter{id: "kimi"}
	b.Flip(true, "")
	p.Swap(b)
	if id := p.Snapshot().AgentID; id != "kimi" {
		t.Fatalf("post-swap snapshot id = %q want kimi", id)
	}
}

func TestBadgePlainAndColored(t *testing.T) {
	plain := Badge(Status{AgentID: "claude", OK: true}, true)
	if plain != "|claude ok" {
		t.Errorf("plain: %q", plain)
	}
	col := Badge(Status{AgentID: "claude", OK: false}, false)
	if col == "" {
		t.Error("colored badge should not be empty for set agent")
	}
}

func TestBadgeEmptyWhenNoAgent(t *testing.T) {
	if Badge(Status{}, false) != "" {
		t.Error("no agent → empty badge")
	}
}
