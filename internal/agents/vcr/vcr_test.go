// SPDX-License-Identifier: MIT

package vcr

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ropeixoto/harnessx/internal/agents"
)

type stubAdapter struct {
	id     string
	calls  int32
	result agents.AgentResult
}

func (s *stubAdapter) ID() string                        { return s.id }
func (s *stubAdapter) Name() string                      { return s.id }
func (s *stubAdapter) Capabilities() agents.Capabilities { return agents.Capabilities{} }
func (s *stubAdapter) Healthcheck(context.Context) agents.HealthcheckResult {
	return agents.HealthcheckResult{OK: true}
}
func (s *stubAdapter) Run(context.Context, agents.AgentRequest) agents.AgentResult {
	atomic.AddInt32(&s.calls, 1)
	return s.result
}
func (s *stubAdapter) ParseUsage(agents.AgentOutput) agents.Usage { return agents.Usage{} }
func (s *stubAdapter) ClassifyFailure(agents.AgentOutput, int, error) agents.FailureType {
	return agents.FailureNone
}

func TestAutoRecordsThenReplays(t *testing.T) {
	dir := t.TempDir()
	stub := &stubAdapter{id: "claude", result: agents.AgentResult{
		Output:   agents.AgentOutput{Stdout: []byte("hi"), FinalMessage: "hello"},
		ExitCode: 0, Duration: 12 * time.Millisecond,
		Usage: agents.Usage{InputTokens: 10, OutputTokens: 5, EstimatedCostUSD: 0.01},
	}}
	a := New(Options{Inner: stub, Dir: dir, Mode: ModeAuto})

	req := agents.AgentRequest{Prompt: "x", Model: "m"}
	first := a.Run(context.Background(), req)
	if first.Output.FinalMessage != "hello" {
		t.Errorf("first call: %+v", first.Output)
	}
	if atomic.LoadInt32(&stub.calls) != 1 {
		t.Errorf("inner should be called once, got %d", stub.calls)
	}

	second := a.Run(context.Background(), req)
	if second.Output.FinalMessage != "hello" || second.ExitCode != 0 {
		t.Errorf("replay differs: %+v", second)
	}
	if atomic.LoadInt32(&stub.calls) != 1 {
		t.Errorf("replay should not re-invoke inner, got %d", stub.calls)
	}
	if second.Duration != 12*time.Millisecond {
		t.Errorf("replay duration lost: %s", second.Duration)
	}
}

func TestReplayMissingErrors(t *testing.T) {
	a := New(Options{Inner: &stubAdapter{id: "claude"}, Dir: t.TempDir(), Mode: ModeReplay})
	res := a.Run(context.Background(), agents.AgentRequest{Prompt: "absent"})
	if res.Err == nil {
		t.Fatal("expected error")
	}
	if !errorContains(res.Err, "replay missing") {
		t.Errorf("want replay-missing err, got %v", res.Err)
	}
}

func TestRecordModeAlwaysRewrites(t *testing.T) {
	dir := t.TempDir()
	stub := &stubAdapter{id: "claude", result: agents.AgentResult{
		Output: agents.AgentOutput{FinalMessage: "v1"},
	}}
	a := New(Options{Inner: stub, Dir: dir, Mode: ModeRecord})
	req := agents.AgentRequest{Prompt: "p"}

	a.Run(context.Background(), req)
	stub.result.Output.FinalMessage = "v2"
	res := a.Run(context.Background(), req)
	if res.Output.FinalMessage != "v2" {
		t.Errorf("ModeRecord should rewrite, got %q", res.Output.FinalMessage)
	}
	if atomic.LoadInt32(&stub.calls) != 2 {
		t.Errorf("inner expected 2 calls, got %d", stub.calls)
	}
}

func TestFingerprintStableForSameRequest(t *testing.T) {
	r := agents.AgentRequest{Prompt: "hello", Model: "m", WorkingDir: "/path/to/p"}
	a := fingerprint(r, "claude")
	b := fingerprint(r, "claude")
	if a != b {
		t.Errorf("fingerprint not stable: %s vs %s", a, b)
	}
	r.Prompt = "hello "
	if fingerprint(r, "claude") == a {
		t.Error("prompt change must change fingerprint")
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "subdir", "cassette.json")
	original := cassette{
		Fingerprint: "abc", AdapterID: "claude",
		RecordedAt: time.Now().UTC().Truncate(time.Second),
		ExitCode:   1, DurationMS: 42,
		Stdout: []byte("out"), Stderr: []byte("err"),
		FinalMessage: "fm", ErrMessage: "boom",
	}
	if err := save(path, original); err != nil {
		t.Fatal(err)
	}
	got, err := load(path)
	if err != nil {
		t.Fatal(err)
	}
	if got.Fingerprint != "abc" || got.ExitCode != 1 || got.DurationMS != 42 ||
		string(got.Stdout) != "out" || got.FinalMessage != "fm" || got.ErrMessage != "boom" {
		t.Errorf("round-trip mismatch: %+v", got)
	}
}

func TestCassetteToResultPreservesError(t *testing.T) {
	c := cassette{ErrMessage: "kaboom", ExitCode: 7}
	res := c.toResult()
	if res.Err == nil || res.Err.Error() != "kaboom" {
		t.Errorf("err not restored: %v", res.Err)
	}
	if res.ExitCode != 7 {
		t.Errorf("exit code not restored: %d", res.ExitCode)
	}
}

func TestAdapterDelegatesParseUsageAndClassify(t *testing.T) {
	stub := &stubAdapter{id: "claude"}
	a := New(Options{Inner: stub, Dir: t.TempDir(), Mode: ModeAuto})
	if _ = a.ParseUsage(agents.AgentOutput{}); a.Name() != "vcr(claude)" {
		t.Errorf("Name wrap broken: %s", a.Name())
	}
	if a.ClassifyFailure(agents.AgentOutput{}, 0, nil) != agents.FailureNone {
		t.Error("classify delegation broken")
	}
}

func TestRecordedAtIsRecent(t *testing.T) {
	dir := t.TempDir()
	stub := &stubAdapter{id: "claude"}
	a := New(Options{Inner: stub, Dir: dir, Mode: ModeAuto})
	a.Run(context.Background(), agents.AgentRequest{Prompt: "p"})
	files, _ := os.ReadDir(dir)
	if len(files) != 1 {
		t.Fatalf("want 1 cassette, got %d", len(files))
	}
}

func errorContains(err error, want string) bool {
	if err == nil {
		return false
	}
	return errors.Unwrap(err) != nil || strContains(err.Error(), want)
}

func strContains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || (len(s) > len(sub) && (indexOf(s, sub) >= 0)))
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
