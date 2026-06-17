package planscope

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestGitChangedFilesAgainstRealRepo(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}
	dir := t.TempDir()
	for _, args := range [][]string{
		{"init", "-q"},
		{"-c", "user.email=t@t", "-c", "user.name=t", "commit", "--allow-empty", "-m", "init"},
	} {
		c := exec.Command("git", args...)
		c.Dir = dir
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	_ = os.WriteFile(filepath.Join(dir, "a.txt"), []byte("x"), 0o644)
	files, err := GitChangedFiles(context.Background(), dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 || files[0] != "a.txt" {
		t.Errorf("want [a.txt], got %v", files)
	}
}

func TestGitChangedFilesPropagatesError(t *testing.T) {
	_, err := GitChangedFiles(context.Background(), "/nonexistent/repo/xxxxxxx")
	if err == nil {
		t.Fatal("want error")
	}
}
