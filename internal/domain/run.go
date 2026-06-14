// SPDX-License-Identifier: MIT

package domain

import "time"

type Stage string

const (
	StageInit         Stage = "init"
	StageDoctor       Stage = "doctor"
	StageSpec         Stage = "spec"
	StagePlan         Stage = "plan"
	StageContextBuild Stage = "context_build"
	StageExecution    Stage = "execution"
	StageSensors      Stage = "sensors"
	StageReport       Stage = "report"
)

type Run struct {
	ID                string
	SessionID         string
	Stage             Stage
	Agent             string // empty when no agent (e.g. init, doctor)
	Model             string
	Status            Status
	PromptHash        string
	ContextHash       string
	StartedAt         time.Time
	FinishedAt        *time.Time
	LatencyMs         int64
	InputTokens       int64
	CachedInputTokens int64
	OutputTokens      int64
	ReasoningTokens   int64
	EstimatedCostUSD  float64
	ExitCode          int
	FallbackFrom      string
	ErrorType         string
}
