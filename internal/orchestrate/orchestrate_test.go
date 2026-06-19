package orchestrate

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestValidateRejectsUnknownRole(t *testing.T) {
	f := Flow{Name: "x", Topology: TopologyChain, Steps: []Step{{Role: "alien", Command: []string{"true"}}}}
	if err := f.Validate(); err == nil {
		t.Fatal("expected error")
	}
}

func TestValidateRejectsCyclicWithoutMaxCycles(t *testing.T) {
	f := Flow{Name: "x", Topology: TopologyCyclic, Steps: []Step{{Role: RoleCoder, Command: []string{"true"}}}}
	if err := f.Validate(); err == nil {
		t.Fatal("expected error")
	}
}

func TestValidateAcceptsKnownRoles(t *testing.T) {
	for _, role := range KnownRoles() {
		f := Flow{Name: "x", Topology: TopologyChain, Steps: []Step{{Role: role, Command: []string{"true"}}}}
		if err := f.Validate(); err != nil {
			t.Errorf("role %s rejected: %v", role, err)
		}
	}
}

func TestLoadParsesFile(t *testing.T) {
	dir := t.TempDir()
	d := filepath.Join(dir, ".harness", "orchestrations")
	_ = os.MkdirAll(d, 0o755)
	body := `name: review
topology: chain
steps:
  - role: coder
    command: [echo, "wrote code"]
  - role: tester
    command: [echo, "tests pass"]
`
	_ = os.WriteFile(filepath.Join(d, "review.yaml"), []byte(body), 0o644)
	f, err := Load(dir, "review")
	if err != nil {
		t.Fatal(err)
	}
	if len(f.Steps) != 2 || f.Steps[0].Role != RoleCoder {
		t.Errorf("unexpected steps: %+v", f.Steps)
	}
}

func TestRunChainProducesBlackboard(t *testing.T) {
	dir := t.TempDir()
	flow := Flow{Name: "chain", Topology: TopologyChain, Steps: []Step{
		{Role: RoleCoder, Command: []string{"echo", "step1"}},
		{Role: RoleTester, Command: []string{"echo", "step2"}},
	}}
	var buf bytes.Buffer
	res, err := Run(context.Background(), RunOptions{Root: dir, Flow: flow}, &buf)
	if err != nil {
		t.Fatal(err)
	}
	if !res.OK || len(res.Entries) != 2 {
		t.Fatalf("unexpected res: %+v", res)
	}
	bb := filepath.Join(dir, ".harness", "artifacts", "runs", res.RunID, "blackboard.json")
	if _, err := os.Stat(bb); err != nil {
		t.Fatalf("blackboard missing: %v", err)
	}
}

func TestRunStreamsChildStdoutWithRolePrefix(t *testing.T) {
	dir := t.TempDir()
	flow := Flow{Name: "stream", Topology: TopologyChain, Steps: []Step{
		{Role: RoleCoder, Command: []string{"echo", "hello world"}},
		{Role: RoleTester, Command: []string{"echo", "two\nlines"}},
	}}
	var buf bytes.Buffer
	if _, err := Run(context.Background(), RunOptions{Root: dir, Flow: flow}, &buf); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	for _, want := range []string{"  [coder] hello world", "  [tester] two", "  [tester] lines"} {
		if !bytes.Contains([]byte(got), []byte(want)) {
			t.Errorf("missing %q in stream\n--- output ---\n%s", want, got)
		}
	}
}

func TestRunDryDoesNotExecute(t *testing.T) {
	dir := t.TempDir()
	flow := Flow{Name: "x", Topology: TopologyChain, Steps: []Step{
		{Role: RoleCoder, Command: []string{"false"}},
	}}
	var buf bytes.Buffer
	res, err := Run(context.Background(), RunOptions{Root: dir, Flow: flow, DryRun: true}, &buf)
	if err != nil {
		t.Fatal(err)
	}
	if !res.OK {
		t.Errorf("dry-run should not propagate child failure")
	}
}
