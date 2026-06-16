// SPDX-License-Identifier: MIT

package auditsolid

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultOpts(t *testing.T) {
	o := Default()
	if o.LimitLOC != 400 || o.LimitFan != 15 {
		t.Fatalf("unexpected defaults: %+v", o)
	}
}

func TestScanCleanRepo(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "ok.go"),
		[]byte("package x\n\nfunc F() {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	v, err := Scan(dir, Default())
	if err != nil {
		t.Fatal(err)
	}
	if len(v) != 0 {
		t.Fatalf("want 0, got %d", len(v))
	}
}

func TestScanGodFile(t *testing.T) {
	dir := t.TempDir()
	var b strings.Builder
	b.WriteString("package x\n")
	for i := 0; i < 500; i++ {
		b.WriteString("// line\n")
	}
	if err := os.WriteFile(filepath.Join(dir, "big.go"), []byte(b.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	v, err := Scan(dir, Opts{LimitLOC: 400, LimitFan: 15})
	if err != nil {
		t.Fatal(err)
	}
	if len(v) == 0 || v[0].Kind != "loc" {
		t.Fatalf("expected loc violation, got %+v", v)
	}
}

func TestScanFanOut(t *testing.T) {
	dir := t.TempDir()
	var b strings.Builder
	b.WriteString("package x\nimport (\n")
	for i := 0; i < 20; i++ {
		b.WriteString("\t\"fmt\"\n")
	}
	b.WriteString(")\n")
	if err := os.WriteFile(filepath.Join(dir, "fan.go"), []byte(b.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	v, err := Scan(dir, Opts{LimitLOC: 99999, LimitFan: 5})
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, x := range v {
		if x.Kind == "fan-out" {
			found = true
		}
	}
	if !found {
		t.Fatalf("want fan-out violation, got %+v", v)
	}
}

func TestSkipDirs(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "vendor")
	_ = os.Mkdir(sub, 0o755)
	if err := os.WriteFile(filepath.Join(sub, "x.go"),
		[]byte("package x\n"+strings.Repeat("// x\n", 1000)), 0o644); err != nil {
		t.Fatal(err)
	}
	v, _ := Scan(dir, Default())
	if len(v) != 0 {
		t.Fatalf("vendor should be skipped, got %+v", v)
	}
}

func TestReport(t *testing.T) {
	if !strings.Contains(Report(nil), "0 SOLID") {
		t.Fatal("clean report wrong")
	}
	s := Report([]Violation{{Path: "a.go", Kind: "loc", Metric: 500, Limit: 400}})
	if !strings.Contains(s, "a.go") || !strings.Contains(s, "loc") {
		t.Fatalf("report missing fields: %s", s)
	}
}
