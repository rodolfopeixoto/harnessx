package orchestrate

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestValidateMissingName(t *testing.T) {
	if err := (Flow{Topology: TopologyChain, Steps: []Step{{Role: RoleCoder, Command: []string{"x"}}}}).Validate(); err == nil {
		t.Fatal("want error")
	}
}

func TestValidateInvalidTopology(t *testing.T) {
	if err := (Flow{Name: "x", Topology: "weird", Steps: []Step{{Role: RoleCoder, Command: []string{"x"}}}}).Validate(); err == nil {
		t.Fatal("want error")
	}
}

func TestValidateStepWithoutCommandOrAdapter(t *testing.T) {
	f := Flow{Name: "x", Topology: TopologyChain, Steps: []Step{{Role: RoleCoder}}}
	if err := f.Validate(); err == nil {
		t.Fatal("want error")
	}
}

func TestLoadFileMissingFile(t *testing.T) {
	if _, err := LoadFile(filepath.Join(t.TempDir(), "nope.yaml")); err == nil {
		t.Fatal("want error")
	}
}

func TestListPropagatesReadDirError(t *testing.T) {
	dir := t.TempDir()
	_ = os.MkdirAll(filepath.Join(dir, ".harness"), 0o755)
	path := filepath.Join(dir, ".harness", "orchestrations")
	_ = os.WriteFile(path, []byte("not a directory"), 0o644)
	if _, err := List(dir); err == nil {
		t.Fatal("want error")
	}
}

func TestRunWritesAdapterStepStub(t *testing.T) {
	dir := t.TempDir()
	flow := Flow{Name: "x", Topology: TopologyChain, Steps: []Step{
		{Role: RoleCoder, Adapter: "ollama"},
	}}
	var buf bytes.Buffer
	res, err := Run(context.Background(), RunOptions{Root: dir, Flow: flow}, &buf)
	if err != nil {
		t.Fatal(err)
	}
	if res.Entries[0].Status != "adapter-step-not-executed-yet" {
		t.Errorf("want stub status, got %q", res.Entries[0].Status)
	}
}
