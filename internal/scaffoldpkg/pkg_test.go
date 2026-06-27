// SPDX-License-Identifier: MIT

package scaffoldpkg

import (
	"os"
	"path/filepath"
	"testing"
)

func TestListAllBundledLanguages(t *testing.T) {
	got, err := List()
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"dotnet", "elixir", "go", "java", "kotlin", "php", "python", "python-ecommerce", "rails", "react", "ruby", "rust", "swift"}
	if len(got) != len(want) {
		t.Fatalf("want %d langs, got %d: %v", len(want), len(got), got)
	}
	for i, w := range want {
		if got[i] != w {
			t.Errorf("lang[%d]: want %q, got %q", i, w, got[i])
		}
	}
}

func TestLoadEveryBundled(t *testing.T) {
	langs, _ := List()
	for _, l := range langs {
		m, err := Load(l)
		if err != nil {
			t.Fatalf("load %s: %v", l, err)
		}
		if m.Language != l {
			t.Errorf("%s: meta.Language=%q", l, m.Language)
		}
		if len(m.Files) == 0 {
			t.Errorf("%s: no files declared", l)
		}
		if m.LintCommand == "" || m.TestCommand == "" {
			t.Errorf("%s: missing lint/test command", l)
		}
	}
}

func TestLoadUnknownLanguage(t *testing.T) {
	_, err := Load("cobol-1959")
	if err == nil {
		t.Fatal("expected error for unknown language")
	}
}

func TestApplyDryRun(t *testing.T) {
	dir := t.TempDir()
	m, _ := Load("python")
	res, err := Apply(m, ApplyOptions{Root: dir, Name: "demo", Dry: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Created) == 0 {
		t.Fatal("dry-run reported zero created files")
	}
	// Dry run must not actually write
	entries, _ := os.ReadDir(dir)
	if len(entries) != 0 {
		t.Fatalf("dry-run wrote %d entries", len(entries))
	}
}

func TestApplyWritesFiles(t *testing.T) {
	dir := t.TempDir()
	m, _ := Load("python")
	res, err := Apply(m, ApplyOptions{Root: dir, Name: "demo"})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Created) == 0 {
		t.Fatal("apply created zero files")
	}
	if _, err := os.Stat(filepath.Join(dir, "app.py")); err != nil {
		t.Errorf("app.py missing: %v", err)
	}
}

func TestApplySkipsExisting(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "app.py"), []byte("custom"), 0o644); err != nil {
		t.Fatal(err)
	}
	m, _ := Load("python")
	res, _ := Apply(m, ApplyOptions{Root: dir, Name: "demo"})
	found := false
	for _, s := range res.Skipped {
		if s == "app.py" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected app.py in skipped: %+v", res.Skipped)
	}
}

func TestApplyForceOverwrites(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "app.py"), []byte("custom"), 0o644)
	m, _ := Load("python")
	_, err := Apply(m, ApplyOptions{Root: dir, Name: "demo", Force: true})
	if err != nil {
		t.Fatal(err)
	}
	body, _ := os.ReadFile(filepath.Join(dir, "app.py"))
	if string(body) == "custom" {
		t.Errorf("--force did not overwrite")
	}
}

func TestApplyMissingRoot(t *testing.T) {
	_, err := Apply(Meta{Language: "python"}, ApplyOptions{})
	if err == nil {
		t.Fatal("expected error for empty Root")
	}
}

func TestReadTemplateSubstitutesName(t *testing.T) {
	body, err := ReadTemplate("python", "app.py.tmpl", "myapi")
	if err != nil {
		t.Fatal(err)
	}
	if !containsBytes(body, "myapi") {
		t.Errorf("template did not substitute $NAME → myapi: %s", string(body))
	}
}

func containsBytes(haystack []byte, needle string) bool {
	return string(haystack) != "" && len(needle) > 0 && bytesIndex(haystack, needle) >= 0
}

func bytesIndex(haystack []byte, needle string) int {
	n := len(needle)
	for i := 0; i+n <= len(haystack); i++ {
		if string(haystack[i:i+n]) == needle {
			return i
		}
	}
	return -1
}
