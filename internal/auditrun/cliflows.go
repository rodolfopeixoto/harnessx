// SPDX-License-Identifier: MIT

package auditrun

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

type CLIFlow struct {
	Name         string   `json:"name"`
	Args         []string `json:"args"`
	ExitCode     int      `json:"exit_code"`
	DurationMs   int64    `json:"duration_ms"`
	Stdout       string   `json:"stdout"`
	Stderr       string   `json:"stderr"`
	NeedsTmpRoot bool     `json:"needs_tmp_root"`
}

type CLIReport struct {
	GeneratedAt time.Time `json:"generated_at"`
	Binary      string    `json:"binary"`
	Flows       []CLIFlow `json:"flows"`
}

func DefaultCLIFlows() []CLIFlow {
	return []CLIFlow{
		{Name: "version", Args: []string{"version"}},
		{Name: "help", Args: []string{"--help"}},
		{Name: "doctor --plain", Args: []string{"doctor", "--plain"}},
		{Name: "init", Args: []string{"init"}, NeedsTmpRoot: true},
		{Name: "project list (empty)", Args: []string{"project", "list"}, NeedsTmpRoot: true},
		{Name: "catalog list", Args: []string{"catalog", "list"}, NeedsTmpRoot: true},
		{Name: "catalog show mcp filesystem", Args: []string{"catalog", "show", "mcp", "filesystem"}, NeedsTmpRoot: true},
		{Name: "cleanup scan", Args: []string{"cleanup", "scan"}, NeedsTmpRoot: true},
		{Name: "autonomy get", Args: []string{"autonomy", "get"}},
		{Name: "health show", Args: []string{"health", "show"}},
		{Name: "palette search settings", Args: []string{"palette", "search", "settings"}, NeedsTmpRoot: true},
		{Name: "stack status (offline)", Args: []string{"stack", "status", "--addr", "127.0.0.1:1"}},
	}
}

func RunCLIFlows(ctx context.Context, binary, tmpRoot string, flows []CLIFlow) CLIReport {
	report := CLIReport{GeneratedAt: time.Now().UTC(), Binary: binary, Flows: make([]CLIFlow, 0, len(flows))}
	for _, flow := range flows {
		report.Flows = append(report.Flows, runFlow(ctx, binary, tmpRoot, flow))
	}
	return report
}

func runFlow(ctx context.Context, binary, tmpRoot string, flow CLIFlow) CLIFlow {
	start := time.Now()
	cmd := exec.CommandContext(ctx, binary, flow.Args...)
	if flow.NeedsTmpRoot {
		cmd.Dir = tmpRoot
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	flow.DurationMs = time.Since(start).Milliseconds()
	flow.Stdout = truncateOutput(stdout.String())
	flow.Stderr = truncateOutput(stderr.String())
	flow.ExitCode = exitCodeOf(err, cmd)
	return flow
}

func exitCodeOf(err error, cmd *exec.Cmd) int {
	if err == nil {
		if cmd.ProcessState != nil {
			return cmd.ProcessState.ExitCode()
		}
		return 0
	}
	if ee, ok := err.(*exec.ExitError); ok {
		return ee.ExitCode()
	}
	return -1
}

func truncateOutput(s string) string {
	const limit = 4 * 1024
	if len(s) <= limit {
		return s
	}
	return s[:limit] + "\n…(truncated)"
}

func PrepareCLITmpRoot() (string, func(), error) {
	dir, err := os.MkdirTemp("", "harness-audit-cli-")
	if err != nil {
		return "", nil, err
	}
	cmd := exec.Command("git", "init", "-q")
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		_ = os.RemoveAll(dir)
		return "", nil, err
	}
	_ = os.MkdirAll(filepath.Join(dir, "templates"), 0o755)
	return dir, func() { _ = os.RemoveAll(dir) }, nil
}
