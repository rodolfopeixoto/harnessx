// SPDX-License-Identifier: MIT

// Package sensorcmd wires `harness sensor list|run`, `harness check`, and
// `harness ci` on top of internal/sensors.
package sensorcmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/ropeixoto/harnessx/internal/index"
	"github.com/ropeixoto/harnessx/internal/platform/config"
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

	st := openSessionState(ctx, rc)
	defer st.close()

	rcOut := filepath.Join(rc.root, ".harness", "artifacts", "sensors")
	if st.hasDB {
		rcOut = filepath.Join(rcOut, st.run.ID)
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
			st.recordResult(ctx, res)
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

	st.finish(ctx, sum.Failed > 0)

	if opts.FailOnError && sum.Failed > 0 {
		return results, errors.New("one or more sensors failed")
	}
	return results, nil
}
