package venvinstall

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestResultOKHappyPath(t *testing.T) {
	r := Result{Strategy: "x", Steps: []StepOutcome{{Err: nil}}}
	if !r.OK() {
		t.Error("want OK")
	}
}

func TestResultOKFailsOnAnyError(t *testing.T) {
	r := Result{Strategy: "x", Steps: []StepOutcome{{Err: nil}, {Err: errors.New("boom")}}}
	if r.OK() {
		t.Error("want not OK")
	}
}

func TestResultOKFailsEmpty(t *testing.T) {
	r := Result{}
	if r.OK() {
		t.Error("empty strategy not OK")
	}
}

func TestDetectInterpreterReturnsPathOrEmpty(t *testing.T) {
	got, _ := DetectInterpreter()
	if got != "" {
		if _, err := os.Stat("/dev/null"); err != nil {
			t.Skip("no fs")
		}
	}
}

func TestRunStrategySucceedsForTrueCommand(t *testing.T) {
	dir := t.TempDir()
	s := Strategy{Name: "noop", Cmds: [][]string{{"true"}}}
	var buf bytes.Buffer
	res, err := runStrategy(context.Background(), dir, &buf, s)
	if err != nil {
		t.Fatal(err)
	}
	if !res.OK() {
		t.Errorf("want OK: %+v", res)
	}
}

func TestRunStrategyFailsForFalseCommand(t *testing.T) {
	dir := t.TempDir()
	s := Strategy{Name: "boom", Cmds: [][]string{{"false"}}}
	var buf bytes.Buffer
	res, err := runStrategy(context.Background(), dir, &buf, s)
	if err == nil {
		t.Fatal("want error")
	}
	if res.OK() {
		t.Error("res should not be OK after error")
	}
}

func TestRunStrategyHaltsAfterFirstFailure(t *testing.T) {
	dir := t.TempDir()
	s := Strategy{Name: "halt", Cmds: [][]string{{"false"}, {"true"}}}
	var buf bytes.Buffer
	res, _ := runStrategy(context.Background(), dir, &buf, s)
	if len(res.Steps) != 1 {
		t.Errorf("want 1 step before halt, got %d", len(res.Steps))
	}
}

func TestCleanVenvIsNoopOnAbsentDir(t *testing.T) {
	dir := t.TempDir()
	cleanVenv(filepath.Join(dir, "nonexistent"))
}

func TestInstallErrorsWithoutInterpreter(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("PATH", filepath.Join(dir, "empty-bin"))
	_ = os.MkdirAll(filepath.Join(dir, "empty-bin"), 0o755)
	_, err := Install(context.Background(), dir, "requirements.txt", &bytes.Buffer{})
	if err == nil {
		t.Fatal("want error when PATH empty")
	}
}
