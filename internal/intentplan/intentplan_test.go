package intentplan

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestValidateAcceptsKnownGoals(t *testing.T) {
	for _, g := range KnownGoals() {
		p := Plan{Goal: g, Intent: "x", Steps: []Step{{Kind: StepHarness, Cmd: []string{GoalPalette(g)[0]}}}}
		if err := p.Validate(); err != nil {
			t.Errorf("goal %s rejected: %v", g, err)
		}
	}
}

func TestValidateRejectsUnknownGoal(t *testing.T) {
	if err := (Plan{Goal: "alien", Intent: "x", Steps: []Step{{Kind: StepHarness, Cmd: []string{"x"}}}}).Validate(); err == nil {
		t.Fatal("want error")
	}
}

func TestValidateRejectsMissingIntent(t *testing.T) {
	if err := (Plan{Goal: GoalDev, Steps: []Step{{Kind: StepHarness, Cmd: []string{"ci"}}}}).Validate(); err == nil {
		t.Fatal("want error")
	}
}

func TestValidateRejectsNoSteps(t *testing.T) {
	if err := (Plan{Goal: GoalDev, Intent: "x"}).Validate(); err == nil {
		t.Fatal("want error")
	}
}

func TestValidateRejectsCmdNotInPalette(t *testing.T) {
	p := Plan{Goal: GoalDev, Intent: "x", Steps: []Step{{Kind: StepHarness, Cmd: []string{"runtime"}}}}
	if err := p.Validate(); err == nil {
		t.Fatal("runtime not in dev palette; want error")
	}
}

func TestStepShellAcceptsAnyCmd(t *testing.T) {
	p := Plan{Goal: GoalDev, Intent: "x", Steps: []Step{{Kind: StepShell, Cmd: []string{"echo", "hi"}}}}
	if err := p.Validate(); err != nil {
		t.Fatalf("shell should accept arbitrary cmd: %v", err)
	}
}

func TestStepRejectsUnknownKind(t *testing.T) {
	p := Plan{Goal: GoalDev, Intent: "x", Steps: []Step{{Kind: "weird", Cmd: []string{"x"}}}}
	if err := p.Validate(); err == nil {
		t.Fatal("want error")
	}
}

func TestParseStringRoundtrip(t *testing.T) {
	src := `{"goal":"dev","intent":"add /healthz","steps":[{"kind":"harness","cmd":["ci"]}]}`
	p, err := ParseString(src)
	if err != nil {
		t.Fatal(err)
	}
	if p.Goal != GoalDev || p.Steps[0].Cmd[0] != "ci" {
		t.Errorf("bad parse: %+v", p)
	}
}

func TestParseStringRejectsBadJSON(t *testing.T) {
	if _, err := ParseString(":::"); err == nil {
		t.Fatal("want error")
	}
}

func TestMarshalPrettyContainsIntent(t *testing.T) {
	p := Plan{Goal: GoalDev, Intent: "hi", Steps: []Step{{Kind: StepHarness, Cmd: []string{"ci"}}}}
	body, err := p.MarshalPretty()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), `"intent": "hi"`) {
		t.Errorf("missing intent: %s", body)
	}
}

func writeFakeBin(t *testing.T, body string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "harness-fake.sh")
	if err := os.WriteFile(p, []byte(body), 0o755); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestExecuteRunsHarnessStep(t *testing.T) {
	bin := writeFakeBin(t, "#!/bin/sh\necho ran $@\nexit 0\n")
	plan := Plan{Goal: GoalDev, Intent: "x", Steps: []Step{
		{Kind: StepHarness, Cmd: []string{"ci"}},
	}}
	var buf bytes.Buffer
	res, err := Execute(context.Background(), plan, ExecOptions{HarnessBin: bin, WorkingDir: t.TempDir(), Out: &buf, StepTimeout: 5 * time.Second})
	if err != nil {
		t.Fatal(err)
	}
	if !res.OK {
		t.Errorf("expected OK, got %+v", res)
	}
}

func TestExecuteStopsOnFailure(t *testing.T) {
	bin := writeFakeBin(t, "#!/bin/sh\nexit 1\n")
	plan := Plan{Goal: GoalDev, Intent: "x", Steps: []Step{
		{Kind: StepHarness, Cmd: []string{"ci"}},
		{Kind: StepHarness, Cmd: []string{"test"}},
	}}
	res, _ := Execute(context.Background(), plan, ExecOptions{HarnessBin: bin, WorkingDir: t.TempDir(), StepTimeout: 5 * time.Second})
	if res.OK {
		t.Error("expected OK=false")
	}
	if len(res.Steps) != 1 {
		t.Errorf("should stop after first failure, got %d steps", len(res.Steps))
	}
}

func TestExecuteShellStep(t *testing.T) {
	plan := Plan{Goal: GoalDev, Intent: "x", Steps: []Step{
		{Kind: StepShell, Cmd: []string{"true"}},
	}}
	res, err := Execute(context.Background(), plan, ExecOptions{HarnessBin: "/bin/true", WorkingDir: t.TempDir(), StepTimeout: 5 * time.Second})
	if err != nil {
		t.Fatal(err)
	}
	if !res.OK {
		t.Errorf("got %+v", res)
	}
}

func TestExecuteWaitStep(t *testing.T) {
	plan := Plan{Goal: GoalDev, Intent: "x", Steps: []Step{
		{Kind: StepWait, Cmd: []string{"5ms"}},
	}}
	start := time.Now()
	res, _ := Execute(context.Background(), plan, ExecOptions{HarnessBin: "/bin/true", WorkingDir: t.TempDir()})
	if time.Since(start) < 5*time.Millisecond {
		t.Errorf("wait did not honour duration")
	}
	if !res.OK {
		t.Errorf("wait should succeed")
	}
}

func TestAllowedHarnessCmdsUnion(t *testing.T) {
	got := AllowedHarnessCmds()
	for _, want := range []string{"ci", "ship", "agent", "doctor"} {
		if !contains(got, want) {
			t.Errorf("missing %s in union: %v", want, got)
		}
	}
}

func TestParseDurationFallsBack(t *testing.T) {
	if parseDuration([]string{"junk"}) != 0 {
		t.Error("invalid duration should fall back to 0")
	}
	if parseDuration(nil) != 0 {
		t.Error("nil should be 0")
	}
}

func TestTruncateLongOutput(t *testing.T) {
	long := strings.Repeat("x", 5000)
	got := truncate(long, 100)
	if !strings.HasSuffix(got, "[truncated]") {
		t.Errorf("missing marker")
	}
}
