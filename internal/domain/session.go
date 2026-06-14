// SPDX-License-Identifier: MIT

package domain

import "time"

type Mode string

const (
	ModeBootstrap       Mode = "bootstrap"
	ModeQuestion        Mode = "question"
	ModeFeature         Mode = "feature"
	ModeBugfix          Mode = "bugfix"
	ModeDesignToProduct Mode = "design_to_product"
	ModeOptimization    Mode = "optimization"
	ModeAudit           Mode = "audit"
	ModeReview          Mode = "review"
	ModeSetup           Mode = "setup"
)

type Status string

const (
	StatusPending   Status = "pending"
	StatusRunning   Status = "running"
	StatusSucceeded Status = "succeeded"
	StatusFailed    Status = "failed"
	StatusCancelled Status = "cancelled"
)

type Session struct {
	ID                string
	ProjectPath       string
	Mode              Mode
	Status            Status
	StartedAt         time.Time
	FinishedAt        *time.Time
	TotalCostUSD      float64
	TotalLatencyMs    int64
	TotalInputTokens  int64
	TotalOutputTokens int64
}
