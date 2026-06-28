// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/ropeixoto/harnessx/internal/app/memorycmd"
	"github.com/ropeixoto/harnessx/internal/platform/constants"
	"github.com/ropeixoto/harnessx/internal/recall"
)

func newMemoryCmd() *cobra.Command {
	c := &cobra.Command{Use: "memory", Short: "Project memory commands"}

	var limit int
	var scope, kind string
	listC := &cobra.Command{
		Use:   "list",
		Short: "List recent project memory entries",
		RunE: func(cmd *cobra.Command, _ []string) error {
			dir, err := cwd()
			if err != nil {
				return err
			}
			return memorycmd.List(cmd.OutOrStdout(), memorycmd.ListOptions{
				StartDir: dir, Limit: limit, Scope: scope, Kind: kind,
			})
		},
	}
	listC.Flags().IntVar(&limit, "limit", constants.DefaultListLimit, "max entries to return")
	listC.Flags().StringVar(&scope, "scope", "", "filter by scope")
	listC.Flags().StringVar(&kind, "kind", "", "filter by kind (working|semantic|experiential|long_term|multi_agent)")

	var (
		mScope, mKind, mContent, mRunID string
		mConf                           float64
	)
	promoteC := &cobra.Command{
		Use:   "promote",
		Short: "Promote an evidence-backed memory entry (gate-checked)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			dir, err := cwd()
			if err != nil {
				return err
			}
			return memorycmd.Promote(cmd.Context(), cmd.OutOrStdout(), memorycmd.PromoteOptions{
				StartDir: dir, Scope: mScope, Kind: mKind,
				Content: mContent, RunID: mRunID, Confidence: mConf,
			})
		},
	}
	promoteC.Flags().StringVar(&mScope, "scope", "project", "scope (project|session|global)")
	promoteC.Flags().StringVar(&mKind, "kind", "semantic", "kind (working|semantic|experiential|long_term|multi_agent — paper §3.2)")
	promoteC.Flags().StringVar(&mContent, "content", "", "memory content (required)")
	promoteC.Flags().StringVar(&mRunID, "run-id", "", "evidence run id (required)")
	promoteC.Flags().Float64Var(&mConf, "confidence", 0.7, "confidence 0..1 (>= 0.4 required)")

	var recallLimit int
	recallC := &cobra.Command{
		Use:   "recall \"<query>\"",
		Short: "Search past run reports by keyword overlap (no LLM)",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := cwd()
			if err != nil {
				return err
			}
			hits, err := recall.Recall(dir, strings.Join(args, " "), recallLimit)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			if len(hits) == 0 {
				fmt.Fprintln(out, "no matches")
				return nil
			}
			tw := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
			fmt.Fprintln(tw, "SCORE\tRUN\tSNIPPET")
			for _, h := range hits {
				fmt.Fprintf(tw, "%.2f\t%s\t%s\n", h.Score, h.RunID, h.Snippet)
			}
			return tw.Flush()
		},
	}
	recallC.Flags().IntVar(&recallLimit, "limit", 5, "max matches")

	c.AddCommand(listC, promoteC, recallC, newMemoryLearnCmd(), newMemoryConsolidateCmd(), newMemoryScoreCmd())
	return c
}
