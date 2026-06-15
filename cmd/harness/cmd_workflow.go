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
		budget     float64
		autoYes    bool
		exec       bool
		agentID    string
		apply      bool
		planOnly   bool
		autonomy   string
		promptFile string
		pdf        string
		image      string
	)
	c := &cobra.Command{
		Use:   use + " <prompt>",
		Short: short,
		Args:  cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := cwd()
			if err != nil {
				return err
			}
			executeStep := exec || agentID != "" || apply || planOnly
			_, err = fn(cmd.Context(), workflow.Options{
				StartDir: dir, Prompt: strings.Join(args, " "),
				BudgetUSD: budget, AutoYes: autoYes || agentID != "", Execute: executeStep,
				AgentID: agentID, Apply: apply, PlanOnly: planOnly, Autonomy: autonomy,
				PromptFile: promptFile, PDF: pdf, Image: image,
			}, cmd.OutOrStdout())
			return err
		},
	}
	c.Flags().Float64Var(&budget, "budget-usd", 1.0, "max USD budget for the run")
	c.Flags().Float64Var(&budget, "budget", 1.0, "deprecated: use --budget-usd")
	_ = c.Flags().MarkHidden("budget")
	c.Flags().BoolVar(&autoYes, "yes", false, "auto-approve the plan")
	c.Flags().BoolVar(&exec, "execute", false, "run legacy agent chain (deprecated when --agent set)")
	c.Flags().StringVar(&agentID, "agent", "", "agent id to drive the real execution loop (claude, codex, gemini, fake-real)")
	c.Flags().BoolVar(&apply, "apply", false, "apply diff to project root after gate allow")
	c.Flags().BoolVar(&planOnly, "plan-only", false, "skip agent invocation")
	c.Flags().StringVar(&autonomy, "autonomy", "safe_execute", "manual|plan_and_ask|safe_execute|full_project_loop|scheduled_maintenance")
	c.Flags().StringVar(&promptFile, "prompt-file", "", "read prompt text from file (concatenated before positional)")
	c.Flags().StringVar(&pdf, "pdf", "", "extract text from PDF via pdftotext (requires poppler-utils)")
	c.Flags().StringVar(&image, "image", "", "attach image for vision-capable agents")
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
