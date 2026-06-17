package commentscan

import (
	"os"
	"path/filepath"
	"testing"
)

func write(t *testing.T, dir, name, body string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestScanAllowsSPDX(t *testing.T) {
	dir := t.TempDir()
	p := write(t, dir, "a.go", "// SPDX-License-Identifier: MIT\n\npackage a\n")
	got := Scan([]string{p}, DefaultAllowlist())
	if len(got) != 0 {
		t.Fatalf("want 0 findings, got %d: %+v", len(got), got)
	}
}

func TestScanAllowsPackageDoc(t *testing.T) {
	dir := t.TempDir()
	p := write(t, dir, "a.go", "// SPDX-License-Identifier: MIT\n\n// Package a does things.\npackage a\n")
	got := Scan([]string{p}, DefaultAllowlist())
	if len(got) != 0 {
		t.Errorf("want 0 findings, got %d: %+v", len(got), got)
	}
}

func TestScanAllowsGodocOnExported(t *testing.T) {
	body := `// SPDX-License-Identifier: MIT
package a

// Foo does things.
func Foo() {}
`
	dir := t.TempDir()
	p := write(t, dir, "a.go", body)
	got := Scan([]string{p}, DefaultAllowlist())
	if len(got) != 0 {
		t.Errorf("want 0, got %+v", got)
	}
}

func TestScanFlagsNarrativeBareComment(t *testing.T) {
	body := `// SPDX-License-Identifier: MIT
package a

func foo() {
    // workaround for upstream
    _ = 1
}
`
	dir := t.TempDir()
	p := write(t, dir, "a.go", body)
	got := Scan([]string{p}, DefaultAllowlist())
	if len(got) == 0 {
		t.Fatal("want narrative comment flagged")
	}
	if got[0].Line != 5 {
		t.Errorf("line: want 5 got %d", got[0].Line)
	}
}

func TestScanFlagsBlockComment(t *testing.T) {
	body := "// SPDX-License-Identifier: MIT\npackage a\n\n/* used to be x */\nfunc Bar() {}\n"
	dir := t.TempDir()
	p := write(t, dir, "a.go", body)
	got := Scan([]string{p}, DefaultAllowlist())
	if len(got) == 0 {
		t.Fatal("want block comment flagged (not godoc on Bar)")
	}
}

func TestScanIgnoresNonGo(t *testing.T) {
	dir := t.TempDir()
	p := write(t, dir, "a.txt", "// random\n")
	got := Scan([]string{p}, DefaultAllowlist())
	if len(got) != 0 {
		t.Errorf("non-go must not be scanned: %+v", got)
	}
}

func TestScanSkipsUnparseable(t *testing.T) {
	dir := t.TempDir()
	p := write(t, dir, "broken.go", "this is not go")
	got := Scan([]string{p}, DefaultAllowlist())
	if len(got) != 0 {
		t.Errorf("unparseable must be skipped: %+v", got)
	}
}

func TestTruncateCutsLongSnippet(t *testing.T) {
	long := "x"
	for i := 0; i < 200; i++ {
		long += "y"
	}
	got := truncate(long, 80)
	if len(got) != 83 {
		t.Errorf("want 83 (80+...), got %d", len(got))
	}
}

func TestStartsWithExportedNameHandlesExactMatch(t *testing.T) {
	if !startsWithExportedName("Foo", map[string]bool{"Foo": true}) {
		t.Error("exact match should pass")
	}
	if startsWithExportedName("foo bar", map[string]bool{"Foo": true}) {
		t.Error("lower-case must not match")
	}
}
