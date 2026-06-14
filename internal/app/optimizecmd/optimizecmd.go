// SPDX-License-Identifier: MIT

// Package optimizecmd wires the §21 resource-optimization commands:
// optimize/perf-snapshot/perf-compare/image-audit/dependency-audit/log-audit/security-audit.
package optimizecmd

import (
	stdctx "context"
	"fmt"
	"io"
	"os"

	"github.com/ropeixoto/harnessx/internal/app/sensorcmd"
	"github.com/ropeixoto/harnessx/internal/optimize"
	"github.com/ropeixoto/harnessx/internal/platform/paths"
)

func resolveRoot(start string) (string, error) {
	return paths.FindProjectRoot(start)
}

// PerfSnapshot — Cycle A. Capture a baseline.
func PerfSnapshot(opts SnapshotOptions, out io.Writer) error {
	root, err := resolveRoot(opts.StartDir)
	if err != nil {
		return err
	}
	s, path, err := optimize.Capture(optimize.SnapshotOptions{Root: root, Label: opts.Label})
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "snapshot: %s\n", path)
	fmt.Fprintf(out, "  deps=%d  noisy_logs=%d  harness_dir=%d bytes\n",
		s.Deps.Total, s.Logs.TotalCallSites, s.Disk.HarnessBytes)
	if opts.Report {
		rep, err := optimize.WriteSnapshotReport(root, s)
		if err != nil {
			return err
		}
		fmt.Fprintf(out, "report: %s\n", rep)
	}
	return nil
}

type SnapshotOptions struct {
	StartDir string
	Label    string
	Report   bool
}

// PerfCompare — diff the two most-recent snapshots (or explicit paths).
func PerfCompare(opts CompareOptions, out io.Writer) error {
	root, err := resolveRoot(opts.StartDir)
	if err != nil {
		return err
	}
	var from, to optimize.Snapshot
	var fromName, toName string
	if opts.From != "" && opts.To != "" {
		from, err = optimize.LoadSnapshot(opts.From)
		if err != nil {
			return err
		}
		to, err = optimize.LoadSnapshot(opts.To)
		if err != nil {
			return err
		}
		fromName, toName = opts.From, opts.To
	} else {
		from, to, fromName, toName, err = optimize.LatestTwo(root)
		if err != nil {
			return err
		}
	}
	d := optimize.Compare(from, to)
	fmt.Fprintf(out, "compare: %s → %s\n", fromName, toName)
	for _, r := range d.Rows {
		fmt.Fprintf(out, "  %-25s before=%v after=%v delta=%s (%s)\n",
			r.Metric, r.Before, r.After, r.Delta, r.Status)
	}
	rep, err := optimize.WriteCompareReport(root, d)
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "report: %s\n", rep)
	return nil
}

type CompareOptions struct {
	StartDir string
	From, To string
}

// ImageAudit — Cycle B static Dockerfile audit.
func ImageAudit(opts AuditOptions, out io.Writer) error {
	root, err := resolveRoot(opts.StartDir)
	if err != nil {
		return err
	}
	s, _, err := optimize.Capture(optimize.SnapshotOptions{Root: root, Label: "image-audit"})
	if err != nil {
		return err
	}
	if s.Dockerfile == nil {
		fmt.Fprintln(out, "no Dockerfile detected — image audit skipped")
		return nil
	}
	fmt.Fprintf(out, "Dockerfile: %s\n", s.Dockerfile.Path)
	fmt.Fprintf(out, "  base_image=%q stages=%d run_steps=%d copy_steps=%d\n",
		s.Dockerfile.BaseImage, s.Dockerfile.Stages, s.Dockerfile.RunSteps, s.Dockerfile.CopySteps)
	fmt.Fprintf(out, "  has_user=%v has_healthcheck=%v has_cache_cleanup=%v latest_tag=%v\n",
		s.Dockerfile.HasUSER, s.Dockerfile.HasHealthcheck, s.Dockerfile.HasCacheCleanup, s.Dockerfile.UsesLatestTag)
	if len(s.Dockerfile.Findings) == 0 {
		fmt.Fprintln(out, "findings: none")
		return nil
	}
	fmt.Fprintln(out, "findings:")
	for _, f := range s.Dockerfile.Findings {
		fmt.Fprintf(out, "  [%s] %s — %s\n", f.Severity, f.ID, f.Message)
	}
	return nil
}

type AuditOptions struct {
	StartDir string
}

// DependencyAudit — Cycle C.
func DependencyAudit(opts AuditOptions, out io.Writer) error {
	root, err := resolveRoot(opts.StartDir)
	if err != nil {
		return err
	}
	s, _, err := optimize.Capture(optimize.SnapshotOptions{Root: root, Label: "dependency-audit"})
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "dependencies: total=%d\n", s.Deps.Total)
	for eco, n := range s.Deps.ByEcosystem {
		fmt.Fprintf(out, "  %s: %d\n", eco, n)
	}
	if len(s.Deps.Candidates) == 0 {
		fmt.Fprintln(out, "removal candidates: none")
	} else {
		fmt.Fprintln(out, "removal candidates (review before removing):")
		for _, c := range s.Deps.Candidates {
			fmt.Fprintf(out, "  - [%s] %s — %s\n", c.Ecosystem, c.Name, c.Reason)
		}
	}
	if len(s.Deps.KeepReasons) > 0 {
		fmt.Fprintln(out, "kept for operational safety:")
		for _, k := range s.Deps.KeepReasons {
			fmt.Fprintf(out, "  - %s — %s\n", k.Name, k.Reason)
		}
	}
	return nil
}

// LogAudit — Cycle D.
func LogAudit(opts AuditOptions, out io.Writer) error {
	root, err := resolveRoot(opts.StartDir)
	if err != nil {
		return err
	}
	s, _, err := optimize.Capture(optimize.SnapshotOptions{Root: root, Label: "log-audit"})
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "noisy log call sites: %d (jsonl=%d bytes)\n",
		s.Logs.TotalCallSites, s.Logs.JSONLBytes)
	for _, h := range s.Logs.NoisyCallSites {
		fmt.Fprintf(out, "  %s:%d [%s] %s\n", h.Path, h.Line, h.Kind, h.Snippet)
	}
	return nil
}

// SecurityAudit — runs every Phase 4 sensor whose category == security.
func SecurityAudit(ctx stdctx.Context, opts AuditOptions, out io.Writer) error {
	_, err := sensorcmd.Run(ctx, sensorcmd.RunOptions{
		StartDir: opts.StartDir,
		IDs: []string{
			"forbidden_files", "forbidden_commands", "secrets_scan",
			"go_vuln", "ruby_brakeman", "py_bandit", "rust_audit",
		},
		FailOnError: false,
	}, out)
	return err
}

// Optimize — meta command running the full A→G cycle.
func Optimize(ctx stdctx.Context, opts AuditOptions, out io.Writer) error {
	fmt.Fprintln(out, "Cycle A — measurement")
	if err := PerfSnapshot(SnapshotOptions{StartDir: opts.StartDir, Label: "optimize-baseline", Report: true}, out); err != nil {
		return err
	}
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Cycle B — image audit")
	if err := ImageAudit(opts, out); err != nil {
		return err
	}
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Cycle C — dependency audit")
	if err := DependencyAudit(opts, out); err != nil {
		return err
	}
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Cycle D — log audit")
	if err := LogAudit(opts, out); err != nil {
		return err
	}
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Cycle F — security audit (sensor pass)")
	if err := SecurityAudit(ctx, opts, out); err != nil {
		fmt.Fprintln(out, "  (some sensors skipped — install stack tooling for full coverage)")
	}
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Cycle G — report")
	fmt.Fprintln(out, "Use `harness perf-compare` after applying changes to capture deltas.")
	// keep dummy reference to os to satisfy import lint if file shrinks later
	_ = os.Getenv
	return nil
}
