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

	"github.com/ropeixoto/harnessx/internal/auditrun"
	"github.com/ropeixoto/harnessx/internal/runtime/containers"
	"github.com/ropeixoto/harnessx/internal/stack"
)

func newStackCmd() *cobra.Command {
	c := &cobra.Command{Use: "stack", Short: "Deterministic stack tour + status"}
	tour := &cobra.Command{
		Use:   "tour",
		Short: "Walk every Harness feature against a temp project",
		RunE:  runTour,
	}
	tour.Flags().String("root", "", "project root (default: mktemp)")
	tour.Flags().Bool("keep", false, "keep the temp project + registry after the tour")
	tour.Flags().Bool("dashboard", false, "spawn the dashboard and probe /api/health")
	tour.Flags().String("addr", "127.0.0.1:17820", "dashboard listen address")
	tour.Flags().String("templates", "", "templates dir to copy into the project (default: <repo>/templates)")

	status := &cobra.Command{
		Use:   "status",
		Short: "Probe the dashboard health endpoint (exits 0 when offline; --strict to fail)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			addr, _ := cmd.Flags().GetString("addr")
			strict, _ := cmd.Flags().GetBool("strict")
			probe := containers.HealthProbe{
				URL:     fmt.Sprintf("http://%s/api/health", addr),
				Client:  &http.Client{Timeout: 2 * time.Second},
				Timeout: 3 * time.Second,
				Backoff: 200 * time.Millisecond,
			}
			if err := probe.Wait(cmd.Context()); err != nil {
				fmt.Fprintf(cmd.OutOrStdout(), "dashboard: offline at http://%s — start with `harness dashboard --addr %s`\n", addr, addr)
				if strict {
					return err
				}
				return nil
			}
			fmt.Fprintf(cmd.OutOrStdout(), "dashboard: online at http://%s\n", addr)
			return nil
		},
	}
	status.Flags().String("addr", "127.0.0.1:7373", "dashboard listen address")
	status.Flags().Bool("strict", false, "exit non-zero when dashboard is offline (for CI gates)")

	audit := &cobra.Command{
		Use:   "audit",
		Short: "Run the deterministic visual + functional audit pipeline",
		RunE: func(cmd *cobra.Command, _ []string) error {
			repoRoot, err := cwd()
			if err != nil {
				return err
			}
			opts := auditrun.DefaultOptionsFromEnv(repoRoot)
			opts.Out = cmd.OutOrStdout()
			_, err = auditrun.New(opts).Run(cmd.Context())
			return err
		},
	}

	c.AddCommand(tour, status, audit)
	return c
}

func runTour(cmd *cobra.Command, _ []string) error {
	root, _ := cmd.Flags().GetString("root")
	keep, _ := cmd.Flags().GetBool("keep")
	withDashboard, _ := cmd.Flags().GetBool("dashboard")
	addr, _ := cmd.Flags().GetString("addr")
	templates, _ := cmd.Flags().GetString("templates")

	root, cleanupRoot, err := resolveTourRoot(root, keep)
	if err != nil {
		return err
	}
	if cleanupRoot != nil {
		defer cleanupRoot()
	}
	repoRoot, err := cwd()
	if err != nil {
		return err
	}
	if templates == "" {
		templates = filepath.Join(repoRoot, "templates")
	}
	t := &stack.Tour{
		Root:         root,
		TemplatesSrc: templates,
		RegistryPath: filepath.Join(root, ".harness", "registry.sqlite"),
	}
	if withDashboard {
		cleanupDash, err := attachDashboard(cmd.Context(), root, addr, t)
		if err != nil {
			return err
		}
		defer cleanupDash()
	}
	_, err = t.Run(cmd.Context(), cmd.OutOrStdout())
	return err
}

func resolveTourRoot(root string, keep bool) (string, func(), error) {
	if root == "" {
		tmp, err := os.MkdirTemp("", "harnessx-tour-")
		if err != nil {
			return "", nil, err
		}
		root = tmp
	}
	if keep {
		return root, nil, nil
	}
	return root, func() { _ = os.RemoveAll(root) }, nil
}

func attachDashboard(ctx context.Context, root, addr string, t *stack.Tour) (func(), error) {
	if err := os.MkdirAll(root, 0o755); err != nil {
		return nil, err
	}
	binary, err := os.Executable()
	if err != nil {
		return nil, err
	}
	probeURL := fmt.Sprintf("http://%s/api/health", addr)
	dashCtx, cancel := context.WithCancel(ctx)
	dash := exec.CommandContext(dashCtx, binary, "dashboard", "--addr", addr)
	dash.Dir = root
	dash.Stdout = os.Stderr
	dash.Stderr = os.Stderr
	if err := dash.Start(); err != nil {
		cancel()
		return nil, err
	}
	t.DashboardProbe = probeURL
	t.Probe = containers.HealthProbe{
		URL:     probeURL,
		Client:  &http.Client{Timeout: 2 * time.Second},
		Timeout: 10 * time.Second,
		Backoff: 200 * time.Millisecond,
	}
	return func() {
		_ = dash.Process.Kill()
		_, _ = dash.Process.Wait()
		cancel()
	}, nil
}
