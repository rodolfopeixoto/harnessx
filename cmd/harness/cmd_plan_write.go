// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/ropeixoto/harnessx/internal/platform/ids"
	"github.com/ropeixoto/harnessx/internal/sensors/planscope"
)

func newPlanCheckCmd() *cobra.Command {
	var planID string
	c := &cobra.Command{
		Use:   "check",
		Short: "Verify changed files stay within PLAN contract scope (paper §3.4.2)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			dir, err := cwd()
			if err != nil {
				return err
			}
			if planID == "" {
				return fmt.Errorf("plan check: --plan id required")
			}
			res, err := planscope.Check(cmd.Context(), planscope.Options{Root: dir, PlanID: planID})
			if err != nil {
				return err
			}
			fmt.Fprint(cmd.OutOrStdout(), planscope.FormatResult(res))
			if !res.Pass() {
				return fmt.Errorf("plan check: %d violation(s)", len(res.Violations))
			}
			return nil
		},
	}
	c.Flags().StringVar(&planID, "plan", "", "PLAN ulid or filename")
	return c
}

func newPlanWriteCmd() *cobra.Command {
	var (
		files      []string
		invariants []string
		validation []string
		rollback   string
		risk       string
	)
	c := &cobra.Command{
		Use:   "write <prompt>",
		Short: "Write a PLAN-as-contract artifact (paper §3.4.2)",
		Long: `Materialises planning as a contract: intent, file scope, invariants,
validation commands, rollback, and risk tier. Written to
.harness/artifacts/plans/PLAN-<ulid>.md with no LLM call.

Downstream 'harness do --plan <id>' (future) reads the contract and
sensors verify that edits stay within scope.`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := cwd()
			if err != nil {
				return err
			}
			prompt := strings.Join(args, " ")
			id := ids.New()
			body := renderPlanContract(planContract{
				ID:         id,
				Intent:     prompt,
				Files:      files,
				Invariants: invariants,
				Validation: validation,
				Rollback:   rollback,
				Risk:       risk,
				CreatedAt:  time.Now().UTC(),
			})
			dst := filepath.Join(root, ".harness", "artifacts", "plans", "PLAN-"+id+".md")
			if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
				return err
			}
			if err := os.WriteFile(dst, []byte(body), 0o644); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "plan: wrote %s\n", dst)
			return nil
		},
	}
	c.Flags().StringSliceVar(&files, "file", nil, "file expected to be touched (repeatable)")
	c.Flags().StringSliceVar(&invariants, "invariant", nil, "invariant the change must preserve (repeatable)")
	c.Flags().StringSliceVar(&validation, "validate", []string{"harness ci"}, "validation command (repeatable)")
	c.Flags().StringVar(&rollback, "rollback", "git revert HEAD", "rollback command")
	c.Flags().StringVar(&risk, "risk", "medium", "risk tier (low|medium|high)")
	return c
}

type planContract struct {
	ID         string
	Intent     string
	Files      []string
	Invariants []string
	Validation []string
	Rollback   string
	Risk       string
	CreatedAt  time.Time
}

func renderPlanContract(p planContract) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "# PLAN-%s\n\n", p.ID)
	fmt.Fprintf(&sb, "*Created %s — paper Code as Agent Harness §3.4.2*\n\n", p.CreatedAt.Format(time.RFC3339))
	fmt.Fprintf(&sb, "## Intent\n\n%s\n\n", p.Intent)
	fmt.Fprintf(&sb, "## Files in scope\n\n")
	if len(p.Files) == 0 {
		sb.WriteString("- _unconstrained_\n")
	}
	for _, f := range p.Files {
		fmt.Fprintf(&sb, "- `%s`\n", f)
	}
	sb.WriteString("\n## Invariants\n\n")
	if len(p.Invariants) == 0 {
		sb.WriteString("- _none declared_\n")
	}
	for _, inv := range p.Invariants {
		fmt.Fprintf(&sb, "- %s\n", inv)
	}
	sb.WriteString("\n## Validation\n\n```sh\n")
	for _, v := range p.Validation {
		fmt.Fprintf(&sb, "%s\n", v)
	}
	sb.WriteString("```\n\n## Rollback\n\n```sh\n")
	fmt.Fprintf(&sb, "%s\n```\n\n", p.Rollback)
	fmt.Fprintf(&sb, "## Risk tier\n\n%s\n", p.Risk)
	return sb.String()
}
