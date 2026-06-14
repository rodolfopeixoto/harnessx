// SPDX-License-Identifier: MIT

package main

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/ropeixoto/harnessx/internal/app/explaincmd"
)

func newExplainCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "explain <prompt>",
		Short: "Dry-run intent classifier + router for a prompt",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := cwd()
			if err != nil {
				return err
			}
			return explaincmd.Run(cmd.OutOrStdout(), explaincmd.Options{
				StartDir: dir, Prompt: strings.Join(args, " "),
			})
		},
	}
}
