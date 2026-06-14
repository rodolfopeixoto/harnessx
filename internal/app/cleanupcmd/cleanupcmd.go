// SPDX-License-Identifier: MIT

package cleanupcmd

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/ropeixoto/harnessx/internal/cleanup"
	"github.com/ropeixoto/harnessx/internal/cleanup/detectors"
	"github.com/ropeixoto/harnessx/internal/platform/constants"
)

func DefaultScanner() *cleanup.Scanner {
	return cleanup.New(
		detectors.Worktrees{},
		detectors.Caches{},
		detectors.AbandonedHarness{},
		detectors.LargeFiles{},
		detectors.VMLeftovers{},
		detectors.ClaudeLeftovers{},
		detectors.Containers{},
	)
}

func Scan(ctx context.Context, root string, asJSON bool, out io.Writer) error {
	findings, err := DefaultScanner().Scan(ctx, root)
	if err != nil {
		return err
	}
	if asJSON {
		enc := json.NewEncoder(out)
		enc.SetIndent("", "  ")
		return enc.Encode(findings)
	}
	if len(findings) == 0 {
		fmt.Fprintln(out, "no cleanup candidates")
		return nil
	}
	tw := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "RISK\tKIND\tSIZE\tPATH\tREASON")
	for _, f := range findings {
		fmt.Fprintf(tw, "%s\t%s\t%dB\t%s\t%s\n", f.Risk, f.Kind, f.SizeBytes, f.Path, f.Reason)
	}
	return tw.Flush()
}

func Apply(ctx context.Context, root, policyPath string, yes bool, in io.Reader, out io.Writer) error {
	policy, err := loadPolicy(root, policyPath)
	if err != nil {
		return err
	}
	findings, err := DefaultScanner().Scan(ctx, root)
	if err != nil {
		return err
	}
	if len(findings) == 0 {
		fmt.Fprintln(out, "no cleanup candidates")
		return nil
	}
	executor := cleanup.NewExecutor(policy, cleanup.NoopAudit{})
	if yes {
		executor.Acknowledgement = "1"
	}
	if !yes {
		executor.Interactive = interactivePrompt(in, out)
	}
	var (
		applied int
		errs    []error
	)
	for _, f := range findings {
		outcome, err := executor.Apply(ctx, f)
		if err != nil {
			if errors.Is(err, cleanup.ErrPolicyMissing) || errors.Is(err, cleanup.ErrUserDenied) {
				fmt.Fprintf(out, "skip  %s %s (%v)\n", f.Risk, f.Path, err)
				continue
			}
			errs = append(errs, err)
			fmt.Fprintf(out, "error %s %s (%v)\n", f.Risk, f.Path, err)
			continue
		}
		applied++
		fmt.Fprintf(out, "ok    %s %s (%d bytes, hash %s)\n", f.Risk, f.Path, outcome.SizeBytes, shortHash(outcome.ContentHash))
	}
	fmt.Fprintf(out, "applied %d/%d findings\n", applied, len(findings))
	if len(errs) > 0 {
		return fmt.Errorf("cleanup: %d errors", len(errs))
	}
	return nil
}

func PolicyInit(root string, out io.Writer) error {
	target, err := safePolicyTarget(root)
	if err != nil {
		return err
	}
	if _, err := os.Stat(target); err == nil {
		return fmt.Errorf("cleanup: %s already exists", target)
	}
	body := loadPolicyTemplate()
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(target, body, 0o644); err != nil {
		return err
	}
	fmt.Fprintf(out, "wrote %s\n", target)
	return nil
}

func safePolicyTarget(root string) (string, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	return filepath.Join(abs, constants.HarnessDir, constants.CleanupSubdir, constants.CleanupPolicyFilename), nil
}

func loadPolicyTemplate() []byte {
	body, err := os.ReadFile(filepath.Join("templates", "cleanup", "policy.example.yaml"))
	if err != nil || len(body) == 0 {
		return []byte(defaultPolicyTemplate)
	}
	return body
}

func loadPolicy(root, override string) (cleanup.Policy, error) {
	if override != "" {
		return cleanup.LoadPolicyFile(override)
	}
	return cleanup.LoadPolicy(root)
}

func interactivePrompt(in io.Reader, out io.Writer) cleanup.Approver {
	reader := bufio.NewReader(in)
	return func(f cleanup.Finding) (bool, error) {
		fmt.Fprintf(out, "apply %s %s (%s)? [y/N] ", f.Risk, f.Path, f.Reason)
		line, err := reader.ReadString('\n')
		if err != nil {
			return false, nil
		}
		if len(line) > 0 && (line[0] == 'y' || line[0] == 'Y') {
			return true, nil
		}
		return false, nil
	}
}

func shortHash(h string) string {
	if len(h) <= 12 {
		return h
	}
	return h[:12]
}

const defaultPolicyTemplate = `version: 1
globals:
  require_acknowledgement: true
rules: []
`
