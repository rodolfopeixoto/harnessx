// SPDX-License-Identifier: MIT

package main

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/ropeixoto/harnessx/internal/app/designcmd"
)

func newDesignToProductCmd() *cobra.Command {
	var src string
	c := &cobra.Command{
		Use:   "design-to-product <prompt>",
		Short: "Convert a design (ZIP/folder) to React parity + toggles + roadmap",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := cwd()
			if err != nil {
				return err
			}
			_, err = designcmd.Run(cmd.Context(), designcmd.Options{
				StartDir: dir, Prompt: strings.Join(args, " "), Source: src,
			}, cmd.OutOrStdout())
			return err
		},
	}
	c.Flags().StringVar(&src, "source", "", "explicit path to design ZIP or folder")
	return c
}
