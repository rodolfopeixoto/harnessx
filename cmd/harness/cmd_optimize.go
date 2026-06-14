// SPDX-License-Identifier: MIT

package main

import (
	"github.com/spf13/cobra"

	"github.com/ropeixoto/harnessx/internal/app/optimizecmd"
)

func newOptimizeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "optimize [resources]",
		Short: "Run the full A→G resource-optimization cycle",
		RunE: func(cmd *cobra.Command, _ []string) error {
			dir, err := cwd()
			if err != nil {
				return err
			}
			return optimizecmd.Optimize(cmd.Context(), optimizecmd.AuditOptions{StartDir: dir}, cmd.OutOrStdout())
		},
	}
}

func newPerfSnapshotCmd() *cobra.Command {
	var label string
	var report bool
	c := &cobra.Command{
		Use:   "perf-snapshot",
		Short: "Capture a resource snapshot (Cycle A)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			dir, err := cwd()
			if err != nil {
				return err
			}
			return optimizecmd.PerfSnapshot(optimizecmd.SnapshotOptions{
				StartDir: dir, Label: label, Report: report,
			}, cmd.OutOrStdout())
		},
	}
	c.Flags().StringVar(&label, "label", "", "human label for the snapshot")
	c.Flags().BoolVar(&report, "report", false, "also write a markdown report")
	return c
}

func newPerfCompareCmd() *cobra.Command {
	var from, to string
	c := &cobra.Command{
		Use:   "perf-compare [from] [to]",
		Short: "Diff two snapshots (default: the two most recent)",
		Args:  cobra.MaximumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := cwd()
			if err != nil {
				return err
			}
			if len(args) >= 2 && from == "" && to == "" {
				from, to = args[0], args[1]
			}
			return optimizecmd.PerfCompare(optimizecmd.CompareOptions{
				StartDir: dir, From: from, To: to,
			}, cmd.OutOrStdout())
		},
	}
	c.Flags().StringVar(&from, "from", "", "path to older snapshot json")
	c.Flags().StringVar(&to, "to", "", "path to newer snapshot json")
	return c
}

func newImageAuditCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "image-audit",
		Short: "Static Dockerfile audit (Cycle B)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			dir, err := cwd()
			if err != nil {
				return err
			}
			return optimizecmd.ImageAudit(optimizecmd.AuditOptions{StartDir: dir}, cmd.OutOrStdout())
		},
	}
}

func newDependencyAuditCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "dependency-audit",
		Short: "Classify dependencies + flag removal candidates (Cycle C)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			dir, err := cwd()
			if err != nil {
				return err
			}
			return optimizecmd.DependencyAudit(optimizecmd.AuditOptions{StartDir: dir}, cmd.OutOrStdout())
		},
	}
}

func newLogAuditCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "log-audit",
		Short: "Surface noisy log call sites (Cycle D)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			dir, err := cwd()
			if err != nil {
				return err
			}
			return optimizecmd.LogAudit(optimizecmd.AuditOptions{StartDir: dir}, cmd.OutOrStdout())
		},
	}
}

func newSecurityAuditCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "security-audit",
		Short: "Run the security-category sensors",
		RunE: func(cmd *cobra.Command, _ []string) error {
			dir, err := cwd()
			if err != nil {
				return err
			}
			return optimizecmd.SecurityAudit(cmd.Context(), optimizecmd.AuditOptions{StartDir: dir}, cmd.OutOrStdout())
		},
	}
}
