// SPDX-License-Identifier: MIT

package main

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/ropeixoto/harnessx/internal/app/speccmd"
	"github.com/ropeixoto/harnessx/internal/domain"
)

func newSpecCmd() *cobra.Command {
	c := &cobra.Command{Use: "spec", Short: "Spec-driven-development helpers"}

	var name, mode string
	initC := &cobra.Command{
		Use:   "init [prompt]",
		Short: "Scaffold a fresh spec under .harness/artifacts/specs/",
		Args:  cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := cwd()
			if err != nil {
				return err
			}
			prompt := strings.Join(args, " ")
			return speccmd.Init(cmd.OutOrStdout(), speccmd.InitOptions{
				StartDir: dir, Name: name, Prompt: prompt, Mode: domain.Mode(mode),
			})
		},
	}
	initC.Flags().StringVar(&name, "name", "", "spec name (required, used as filename slug)")
	initC.Flags().StringVar(&mode, "mode", string(domain.ModeFeature),
		"mode (feature|bugfix|design_to_product|optimization|audit|review|setup)")

	c.AddCommand(initC, newSpecAuthorCmd())
	return c
}
