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

func TestRenderIncludesCurrentClaudeModels(t *testing.T) {
	want := []string{"claude-opus-4-7-1m", "claude-opus-4-6-fast", "claude-sonnet-4-6", "claude-sonnet-4-6-1m", "claude-haiku-4-5"}
	var buf bytes.Buffer
	_ = Render(&buf, Result{ContextTokens: 1, Files: 1, Duration: time.Millisecond}, Options{OutputTokens: 1})
	for _, w := range want {
		if !strings.Contains(buf.String(), w) {
			t.Errorf("expected %s in output", w)
		}
	}
}

func TestRenderShowsEffortTier(t *testing.T) {
	var buf bytes.Buffer
	_ = Render(&buf, Result{ContextTokens: 100, Files: 1, Duration: time.Millisecond}, Options{OutputTokens: 1000, Effort: EffortHigh})
	out := buf.String()
	if !strings.Contains(out, "effort tier:") {
		t.Fatalf("missing effort tier label: %s", out)
	}
	if !strings.Contains(out, "high") {
		t.Fatalf("effort should mention high: %s", out)
	}
	if !strings.Contains(out, "× 4.0") && !strings.Contains(out, "x 4.0") {
		t.Fatalf("high effort multiplier should be 4.0, got: %s", out)
	}
}

func TestRenderShowsFormula(t *testing.T) {
	var buf bytes.Buffer
	_ = Render(&buf, Result{ContextTokens: 1, Files: 1, Duration: time.Millisecond}, Options{OutputTokens: 1})
	if !strings.Contains(buf.String(), "Formula:") {
		t.Fatalf("formula must be printed for transparency, got: %s", buf.String())
	}
	if !strings.Contains(buf.String(), "in_tokens/1M") {
		t.Fatalf("formula must show per-Mtok division, got: %s", buf.String())
	}
}

func TestRenderCriteriaTable(t *testing.T) {
	var buf bytes.Buffer
	_ = Render(&buf, Result{ContextTokens: 1, Files: 1, Duration: time.Millisecond}, Options{OutputTokens: 1, ShowCriteria: true})
	out := buf.String()
	for _, want := range []string{"feature", "planning", "security", "kimi-k2", "WHY"} {
		if !strings.Contains(out, want) {
			t.Errorf("criteria table missing %q\noutput:\n%s", want, out)
		}
	}
}

func TestRenderUsesCustomModels(t *testing.T) {
	var buf bytes.Buffer
	custom := []ModelPrice{{Model: "fake-model", InputUSDPerM: 1.0, OutputUSDPerM: 2.0}}
	err := Render(&buf, Result{ContextTokens: 1_000_000, Files: 1, Duration: time.Millisecond}, Options{OutputTokens: 500_000, Models: custom, Effort: EffortLow})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "fake-model") {
		t.Fatalf("custom model missing: %s", buf.String())
	}
	if !strings.Contains(buf.String(), "$2.0000") {
		t.Fatalf("expected $2 total at effort=low (1M*$1 + 500k*$2), got: %s", buf.String())
	}
}

func TestProfileForUnknownReturnsMedium(t *testing.T) {
	p := profileFor("nope")
	if p.Level != EffortMedium {
		t.Errorf("unknown effort must fall back to medium, got %s", p.Level)
	}
}

func TestRecommendationForKnownTasks(t *testing.T) {
	for _, want := range []TaskKind{TaskFeature, TaskPlanning, TaskSecurity, TaskCheapReview, TaskCodebaseScan} {
		if _, ok := RecommendationFor(want); !ok {
			t.Errorf("missing recommendation for %s", want)
		}
	}
}
