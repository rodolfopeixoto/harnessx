// SPDX-License-Identifier: MIT

package smokecmd

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func writeFakeBin(t *testing.T, body string) string {
	t.Helper()
	dir := t.TempDir()
	name := "harness-fake.sh"
	if runtime.GOOS == "windows" {
		t.Skip("fake bin uses sh; skipping on windows")
	}
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(body), 0o755); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestDefaultStepsContainsCanonicalCommands(t *testing.T) {
	steps := DefaultSteps("python")
	wantNames := []string{
		"git init", "harness init", "harness install-git-hooks",
		"harness scaffold apply", "harness doctor", "harness sensor list",
		"harness check", "harness memory list", "harness flow list", "harness routes",
	}
	if len(steps) != len(wantNames) {
		t.Fatalf("step count: want %d got %d", len(wantNames), len(steps))
	}
	for i, w := range wantNames {
		if steps[i].Name != w {
			t.Errorf("step[%d]: want %q got %q", i, w, steps[i].Name)
		}
	}
}

func TestRunHonorsLangsFilter(t *testing.T) {
	bin := writeFakeBin(t, "#!/bin/sh\nexit 0\n")
	var buf bytes.Buffer
	res, err := Run(context.Background(), Options{
		HarnessBin:  bin,
		Langs:       []string{"go", "python"},
		StepTimeout: 5 * time.Second,
	}, &buf)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Stacks) != 2 {
		t.Fatalf("stack count: want 2 got %d", len(res.Stacks))
	}
	if !(res.Stacks[0].Stack == "go" && res.Stacks[1].Stack == "python") {
		t.Errorf("stacks not sorted: %v", []string{res.Stacks[0].Stack, res.Stacks[1].Stack})
	}
	if !res.OK {
		t.Errorf("expected OK with passing bin; got fail")
	}
}

func TestRunMarksCLIFailuresButTolersToolFailures(t *testing.T) {
	bin := writeFakeBin(t, "#!/bin/sh\nexit 1\n")
	var buf bytes.Buffer
	res, err := Run(context.Background(), Options{
		HarnessBin:  bin,
		Langs:       []string{"go"},
		StepTimeout: 5 * time.Second,
	}, &buf)
	if err != nil {
		t.Fatal(err)
	}
	if res.OK {
		t.Fatal("matrix should fail when CLI steps exit non-zero")
	}
	failures := FailedSteps(res)
	if len(failures) == 0 {
		t.Fatal("expected failed CLI steps recorded")
	}
}

func TestRunRejectsMissingBinary(t *testing.T) {
	_, err := Run(context.Background(), Options{HarnessBin: "/nonexistent/harness-bin-xyz"}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected error for missing binary")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Fatalf("error must mention not-found: %v", err)
	}
}

func TestFormatJSONRoundTrips(t *testing.T) {
	bin := writeFakeBin(t, "#!/bin/sh\nexit 0\n")
	var buf bytes.Buffer
	res, err := Run(context.Background(), Options{HarnessBin: bin, Langs: []string{"go"}, StepTimeout: 5 * time.Second}, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	if err := FormatJSON(res, &buf); err != nil {
		t.Fatal(err)
	}
	var back MatrixResult
	if err := json.Unmarshal(buf.Bytes(), &back); err != nil {
		t.Fatalf("json invalid: %v", err)
	}
	if back.HarnessBin != res.HarnessBin {
		t.Errorf("bin mismatch")
	}
}

func TestFormatTableEmitsPassFail(t *testing.T) {
	res := MatrixResult{HarnessBin: "/x/harness", OK: true, Stacks: []StackResult{{
		Stack: "go", OK: true, Steps: []StepResult{{Name: "harness init", ExitCode: 0, DurationMs: 12}},
	}}}
	var buf bytes.Buffer
	FormatTable(res, &buf)
	got := buf.String()
	if !strings.Contains(got, "matrix: PASS") {
		t.Errorf("missing PASS verdict: %q", got)
	}
	if !strings.Contains(got, "harness init") {
		t.Errorf("missing step name: %q", got)
	}
}

func TestRunStepTimeoutKillsHangingProcess(t *testing.T) {
	bin := writeFakeBin(t, "#!/bin/sh\nsleep 30\n")
	var buf bytes.Buffer
	res, err := Run(context.Background(), Options{
		HarnessBin:  bin,
		Langs:       []string{"go"},
		StepTimeout: 200 * time.Millisecond,
	}, &buf)
	if err != nil {
		t.Fatal(err)
	}
	if res.OK {
		t.Fatal("expected fail due to timeout")
	}
}
