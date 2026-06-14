// SPDX-License-Identifier: MIT

package auditrun

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/ropeixoto/harnessx/internal/platform/constants"
	"github.com/ropeixoto/harnessx/internal/runtime/containers"
)

type Options struct {
	RepoRoot          string
	BaseURL           string
	Timestamp         string
	Keep              bool
	Headed            bool
	OnlyMobile        bool
	OnlyVisual        bool
	OnlyFeature       string
	OnlyRole          string
	ReferencePath     string
	PreviousReport    string
	Out               io.Writer
	PlaywrightSkip    bool
	DashboardLauncher func(ctx context.Context, addr string) (cleanup func(), err error)
}

func DefaultOptionsFromEnv(repoRoot string) Options {
	timestamp := time.Now().UTC().Format("20060102T150405Z")
	opts := Options{
		RepoRoot:       repoRoot,
		BaseURL:        envOrDefault(constants.EnvAuditBaseURL, constants.AuditDefaultBaseURL),
		Timestamp:      timestamp,
		Keep:           os.Getenv(constants.EnvAuditKeep) == "1",
		Headed:         os.Getenv(constants.EnvAuditHeaded) == "1",
		OnlyMobile:     os.Getenv(constants.EnvAuditMobile) == "1",
		OnlyVisual:     os.Getenv(constants.EnvAuditVisual) == "1",
		OnlyFeature:    os.Getenv(constants.EnvAuditFeature),
		OnlyRole:       os.Getenv(constants.EnvAuditRole),
		ReferencePath:  os.Getenv(constants.EnvAuditReference),
		PreviousReport: os.Getenv(constants.EnvAuditPrevReport),
		PlaywrightSkip: os.Getenv(constants.EnvAuditPlaywrightSkip) == "1",
	}
	return opts
}

type Runner struct {
	opts Options
}

func New(opts Options) *Runner {
	if opts.Out == nil {
		opts.Out = os.Stdout
	}
	if opts.Timestamp == "" {
		opts.Timestamp = time.Now().UTC().Format("20060102T150405Z")
	}
	return &Runner{opts: opts}
}

func (r *Runner) Layout() Paths {
	base := filepath.Join(r.opts.RepoRoot, constants.AuditRootDir, r.opts.Timestamp)
	return Paths{
		Base:                 base,
		JSON:                 filepath.Join(base, "json"),
		ReportDir:            filepath.Join(base, "report"),
		CurrentScreenshots:   filepath.Join(base, "current", "screenshots"),
		ReferenceScreenshots: filepath.Join(base, "reference", "screenshots"),
		Diff:                 filepath.Join(base, "diff"),
		RunLog:               filepath.Join(base, constants.AuditRunLogFile),
	}
}

type Paths struct {
	Base                 string
	JSON                 string
	ReportDir            string
	CurrentScreenshots   string
	ReferenceScreenshots string
	Diff                 string
	RunLog               string
}

func (r *Runner) Run(ctx context.Context) (Summary, error) {
	paths := r.Layout()
	if err := mkdirs(paths); err != nil {
		return Summary{}, err
	}
	logFile, err := os.Create(paths.RunLog)
	if err != nil {
		return Summary{}, err
	}
	defer logFile.Close()
	logf := newLogger(logFile)

	features, viewports, err := r.prepareScope(paths, logf)
	if err != nil {
		return Summary{}, err
	}
	cleanup, err := r.startDashboard(ctx, paths.Base, logf)
	if err != nil {
		return Summary{}, err
	}
	if !r.opts.Keep && cleanup != nil {
		defer cleanup()
	}
	results := r.collectResults(ctx, paths, features, viewports, logf)
	if err := r.persistResults(paths, results); err != nil {
		return Summary{}, err
	}
	summary := Aggregate(results, features)
	summary.BaseURL = r.opts.BaseURL
	if err := WriteJSONFile(filepath.Join(paths.JSON, constants.AuditSummaryFile), summary); err != nil {
		return Summary{}, err
	}
	if err := r.renderReports(ctx, paths, summary, features, results, logf); err != nil {
		return Summary{}, err
	}
	logf("audit finished base=%s pass_rate=%.2f", r.opts.BaseURL, summary.PassRate)
	return summary, nil
}

type logger func(format string, a ...any)

func newLogger(out io.Writer) logger {
	return func(format string, a ...any) {
		fmt.Fprintf(out, time.Now().UTC().Format(time.RFC3339)+" "+format+"\n", a...)
	}
}

func (r *Runner) prepareScope(paths Paths, logf logger) ([]Feature, []Viewport, error) {
	features := DefaultFeatures(r.opts.BaseURL)
	features = filterFeatures(features, r.opts.OnlyFeature, r.opts.OnlyRole)
	viewports := DefaultViewports()
	if r.opts.OnlyMobile {
		viewports = filterViewport(viewports, constants.AuditViewportMob)
	}
	if _, err := WriteFeatureMap(paths.JSON, r.opts.BaseURL, features, viewports); err != nil {
		return nil, nil, err
	}
	logf("feature map written (%d features, %d viewports)", len(features), len(viewports))
	return features, viewports, nil
}

func (r *Runner) startDashboard(ctx context.Context, base string, logf logger) (func(), error) {
	cleanup, err := r.ensureDashboard(ctx, base)
	if err != nil {
		logf("dashboard launch failed: %v", err)
	}
	return cleanup, err
}

func (r *Runner) collectResults(ctx context.Context, paths Paths, features []Feature, viewports []Viewport, logf logger) []Result {
	if r.opts.PlaywrightSkip {
		logf("playwright skipped via env")
		return synthesiseSkipped(features, viewports, "AUDIT_PLAYWRIGHT_SKIP=1")
	}
	runResults, runErr := r.runPlaywright(ctx, paths)
	if runErr != nil {
		logf("playwright run failed: %v", runErr)
		return synthesiseSkipped(features, viewports, runErr.Error())
	}
	return runResults
}

func (r *Runner) persistResults(paths Paths, results []Result) error {
	if err := WriteJSONFile(filepath.Join(paths.JSON, constants.AuditResultsFile),
		Results{GeneratedAt: time.Now().UTC(), BaseURL: r.opts.BaseURL, Results: results}); err != nil {
		return err
	}
	return writeAuxJSON(paths, results)
}

func (r *Runner) renderReports(ctx context.Context, paths Paths, summary Summary, features []Feature, results []Result, logf logger) error {
	backlog := BuildBacklog(results, features)
	if err := WriteBacklog(filepath.Join(paths.ReportDir, constants.AuditBacklogFile), backlog); err != nil {
		return err
	}
	htmlPath := filepath.Join(paths.ReportDir, constants.AuditHTMLFile)
	report := ReportInput{
		Timestamp: r.opts.Timestamp,
		BaseURL:   r.opts.BaseURL,
		Summary:   summary,
		Features:  features,
		Results:   results,
		Backlog:   backlog,
	}
	if err := WriteHTMLReport(htmlPath, report); err != nil {
		return err
	}
	pdfPath := filepath.Join(paths.ReportDir, constants.AuditPDFFile)
	if err := writePDF(ctx, htmlPath, pdfPath); err != nil {
		logf("pdf render skipped: %v", err)
	}
	artifacts := map[string]string{
		"PDF":                   pdfPath,
		"HTML":                  htmlPath,
		"Screenshots current":   paths.CurrentScreenshots,
		"Screenshots reference": paths.ReferenceScreenshots,
		"Diffs":                 paths.Diff,
		"JSON":                  filepath.Join(paths.JSON, constants.AuditResultsFile),
		"Backlog":               filepath.Join(paths.ReportDir, constants.AuditBacklogFile),
		"Log":                   paths.RunLog,
	}
	fmt.Fprint(r.opts.Out, TerminalSummary(report, artifacts))
	return nil
}

func mkdirs(p Paths) error {
	for _, d := range []string{p.Base, p.JSON, p.ReportDir, p.CurrentScreenshots, p.ReferenceScreenshots, p.Diff} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return err
		}
	}
	return nil
}

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

func synthesiseSkipped(features []Feature, viewports []Viewport, reason string) []Result {
	now := time.Now().UTC()
	var out []Result
	for _, f := range features {
		for _, v := range viewports {
			if len(f.Viewports) > 0 && !contains(f.Viewports, v.Name) {
				continue
			}
			out = append(out, Result{
				FeatureID:  f.ID,
				Viewport:   v.Name,
				Status:     constants.AuditStatusNotImplemented,
				Reason:     reason,
				RecordedAt: now,
			})
		}
	}
	return out
}

func filterFeatures(in []Feature, onlyFeature, onlyRole string) []Feature {
	if onlyFeature == "" && onlyRole == "" {
		return in
	}
	var out []Feature
	for _, f := range in {
		if onlyFeature != "" && f.ID != onlyFeature {
			continue
		}
		if onlyRole != "" && string(f.Role) != onlyRole {
			continue
		}
		out = append(out, f)
	}
	return out
}

func filterViewport(in []Viewport, want string) []Viewport {
	for _, v := range in {
		if v.Name == want {
			return []Viewport{v}
		}
	}
	return in
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func baseAddr(url string) string {
	url = strings.TrimPrefix(url, "http://")
	url = strings.TrimPrefix(url, "https://")
	if i := strings.Index(url, "/"); i >= 0 {
		url = url[:i]
	}
	return url
}

func contains(list []string, want string) bool {
	for _, v := range list {
		if v == want {
			return true
		}
	}
	return false
}
