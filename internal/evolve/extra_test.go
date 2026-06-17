package evolve

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSignatureUncategorisedFallback(t *testing.T) {
	got := signature(map[string]any{"random": "x"})
	if got != "uncategorised" {
		t.Errorf("want uncategorised, got %q", got)
	}
}

func TestIsFailureFieldsDetectsErrText(t *testing.T) {
	if !isFailureFields(map[string]any{"err": "boom"}) {
		t.Error("err field must mark failure")
	}
	if !isFailureFields(map[string]any{"error": "boom"}) {
		t.Error("error field must mark failure")
	}
	if isFailureFields(map[string]any{"status": "ok"}) {
		t.Error("status=ok must not mark failure")
	}
}

func TestProposeRoundTrip(t *testing.T) {
	dir := t.TempDir()
	id, err := Propose(dir, Mutation{Component: "router", Description: "x"})
	if err != nil {
		t.Fatal(err)
	}
	body, _ := os.ReadFile(filepath.Join(dir, ".harness", "logs", "mutations.jsonl"))
	if !strings.Contains(string(body), id) {
		t.Errorf("log missing id")
	}
}

func TestPromoteWritesMutation(t *testing.T) {
	dir := t.TempDir()
	if err := Promote(dir, PromoteOptions{MutationID: "abc", HITL: true, Reason: "good"}); err != nil {
		t.Fatal(err)
	}
	body, _ := os.ReadFile(filepath.Join(dir, ".harness", "logs", "mutations.jsonl"))
	if !strings.Contains(string(body), "promoted") {
		t.Errorf("status promoted missing: %s", body)
	}
}

func TestWriteJSONEncodes(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteJSON(&buf, map[string]int{"a": 1}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), `"a": 1`) {
		t.Errorf("not pretty: %s", buf.String())
	}
}

func TestReplayMissingFileErrors(t *testing.T) {
	_, err := Replay(t.TempDir(), "/nonexistent/trace.jsonl", Diagnosis{})
	if err == nil {
		t.Fatal("want error")
	}
}

func TestTruncateShortPassthrough(t *testing.T) {
	if truncate("abc", 100) != "abc" {
		t.Error("short string must pass through")
	}
}

func TestDiagnoseSkipsInvalidLines(t *testing.T) {
	dir := t.TempDir()
	writeEvents(t, dir, []string{
		"not json",
		`{"level":"error","fields":{"status":"failed","stage":"x"}}`,
	})
	d, err := Diagnose(dir)
	if err != nil {
		t.Fatal(err)
	}
	if d.Failures != 1 {
		t.Errorf("want 1 failure, got %d", d.Failures)
	}
}
