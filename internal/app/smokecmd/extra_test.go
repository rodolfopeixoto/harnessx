package smokecmd

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"
)

func TestFormatTableRendersFailingStack(t *testing.T) {
	res := MatrixResult{HarnessBin: "/x", OK: false, Stacks: []StackResult{{
		Stack: "go", OK: false, Steps: []StepResult{
			{Name: "harness init", ExitCode: 0, DurationMs: 5, Kind: KindCLI},
			{Name: "harness ci", ExitCode: 1, DurationMs: 12, Kind: KindCLI},
		},
	}}}
	var buf bytes.Buffer
	FormatTable(res, &buf)
	got := buf.String()
	if !strings.Contains(got, "go: FAIL") {
		t.Errorf("missing FAIL header: %s", got)
	}
	if !strings.Contains(got, "matrix: FAIL") {
		t.Errorf("missing matrix verdict: %s", got)
	}
}

func TestFailedStepsReportsCliFailuresOnly(t *testing.T) {
	res := MatrixResult{Stacks: []StackResult{{
		Stack: "go", Steps: []StepResult{
			{Name: "ok step", ExitCode: 0, Kind: KindCLI},
			{Name: "tool fail", ExitCode: 1, Kind: KindTool},
			{Name: "cli fail", ExitCode: 1, Kind: KindCLI},
		},
	}}}
	got := FailedSteps(res)
	if len(got) != 1 {
		t.Fatalf("want 1 cli failure, got %d: %v", len(got), got)
	}
	if !strings.Contains(got[0], "cli fail") {
		t.Errorf("missing cli fail: %v", got)
	}
}

func TestRunNoLangsListsBundled(t *testing.T) {
	bin := writeFakeBin(t, "#!/bin/sh\nexit 0\n")
	res, err := Run(context.Background(), Options{
		HarnessBin: bin, StepTimeout: 5 * time.Second,
	}, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Stacks) < 2 {
		t.Errorf("default langs should pick all bundled stacks, got %d", len(res.Stacks))
	}
}

func TestRunErrorsOnMissingBinaryAtPath(t *testing.T) {
	_, err := Run(context.Background(), Options{HarnessBin: "/definitely/not/here"}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("want error")
	}
}

func TestFormatJSONIncludesStackEntries(t *testing.T) {
	res := MatrixResult{HarnessBin: "/x", Stacks: []StackResult{{Stack: "go"}}, OK: true}
	var buf bytes.Buffer
	if err := FormatJSON(res, &buf); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), `"stack": "go"`) {
		t.Errorf("missing stack: %s", buf.String())
	}
}
