// SPDX-License-Identifier: MIT

package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"

	"github.com/ropeixoto/harnessx/internal/sensors/coverage"
)

func newCoverageCmd() *cobra.Command {
	var threshold float64
	var pkg string
	c := &cobra.Command{
		Use:   "coverage",
		Short: "Run go test -cover and gate against a threshold (paper §3.4.4)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			dir, err := cwd()
			if err != nil {
				return err
			}
			out, err := runGoCover(dir, pkg)
			if err != nil {
				fmt.Fprintln(cmd.ErrOrStderr(), string(out))
				return err
			}
			r, err := coverage.ParseGoCoverString(string(out), threshold)
			if err != nil {
				return err
			}
			fmt.Fprint(cmd.OutOrStdout(), coverage.FormatResult(r))
			if !r.Pass() {
				return fmt.Errorf("coverage: threshold %.0f%% not met", threshold*100)
			}
			return nil
		},
	}
	c.Flags().Float64Var(&threshold, "threshold", coverage.DefaultThreshold, "minimum coverage ratio (0..1)")
	c.Flags().StringVar(&pkg, "pkg", "./...", "go package selector")
	return c
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
