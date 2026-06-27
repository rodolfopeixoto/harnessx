package repl

import (
	"bytes"
	"strings"
	"testing"
)

func TestSwitchModelEmptyError(t *testing.T) {
	var buf bytes.Buffer
	opts := &Options{Out: &buf}
	switchModel(opts, "")
	if !strings.Contains(buf.String(), "/model needs a name") {
		t.Fatalf("missing usage message: %s", buf.String())
	}
}

func TestSwitchModelSetsField(t *testing.T) {
	var buf bytes.Buffer
	opts := &Options{Out: &buf}
	switchModel(opts, "claude-sonnet-4-6")
	if opts.Model != "claude-sonnet-4-6" {
		t.Fatalf("model not stored, got %q", opts.Model)
	}
	if !strings.Contains(buf.String(), "claude-sonnet-4-6") {
		t.Fatalf("output should mention model: %s", buf.String())
	}
}

func TestPrintModelDefault(t *testing.T) {
	var buf bytes.Buffer
	opts := &Options{Out: &buf}
	printModel(opts)
	if !strings.Contains(buf.String(), "adapter default") {
		t.Fatalf("default label missing: %s", buf.String())
	}
}

func TestPrintModelExplicit(t *testing.T) {
	var buf bytes.Buffer
	opts := &Options{Out: &buf, Model: "gpt-5-mini"}
	printModel(opts)
	if !strings.Contains(buf.String(), "gpt-5-mini") {
		t.Fatalf("expected gpt-5-mini in output, got: %s", buf.String())
	}
}

func TestOptionsRouteEnabledDefault(t *testing.T) {
	o := Options{}
	if o.RouteEnabled {
		t.Fatal("RouteEnabled should default to false")
	}
	if o.OneShot {
		t.Fatal("OneShot should default to false")
	}
}
