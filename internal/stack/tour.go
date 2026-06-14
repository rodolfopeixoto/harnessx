// SPDX-License-Identifier: MIT

package stack

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/ropeixoto/harnessx/internal/app/catalogcmd"
	"github.com/ropeixoto/harnessx/internal/autonomy"
	"github.com/ropeixoto/harnessx/internal/catalog"
	"github.com/ropeixoto/harnessx/internal/cleanup"
	"github.com/ropeixoto/harnessx/internal/cleanup/detectors"
	"github.com/ropeixoto/harnessx/internal/domain"
	"github.com/ropeixoto/harnessx/internal/health"
	"github.com/ropeixoto/harnessx/internal/runtime/containers"
	"github.com/ropeixoto/harnessx/internal/workspace"
)

type StepResult struct {
	Name    string
	Detail  string
	Latency time.Duration
	Err     error
}

type Tour struct {
	Root           string
	TemplatesSrc   string
	RegistryPath   string
	DashboardProbe string
	Probe          containers.HealthProbe
	Now            func() time.Time
}

func (t *Tour) clock() func() time.Time {
	if t.Now != nil {
		return t.Now
	}
	return time.Now
}

func (t *Tour) Run(ctx context.Context, out io.Writer) ([]StepResult, error) {
	results := make([]StepResult, 0, 6)
	addStep := func(name, detail string, start time.Time, err error) bool {
		latency := t.clock()().Sub(start)
		r := StepResult{Name: name, Detail: detail, Latency: latency, Err: err}
		results = append(results, r)
		printStep(out, r)
		return err == nil
	}

	start := t.clock()()
	if err := os.MkdirAll(t.Root, 0o755); err != nil {
		addStep("ensure_root", t.Root, start, err)
		return results, err
	}
	addStep("ensure_root", t.Root, start, nil)

	start = t.clock()()
	if t.TemplatesSrc != "" {
		if err := copyTree(t.TemplatesSrc, filepath.Join(t.Root, "templates")); err != nil {
			addStep("copy_templates", t.TemplatesSrc, start, err)
			return results, err
		}
	}
	addStep("copy_templates", t.TemplatesSrc, start, nil)

	start = t.clock()()
	reg, err := workspace.Open(t.RegistryPath)
	if err != nil {
		addStep("workspace_open", t.RegistryPath, start, err)
		return results, err
	}
	defer reg.Close()
	addStep("workspace_open", t.RegistryPath, start, nil)

	start = t.clock()()
	project, err := reg.Add(ctx, t.Root, "Stack Tour", "stack-tour")
	if err != nil {
		addStep("workspace_add", t.Root, start, err)
		return results, err
	}
	addStep("workspace_add", project.Slug, start, nil)

	start = t.clock()()
	cat := catalogcmd.New()
	ops, err := cat.Plan(ctx, t.Root, domain.KindMCP, "filesystem")
	if err != nil {
		addStep("catalog_plan", "mcp/filesystem", start, err)
		return results, err
	}
	if _, err := catalog.Apply(ctx, t.Root, ops); err != nil {
		addStep("catalog_install", "mcp/filesystem", start, err)
		return results, err
	}
	addStep("catalog_install", "mcp/filesystem", start, nil)

	start = t.clock()()
	scanner := cleanup.New(detectors.AbandonedHarness{}, detectors.LargeFiles{})
	findings, err := scanner.Scan(ctx, t.Root)
	if err != nil {
		addStep("cleanup_scan", "", start, err)
		return results, err
	}
	addStep("cleanup_scan", fmt.Sprintf("%d findings", len(findings)), start, nil)

	start = t.clock()()
	_, err = autonomy.Gate(autonomy.SafeExecute, autonomy.OpExecuteLowRisk)
	addStep("autonomy_gate", "safe_execute/execute_low_risk", start, err)

	start = t.clock()()
	score := health.Inputs{
		TestsPassPct: 100, SensorsPassPct: 100, DocsCoverage: 80, DesignParityPct: 70,
		RoadmapClearPct: 60, MemoryFreshDays: 7, OutdatedDeps: 1, InvalidConfigs: 0,
	}.Compute()
	addStep("health_score", fmt.Sprintf("%d/100", score.Total), start, nil)

	if t.DashboardProbe != "" {
		start = t.clock()()
		probe := t.Probe
		if probe.URL == "" {
			probe = containers.NewHealthProbe(t.DashboardProbe)
		}
		err := probe.Wait(ctx)
		addStep("dashboard_probe", t.DashboardProbe, start, err)
		if err != nil {
			return results, err
		}
	}
	return results, nil
}

func printStep(out io.Writer, r StepResult) {
	status := "ok"
	if r.Err != nil {
		status = "FAIL"
	}
	fmt.Fprintf(out, "%-4s %-22s %8d ms  %s", status, r.Name, r.Latency.Milliseconds(), r.Detail)
	if r.Err != nil {
		fmt.Fprintf(out, "  (%v)", r.Err)
	}
	fmt.Fprintln(out)
}

func copyTree(src, dst string) error {
	if src == "" {
		return nil
	}
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, path)
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		body, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		return os.WriteFile(target, body, info.Mode())
	})
}
