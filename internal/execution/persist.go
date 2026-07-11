// SPDX-License-Identifier: MIT

package execution

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ropeixoto/harnessx/internal/hookscan"
	"github.com/ropeixoto/harnessx/internal/mcpscan"
)

// ErrRunIncomplete signals a run directory lacks meta.json — either a
// crash mid-execution or an artifact from an older harness version. The
// caller may fall back to degradedRun for partial telemetry.
var ErrRunIncomplete = errors.New("execution: run incomplete (meta.json missing)")

func writeMeta(runDir string, res Result) {
	path := filepath.Join(runDir, "meta.json")
	data, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return
	}
	_ = os.WriteFile(path, data, 0o644)
}

func ListRuns(projectRoot string) ([]Result, error) {
	dir := filepath.Join(projectRoot, ".harness", "runs")
	entries, err := os.ReadDir(dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	out := make([]Result, 0, len(entries))
	for i := len(entries) - 1; i >= 0; i-- {
		e := entries[i]
		if !e.IsDir() {
			continue
		}
		r, err := LoadRun(projectRoot, e.Name())
		if err != nil {
			if errors.Is(err, ErrRunIncomplete) {
				out = append(out, degradedRun(projectRoot, e.Name()))
			}
			continue
		}
		out = append(out, r)
	}
	return out, nil
}

func degradedRun(projectRoot, runID string) Result {
	r := Result{RunID: runID, Status: StatusIncomplete}
	runDir := filepath.Join(projectRoot, ".harness", "runs", runID)
	if st, err := os.Stat(runDir); err == nil {
		r.StartedAt = st.ModTime()
	}
	reportPath := filepath.Join(runDir, "report.md")
	if _, err := os.Stat(reportPath); err == nil {
		r.ReportPath = reportPath
	}
	return r
}

func LoadRun(projectRoot, runID string) (Result, error) {
	runDir := filepath.Join(projectRoot, ".harness", "runs", runID)
	if _, err := os.Stat(runDir); err != nil {
		return Result{}, err
	}
	path := filepath.Join(runDir, "meta.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Result{}, fmt.Errorf("%w: %s", ErrRunIncomplete, runID)
		}
		return Result{}, err
	}
	var r Result
	if err := json.Unmarshal(data, &r); err != nil {
		return Result{}, err
	}
	return r, nil
}

func detectDeferrals(root string) ([]string, []string) {
	var mcps, hooks []string
	if list, err := mcpscan.Scan(root); err == nil {
		for _, m := range list {
			mcps = append(mcps, m.Source+"/"+m.Name)
		}
	}
	if list, err := hookscan.Scan(root); err == nil {
		for _, h := range list {
			hooks = append(hooks, h.Source+"/"+h.Name)
		}
	}
	return mcps, hooks
}

func writeReport(runDir string, req Request, res Result, summary string) string {
	path := filepath.Join(runDir, "report.md")
	var b []byte
	b = append(b, []byte(fmt.Sprintf("# Run Report — %s\n\n", res.RunID))...)
	b = append(b, []byte(fmt.Sprintf("## Summary\n\n%s\n\n", summary))...)
	b = append(b, []byte(fmt.Sprintf("## Prompt\n\n```\n%s\n```\n\n", req.Prompt))...)
	b = append(b, []byte(fmt.Sprintf("## Mode\n\n%s\n\n", req.Mode))...)
	b = append(b, []byte(fmt.Sprintf("## Agent\n\n%s\n\n", res.AgentID))...)
	b = append(b, []byte(fmt.Sprintf("## Worktree\n\n%s\n\n", res.WorktreePath))...)
	if len(res.ChangedFiles) > 0 {
		b = append(b, []byte("## Changed Files\n\n")...)
		for _, c := range res.ChangedFiles {
			b = append(b, []byte("- "+c+"\n")...)
		}
		b = append(b, '\n')
	}
	if res.DiffPath != "" {
		b = append(b, []byte("## Diff\n\n`"+res.DiffPath+"`\n\n")...)
	}
	if len(res.Sensors) > 0 {
		b = append(b, []byte("## Sensors\n\n| ID | Status | Duration (ms) |\n|---|---|---|\n")...)
		for _, s := range res.Sensors {
			b = append(b, []byte(fmt.Sprintf("| %s | %s | %d |\n", s.ID, s.Status, s.DurationMs))...)
		}
		b = append(b, '\n')
	}
	b = append(b, []byte(fmt.Sprintf("## Cost and Tokens\n\nInput: %d | Output: %d | Estimated: $%.4f | Exact: %t\n\n",
		res.InputTokens, res.OutputTokens, res.EstimatedCostUSD, res.ExactUsageAvailable))...)
	if len(res.MCPInjected) > 0 {
		b = append(b, []byte(fmt.Sprintf("## MCP\n\nInjected %d MCP server(s) via %s\n\n", len(res.MCPInjected), res.MCPConfigPath))...)
		j, _ := json.MarshalIndent(res.MCPInjected, "", "  ")
		b = append(b, j...)
		b = append(b, '\n', '\n')
	} else if len(res.MCPDetectedNotActive) > 0 {
		b = append(b, []byte("## MCP\n\nDetected but not injected (adapter capability mcp=false)\n\n")...)
		j, _ := json.MarshalIndent(res.MCPDetectedNotActive, "", "  ")
		b = append(b, j...)
		b = append(b, '\n', '\n')
	}
	if len(res.Hooks) > 0 {
		b = append(b, []byte("## Hooks\n\n| Name | Event | Exit | Duration (ms) | Skipped |\n|---|---|---|---|---|\n")...)
		for _, h := range res.Hooks {
			b = append(b, []byte(fmt.Sprintf("| %s | %s | %d | %d | %t |\n", h.Name, h.Event, h.ExitCode, h.DurationMs, h.Skipped))...)
		}
		b = append(b, '\n')
	} else if len(res.HooksDetectedNotActive) > 0 {
		b = append(b, []byte("## Hooks\n\nDetected but no pre/post-tool-use scripts executable\n\n")...)
		j, _ := json.MarshalIndent(res.HooksDetectedNotActive, "", "  ")
		b = append(b, j...)
		b = append(b, '\n', '\n')
	}
	b = append(b, []byte(fmt.Sprintf("## Status\n\n%s\n\n", res.Status))...)
	if res.ErrorMessage != "" {
		b = append(b, []byte(fmt.Sprintf("## Error\n\n```\n%s: %s\n```\n", res.ErrorType, res.ErrorMessage))...)
	}
	_ = os.WriteFile(path, b, 0o644)
	return path
}
