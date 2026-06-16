// SPDX-License-Identifier: MIT

package main

import (
	"fmt"

	"github.com/ropeixoto/harnessx/internal/auditsolid"
	"github.com/spf13/cobra"
)

func newAuditSolidCmd() *cobra.Command {
	var root string
	c := &cobra.Command{
		Use:   "audit-solid",
		Short: "Scan Go sources for SOLID smells (god files, fan-out)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if root == "" {
				root = "."
			}
			v, err := auditsolid.Scan(root, auditsolid.Default())
			if err != nil {
				return err
			}
			fmt.Fprint(cmd.OutOrStdout(), auditsolid.Report(v))
			if len(v) > 0 {
				return fmt.Errorf("%d SOLID violation(s)", len(v))
			}
			return nil
		},
	}
	c.Flags().StringVar(&root, "root", ".", "scan root")
	return c
}
