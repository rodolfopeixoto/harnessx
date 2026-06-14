// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"io"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ropeixoto/harnessx/internal/app/reportcmd"
	"github.com/ropeixoto/harnessx/internal/app/workflow"
)

func newAskCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "ask <question>",
		Short: "Question mode — read-only, evidence-first answer",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := cwd()
			if err != nil {
				return err
			}
			_, err = workflow.Ask(cmd.Context(), workflow.Options{
				StartDir: dir, Prompt: strings.Join(args, " "),
			}, cmd.OutOrStdout())
			return err
		},
	}
}

func newPlanCmd() *cobra.Command {
	var budget float64
	c := &cobra.Command{
		Use:   "plan <prompt>",
		Short: "Generate spec + plan; no execution",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := cwd()
			if err != nil {
				return err
			}
			_, err = workflow.Plan(cmd.Context(), workflow.Options{
				StartDir: dir, Prompt: strings.Join(args, " "), BudgetUSD: budget,
			}, cmd.OutOrStdout())
			return err
		},
	}
	c.Flags().Float64Var(&budget, "budget", 1.0, "max USD budget for the planning step")
	return c
}

// workflowFn is the shared shape of workflow.Run/Feature/Bugfix so the
// three near-identical Cobra subcommands can share scaffolding.
type workflowFn func(context.Context, workflow.Options, io.Writer) (workflow.Result, error)

func newWorkflowCmd(use, short string, fn workflowFn) *cobra.Command {
	var (
		budget  float64
		autoYes bool
		exec    bool
	)
	c := &cobra.Command{
		Use:   use + " <prompt>",
		Short: short,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := cwd()
			if err != nil {
				return err
			}
			_, err = fn(cmd.Context(), workflow.Options{
				StartDir: dir, Prompt: strings.Join(args, " "),
				BudgetUSD: budget, AutoYes: autoYes, Execute: exec,
			}, cmd.OutOrStdout())
			return err
		},
	}
	c.Flags().Float64Var(&budget, "budget", 1.0, "max USD budget for the run")
	c.Flags().BoolVar(&autoYes, "yes", false, "auto-approve the plan")
	c.Flags().BoolVar(&exec, "execute", false, "run the agent execution step (requires --yes)")
	return c
}

func newRunCmd() *cobra.Command {
	return newWorkflowCmd("run", "Spec + plan + (optional) execute via the agent chain", workflow.Run)
}

func newFeatureCmd() *cobra.Command {
	return newWorkflowCmd("feature", "Feature mode — spec + plan + tests + sensors", workflow.Feature)
}

func newBugfixCmd() *cobra.Command {
	return newWorkflowCmd("bugfix", "Bugfix mode — reproduce, fix, regression-test", workflow.Bugfix)
}

func newReportCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "report",
		Short: "Show the most recent run report",
		RunE: func(cmd *cobra.Command, _ []string) error {
			dir, err := cwd()
			if err != nil {
				return err
			}
			return reportcmd.PrintLast(dir, cmd.OutOrStdout())
		},
	}
}
