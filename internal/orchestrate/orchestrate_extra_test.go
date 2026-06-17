package orchestrate

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestListEmptyDir(t *testing.T) {
	names, err := List(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if names != nil {
		t.Errorf("want nil, got %v", names)
	}
}

func TestListReturnsSortedYAMLs(t *testing.T) {
	dir := t.TempDir()
	d := filepath.Join(dir, ".harness", "orchestrations")
	_ = os.MkdirAll(d, 0o755)
	for _, n := range []string{"zeta.yaml", "alpha.yml", "ignore.txt"} {
		_ = os.WriteFile(filepath.Join(d, n), []byte("name: x\ntopology: chain\nsteps:\n  - role: coder\n    command: [true]\n"), 0o644)
	}
	names, err := List(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(names) != 2 || names[0] != "alpha" || names[1] != "zeta" {
		t.Errorf("got %v", names)
	}
}

func TestLoadMissingFlowErrors(t *testing.T) {
	_, err := Load(t.TempDir(), "ghost")
	if err == nil {
		t.Fatal("want error")
	}
}

func TestLoadFileBadYAMLErrors(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "x.yaml")
	_ = os.WriteFile(p, []byte(":::bad:::"), 0o644)
	if _, err := LoadFile(p); err == nil {
		t.Fatal("want error")
	}
}

func TestLoadFileInvalidValidationFails(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "x.yaml")
	_ = os.WriteFile(p, []byte("name: \ntopology: chain\nsteps:\n"), 0o644)
	if _, err := LoadFile(p); err == nil {
		t.Fatal("want validation error")
	}
}

func TestRunCyclicLoopsMaxCycles(t *testing.T) {
	flow := Flow{
		Name: "c", Topology: TopologyCyclic, MaxCycles: 3,
		Steps: []Step{{Role: RoleCoder, Command: []string{"true"}}},
	}
	var buf bytes.Buffer
	res, err := Run(context.Background(), RunOptions{Root: t.TempDir(), Flow: flow}, &buf)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Entries) != 3 {
		t.Errorf("want 3 entries, got %d", len(res.Entries))
	}
}

func TestRunChainStopsOnFailure(t *testing.T) {
	flow := Flow{Name: "c", Topology: TopologyChain, Steps: []Step{
		{Role: RoleCoder, Command: []string{"false"}},
		{Role: RoleTester, Command: []string{"true"}},
	}}
	var buf bytes.Buffer
	res, _ := Run(context.Background(), RunOptions{Root: t.TempDir(), Flow: flow}, &buf)
	if res.OK {
		t.Error("OK should be false")
	}
	if len(res.Entries) != 1 {
		t.Errorf("chain must stop after failure, entries=%d", len(res.Entries))
	}
}

func TestRunPersistsBlackboardJSON(t *testing.T) {
	dir := t.TempDir()
	flow := Flow{Name: "x", Topology: TopologyChain, Steps: []Step{
		{Role: RoleCoder, Command: []string{"echo", "hello"}},
	}}
	var buf bytes.Buffer
	res, _ := Run(context.Background(), RunOptions{Root: dir, Flow: flow}, &buf)
	bb := filepath.Join(dir, ".harness", "artifacts", "runs", res.RunID, "blackboard.json")
	body, err := os.ReadFile(bb)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), "blackboard"[:0]+"coder") {
		t.Errorf("blackboard missing coder role: %s", body)
	}
}

func TestRunValidateFails(t *testing.T) {
	_, err := Run(context.Background(), RunOptions{Root: t.TempDir(), Flow: Flow{}}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("want error")
	}
}

func TestTruncateLong(t *testing.T) {
	long := strings.Repeat("x", 9000)
	got := truncate(long, 100)
	if !strings.HasSuffix(got, "[truncated]") {
		t.Errorf("missing marker: %s", got[len(got)-15:])
	}
}
