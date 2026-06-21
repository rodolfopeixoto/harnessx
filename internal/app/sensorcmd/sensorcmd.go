// SPDX-License-Identifier: MIT

// Package sensorcmd wires `harness sensor list|run`, `harness check`, and
// `harness ci` on top of internal/sensors.
package sensorcmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/ropeixoto/harnessx/internal/adapters/logger"
	"github.com/ropeixoto/harnessx/internal/adapters/sqlite"
	"github.com/ropeixoto/harnessx/internal/domain"
	"github.com/ropeixoto/harnessx/internal/index"
	"github.com/ropeixoto/harnessx/internal/platform/config"
	"github.com/ropeixoto/harnessx/internal/platform/ids"
	"github.com/ropeixoto/harnessx/internal/platform/paths"
	"github.com/ropeixoto/harnessx/internal/sensors"
)

type runtimeContext struct {
	root    string
	cfg     config.Config
	dbPath  string
	logPath string
	profile index.Profile
}

func resolve(startDir string) (runtimeContext, error) {
	root, err := paths.FindProjectRoot(startDir)
	if err != nil {
		return runtimeContext{}, err
	}
	cfg, err := config.Load(filepath.Join(root, ".harness", "config", "harness.yaml"), root)
	if err != nil {
		return runtimeContext{}, err
	}
	rc := runtimeContext{
		root: root, cfg: cfg,
		dbPath:  config.Resolve(root, cfg.Database.Path),
		logPath: config.Resolve(root, cfg.Logging.Path),
	}
	if err := index.ReadMap(root, index.MapProfile, &rc.profile); err != nil {
		// Fall back to live detection so sensors still work pre-index.
		stacks := index.DetectStacks(root)
		rc.profile = index.Profile{Root: root, Stacks: stacks}
	}
	return rc, nil
}

func List(out io.Writer, startDir string) error {
	rc, err := resolve(startDir)
	if err != nil {
		return err
	}
	catalog := sensors.Catalog(rc.profile)
	if len(catalog) == 0 {
		fmt.Fprintln(out, "no sensors registered for this project")
		return nil
	}
	fmt.Fprintf(out, "%-22s %-14s %-14s\n", "ID", "CATEGORY", "KIND")
	for _, s := range catalog {
		fmt.Fprintf(out, "%-22s %-14s %-14s\n", s.ID(), s.Category(), s.Kind())
	}
	return nil
}

type RunOptions struct {
	StartDir string
	IDs      []string // empty = run all
	Quiet    bool
	// FailOnError controls whether at least one StatusFailed returns a non-zero error.
	FailOnError bool
	// Fast drops the slowest sensors so the auto-gate flow inside
	// `harness chat` stays snappy. Today: secrets_scan (ripgrep walk).
	Fast bool
	// InstallMissing pip-installs the optional python dev tools that
	// the gate reported as "binary not on PATH" before returning, so
	// the next harness ci run finds them and stops skipping sensors.
	InstallMissing bool
}

var installableBySensorID = map[string]string{
	"py_bandit":    "bandit",
	"py_mypy":      "mypy",
	"py_pip_audit": "pip-audit",
}

// slowSensorIDs is the denylist consulted when Fast is true.
var slowSensorIDs = map[string]bool{
	"secrets_scan": true,
}

func Run(ctx context.Context, opts RunOptions, out io.Writer) ([]sensors.Result, error) {
	rc, err := resolve(opts.StartDir)
	if err != nil {
		return nil, err
	}

	catalog := sensors.Catalog(rc.profile)
	selected := filterByIDs(catalog, opts.IDs)
	if len(opts.IDs) > 0 && len(selected) == 0 {
		return nil, fmt.Errorf("no matching sensors for %v", opts.IDs)
	}
	if opts.Fast {
		filtered := selected[:0]
		skipped := 0
		for _, s := range selected {
			if slowSensorIDs[s.ID()] {
				skipped++
				continue
			}
			filtered = append(filtered, s)
		}
		selected = filtered
		if !opts.Quiet && skipped > 0 {
			fmt.Fprintf(out, "  --fast: skipped %d slow sensor(s)\n", skipped)
		}
	}

	// Open DB + logger if .harness/ initialised.
	var repo *sqlite.Repo
	var lg *logger.JSONL
	var sess domain.Session
	var run domain.Run
	hasDB := false
	if _, err := os.Stat(rc.dbPath); err == nil {
		repo, err = sqlite.Open(rc.dbPath)
		if err == nil {
			hasDB = true
			defer repo.Close()
			lg, _ = logger.Open(rc.logPath, rc.cfg.Logging.RotateMaxBytes)
			if lg != nil {
				defer lg.Close()
			}
			now := time.Now().UTC()
			sess = domain.Session{
				ID: ids.New(), ProjectPath: rc.root, Mode: domain.ModeAudit,
				Status: domain.StatusRunning, StartedAt: now,
			}
			run = domain.Run{
				ID: ids.New(), SessionID: sess.ID, Stage: domain.StageSensors,
				Status: domain.StatusRunning, StartedAt: now,
			}
			_ = repo.CreateSession(ctx, sess)
			_ = repo.CreateRun(ctx, run)
		}
	}

	rcOut := filepath.Join(rc.root, ".harness", "artifacts", "sensors")
	if hasDB {
		rcOut = filepath.Join(rcOut, run.ID)
	}
	runner := &sensors.Runner{
		OnResult: func(res sensors.Result) {
			if !opts.Quiet {
				conf := ""
				if res.Confidence > 0 && res.Confidence < 1.0 {
					conf = fmt.Sprintf(" (~conf %.2f)", res.Confidence)
				}
				fmt.Fprintf(out, "  [%s] %-22s %s %s%s\n", icon(res.Status), res.ID, res.Duration.Round(time.Millisecond), detail(res), conf)
			}
			if hasDB {
				_ = repo.WriteSensorResult(ctx, run.ID, res.ID, string(res.Status), res.Duration.Milliseconds(), res.OutputPath, time.Now().UTC())
				if lg != nil {
					_ = lg.Write("info", map[string]any{
						"stage": "sensor", "session_id": sess.ID, "run_id": run.ID,
						"sensor": res.ID, "status": string(res.Status),
						"duration_ms": res.Duration.Milliseconds(),
					})
				}
			}
		},
	}
	results := runner.Run(ctx, selected, sensors.RunCtx{Root: rc.root, OutputDir: rcOut})

	sum := sensors.Summarize(results)
	fmt.Fprintf(out, "\nsummary: %d passed, %d failed, %d skipped (of %d)\n", sum.Passed, sum.Failed, sum.Skipped, sum.Total)

	if missing := missingInstallables(results); len(missing) > 0 {
		if opts.InstallMissing {
			if err := installPythonTools(ctx, rc.root, missing, out); err != nil {
				fmt.Fprintf(out, "  install-missing: %v\n", err)
			}
		} else {
			fmt.Fprintf(out, "  hint: %d optional python tool(s) missing (%s) — rerun with `harness ci --install-missing` to fix\n",
				len(missing), strings.Join(missing, ", "))
		}
	}

	if hasDB {
		end := time.Now().UTC()
		status := domain.StatusSucceeded
		if sum.Failed > 0 {
			status = domain.StatusFailed
		}
		_ = repo.FinishRun(ctx, run.ID, status, end, ifFail(sum.Failed > 0))
		_ = repo.FinishSession(ctx, sess.ID, status, end)
	}

	if opts.FailOnError && sum.Failed > 0 {
		return results, errors.New("one or more sensors failed")
	}
	return results, nil
}

func filterByIDs(in []sensors.Sensor, ids []string) []sensors.Sensor {
	if len(ids) == 0 {
		return in
	}
	want := map[string]bool{}
	for _, id := range ids {
		want[id] = true
	}
	var out []sensors.Sensor
	for _, s := range in {
		if want[s.ID()] {
			out = append(out, s)
		}
	}
	return out
}

func icon(s sensors.Status) string {
	switch s {
	case sensors.StatusPassed:
		return "✓"
	case sensors.StatusFailed:
		return "✗"
	default:
		return "·"
	}
}

func detail(res sensors.Result) string {
	if res.Detail == "" {
		return ""
	}
	return "— " + res.Detail
}

func ifFail(b bool) int {
	if b {
		return 1
	}
	return 0
}

func missingInstallables(results []sensors.Result) []string {
	seen := map[string]bool{}
	var out []string
	for _, r := range results {
		pkg, ok := installableBySensorID[r.ID]
		if !ok {
			continue
		}
		if !strings.HasPrefix(r.Detail, "binary not on PATH") {
			continue
		}
		if seen[pkg] {
			continue
		}
		seen[pkg] = true
		out = append(out, pkg)
	}
	return out
}

func installPythonTools(ctx context.Context, root string, pkgs []string, out io.Writer) error {
	venvPython := filepath.Join(root, ".venv", "bin", "python")
	if _, err := os.Stat(venvPython); err != nil {
		return fmt.Errorf(".venv missing — run `harness new <stack> --with-deps` or `uv venv .venv && uv pip install -r requirements.txt`")
	}
	fmt.Fprintf(out, "  → installing into .venv: %s\n", strings.Join(pkgs, " "))
	args := append([]string{"pip", "install", "--python", venvPython}, pkgs...)
	cmd := exec.CommandContext(ctx, "uv", args...)
	cmd.Dir = root
	cmd.Stdout = out
	cmd.Stderr = out
	if err := cmd.Run(); err == nil {
		return nil
	}
	pipArgs := append([]string{"-m", "pip", "install"}, pkgs...)
	pipCmd := exec.CommandContext(ctx, venvPython, pipArgs...)
	pipCmd.Dir = root
	pipCmd.Stdout = out
	pipCmd.Stderr = out
	if err := pipCmd.Run(); err != nil {
		return fmt.Errorf("install failed via uv and pip: %w", err)
	}
	return nil
}
