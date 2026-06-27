package main

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestRenderCostCompareShowsHarnessZero(t *testing.T) {
	var buf bytes.Buffer
	if err := renderCostCompare(&buf, 10_000, 2_000, 5, 80*time.Millisecond); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "$0.0000") {
		t.Errorf("harness column should be $0.0000, got: %s", out)
	}
	for _, m := range bundledModelPrices {
		if !strings.Contains(out, m.Model) {
			t.Errorf("output should mention model %s, got: %s", m.Model, out)
		}
	}
}

func TestCostCompareCmdAdvertisesDeterministic(t *testing.T) {
	cmd := newCostCompareCmd()
	if !strings.Contains(cmd.Long, "No agent is invoked") {
		t.Errorf("Long should declare deterministic-only: %s", cmd.Long)
	}
	if !strings.Contains(cmd.Long, "0 LLM tokens") {
		t.Errorf("Long should advertise 0 LLM tokens for harness column")
	}
}
