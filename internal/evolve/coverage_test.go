package evolve

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestAppendMutationFailsWhenLogPathIsDir(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, ".harness", "logs", "mutations.jsonl")
	_ = os.MkdirAll(logPath, 0o755)
	if err := appendMutation(dir, Mutation{}, "proposed"); err == nil {
		t.Fatal("want error")
	}
}

func TestTruncateExactBoundary(t *testing.T) {
	got := truncate("abc", 3)
	if got != "abc" {
		t.Errorf("got %q", got)
	}
}

func TestIsFailureFieldsHandlesStringStatusVariants(t *testing.T) {
	for _, s := range []string{"failed", "FAIL", "error", "red", "denied"} {
		if !isFailureFields(map[string]any{"status": s}) {
			t.Errorf("%s should mark failure", s)
		}
	}
}

func TestIsFailureFieldsIgnoresUnknownStatus(t *testing.T) {
	if isFailureFields(map[string]any{"status": "queued"}) {
		t.Error("queued should not mark failure")
	}
}

func TestRunSandboxUsesProvidedWorkspaceRoot(t *testing.T) {
	bin := writeBin(t, "#!/bin/sh\necho '{\"failures\":0}'\n")
	trace := writeTrace(t, `{"level":"info"}`+"\n")
	ws := t.TempDir()
	res, err := RunSandbox(context.Background(), SandboxOptions{
		HarnessBin: bin, TraceFile: trace, WorkspaceRoot: ws, Timeout: 5 * time.Second,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(res.WorkspaceRoot, ws) {
		t.Errorf("workspace not honoured: %s", res.WorkspaceRoot)
	}
}

func TestRunOnceCapturesNonZeroExit(t *testing.T) {
	bin := writeBin(t, "#!/bin/sh\nexit 7\n")
	trace := writeTrace(t, `{"level":"info"}`+"\n")
	snap, err := runOnce(context.Background(), bin, nil, t.TempDir(), trace, 10*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if snap.Exit == 0 {
		t.Errorf("expected non-zero exit, got %d", snap.Exit)
	}
}
