// SPDX-License-Identifier: MIT

package execution

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestApplyWorktreeDiff_ConflictRejectsPatch is the BUG-13/14 regression
// guard: when an agent emits a diff that touches the same lines the user
// has already modified, the apply step must REFUSE to write to disk,
// surface ErrApplyConflict, and dump the original patch under
// <runDir>/rejects/ so the user can resolve manually.
func TestApplyWorktreeDiff_ConflictRejectsPatch_BUG13(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	root := t.TempDir()
	applyGitInit(t, root)
	// Seed file v1, commit.
	writeFile(t, root, "app.py", "def main():\n    return 1\n")
	gitAddCommit(t, root, "init")
	// User edits to v2 (uncommitted dirty).
	writeFile(t, root, "app.py", "def main():\n    return 99\n")

	// Build a patch that wants to replace v1 with v3 — incompatible with v2.
	runDir := filepath.Join(root, ".harness", "runs", "run_test")
	if err := os.MkdirAll(runDir, 0o755); err != nil {
		t.Fatal(err)
	}
	patch := `diff --git a/app.py b/app.py
index 0000000..1111111 100644
--- a/app.py
+++ b/app.py
@@ -1,2 +1,2 @@
 def main():
-    return 1
+    return 3
`
	if err := os.WriteFile(filepath.Join(runDir, "diff.patch"), []byte(patch), 0o644); err != nil {
		t.Fatal(err)
	}
	wt := Worktree{Kind: "git_worktree", Path: root}

	err := ApplyWorktreeDiff(context.Background(), root, wt, runDir)
	if err == nil {
		t.Fatal("expected ErrApplyConflict, got nil")
	}
	if !errors.Is(err, ErrApplyConflict) {
		t.Fatalf("expected ErrApplyConflict, got %v", err)
	}

	// The user's working tree must NOT contain conflict markers.
	body, _ := os.ReadFile(filepath.Join(root, "app.py"))
	if strings.Contains(string(body), "<<<<<<<") {
		t.Fatalf("apply leaked conflict markers into app.py:\n%s", body)
	}

	// rejects/ must exist and contain the original patch.
	if _, err := os.Stat(filepath.Join(runDir, "rejects", "diff.patch")); err != nil {
		t.Fatalf("expected rejects/diff.patch to exist: %v", err)
	}
}

func applyGitInit(t *testing.T, root string) {
	t.Helper()
	for _, args := range [][]string{
		{"init", "-q"},
		{"-c", "user.email=t@t", "-c", "user.name=t", "config", "commit.gpgsign", "false"},
		{"-c", "user.email=t@t", "-c", "user.name=t", "config", "user.email", "t@t"},
		{"-c", "user.email=t@t", "-c", "user.name=t", "config", "user.name", "t"},
	} {
		c := exec.Command("git", args...)
		c.Dir = root
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
}

func gitAddCommit(t *testing.T, root, msg string) {
	t.Helper()
	for _, args := range [][]string{
		{"add", "-A"},
		{"commit", "-qm", msg},
	} {
		c := exec.Command("git", args...)
		c.Dir = root
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
}

func writeFile(t *testing.T, root, rel, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(root, filepath.Dir(rel)), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, rel), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestUntouchedPromisedFiles_BUG18 covers the regression where a run
// claimed `applied` after editing only ancillary files (e.g. tests) and
// left the requested source file unchanged. Audit BUG-18.
func TestUntouchedPromisedFiles_BUG18(t *testing.T) {
	cases := []struct {
		name     string
		promised []string
		changed  []string
		want     []string
	}{
		{"no promises is no miss", nil, []string{"a.go"}, nil},
		{"all promises touched", []string{"a.go", "b.go"}, []string{"a.go", "b.go"}, nil},
		{"some untouched", []string{"app.py", "test_app.py"}, []string{"test_app.py"}, []string{"app.py"}},
		{"none touched", []string{"a", "b"}, nil, []string{"a", "b"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := untouchedPromisedFiles(tc.promised, tc.changed)
			if len(got) != len(tc.want) {
				t.Fatalf("got %v want %v", got, tc.want)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Fatalf("got %v want %v", got, tc.want)
				}
			}
		})
	}
}
