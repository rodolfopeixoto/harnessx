package orchestrate

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/ropeixoto/harnessx/internal/agents"
)

type fakeAdapter struct {
	id     string
	out    string
	stdout string
	err    error
}

func (f fakeAdapter) ID() string                        { return f.id }
func (f fakeAdapter) Name() string                      { return f.id }
func (f fakeAdapter) Capabilities() agents.Capabilities { return agents.Capabilities{} }
func (f fakeAdapter) Healthcheck(ctx context.Context) agents.HealthcheckResult {
	return agents.HealthcheckResult{OK: true}
}
func (f fakeAdapter) ParseUsage(o agents.AgentOutput) agents.Usage { return agents.Usage{} }
func (f fakeAdapter) ClassifyFailure(o agents.AgentOutput, c int, e error) agents.FailureType {
	return agents.FailureNone
}
func (f fakeAdapter) Run(ctx context.Context, req agents.AgentRequest) agents.AgentResult {
	return agents.AgentResult{
		Output: agents.AgentOutput{FinalMessage: f.out, Stdout: []byte(f.stdout)},
		Err:    f.err,
	}
}

func TestAdapterRunnerErrorsWithoutAdapter(t *testing.T) {
	r := NewAdapterRunner(agents.NewRegistry(), "/", time.Second)
	_, err := r(context.Background(), Step{}, nil)
	if err == nil {
		t.Fatal("want error for empty adapter id")
	}
}

func TestAdapterRunnerErrorsWhenAdapterMissing(t *testing.T) {
	r := NewAdapterRunner(agents.NewRegistry(), "/", time.Second)
	_, err := r(context.Background(), Step{Adapter: "missing", Role: RoleCoder}, nil)
	if err == nil {
		t.Fatal("want error")
	}
}

func TestAdapterRunnerReturnsFinalMessage(t *testing.T) {
	reg := agents.NewRegistry()
	_ = reg.Register(fakeAdapter{id: "ok", out: "patch applied"})
	r := NewAdapterRunner(reg, "/", time.Second)
	got, err := r(context.Background(), Step{Adapter: "ok", Role: RoleCoder, Prompt: "do x"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got != "patch applied" {
		t.Errorf("got %q", got)
	}
}

func TestAdapterRunnerFallsBackToStdout(t *testing.T) {
	reg := agents.NewRegistry()
	_ = reg.Register(fakeAdapter{id: "ok", stdout: "raw output"})
	r := NewAdapterRunner(reg, "/", time.Second)
	got, _ := r(context.Background(), Step{Adapter: "ok", Role: RoleCoder}, nil)
	if got != "raw output" {
		t.Errorf("got %q", got)
	}
}

func TestAdapterRunnerPropagatesAdapterErr(t *testing.T) {
	reg := agents.NewRegistry()
	_ = reg.Register(fakeAdapter{id: "boom", err: errors.New("net")})
	r := NewAdapterRunner(reg, "/", time.Second)
	_, err := r(context.Background(), Step{Adapter: "boom", Role: RoleCoder}, nil)
	if err == nil {
		t.Fatal("want error")
	}
}

func TestBuildRolePromptIncludesRecentBlackboard(t *testing.T) {
	prev := []BlackboardEntry{
		{Step: 0, Role: RoleCoder, Stdout: "first"},
		{Step: 1, Role: RoleTester, Stdout: "second"},
		{Step: 2, Role: RoleReviewer, Stdout: "third"},
	}
	got := buildRolePrompt(Step{Role: RoleManager, Prompt: "decide"}, prev)
	if !strings.Contains(got, "manager") {
		t.Errorf("missing role: %s", got)
	}
	if !strings.Contains(got, "third") {
		t.Errorf("missing latest entry: %s", got)
	}
}

func TestRunAdapterStepExecutesViaRunner(t *testing.T) {
	dir := t.TempDir()
	calls := 0
	runner := AdapterRunner(func(ctx context.Context, step Step, prev []BlackboardEntry) (string, error) {
		calls++
		return "from-adapter", nil
	})
	flow := Flow{Name: "x", Topology: TopologyChain, Steps: []Step{
		{Role: RoleCoder, Adapter: "ollama"},
	}}
	var buf bytes.Buffer
	res, err := Run(context.Background(), RunOptions{Root: dir, Flow: flow, AdapterRunner: runner}, &buf)
	if err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Errorf("runner called %d times", calls)
	}
	if res.Entries[0].Status != "ok" {
		t.Errorf("status: %s", res.Entries[0].Status)
	}
	if !strings.Contains(res.Entries[0].Stdout, "from-adapter") {
		t.Errorf("stdout: %s", res.Entries[0].Stdout)
	}
}

func TestRunAdapterStepRecordsFailure(t *testing.T) {
	dir := t.TempDir()
	runner := AdapterRunner(func(ctx context.Context, step Step, prev []BlackboardEntry) (string, error) {
		return "", errors.New("adapter down")
	})
	flow := Flow{Name: "x", Topology: TopologyChain, Steps: []Step{
		{Role: RoleCoder, Adapter: "x"},
	}}
	var buf bytes.Buffer
	res, _ := Run(context.Background(), RunOptions{Root: dir, Flow: flow, AdapterRunner: runner}, &buf)
	if res.OK {
		t.Error("OK should be false")
	}
}
