// SPDX-License-Identifier: MIT

package main

import (
	"github.com/spf13/cobra"

	"github.com/ropeixoto/harnessx/internal/app/indexcmd"
	"github.com/ropeixoto/harnessx/internal/app/projectcmd"
)

func newProjectCmd() *cobra.Command {
	c := &cobra.Command{Use: "project", Short: "Project index + workspace registry"}
	c.AddCommand(newProjectWorkspaceCmds()...)

	var force bool
	indexC := &cobra.Command{
		Use:   "index",
		Short: "Build or refresh .harness/project/*.json maps",
		RunE: func(cmd *cobra.Command, _ []string) error {
			dir, err := cwd()
			if err != nil {
				return err
			}
			_, err = indexcmd.RunIndex(cmd.Context(), indexcmd.IndexOptions{StartDir: dir, Force: force}, cmd.OutOrStdout())
			return err
		},
	}
	indexC.Flags().BoolVar(&force, "force", false, "rebuild every map even when inputs are unchanged")

	var mapName string
	inspectC := &cobra.Command{
		Use:   "inspect [map]",
		Short: "List project maps or pretty-print one",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := cwd()
			if err != nil {
				return err
			}
			name := mapName
			if len(args) == 1 && name == "" {
				name = args[0]
			}
			return indexcmd.RunInspect(indexcmd.InspectOptions{StartDir: dir, Map: name}, cmd.OutOrStdout())
		},
	}
	inspectC.Flags().StringVar(&mapName, "map", "", "map name (e.g. profile, commands, dependencies)")

	c.AddCommand(indexC, inspectC)
	return c
}

func newProjectWorkspaceCmds() []*cobra.Command {
	opts := projectcmd.Options{}
	var registry string

	add := &cobra.Command{
		Use:   "add [path]",
		Short: "Register a project root in the workspace registry",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := ""
			if len(args) == 1 {
				path = args[0]
			}
			displayName, _ := cmd.Flags().GetString("name")
			slug, _ := cmd.Flags().GetString("slug")
			opts.RegistryPath = registry
			return projectcmd.Add(cmd.Context(), opts, path, displayName, slug, cmd.OutOrStdout())
		},
	}
	add.Flags().String("name", "", "display name (defaults to folder basename)")
	add.Flags().String("slug", "", "kebab-case slug (defaults to folder basename)")
	add.Flags().StringVar(&registry, "registry", "", "registry SQLite path (defaults to ~/.harness/registry.sqlite)")

	list := &cobra.Command{
		Use:   "list",
		Short: "Show registered projects",
		RunE: func(cmd *cobra.Command, _ []string) error {
			showArchived, _ := cmd.Flags().GetBool("archived")
			opts.RegistryPath = registry
			return projectcmd.List(cmd.Context(), opts, showArchived, cmd.OutOrStdout())
		},
	}
	list.Flags().Bool("archived", false, "include archived projects")
	list.Flags().StringVar(&registry, "registry", "", "registry SQLite path")

	switchCmd := &cobra.Command{
		Use:   "switch <slug|path|id>",
		Short: "Set the active project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.RegistryPath = registry
			return projectcmd.Switch(cmd.Context(), opts, args[0], cmd.OutOrStdout())
		},
	}
	switchCmd.Flags().StringVar(&registry, "registry", "", "registry SQLite path")

	current := &cobra.Command{
		Use:   "current",
		Short: "Print the resolved project",
		RunE: func(cmd *cobra.Command, _ []string) error {
			flag, _ := cmd.Flags().GetString("project")
			opts.RegistryPath = registry
			return projectcmd.Current(cmd.Context(), opts, flag, cmd.OutOrStdout())
		},
	}
	current.Flags().String("project", "", "explicit project ref (overrides env + cwd)")
	current.Flags().StringVar(&registry, "registry", "", "registry SQLite path")

	archive := &cobra.Command{
		Use:   "archive <slug|path|id>",
		Short: "Mark a project as archived",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.RegistryPath = registry
			return projectcmd.Archive(cmd.Context(), opts, args[0], cmd.OutOrStdout())
		},
	}
	archive.Flags().StringVar(&registry, "registry", "", "registry SQLite path")

	unarchive := &cobra.Command{
		Use:   "unarchive <slug|path|id>",
		Short: "Restore an archived project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.RegistryPath = registry
			return projectcmd.Unarchive(cmd.Context(), opts, args[0], cmd.OutOrStdout())
		},
	}
	unarchive.Flags().StringVar(&registry, "registry", "", "registry SQLite path")

	scan := &cobra.Command{
		Use:   "scan [root]",
		Short: "Discover .harness/ folders under root and offer to register them",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root := ""
			if len(args) == 1 {
				root = args[0]
			}
			registerAll, _ := cmd.Flags().GetBool("yes")
			opts.RegistryPath = registry
			return projectcmd.Scan(cmd.Context(), opts, root, registerAll, cmd.OutOrStdout())
		},
	}
	scan.Flags().Bool("yes", false, "register every candidate without asking")
	scan.Flags().StringVar(&registry, "registry", "", "registry SQLite path")

	forget := &cobra.Command{
		Use:   "forget <slug|path|id>",
		Short: "Remove a registry row (does not touch project files)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.RegistryPath = registry
			return projectcmd.Forget(cmd.Context(), opts, args[0], cmd.OutOrStdout())
		},
	}
	forget.Flags().StringVar(&registry, "registry", "", "registry SQLite path")

	return []*cobra.Command{add, list, switchCmd, current, archive, unarchive, scan, forget}
}
