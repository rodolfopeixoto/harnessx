// SPDX-License-Identifier: MIT

package recall

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTokeniseDropsStopWordsAndShorts(t *testing.T) {
	got := tokenise("The Quick fix and a healthz endpoint")
	expected := map[string]bool{"quick": true, "fix": true, "healthz": true, "endpoint": true}
	for w := range expected {
		found := false
		for _, g := range got {
			if g == w {
				found = true
			}
		}
		if !found {
			t.Errorf("tokenise did not yield %q (got %v)", w, got)
		}
	}
	for _, g := range got {
		if g == "the" || g == "and" {
			t.Errorf("tokenise leaked stop word %q", g)
		}
	}
}

func TestScoreReportOverlap(t *testing.T) {
	score, snippet := scoreReport("fixed the /healthz endpoint regression", []string{"healthz", "regression"})
	if score != 1.0 {
		t.Errorf("score: want 1.0, got %v", score)
	}
	if snippet == "" {
		t.Error("expected non-empty snippet")
	}
}

func TestRecallReturnsMatchesAndOrders(t *testing.T) {
	root := t.TempDir()
	runsDir := filepath.Join(root, ".harness", "runs")
	for _, d := range []struct {
		name, body string
	}{
		{"01ONE", "# report\n\nadd healthz endpoint with fastapi"},
		{"01TWO", "# report\n\nunrelated migration work"},
		{"01THR", "# report\n\nfix healthz regression in tests"},
	} {
		dir := filepath.Join(runsDir, d.name)
		_ = os.MkdirAll(dir, 0o755)
		_ = os.WriteFile(filepath.Join(dir, "report.md"), []byte(d.body), 0o644)
	}
	hits, err := Recall(root, "healthz", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) != 2 {
		t.Fatalf("want 2 hits, got %d", len(hits))
	}
	if hits[0].Score < hits[1].Score {
		t.Errorf("hits not sorted desc: %+v", hits)
	}
}

func TestRecallEmptyQuery(t *testing.T) {
	hits, err := Recall(t.TempDir(), "", 5)
	if err != nil {
		t.Fatal(err)
	}
	if hits != nil {
		t.Errorf("empty query should return nil, got %v", hits)
	}
}

func TestRecallNoRuns(t *testing.T) {
	hits, err := Recall(t.TempDir(), "anything", 5)
	if err != nil {
		t.Fatal(err)
	}
	if hits != nil {
		t.Errorf("absent runs dir should return nil, got %v", hits)
	}
}
