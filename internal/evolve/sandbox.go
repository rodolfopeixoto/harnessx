// SPDX-License-Identifier: MIT

package evolve

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

type SandboxOptions struct {
	HarnessBin    string
	TraceFile     string
	CandidateBin  string
	Args          []string
	Timeout       time.Duration
	WorkspaceRoot string
}

type SandboxResult struct {
	WorkspaceRoot string            `json:"workspace_root"`
	Baseline      ReplaySnapshot    `json:"baseline"`
	Candidate     ReplaySnapshot    `json:"candidate"`
	Improvement   ReplayImprovement `json:"improvement"`
}

type ReplaySnapshot struct {
	Source   string           `json:"source"`
	Exit     int              `json:"exit"`
	Failures int              `json:"failures"`
	Clusters []FailureCluster `json:"clusters"`
}

type ReplayImprovement struct {
	FailuresDelta int  `json:"failures_delta"`
	Improved      bool `json:"improved"`
}

func RunSandbox(ctx context.Context, opts SandboxOptions) (SandboxResult, error) {
	if opts.HarnessBin == "" {
		return SandboxResult{}, errors.New("evolve: missing baseline harness binary")
	}
	if opts.CandidateBin == "" {
		opts.CandidateBin = opts.HarnessBin
	}
	if opts.TraceFile == "" {
		return SandboxResult{}, errors.New("evolve: missing trace file")
	}
	if opts.Timeout == 0 {
		opts.Timeout = 2 * time.Minute
	}
	wsBase := opts.WorkspaceRoot
	if wsBase == "" {
		var err error
		wsBase, err = os.MkdirTemp("", "evolve-replay-*")
		if err != nil {
			return SandboxResult{}, err
		}
	}
	res := SandboxResult{WorkspaceRoot: wsBase}
	baselineWS := filepath.Join(wsBase, "baseline")
	candidateWS := filepath.Join(wsBase, "candidate")
	if err := os.MkdirAll(baselineWS, 0o755); err != nil {
		return res, err
	}
	if err := os.MkdirAll(candidateWS, 0o755); err != nil {
		return res, err
	}

	base, err := runOnce(ctx, opts.HarnessBin, opts.Args, baselineWS, opts.TraceFile, opts.Timeout)
	if err != nil {
		return res, fmt.Errorf("baseline: %w", err)
	}
	res.Baseline = base

	cand, err := runOnce(ctx, opts.CandidateBin, opts.Args, candidateWS, opts.TraceFile, opts.Timeout)
	if err != nil {
		return res, fmt.Errorf("candidate: %w", err)
	}
	res.Candidate = cand

	res.Improvement = ReplayImprovement{
		FailuresDelta: cand.Failures - base.Failures,
		Improved:      cand.Failures < base.Failures,
	}
	return res, nil
}

func runOnce(ctx context.Context, bin string, args []string, workspace, traceFile string, timeout time.Duration) (ReplaySnapshot, error) {
	tracePath := filepath.Join(workspace, ".harness", "logs", "events.jsonl")
	if err := os.MkdirAll(filepath.Dir(tracePath), 0o755); err != nil {
		return ReplaySnapshot{}, err
	}
	body, err := os.ReadFile(traceFile)
	if err != nil {
		return ReplaySnapshot{}, err
	}
	if err := os.WriteFile(tracePath, body, 0o644); err != nil {
		return ReplaySnapshot{}, err
	}
	cctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	full := append([]string{"evolve", "diagnose", "--json"}, args...)
	cmd := exec.CommandContext(cctx, bin, full...)
	cmd.Dir = workspace
	cmd.Env = append(os.Environ(), "HARNESS_PLAIN=1")
	out, runErr := cmd.Output()
	snap := ReplaySnapshot{Source: traceFile}
	if runErr != nil {
		if ee, ok := runErr.(*exec.ExitError); ok {
			snap.Exit = ee.ExitCode()
		} else {
			snap.Exit = -1
		}
	}
	var diag Diagnosis
	if err := json.Unmarshal(out, &diag); err == nil {
		snap.Failures = diag.Failures
		snap.Clusters = diag.Clusters
	}
	return snap, nil
}
