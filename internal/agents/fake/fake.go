// SPDX-License-Identifier: MIT

// Package fake implements a deterministic AgentAdapter for tests and the
// e2e harness. Behaviour is fully controlled by exported fields — no global
// state, safe for parallel use.
package fake

import (
	"context"
	"time"

	"github.com/ropeixoto/harnessx/internal/agents"
)

type Adapter struct {
	IDValue      string
	NameValue    string
	CapsValue    agents.Capabilities
	HealthOK     bool
	HealthDetail string
	CLIVersion   string
	FinalMessage string
	StdoutBytes  []byte
	StderrBytes  []byte
	ExitCode     int
	RunDelay     time.Duration
	RunErr       error
	UsageValue   agents.Usage
	ForceFailure agents.FailureType
}

func New(id string) *Adapter {
	return &Adapter{
		IDValue:      id,
		NameValue:    "Fake " + id,
		HealthOK:     true,
		CLIVersion:   "fake-1.0.0",
		FinalMessage: "ok",
		CapsValue: agents.Capabilities{
			Text: true, Files: true, Folders: true, Diff: true,
			JSONOutput: true, StreamOutput: true,
			MaxContextTokens: 128000,
			Strengths:        []string{"implementation", "tests"},
			Models:           map[string]string{"default": "fake-default"},
		},
		UsageValue: agents.Usage{
			InputTokens: 100, OutputTokens: 50, Mode: "estimated",
		},
	}
}

func (a *Adapter) ID() string                        { return a.IDValue }
func (a *Adapter) Name() string                      { return a.NameValue }
func (a *Adapter) Capabilities() agents.Capabilities { return a.CapsValue }

func (a *Adapter) Healthcheck(ctx context.Context) agents.HealthcheckResult {
	return agents.HealthcheckResult{
		OK: a.HealthOK, Detail: a.HealthDetail, CLIVersion: a.CLIVersion,
	}
}

func (a *Adapter) Run(ctx context.Context, req agents.AgentRequest) agents.AgentResult {
	start := time.Now()
	if a.RunDelay > 0 {
		select {
		case <-time.After(a.RunDelay):
		case <-ctx.Done():
			return agents.AgentResult{
				Err: ctx.Err(), ExitCode: 124,
				Duration: time.Since(start), Failure: agents.FailureTimeout,
			}
		}
	}
	out := agents.AgentOutput{
		Stdout: a.StdoutBytes, Stderr: a.StderrBytes, FinalMessage: a.FinalMessage,
	}
	failure := a.ForceFailure
	if failure == agents.FailureNone && a.RunErr != nil {
		failure = a.ClassifyFailure(out, a.ExitCode, a.RunErr)
	}
	return agents.AgentResult{
		Output: out, Usage: a.UsageValue, Err: a.RunErr,
		ExitCode: a.ExitCode, Duration: time.Since(start), Failure: failure,
	}
}

func (a *Adapter) ParseUsage(output agents.AgentOutput) agents.Usage {
	return a.UsageValue
}

func (a *Adapter) ClassifyFailure(output agents.AgentOutput, exitCode int, runErr error) agents.FailureType {
	if a.ForceFailure != agents.FailureNone {
		return a.ForceFailure
	}
	if runErr == nil && exitCode == 0 {
		return agents.FailureNone
	}
	return agents.FailurePermanent
}
