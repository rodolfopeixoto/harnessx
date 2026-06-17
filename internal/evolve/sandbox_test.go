package evolve

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func writeBin(t *testing.T, body string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "fake-harness")
	if err := os.WriteFile(p, []byte(body), 0o755); err != nil {
		t.Fatal(err)
	}
	return p
}

func writeTrace(t *testing.T, body string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "trace.jsonl")
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestRunSandboxRequiresBaselineBin(t *testing.T) {
	_, err := RunSandbox(context.Background(), SandboxOptions{TraceFile: "x"})
	if err == nil {
		t.Fatal("want error")
	}
}

func TestRunSandboxRequiresTraceFile(t *testing.T) {
	_, err := RunSandbox(context.Background(), SandboxOptions{HarnessBin: "/bin/true"})
	if err == nil {
		t.Fatal("want error")
	}
}

func TestRunSandboxExecutesBothBinaries(t *testing.T) {
	bin := writeBin(t, `#!/bin/sh
echo '{"events_scanned":5,"failures":3,"clusters":[]}'
`)
	trace := writeTrace(t, `{"level":"error","fields":{"status":"failed"}}`+"\n")
	res, err := RunSandbox(context.Background(), SandboxOptions{
		HarnessBin: bin, CandidateBin: bin, TraceFile: trace, Timeout: 10 * time.Second,
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Baseline.Failures != 3 || res.Candidate.Failures != 3 {
		t.Errorf("failures: base=%d cand=%d", res.Baseline.Failures, res.Candidate.Failures)
	}
	if res.Improvement.Improved {
		t.Errorf("equal binaries should not improve")
	}
}

func TestRunSandboxFlagsImprovement(t *testing.T) {
	base := writeBin(t, `#!/bin/sh
echo '{"failures":5,"clusters":[]}'
`)
	cand := writeBin(t, `#!/bin/sh
echo '{"failures":2,"clusters":[]}'
`)
	trace := writeTrace(t, `{"level":"error","fields":{"status":"failed"}}`+"\n")
	res, err := RunSandbox(context.Background(), SandboxOptions{
		HarnessBin: base, CandidateBin: cand, TraceFile: trace, Timeout: 10 * time.Second,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Improvement.Improved || res.Improvement.FailuresDelta != -3 {
		t.Errorf("want improved (delta -3): %+v", res.Improvement)
	}
}

func TestRunSandboxPropagatesMissingTrace(t *testing.T) {
	bin := writeBin(t, "#!/bin/sh\nexit 0\n")
	_, err := RunSandbox(context.Background(), SandboxOptions{
		HarnessBin: bin, TraceFile: "/nonexistent/trace.jsonl", Timeout: time.Second,
	})
	if err == nil {
		t.Fatal("want error")
	}
}
