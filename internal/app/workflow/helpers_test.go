// SPDX-License-Identifier: MIT

package workflow

import (
	"testing"

	hxcontext "github.com/ropeixoto/harnessx/internal/context"
	"github.com/ropeixoto/harnessx/internal/domain"
	"github.com/ropeixoto/harnessx/internal/index"
)

func TestEstimateCostIsTenPercent(t *testing.T) {
	cases := []struct {
		in, want float64
	}{
		{1.0, 0.10},
		{0.50, 0.05},
		{0.0, 0.0},
		{10.0, 1.0},
	}
	for _, c := range cases {
		if got := estimateCost(c.in); got != c.want {
			t.Errorf("estimateCost(%v)=%v, want %v", c.in, got, c.want)
		}
	}
}

func TestPromptOrError(t *testing.T) {
	if err := PromptOrError(""); err == nil {
		t.Error("empty prompt should error")
	}
	if err := PromptOrError("   "); err == nil {
		t.Error("whitespace prompt should error")
	}
	if err := PromptOrError("real"); err != nil {
		t.Errorf("real prompt: %v", err)
	}
}

func TestTaskForMode(t *testing.T) {
	cases := []struct {
		mode domain.Mode
		want string
	}{
		{domain.ModeBugfix, "implementation"},
		{domain.ModeOptimization, "resource_optimization"},
		{domain.ModeAudit, "dependency_audit"},
		{domain.ModeReview, "security_review"},
		{domain.ModeDesignToProduct, "design_to_product"},
		{domain.Mode("random"), "implementation"},
	}
	for _, c := range cases {
		if got := taskFor(c.mode); got != c.want {
			t.Errorf("taskFor(%q)=%q, want %q", c.mode, got, c.want)
		}
	}
}

func TestFilesOfNilPack(t *testing.T) {
	if files := filesOf(nil); files != nil {
		t.Errorf("nil pack should return nil, got %v", files)
	}
}

func TestFilesOfExtracts(t *testing.T) {
	pack := &hxcontext.Pack{RelevantFiles: []hxcontext.FileEntry{{Path: "a.go"}, {Path: "b.go"}}}
	got := filesOf(pack)
	if len(got) != 2 || got[0] != "a.go" || got[1] != "b.go" {
		t.Errorf("filesOf: got %v", got)
	}
}

func TestStackNames(t *testing.T) {
	p := index.Profile{Stacks: []index.Stack{{Name: "python"}, {Name: "go"}}}
	got := stackNames(p)
	if len(got) != 2 || got[0] != "python" || got[1] != "go" {
		t.Errorf("stackNames: got %v", got)
	}
}

func TestStackNamesEmpty(t *testing.T) {
	if got := stackNames(index.Profile{}); len(got) != 0 {
		t.Errorf("empty profile should return empty slice, got %v", got)
	}
}

func TestEnsureRootReturnsAbs(t *testing.T) {
	dir := t.TempDir()
	got, err := EnsureRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	if got == "" {
		t.Error("EnsureRoot returned empty path")
	}
}
