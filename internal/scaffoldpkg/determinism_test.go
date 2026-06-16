// SPDX-License-Identifier: MIT

package scaffoldpkg

import (
	"bytes"
	"os"
	"path/filepath"
	"sort"
	"testing"
)

// TestApplyByteIdenticalAcrossRuns proves the scaffold output is
// deterministic for every bundled language. Two Apply calls with the
// same (lang, name) must produce byte-identical files.
func TestApplyByteIdenticalAcrossRuns(t *testing.T) {
	langs, err := List()
	if err != nil {
		t.Fatal(err)
	}
	for _, lang := range langs {
		t.Run(lang, func(t *testing.T) {
			m, err := Load(lang)
			if err != nil {
				t.Fatal(err)
			}
			a := writeAndSnapshot(t, m, "demo")
			b := writeAndSnapshot(t, m, "demo")
			if len(a) != len(b) {
				t.Fatalf("file count differs: %d vs %d", len(a), len(b))
			}
			for path, body := range a {
				if !bytes.Equal(body, b[path]) {
					t.Errorf("%s/%s: byte mismatch across runs", lang, path)
				}
			}
		})
	}
}

func TestApplyDifferentNamesProduceDifferentFiles(t *testing.T) {
	m, err := Load("python")
	if err != nil {
		t.Fatal(err)
	}
	a := writeAndSnapshot(t, m, "first")
	b := writeAndSnapshot(t, m, "second")
	if bytes.Equal(a["app.py"], b["app.py"]) {
		t.Error("two different names should produce different app.py")
	}
}

func TestEveryScaffoldHasGitignore(t *testing.T) {
	langs, _ := List()
	for _, lang := range langs {
		m, _ := Load(lang)
		found := false
		for _, f := range m.Files {
			if f.Path == ".gitignore" {
				found = true
			}
		}
		if !found {
			t.Errorf("%s: missing .gitignore in scaffold", lang)
		}
	}
}

func TestEveryScaffoldHasRequiredTools(t *testing.T) {
	langs, _ := List()
	for _, lang := range langs {
		m, _ := Load(lang)
		if len(m.RequiredTools) == 0 {
			t.Errorf("%s: required_tools is empty", lang)
		}
	}
}

func TestEveryScaffoldHasTestedAgainst(t *testing.T) {
	langs, _ := List()
	for _, lang := range langs {
		m, _ := Load(lang)
		if m.TestedAgainst == "" {
			t.Errorf("%s: tested_against is empty", lang)
		}
	}
}

func writeAndSnapshot(t *testing.T, m Meta, name string) map[string][]byte {
	t.Helper()
	dir := t.TempDir()
	if _, err := Apply(m, ApplyOptions{Root: dir, Name: name}); err != nil {
		t.Fatal(err)
	}
	snap := map[string][]byte{}
	_ = filepath.Walk(dir, func(p string, info os.FileInfo, err error) error {
		if err != nil || info == nil || info.IsDir() {
			return nil
		}
		rel, _ := filepath.Rel(dir, p)
		body, rerr := os.ReadFile(p)
		if rerr != nil {
			return rerr
		}
		snap[rel] = body
		return nil
	})
	// sort keys for stable iteration during diffs
	keys := make([]string, 0, len(snap))
	for k := range snap {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return snap
}
