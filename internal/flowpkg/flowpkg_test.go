// SPDX-License-Identifier: MIT

package flowpkg

import (
	"bytes"
	"context"
	"testing"
)

func TestValidateAcceptsGoodFlow(t *testing.T) {
	f := Flow{
		Name:        "demo",
		Description: "demo flow",
		Phases: []Phase{
			{Name: "scaffold", Kind: PhaseDeterministic, Cmd: []string{"echo", "ok"}},
			{Name: "llm", Kind: PhaseLLM, Prompt: "noop"},
		},
	}
	if err := f.Validate(); err != nil {
		t.Errorf("valid flow rejected: %v", err)
	}
}

func TestValidateRejectsMissingName(t *testing.T) {
	if err := (Flow{Phases: []Phase{{Name: "x", Kind: PhaseDeterministic}}}).Validate(); err == nil {
		t.Error("missing name should fail")
	}
}

func TestValidateRejectsNoPhases(t *testing.T) {
	if err := (Flow{Name: "x"}).Validate(); err == nil {
		t.Error("zero phases should fail")
	}
}

func TestValidateRejectsUnknownKind(t *testing.T) {
	if err := (Flow{Name: "x", Phases: []Phase{{Name: "p", Kind: "bogus"}}}).Validate(); err == nil {
		t.Error("unknown kind should fail")
	}
}

func TestListEmptyRegistryNotError(t *testing.T) {
	if _, err := List(); err != nil {
		t.Errorf("empty registry should not error, got %v", err)
	}
}

func TestLoadUnknownFlow(t *testing.T) {
	if _, err := Load("does-not-exist"); err == nil {
		t.Error("expected error for unknown flow")
	}
}

func TestApplyDryRunSkipsPhases(t *testing.T) {
	f := Flow{Name: "x", Phases: []Phase{{Name: "p1", Kind: PhaseDeterministic, Cmd: []string{"echo", "ok"}}}}
	var buf bytes.Buffer
	got, err := Apply(context.Background(), f, ApplyOptions{Dry: true}, &buf)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || !got[0].Skipped {
		t.Errorf("dry-run should skip: %+v", got)
	}
}

func TestApplyExecutesShellPhase(t *testing.T) {
	f := Flow{Name: "x", Phases: []Phase{{Name: "p1", Kind: PhaseDeterministic, Cmd: []string{"true"}}}}
	got, err := Apply(context.Background(), f, ApplyOptions{Root: t.TempDir()}, new(bytes.Buffer))
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Err != nil {
		t.Errorf("expected pass: %+v", got)
	}
}

func TestApplyShellEmptyCmdErrors(t *testing.T) {
	f := Flow{Name: "x", Phases: []Phase{{Name: "p1", Kind: PhaseDeterministic}}}
	got, err := Apply(context.Background(), f, ApplyOptions{Root: t.TempDir()}, new(bytes.Buffer))
	if err != nil {
		t.Fatal(err)
	}
	if got[0].Err == nil {
		t.Error("empty cmd should fail")
	}
}

func TestApplyLLMAndSensorAreSkipped(t *testing.T) {
	f := Flow{Name: "x", Phases: []Phase{
		{Name: "llm", Kind: PhaseLLM, Prompt: "noop"},
		{Name: "sensor", Kind: PhaseSensor, SensorID: "secrets_scan"},
	}}
	got, err := Apply(context.Background(), f, ApplyOptions{Root: t.TempDir()}, new(bytes.Buffer))
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range got {
		if !r.Skipped {
			t.Errorf("llm/sensor should be skipped in v0.92, got %+v", r)
		}
	}
}

func TestApplyRejectsInvalidFlow(t *testing.T) {
	_, err := Apply(context.Background(), Flow{}, ApplyOptions{}, new(bytes.Buffer))
	if err == nil {
		t.Error("invalid flow should error")
	}
}
