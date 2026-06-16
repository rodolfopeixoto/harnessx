// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/spf13/cobra"

	"github.com/ropeixoto/harnessx/internal/app/projectcmd"
	"github.com/ropeixoto/harnessx/internal/execution"
)

func newRunsPruneCmd() *cobra.Command {
	var (
		olderThan string
		keepLast  int
		apply     bool
	)
	c := &cobra.Command{
		Use:   "prune",
		Short: "Delete old run directories (default dry-run; pass --apply to remove)",
		Long: `Lists run directories under .harness/runs that are eligible for
deletion under the retention policy:

  --older-than <dur>   e.g. 7d, 24h, 30d (default 0 = disabled)
  --keep-last <n>      always keep the last N runs (default 0 = disabled)
  --apply              actually delete (default: dry-run)

When both --older-than and --keep-last are set, a run is pruned if it
matches EITHER (older than AND not in keep-last window).`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runRunsPrune(cmd.Context(), cmd.OutOrStdout(), olderThan, keepLast, apply)
		},
	}
	c.Flags().StringVar(&olderThan, "older-than", "", "duration (e.g. 7d, 30d, 24h)")
	c.Flags().IntVar(&keepLast, "keep-last", 0, "always keep the last N runs")
	c.Flags().BoolVar(&apply, "apply", false, "actually delete (default dry-run)")
	return c
}

func runRunsPrune(_ context.Context, out io.Writer, olderThan string, keepLast int, apply bool) error {
	dir, err := cwd()
	if err != nil {
		return err
	}
	dur, err := parseDuration(olderThan)
	if err != nil {
		return err
	}
	if dur == 0 && keepLast == 0 {
		return fmt.Errorf("prune: pass --older-than and/or --keep-last")
	}
	candidates, err := execution.PruneCandidates(dir, dur, keepLast)
	if err != nil {
		return err
	}
	if len(candidates) == 0 {
		fmt.Fprintln(out, "no runs match the retention policy")
		return nil
	}
	mode := "dry-run"
	if apply {
		mode = "applied"
	}
	fmt.Fprintf(out, "runs prune (%s) — %d candidate(s):\n", mode, len(candidates))
	for _, p := range candidates {
		fmt.Fprintf(out, "  · %s\n", p)
	}
	if !apply {
		fmt.Fprintln(out, "pass --apply to delete")
		return nil
	}
	freed, err := execution.DeletePaths(candidates)
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "freed %.2f MiB\n", float64(freed)/(1024*1024))
	return nil
}

func newProjectPruneCmd() *cobra.Command {
	var (
		olderThan string
		apply     bool
	)
	c := &cobra.Command{
		Use:   "prune",
		Short: "Archive projects whose last_seen_at is older than --older-than",
		Long: `Marks projects as archived in the workspace registry when their
last_seen_at is older than the threshold. Archived projects stay in
the DB; use 'harness project unarchive' to restore. Pass --apply to
actually archive (default: dry-run).`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runProjectPrune(cmd.Context(), cmd.OutOrStdout(), olderThan, apply)
		},
	}
	c.Flags().StringVar(&olderThan, "older-than", "", "duration (e.g. 30d, 90d)")
	c.Flags().BoolVar(&apply, "apply", false, "actually archive (default dry-run)")
	return c
}

func runProjectPrune(ctx context.Context, out io.Writer, olderThan string, apply bool) error {
	dur, err := parseDuration(olderThan)
	if err != nil {
		return err
	}
	if dur == 0 {
		return fmt.Errorf("project prune: pass --older-than (e.g. 30d)")
	}
	stale, err := projectcmd.StaleSince(ctx, projectcmd.Options{}, time.Now().Add(-dur))
	if err != nil {
		return err
	}
	if len(stale) == 0 {
		fmt.Fprintln(out, "no projects older than threshold")
		return nil
	}
	mode := "dry-run"
	if apply {
		mode = "applied"
	}
	fmt.Fprintf(out, "project prune (%s) — %d candidate(s):\n", mode, len(stale))
	for _, p := range stale {
		fmt.Fprintf(out, "  · %s  last_seen=%s\n", p.Slug, p.LastSeenAt)
	}
	if !apply {
		fmt.Fprintln(out, "pass --apply to archive")
		return nil
	}
	for _, p := range stale {
		if err := projectcmd.Archive(ctx, projectcmd.Options{}, p.Slug, out); err != nil {
			fmt.Fprintf(out, "  ✗ %s: %v\n", p.Slug, err)
		}
	}
	return nil
}

func parseDuration(s string) (time.Duration, error) {
	if s == "" {
		return 0, nil
	}
	// support Nd suffix
	if l := len(s); l > 1 && (s[l-1] == 'd' || s[l-1] == 'D') {
		var days int
		_, err := fmt.Sscanf(s, "%dd", &days)
		if err != nil {
			return 0, fmt.Errorf("parse duration %q: %w", s, err)
		}
		return time.Duration(days) * 24 * time.Hour, nil
	}
	return time.ParseDuration(s)
}
