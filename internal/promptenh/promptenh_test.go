// SPDX-License-Identifier: MIT

package promptenh

import (
	"strings"
	"testing"

	hxcontext "github.com/ropeixoto/harnessx/internal/context"
	"github.com/ropeixoto/harnessx/internal/domain"
)

type fakeSkills struct{ items []SkillSnippet }

func (f fakeSkills) List() ([]SkillSnippet, error) { return f.items, nil }

func TestEnhance_AddsTaskHeaderAndContext(t *testing.T) {
	pack := &hxcontext.Pack{
		RelevantFiles: []hxcontext.FileEntry{
			{Path: "internal/foo.go", Reason: "modified"},
			{Path: "internal/bar.go", Reason: "imports foo"},
		},
	}
	e := Enhance("add a /healthz endpoint", domain.ModeFeature, pack, nil)
	if !strings.Contains(e.Enhanced, "## Task") {
		t.Fatal("missing Task header")
	}
	if !strings.Contains(e.Enhanced, "## Project context") {
		t.Fatal("missing Project context")
	}
	if !strings.Contains(e.Enhanced, "internal/foo.go") {
		t.Fatal("missing relevant file")
	}
	if e.TokensAdded <= 0 {
		t.Fatalf("expected positive tokens added, got %d", e.TokensAdded)
	}
}

func TestEnhance_PrefixesMatchedSkills(t *testing.T) {
	fs := fakeSkills{items: []SkillSnippet{
		{ID: "go-test-loop", Mode: "feature", Body: "Always add a table-driven test.", Score: 0.9},
		{ID: "ruby-style", Mode: "bugfix", Body: "n/a", Score: 0.5},
		{ID: "universal", Mode: "*", Body: "Prefer small commits.", Score: 0.7},
	}}
	e := Enhance("add a /healthz endpoint", domain.ModeFeature, nil, fs)
	if len(e.SkillPrefixes) != 2 {
		t.Fatalf("expected 2 matched skills, got %v", e.SkillPrefixes)
	}
	if !strings.Contains(e.Enhanced, "Always add a table-driven test.") {
		t.Fatal("missing go-test-loop body")
	}
	if !strings.Contains(e.Enhanced, "Prefer small commits.") {
		t.Fatal("missing universal body")
	}
	if strings.Contains(e.Enhanced, "n/a") {
		t.Fatal("bugfix-only skill leaked into feature enhancement")
	}
}

func TestEnhance_DeterministicAcrossRuns(t *testing.T) {
	pack := &hxcontext.Pack{RelevantFiles: []hxcontext.FileEntry{{Path: "a", Reason: "r"}}}
	a := Enhance("p", domain.ModeFeature, pack, nil)
	b := Enhance("p", domain.ModeFeature, pack, nil)
	if a.Enhanced != b.Enhanced {
		t.Fatal("enhancement is not deterministic")
	}
}

func TestWrite_PersistsJSON(t *testing.T) {
	dir := t.TempDir()
	path, err := Write(dir, Enhancement{Original: "p", Enhanced: "p"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(path, "enhancement.json") {
		t.Fatalf("unexpected path %s", path)
	}
}
