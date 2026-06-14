// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/ropeixoto/harnessx/internal/runtime/containers"
	"github.com/ropeixoto/harnessx/internal/stack"
)

func newStackCmd() *cobra.Command {
	c := &cobra.Command{Use: "stack", Short: "Deterministic stack tour + status"}
	tour := &cobra.Command{
		Use:   "tour",
		Short: "Walk every Harness feature against a temp project",
		RunE: func(cmd *cobra.Command, _ []string) error {
			root, _ := cmd.Flags().GetString("root")
			keep, _ := cmd.Flags().GetBool("keep")
			withDashboard, _ := cmd.Flags().GetBool("dashboard")
			addr, _ := cmd.Flags().GetString("addr")
			templates, _ := cmd.Flags().GetString("templates")

			if root == "" {
				tmp, err := os.MkdirTemp("", "harnessx-tour-")
				if err != nil {
					return err
				}
				root = tmp
			}
			if !keep {
				defer os.RemoveAll(root)
			}
			repoRoot, err := cwd()
			if err != nil {
				return err
			}
			if templates == "" {
				templates = filepath.Join(repoRoot, "templates")
			}
			registry := filepath.Join(root, ".harness", "registry.sqlite")

			t := &stack.Tour{
				Root:         root,
				TemplatesSrc: templates,
				RegistryPath: registry,
			}

			if withDashboard {
				if err := os.MkdirAll(root, 0o755); err != nil {
					return err
				}
				binary, err := os.Executable()
				if err != nil {
					return err
				}
				probeURL := fmt.Sprintf("http://%s/api/health", addr)
				ctx, cancel := context.WithCancel(cmd.Context())
				defer cancel()
				dash := exec.CommandContext(ctx, binary, "dashboard", "--addr", addr)
				dash.Dir = root
				dash.Stdout = os.Stderr
				dash.Stderr = os.Stderr
				if err := dash.Start(); err != nil {
					return err
				}
				defer func() {
					_ = dash.Process.Kill()
					_, _ = dash.Process.Wait()
				}()
				t.DashboardProbe = probeURL
				t.Probe = containers.HealthProbe{
					URL:     probeURL,
					Client:  &http.Client{Timeout: 2 * time.Second},
					Timeout: 10 * time.Second,
					Backoff: 200 * time.Millisecond,
				}
			}

			_, err = t.Run(cmd.Context(), cmd.OutOrStdout())
			return err
		},
	}
	tour.Flags().String("root", "", "project root (default: mktemp)")
	tour.Flags().Bool("keep", false, "keep the temp project + registry after the tour")
	tour.Flags().Bool("dashboard", false, "spawn the dashboard and probe /api/health")
	tour.Flags().String("addr", "127.0.0.1:17820", "dashboard listen address")
	tour.Flags().String("templates", "", "templates dir to copy into the project (default: <repo>/templates)")

	status := &cobra.Command{
		Use:   "status",
		Short: "Probe the dashboard health endpoint",
		RunE: func(cmd *cobra.Command, _ []string) error {
			addr, _ := cmd.Flags().GetString("addr")
			probe := containers.HealthProbe{
				URL:     fmt.Sprintf("http://%s/api/health", addr),
				Client:  &http.Client{Timeout: 2 * time.Second},
				Timeout: 3 * time.Second,
				Backoff: 200 * time.Millisecond,
			}
			if err := probe.Wait(cmd.Context()); err != nil {
				fmt.Fprintf(cmd.OutOrStdout(), "dashboard: offline (%v)\n", err)
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "dashboard: online at http://%s\n", addr)
			return nil
		},
	}
	status.Flags().String("addr", "127.0.0.1:7373", "dashboard listen address")

	c.AddCommand(tour, status)
	return c
}
