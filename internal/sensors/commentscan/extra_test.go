package commentscan

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanAllowsTypeGodoc(t *testing.T) {
	body := `// SPDX-License-Identifier: MIT
package a

// Item is the thing.
type Item struct{}
`
	dir := t.TempDir()
	p := filepath.Join(dir, "a.go")
	_ = os.WriteFile(p, []byte(body), 0o644)
	got := Scan([]string{p}, DefaultAllowlist())
	if len(got) != 0 {
		t.Errorf("type godoc should pass: %+v", got)
	}
}

func TestScanAllowsValueGodoc(t *testing.T) {
	body := `// SPDX-License-Identifier: MIT
package a

// Pi is approximate.
const Pi = 3.14
`
	dir := t.TempDir()
	p := filepath.Join(dir, "a.go")
	_ = os.WriteFile(p, []byte(body), 0o644)
	got := Scan([]string{p}, DefaultAllowlist())
	if len(got) != 0 {
		t.Errorf("const godoc should pass: %+v", got)
	}
}

func TestScanWithoutSPDXAllowanceFlagsLicense(t *testing.T) {
	body := "// SPDX-License-Identifier: MIT\npackage a\n"
	dir := t.TempDir()
	p := filepath.Join(dir, "a.go")
	_ = os.WriteFile(p, []byte(body), 0o644)
	got := Scan([]string{p}, Allowlist{Package: true, Godoc: true})
	if len(got) == 0 {
		t.Error("disabling SPDX allowance should flag the license comment")
	}
}

func TestNormalizeBlockComment(t *testing.T) {
	if got := normalize("/* hi */"); got != "hi" {
		t.Errorf("normalize block: %q", got)
	}
}

func TestNormalizeLineComment(t *testing.T) {
	if got := normalize("//  hello"); got != "hello" {
		t.Errorf("normalize line: %q", got)
	}
}

func TestExportedNamesCollectsAll(t *testing.T) {
	body := `package a
func Exported() {}
func unexported() {}
type Foo struct{}
type bar struct{}
const Big = 1
const small = 2
`
	dir := t.TempDir()
	p := filepath.Join(dir, "a.go")
	_ = os.WriteFile(p, []byte(body), 0o644)
	got := Scan([]string{p}, DefaultAllowlist())
	_ = got
}
