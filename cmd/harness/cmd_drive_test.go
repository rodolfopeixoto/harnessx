// SPDX-License-Identifier: MIT

package main

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ropeixoto/harnessx/internal/platform/constants"
)

func TestSanitisePyIdent(t *testing.T) {
	cases := map[string]string{
		"add-products":   "add_products",
		"AddProducts":    "addproducts",
		"":               "x_",
		"123abc":         "x_123abc",
		"with space":     "with_space",
		"weird/path.py":  "weird_path_py",
		"alpha_numeric9": "alpha_numeric9",
	}
	for in, want := range cases {
		if got := sanitisePyIdent(in); got != want {
			t.Errorf("sanitisePyIdent(%q)=%q want %q", in, got, want)
		}
	}
}

func TestTruncSubject(t *testing.T) {
	cases := []struct {
		in   string
		max  int
		want string
	}{
		{"hello", 10, "hello"},
		{"helloworld", 5, "hello"},
		{"x", 0, ""},
		{"x", -3, ""},
	}
	for _, c := range cases {
		if got := truncSubject(c.in, c.max); got != c.want {
			t.Errorf("truncSubject(%q,%d)=%q want %q", c.in, c.max, got, c.want)
		}
	}
}

func TestConventionalDriveSubject(t *testing.T) {
	subj := conventionalDriveSubject("add /products endpoint")
	if !strings.HasPrefix(subj, constants.DriveCommitTypeFeat+": ") {
		t.Errorf("missing feat prefix: %q", subj)
	}
	if len(subj) > constants.DriveCommitSubjectMax {
		t.Errorf("subject %d > %d", len(subj), constants.DriveCommitSubjectMax)
	}
}

func TestRenderPlaceholderTestUsesSlug(t *testing.T) {
	body := renderPlaceholderTest("add stock", "add-stock")
	for _, want := range []string{
		"# harness drive — failing placeholder test for: add stock",
		"def test_drive_placeholder_for_add_stock()",
		"pytest.fail",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("missing %q\n%s", want, body)
		}
	}
}

func TestWritePlaceholderTestCreatesFile(t *testing.T) {
	dir := t.TempDir()
	opts := driveOpts{root: dir, prompt: "add stock", slug: "add-stock"}
	path, err := writePlaceholderTest(opts)
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(dir, "tests", constants.DriveTestFilePrefix+"add-stock"+constants.DriveTestFileSuffix)
	if path != want {
		t.Errorf("want %s got %s", want, path)
	}
	body, _ := os.ReadFile(path)
	if !bytes.Contains(body, []byte("pytest.fail")) {
		t.Errorf("placeholder body missing pytest.fail: %s", body)
	}
}

func TestRunHarnessChildPropagatesExitCode(t *testing.T) {
	out := &bytes.Buffer{}
	err := runHarnessChild(context.Background(), "/usr/bin/false", t.TempDir(), out, nil)
	if err == nil {
		t.Fatal("expected error from false")
	}
	if !strings.Contains(err.Error(), "exit 1") {
		t.Errorf("want exit 1 in %v", err)
	}
}

func TestRunHarnessChildSucceedsForTrue(t *testing.T) {
	if err := runHarnessChild(context.Background(), "/usr/bin/true", t.TempDir(), &bytes.Buffer{}, nil); err != nil {
		t.Fatal(err)
	}
}

func TestRunGitInDirReportsFailure(t *testing.T) {
	if err := runGitInDir(context.Background(), t.TempDir(), "not-a-real-subcommand"); err == nil {
		t.Fatal("expected error")
	}
}

func TestRunDriveSpecStepFailsBubblesUp(t *testing.T) {
	out := &bytes.Buffer{}
	err := runDrive(context.Background(), out, driveOpts{
		root: t.TempDir(), bin: "/usr/bin/false",
		prompt: "x", slug: "x", autonomy: "safe_execute", maxAttempts: 1,
	})
	if err == nil || !strings.Contains(err.Error(), "spec step") {
		t.Errorf("want spec-step error, got %v", err)
	}
}

func TestRunDriveTestsAlreadyGreenShortCircuits(t *testing.T) {
	dir := t.TempDir()
	out := &bytes.Buffer{}
	err := runDrive(context.Background(), out, driveOpts{
		root: dir, bin: "/usr/bin/true",
		prompt: "x", slug: "x", autonomy: "safe_execute", maxAttempts: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "tests already green") {
		t.Errorf("missing already-green branch: %s", out.String())
	}
}

func TestDriveExpectRedTestsTrueWhenChildFails(t *testing.T) {
	out := &bytes.Buffer{}
	if !driveExpectRedTests(context.Background(), out, driveOpts{
		root: t.TempDir(), bin: "/usr/bin/false",
	}) {
		t.Error("want true when test step fails")
	}
}

func TestDriveExpectRedTestsFalseWhenChildPasses(t *testing.T) {
	out := &bytes.Buffer{}
	if driveExpectRedTests(context.Background(), out, driveOpts{
		root: t.TempDir(), bin: "/usr/bin/true",
	}) {
		t.Error("want false when test step passes")
	}
}

func TestDriveImplLoopAbortsOnNoChanges(t *testing.T) {
	out := &bytes.Buffer{}
	err := driveImplLoop(context.Background(), out, driveOpts{
		root: t.TempDir(), bin: "/usr/bin/true",
		prompt: "x", slug: "x", autonomy: "safe_execute",
		maxAttempts: 1, skipCommit: true,
	})
	if err == nil || !strings.Contains(err.Error(), "no changes") {
		t.Fatalf("expected no-changes abort, got %v", err)
	}
}

func TestDriveImplLoopExhaustsAttempts(t *testing.T) {
	out := &bytes.Buffer{}
	err := driveImplLoop(context.Background(), out, driveOpts{
		root: t.TempDir(), bin: "/usr/bin/false",
		prompt: "x", slug: "x", autonomy: "safe_execute", maxAttempts: 2,
	})
	if err == nil {
		t.Fatal("want error after exhausting attempts")
	}
}

func TestDriveTestEmitWritesPlaceholderEvenWithoutAgents(t *testing.T) {
	dir := t.TempDir()
	out := &bytes.Buffer{}
	path, err := driveTestEmit(context.Background(), out, driveOpts{root: dir, slug: "abc", prompt: "x"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Errorf("placeholder not on disk: %v", err)
	}
}

func TestNewDriveCmdRegistersFlags(t *testing.T) {
	c := newDriveCmd()
	for _, f := range []string{"slug", "autonomy", "max-attempts", "skip-commit"} {
		if c.Flag(f) == nil {
			t.Errorf("flag %q not registered", f)
		}
	}
}

func TestDriveCommitWritesCommit(t *testing.T) {
	dir := t.TempDir()
	for _, args := range [][]string{
		{"init", "-q"},
		{"config", "user.email", "x@y.z"},
		{"config", "user.name", "t"},
	} {
		c := exec.Command("git", args...)
		c.Dir = dir
		if err := c.Run(); err != nil {
			t.Skipf("git not available: %v", err)
		}
	}
	if err := os.WriteFile(filepath.Join(dir, "x.txt"), []byte("hi"), 0o644); err != nil {
		t.Fatal(err)
	}
	out := &bytes.Buffer{}
	if err := driveCommit(context.Background(), out, driveOpts{root: dir, prompt: "add stock"}); err != nil {
		t.Fatal(err)
	}
	c := exec.Command("git", "log", "--oneline", "-1")
	c.Dir = dir
	body, err := c.Output()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), "feat: add stock") {
		t.Errorf("commit subject not found in log: %s", body)
	}
}
