package learncmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/ropeixoto/harnessx/internal/workspace"
)

type Global struct {
	GeneratedAt   time.Time              `json:"generated_at"`
	Projects      int                    `json:"projects"`
	RunsTotal     int                    `json:"runs_total"`
	TokensTotal   int                    `json:"tokens_total"`
	CostTotal     float64                `json:"cost_total"`
	ByAdapter     map[string]int         `json:"by_adapter"`
	ByStatus      map[string]int         `json:"by_status"`
	TokensPerAdpt map[string]int         `json:"tokens_per_adapter,omitempty"`
	PerProject    []ProjectMemorySummary `json:"per_project"`
}

type ProjectMemorySummary struct {
	Slug     string  `json:"slug"`
	RootPath string  `json:"root_path"`
	Runs     int     `json:"runs"`
	Tokens   int     `json:"tokens"`
	CostUSD  float64 `json:"cost_usd"`
}

type ConsolidateOptions struct {
	OutputPath string
}

func Consolidate(ctx context.Context, out io.Writer, opts ConsolidateOptions) (Global, string, error) {
	reg, err := workspace.Open("")
	if err != nil {
		return Global{}, "", fmt.Errorf("open workspace: %w", err)
	}
	defer reg.Close()
	projects, err := reg.List(ctx, false)
	if err != nil {
		return Global{}, "", err
	}
	global := Global{
		GeneratedAt:   time.Now().UTC(),
		ByAdapter:     map[string]int{},
		ByStatus:      map[string]int{},
		TokensPerAdpt: map[string]int{},
	}
	for _, p := range projects {
		inc, err := LoadIncremental(p.RootPath)
		if err != nil {
			fmt.Fprintf(out, "  skip %s: %v\n", p.Slug, err)
			continue
		}
		if inc.RunsSeen == 0 {
			continue
		}
		global.Projects++
		global.RunsTotal += inc.RunsSeen
		global.TokensTotal += inc.TokensTotal
		global.CostTotal += inc.CostTotal
		for adapter, count := range inc.ByAdapter {
			global.ByAdapter[adapter] += count
		}
		for status, count := range inc.ByStatus {
			global.ByStatus[status] += count
		}
		for adapter, tokens := range inc.TokensPerAdpt {
			global.TokensPerAdpt[adapter] += tokens
		}
		global.PerProject = append(global.PerProject, ProjectMemorySummary{
			Slug:     p.Slug,
			RootPath: p.RootPath,
			Runs:     inc.RunsSeen,
			Tokens:   inc.TokensTotal,
			CostUSD:  inc.CostTotal,
		})
	}
	path := opts.OutputPath
	if path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return global, "", err
		}
		path = filepath.Join(home, ".config", "harness", "memory.json")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return global, "", err
	}
	body, err := json.MarshalIndent(global, "", "  ")
	if err != nil {
		return global, "", err
	}
	if err := os.WriteFile(path, body, 0o644); err != nil {
		return global, "", err
	}
	render(out, Result{
		RunsAnalyzed: global.RunsTotal,
		TokensTotal:  global.TokensTotal,
		CostTotal:    global.CostTotal,
	})
	fmt.Fprintf(out, "consolidated %d project(s) → %s\n", global.Projects, path)
	return global, path, nil
}
