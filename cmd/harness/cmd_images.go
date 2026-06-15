// SPDX-License-Identifier: MIT

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/ropeixoto/harnessx/internal/runtime/containers"
)

func newImagesCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "images",
		Short: "List and prune container images via the selected runtime",
	}
	c.AddCommand(newImagesListCmd(), newImagesPruneCmd())
	return c
}

func newImagesListCmd() *cobra.Command {
	var jsonOut bool
	c := &cobra.Command{
		Use:   "list",
		Short: "List container images",
		RunE: func(cmd *cobra.Command, _ []string) error {
			root, err := cwd()
			if err != nil {
				return err
			}
			ctx := cmd.Context()
			rt, _, err := containers.Resolve(ctx, root)
			if err != nil {
				return err
			}
			images, err := rt.ListImages(ctx)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			if jsonOut {
				return json.NewEncoder(out).Encode(images)
			}
			w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "REPOSITORY\tTAG\tID\tCREATED")
			for _, img := range images {
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", img.Repository, img.Tag, short(img.ID, 12), img.CreatedAt.Format("2006-01-02"))
			}
			return w.Flush()
		},
	}
	c.Flags().BoolVar(&jsonOut, "json", false, "emit JSON")
	return c
}

func newImagesPruneCmd() *cobra.Command {
	var (
		olderThan time.Duration
	)
	c := &cobra.Command{
		Use:   "prune",
		Short: "Prune dangling images (two-key safety)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			root, err := cwd()
			if err != nil {
				return err
			}
			ctx := cmd.Context()
			rt, _, err := containers.Resolve(ctx, root)
			if err != nil {
				return err
			}
			opts := containers.ImagePruneOptions{OlderThan: olderThan}
			opts.IUnderstand = os.Getenv(containersConfirmEnv) == "1"
			if !opts.IUnderstand {
				fmt.Fprintf(cmd.OutOrStdout(), "About to prune dangling images via %s.\nSet %s=1 to confirm.\n", rt.ID(), containersConfirmEnv)
				return fmt.Errorf("prune aborted (set %s=1)", containersConfirmEnv)
			}
			if _, err := rt.PruneImages(ctx, opts); err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), "images pruned")
			return nil
		},
	}
	c.Flags().DurationVar(&olderThan, "older-than", 0, "only prune images older than this duration")
	return c
}
