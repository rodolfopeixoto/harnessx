// SPDX-License-Identifier: MIT

package main

import (
	"github.com/spf13/cobra"

	"github.com/ropeixoto/harnessx/internal/app/memorycmd"
	"github.com/ropeixoto/harnessx/internal/platform/constants"
)

func newMemoryCmd() *cobra.Command {
	c := &cobra.Command{Use: "memory", Short: "Project memory commands"}

	var limit int
	var scope string
	listC := &cobra.Command{
		Use:   "list",
		Short: "List recent project memory entries",
		RunE: func(cmd *cobra.Command, _ []string) error {
			dir, err := cwd()
			if err != nil {
				return err
			}
			return memorycmd.List(cmd.OutOrStdout(), memorycmd.ListOptions{
				StartDir: dir, Limit: limit, Scope: scope,
			})
		},
	}
	listC.Flags().IntVar(&limit, "limit", constants.DefaultListLimit, "max entries to return")
	listC.Flags().StringVar(&scope, "scope", "", "filter by scope")

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
	promoteC.Flags().StringVar(&mKind, "kind", "fact", "kind (fact|convention|failure|success)")
	promoteC.Flags().StringVar(&mContent, "content", "", "memory content (required)")
	promoteC.Flags().StringVar(&mRunID, "run-id", "", "evidence run id (required)")
	promoteC.Flags().Float64Var(&mConf, "confidence", 0.7, "confidence 0..1 (>= 0.4 required)")

	c.AddCommand(listC, promoteC)
	return c
}
