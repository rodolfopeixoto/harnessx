// SPDX-License-Identifier: MIT

package sensorcmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ropeixoto/harnessx/internal/sensors"
)

// installableBySensorID maps failing sensor ids to the pip package that
// provides them — used by `harness ci --install-missing` to auto-install
// dev tooling the venv is missing.
var installableBySensorID = map[string]string{
	"py_bandit":    "bandit",
	"py_mypy":      "mypy",
	"py_pip_audit": "pip-audit",
}

func missingInstallables(results []sensors.Result) []string {
	seen := map[string]bool{}
	var out []string
	for _, r := range results {
		pkg, ok := installableBySensorID[r.ID]
		if !ok {
			continue
		}
		if !strings.HasPrefix(r.Detail, "binary not on PATH") {
			continue
		}
		if seen[pkg] {
			continue
		}
		seen[pkg] = true
		out = append(out, pkg)
	}
	return out
}

func installPythonTools(ctx context.Context, root string, pkgs []string, out io.Writer) error {
	venvPython := filepath.Join(root, ".venv", "bin", "python")
	if _, err := os.Stat(venvPython); err != nil {
		return fmt.Errorf(".venv missing — run `harness new <stack> --with-deps` or `uv venv .venv && uv pip install -r requirements.txt`")
	}
	fmt.Fprintf(out, "  → installing into .venv: %s\n", strings.Join(pkgs, " "))
	args := append([]string{"pip", "install", "--python", venvPython}, pkgs...)
	cmd := exec.CommandContext(ctx, "uv", args...)
	cmd.Dir = root
	cmd.Stdout = out
	cmd.Stderr = out
	if err := cmd.Run(); err == nil {
		return nil
	}
	pipArgs := append([]string{"-m", "pip", "install"}, pkgs...)
	pipCmd := exec.CommandContext(ctx, venvPython, pipArgs...)
	pipCmd.Dir = root
	pipCmd.Stdout = out
	pipCmd.Stderr = out
	if err := pipCmd.Run(); err != nil {
		return fmt.Errorf("install failed via uv and pip: %w", err)
	}
	return nil
}
