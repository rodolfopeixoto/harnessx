package intentplan

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestGoalPaletteUnknownReturnsNil(t *testing.T) {
	if GoalPalette("alien") != nil {
		t.Error("unknown goal should yield nil palette")
	}
}

func TestValidateRejectsHarnessEmptyCmd(t *testing.T) {
	p := Plan{Goal: GoalDev, Intent: "x", Steps: []Step{{Kind: StepHarness}}}
	if err := p.Validate(); err == nil {
		t.Fatal("want error")
	}
}

func TestValidateRejectsShellEmptyCmd(t *testing.T) {
	p := Plan{Goal: GoalDev, Intent: "x", Steps: []Step{{Kind: StepShell}}}
	if err := p.Validate(); err == nil {
		t.Fatal("want error")
	}
}

func TestStepValidateWaitAcceptsEmpty(t *testing.T) {
	s := Step{Kind: StepWait}
	if err := s.validate(nil); err != nil {
		t.Fatal(err)
	}
}

func TestParseJSONInvalidPayload(t *testing.T) {
	_, err := ParseJSON(bytes.NewReader([]byte(`{"goal":""}`)))
	if err == nil {
		t.Fatal("want validate error")
	}
}

func TestExecuteUsesGetwdWhenEmpty(t *testing.T) {
	plan := Plan{Goal: GoalDev, Intent: "x", Steps: []Step{{Kind: StepWait, Cmd: []string{"1ms"}}}}
	res, err := Execute(context.Background(), plan, ExecOptions{HarnessBin: "/bin/true"})
	if err != nil {
		t.Fatal(err)
	}
	if !res.OK {
		t.Errorf("want OK")
	}
}

func TestExecuteUsesOsExecutableWhenBinEmpty(t *testing.T) {
	plan := Plan{Goal: GoalDev, Intent: "x", Steps: []Step{{Kind: StepWait, Cmd: []string{"1ms"}}}}
	res, err := Execute(context.Background(), plan, ExecOptions{WorkingDir: t.TempDir(), StepTimeout: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	if !res.OK {
		t.Errorf("want OK")
	}
}

func TestExecuteUnknownStepKindRecordsError(t *testing.T) {
	plan := Plan{Goal: GoalDev, Intent: "x", Steps: []Step{
		{Kind: "x", Cmd: []string{"x"}},
	}}
	res, _ := Execute(context.Background(), plan, ExecOptions{HarnessBin: "/bin/true", WorkingDir: t.TempDir()})
	if res.OK {
		t.Error("unknown kind should not succeed")
	}
	if res.Steps[0].ExitCode != -1 {
		t.Errorf("want -1 exit, got %d", res.Steps[0].ExitCode)
	}
}

func TestRunStepNonExitErrorRecorded(t *testing.T) {
	dir := t.TempDir()
	bad := filepath.Join(dir, "nope")
	plan := Plan{Goal: GoalDev, Intent: "x", Steps: []Step{
		{Kind: StepHarness, Cmd: []string{"ci"}},
	}}
	res, err := Execute(context.Background(), plan, ExecOptions{HarnessBin: bad, WorkingDir: dir, StepTimeout: time.Second})
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			t.Fatal(err)
		}
	}
	if res.OK {
		t.Errorf("OK should be false")
	}
}
