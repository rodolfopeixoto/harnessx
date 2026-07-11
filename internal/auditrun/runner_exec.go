// SPDX-License-Identifier: MIT

package auditrun

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/ropeixoto/harnessx/internal/platform/constants"
	"github.com/ropeixoto/harnessx/internal/runtime/containers"
)

func (r *Runner) ensureDashboard(ctx context.Context, base string) (func(), error) {
	probe := containers.HealthProbe{
		URL:     r.opts.BaseURL + "/api/health",
		Client:  &http.Client{Timeout: 2 * time.Second},
		Timeout: 2 * time.Second,
		Backoff: 200 * time.Millisecond,
	}
	if err := probe.Wait(ctx); err == nil {
		return nil, nil
	}
	if r.opts.DashboardLauncher != nil {
		return r.opts.DashboardLauncher(ctx, baseAddr(r.opts.BaseURL))
	}
	binary, err := os.Executable()
	if err != nil {
		return nil, err
	}
	cmd := exec.CommandContext(ctx, binary, "dashboard", "--addr", baseAddr(r.opts.BaseURL))
	cmd.Dir = base
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	if err := probe.Wait(ctx); err != nil {
		_ = cmd.Process.Kill()
		return nil, fmt.Errorf("audit: dashboard never became healthy: %w", err)
	}
	return func() {
		_ = cmd.Process.Kill()
		_, _ = cmd.Process.Wait()
	}, nil
}

func (r *Runner) runPlaywright(ctx context.Context, p Paths) ([]Result, error) {
	dashboardDir := filepath.Join(r.opts.RepoRoot, "web", "dashboard")
	if _, err := os.Stat(filepath.Join(dashboardDir, "node_modules", "@playwright", "test")); err != nil {
		return nil, errors.New("playwright not installed under web/dashboard/node_modules")
	}
	args := []string{"playwright", "test", "audit/audit.spec.ts", "--reporter=list"}
	if r.opts.Headed {
		args = append(args, "--headed")
	}
	cmd := exec.CommandContext(ctx, "npx", args...)
	cmd.Dir = dashboardDir
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("AUDIT_BASE_URL=%s", r.opts.BaseURL),
		fmt.Sprintf("AUDIT_OUT=%s", p.Base),
	)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return nil, err
	}
	results, err := ReadResults(filepath.Join(p.JSON, constants.AuditResultsFile))
	if err != nil {
		return nil, err
	}
	return results.Results, nil
}

func writeAuxJSON(p Paths, results []Result) error {
	type bundles struct {
		Console   []ConsoleError    `json:"console_errors"`
		Network   []NetworkError    `json:"network_errors"`
		Selectors []MissingSelector `json:"missing_selectors"`
		Visual    []VisualDiff      `json:"visual_diffs"`
		Layout    []LayoutMetric    `json:"layout_metrics"`
	}
	var b bundles
	for _, r := range results {
		b.Console = append(b.Console, r.ConsoleErrors...)
		b.Network = append(b.Network, r.NetworkErrors...)
		for _, sel := range r.MissingSelectors {
			b.Selectors = append(b.Selectors, MissingSelector{FeatureID: r.FeatureID, Viewport: r.Viewport, Selector: sel})
		}
		if r.Visual != nil {
			b.Visual = append(b.Visual, *r.Visual)
		}
		if r.Layout != nil {
			b.Layout = append(b.Layout, *r.Layout)
		}
	}
	pairs := map[string]any{
		constants.AuditConsoleFile:    b.Console,
		constants.AuditNetworkFile:    b.Network,
		constants.AuditSelectorsFile:  b.Selectors,
		constants.AuditVisualDiffFile: b.Visual,
		constants.AuditLayoutFile:     b.Layout,
	}
	for name, payload := range pairs {
		if err := WriteJSONFile(filepath.Join(p.JSON, name), payload); err != nil {
			return err
		}
	}
	return nil
}

func writePDF(_ context.Context, htmlPath, pdfPath string) error {
	if _, err := exec.LookPath("npx"); err != nil {
		return errors.New("npx not on PATH; pdf rendering skipped")
	}
	cmd := exec.Command("npx", "playwright", "screenshot", "--full-page", "--device=Desktop Chrome", "file://"+htmlPath, pdfPath)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}
