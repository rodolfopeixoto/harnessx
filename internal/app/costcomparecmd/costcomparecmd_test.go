package costcomparecmd

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestRenderShowsHarnessZero(t *testing.T) {
	var buf bytes.Buffer
	err := Render(&buf, Result{ContextTokens: 10000, Files: 5, Duration: 80 * time.Millisecond}, Options{OutputTokens: 2000})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "$0.0000") {
		t.Errorf("missing $0 harness row")
	}
	for _, m := range BundledModelPrices {
		if !strings.Contains(buf.String(), m.Model) {
			t.Errorf("model %s missing", m.Model)
		}
	}
}

func TestRenderUsesCustomModels(t *testing.T) {
	var buf bytes.Buffer
	custom := []ModelPrice{{Model: "fake-model", InputUSDPerM: 1.0, OutputUSDPerM: 2.0}}
	err := Render(&buf, Result{ContextTokens: 1_000_000, Files: 1, Duration: time.Millisecond}, Options{OutputTokens: 500_000, Models: custom})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "fake-model") {
		t.Fatalf("custom model missing: %s", buf.String())
	}
	if !strings.Contains(buf.String(), "$2.0000") {
		t.Fatalf("expected $2 total (1M*$1 + 500k*$2), got: %s", buf.String())
	}
}
