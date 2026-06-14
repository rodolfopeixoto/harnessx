// SPDX-License-Identifier: MIT

// Package optimize implements spec §21's resource-optimization skill.
// Every cycle (measure → image → deps → logs → runtime → budget → report)
// produces evidence on disk. Nothing is removed automatically — findings
// are recommendations the user accepts via their normal change workflow.
package optimize

import "time"

type Severity string

const (
	SeverityInfo Severity = "info"
	SeverityWarn Severity = "warn"
	SeverityFail Severity = "fail"
)

// Snapshot captures the project's resource posture at a point in time.
// Snapshots are append-only — `harness perf-compare` diffs two of them.
type Snapshot struct {
	ID         string             `json:"id"`
	Label      string             `json:"label,omitempty"`
	CapturedAt time.Time          `json:"captured_at"`
	Root       string             `json:"root"`
	Project    Project            `json:"project"`
	Dockerfile *DockerfileMetrics `json:"dockerfile,omitempty"`
	Deps       DepsMetrics        `json:"deps"`
	Logs       LogsMetrics        `json:"logs"`
	Disk       DiskMetrics        `json:"disk"`
	Runtime    RuntimeMetrics     `json:"runtime"`
	System     SystemInfo         `json:"system"`
}

type Project struct {
	Name   string   `json:"name"`
	Stacks []string `json:"stacks,omitempty"`
}

type DockerfileMetrics struct {
	Path            string    `json:"path"`
	BaseImage       string    `json:"base_image,omitempty"`
	Stages          int       `json:"stages"`
	RunSteps        int       `json:"run_steps"`
	CopySteps       int       `json:"copy_steps"`
	HasUSER         bool      `json:"has_user"`
	HasHealthcheck  bool      `json:"has_healthcheck"`
	HasCacheCleanup bool      `json:"has_cache_cleanup"`
	UsesLatestTag   bool      `json:"uses_latest_tag"`
	Findings        []Finding `json:"findings,omitempty"`
}

type DepsMetrics struct {
	Total       int            `json:"total"`
	ByEcosystem map[string]int `json:"by_ecosystem,omitempty"`
	Candidates  []Candidate    `json:"removal_candidates,omitempty"`
	KeepReasons []KeepReason   `json:"kept_for_operational_safety,omitempty"`
}

type Candidate struct {
	Ecosystem string `json:"ecosystem"`
	Name      string `json:"name"`
	Reason    string `json:"reason"`
}

type KeepReason struct {
	Name   string `json:"name"`
	Reason string `json:"reason"`
}

type LogsMetrics struct {
	NoisyCallSites []LogCallSite `json:"noisy_call_sites,omitempty"`
	TotalCallSites int           `json:"total_call_sites"`
	JSONLBytes     int64         `json:"jsonl_bytes"`
}

type LogCallSite struct {
	Path    string `json:"path"`
	Line    int    `json:"line"`
	Snippet string `json:"snippet"`
	Kind    string `json:"kind"` // console.log | puts | println! | fmt.Println
}

type DiskMetrics struct {
	HarnessBytes int64 `json:"harness_dir_bytes"`
	ProjectBytes int64 `json:"project_bytes"`
}

type SystemInfo struct {
	OS   string `json:"os"`
	Arch string `json:"arch"`
}

type Finding struct {
	ID       string   `json:"id"`
	Severity Severity `json:"severity"`
	Message  string   `json:"message"`
	Detail   string   `json:"detail,omitempty"`
}

// Diff is the structured comparison between two snapshots.
type Diff struct {
	From Snapshot  `json:"from"`
	To   Snapshot  `json:"to"`
	Rows []DiffRow `json:"rows"`
}

type DiffRow struct {
	Metric string `json:"metric"`
	Before any    `json:"before"`
	After  any    `json:"after"`
	Delta  string `json:"delta"`
	Status string `json:"status"` // improved | regressed | unchanged
}
