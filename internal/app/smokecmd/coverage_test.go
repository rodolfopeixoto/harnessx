package smokecmd

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"
)

func TestRunWithDefaultStepTimeout(t *testing.T) {
	bin := writeFakeBin(t, "#!/bin/sh\nexit 0\n")
	res, err := Run(context.Background(), Options{
		HarnessBin: bin, Langs: []string{"go"},
	}, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	if !res.OK {
		t.Errorf("want OK")
	}
}

func TestRunKeepsTempDirWhenRequested(t *testing.T) {
	bin := writeFakeBin(t, "#!/bin/sh\nexit 0\n")
	res, err := Run(context.Background(), Options{
		HarnessBin: bin, Langs: []string{"go"}, Keep: true, StepTimeout: 5 * time.Second,
	}, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	if res.Stacks[0].WorkDir == "" {
		t.Errorf("expected work dir recorded")
	}
}

func TestTruncateSnapshotsLargeOutput(t *testing.T) {
	long := strings.Repeat("y", 9000)
	got := truncate(long, 80)
	if !strings.HasSuffix(got, "[truncated]") {
		t.Errorf("missing marker")
	}
}

func TestRunStepHandlesNonExecBinary(t *testing.T) {
	dir := t.TempDir()
	r := runStep(context.Background(), "/totally/missing/binary", dir,
		Step{Name: "boom", Args: []string{"x"}, Kind: KindCLI}, time.Second)
	if r.ExitCode == 0 {
		t.Errorf("want non-zero")
	}
}
