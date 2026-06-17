package repl

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ropeixoto/harnessx/internal/intentplan"
)

func TestDefaultPlanUnknownGoalReturnsEmpty(t *testing.T) {
	p := DefaultPlan("alien", "x")
	if len(p.Steps) != 0 {
		t.Errorf("unknown goal should yield empty steps: %+v", p.Steps)
	}
}

type erroringReader struct{}

func (erroringReader) Read(p []byte) (int, error) { return 0, errors.New("boom") }

func TestRunPropagatesNonEOFReadError(t *testing.T) {
	bin := writeFakeBin(t, "#!/bin/sh\nexit 0\n")
	err := Run(context.Background(), Options{
		Root:       t.TempDir(),
		HarnessBin: bin,
		Goal:       intentplan.GoalDev,
		In:         erroringReader{},
		Out:        &bytes.Buffer{},
	})
	if err == nil {
		t.Fatal("want propagated error")
	}
}

func TestRunDefaultsRootToGetwd(t *testing.T) {
	cwd, _ := os.Getwd()
	defer os.Chdir(cwd)
	dir := t.TempDir()
	_ = os.Chdir(dir)
	bin := writeFakeBin(t, "#!/bin/sh\nexit 0\n")
	err := Run(context.Background(), Options{
		HarnessBin: bin,
		Goal:       intentplan.GoalDev,
		In:         strings.NewReader("/exit\n"),
		Out:        io.Discard,
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestPersistFailsWhenDirIsFile(t *testing.T) {
	dir := t.TempDir()
	blocked := filepath.Join(dir, ".harness")
	_ = os.WriteFile(blocked, []byte("x"), 0o644)
	if err := persist(dir, Session{ID: "a", Turns: []Turn{{Action: "x"}}}); err == nil {
		t.Fatal("want error when path is blocked")
	}
}
