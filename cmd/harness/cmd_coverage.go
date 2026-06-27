// SPDX-License-Identifier: MIT

package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/ropeixoto/harnessx/internal/projectcfg"
	"github.com/ropeixoto/harnessx/internal/sensors/coverage"
	"github.com/ropeixoto/harnessx/internal/ui"
)

func newCoverageCmd() *cobra.Command {
	var threshold float64
	var pkg string
	c := &cobra.Command{
		Use:   "coverage",
		Short: "Stack-aware test coverage gate (paper §3.4.4)",
		Long: `Detects the project stack from .harness/config/project.yaml or
manifest probes (go.mod, pyproject.toml, Cargo.toml, Gemfile,
package.json) and runs the matching coverage tool. Default threshold
is 90%.

Supported stacks:
  go     → go test -cover
  python → .venv/bin/pytest --cov
  rust   → cargo tarpaulin
  ruby/rails → bundle exec rake coverage`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			dir, err := cwd()
			if err != nil {
				return err
			}
			stack, _ := projectcfg.Detect(dir)
			if stack == "" {
				return fmt.Errorf("coverage: cannot detect stack (no go.mod / pyproject.toml / Cargo.toml / Gemfile / package.json)")
			}
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "%s coverage: stack=%s threshold=%.0f%%\n", ui.MarkInfo(), stack, threshold*100)
			switch stack {
			case "go":
				return runGoCoverage(dir, pkg, threshold, out)
			case "python":
				return runPyCoverage(dir, threshold, out)
			case "rust":
				return runStreamProgram(dir, "cargo", []string{"tarpaulin", "--out", "Stdout"}, out)
			case "ruby", "rails":
				return runStreamProgram(dir, "bundle", []string{"exec", "rake", "coverage"}, out)
			case "node":
				return runStreamProgram(dir, "npx", []string{"c8", "--reporter=text", "--lines", fmt.Sprintf("%.0f", threshold*100), "npm", "test", "--", "--run"}, out)
			}
			return fmt.Errorf("coverage: stack %q not wired (PRs welcome)", stack)
		},
	}
	c.Flags().Float64Var(&threshold, "threshold", coverage.DefaultThreshold, "minimum coverage ratio (0..1)")
	c.Flags().StringVar(&pkg, "pkg", "./...", "go package selector (Go only)")
	return c
}

func runGoCoverage(dir, pkg string, threshold float64, out io.Writer) error {
	body, err := runGoCover(dir, pkg)
	if err != nil {
		fmt.Fprintln(out, string(body))
		return err
	}
	r, err := coverage.ParseGoCoverString(string(body), threshold)
	if err != nil {
		return err
	}
	fmt.Fprint(out, coverage.FormatResult(r))
	if !r.Pass() {
		return fmt.Errorf("coverage: threshold %.0f%% not met", threshold*100)
	}
	return nil
}

func runGoCover(dir, pkg string) ([]byte, error) {
	cmd := exec.Command("go", "test", "-cover", pkg)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "HARNESS_PLAIN=1")
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	if err := cmd.Run(); err != nil {
		return buf.Bytes(), err
	}
	return buf.Bytes(), nil
}

var pyCovTotalRE = regexp.MustCompile(`TOTAL\s+\d+\s+\d+\s+([0-9.]+)%`)

func runPyCoverage(dir string, threshold float64, out io.Writer) error {
	pytest := dir + "/.venv/bin/pytest"
	if _, err := os.Stat(pytest); err != nil {
		pytest = "pytest"
	}
	if err := ensurePytestCov(pytest); err != nil {
		fmt.Fprintf(out, "%s coverage: %v\n", ui.MarkInfo(), err)
		return err
	}
	cmd := exec.Command(pytest, "--cov", "--cov-report=term", "-q")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "HARNESS_PLAIN=1", "PYTHONUNBUFFERED=1")
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	err := cmd.Run()
	body := buf.String()
	fmt.Fprintln(out, body)
	if err != nil {
		return fmt.Errorf("coverage: pytest failed: %w", err)
	}
	m := pyCovTotalRE.FindStringSubmatch(body)
	if m == nil {
		return fmt.Errorf("coverage: no TOTAL line in pytest output (install pytest-cov?)")
	}
	pct, _ := strconv.ParseFloat(m[1], 64)
	ratio := pct / 100.0
	if ratio < threshold {
		return fmt.Errorf("coverage: %.1f%% < threshold %.0f%%", pct, threshold*100)
	}
	fmt.Fprintf(out, "%s coverage %.1f%% ≥ %.0f%%\n", ui.MarkSuccess(), pct, threshold*100)
	return nil
}

func ensurePytestCov(pytest string) error {
	probe := exec.Command(pytest, "--help")
	body, err := probe.CombinedOutput()
	if err != nil {
		return fmt.Errorf("pytest --help failed: %w", err)
	}
	if !bytes.Contains(body, []byte("--cov")) {
		return fmt.Errorf("pytest-cov plugin not installed (add `pytest-cov` to requirements.txt and `pip install -r requirements.txt`)")
	}
	return nil
}

func runStreamProgram(dir, bin string, args []string, out io.Writer) error {
	cmd := exec.Command(bin, args...)
	cmd.Dir = dir
	cmd.Stdout = out
	cmd.Stderr = out
	return cmd.Run()
}
