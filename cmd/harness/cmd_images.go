package main

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/ropeixoto/harnessx/internal/app/imagescmd"
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
			return imagescmd.List(cmd.Context(), cmd.OutOrStdout(), imagescmd.Options{Root: root, JSON: jsonOut})
		},
	}
	c.Flags().BoolVar(&jsonOut, "json", false, "emit JSON")
	return c
}

func newImagesPruneCmd() *cobra.Command {
	var olderThan time.Duration
	c := &cobra.Command{
		Use:   "prune",
		Short: "Prune dangling images (two-key safety)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			root, err := cwd()
			if err != nil {
				return err
			}
			confirmed := os.Getenv(containersConfirmEnv) == "1"
			if !confirmed {
				fmt.Fprintf(cmd.OutOrStdout(), "About to prune dangling images.\nSet %s=1 to confirm.\n", containersConfirmEnv)
				return fmt.Errorf("prune aborted (set %s=1)", containersConfirmEnv)
			}
			return imagescmd.Prune(cmd.Context(), cmd.OutOrStdout(), imagescmd.PruneOptions{Root: root, OlderThan: olderThan, Confirmed: confirmed})
		},
	}
	c.Flags().DurationVar(&olderThan, "older-than", 0, "only prune images older than this duration")
	return c
}
