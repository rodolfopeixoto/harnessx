// SPDX-License-Identifier: MIT

package learncmd

import (
	"os"
	"path/filepath"

	"github.com/ropeixoto/harnessx/internal/execution"
)

func pruneOrphanRuns(root string, runs []execution.Result) int {
	n := 0
	for _, r := range runs {
		if r.Status == execution.StatusIncomplete {
			path := filepath.Join(root, ".harness", "runs", r.RunID)
			if err := os.RemoveAll(path); err == nil {
				n++
			}
		}
	}
	return n
}

func refreshWorktreeExcludes(root string) {
	wtRoot := filepath.Join(root, ".harness", "worktrees")
	entries, err := os.ReadDir(wtRoot)
	if err != nil {
		return
	}
	body := "# harness memory-learn refresh\n"
	for _, d := range []string{".venv", "venv", "__pycache__", "node_modules", "target", "dist", "build", "_build", "deps", ".gradle", ".idea", ".pytest_cache", ".mypy_cache", ".ruff_cache", "bin", "obj", "coverage"} {
		body += d + "/\n"
	}
	body += "*.tsbuildinfo\n*.log\n"
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		info := filepath.Join(wtRoot, e.Name(), ".git", "info")
		_ = os.MkdirAll(info, 0o755)
		path := filepath.Join(info, "exclude")
		existing, _ := os.ReadFile(path)
		_ = os.WriteFile(path, []byte(string(existing)+"\n"+body), 0o644)
	}
}

func writeFile(path, body string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(body), 0o644)
}
