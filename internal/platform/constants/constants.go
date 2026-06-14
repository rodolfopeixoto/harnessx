// SPDX-License-Identifier: MIT

// Package constants centralises every magic number, default path, and
// well-known string used across HarnessX. Touching one constant must
// flow through this file so behaviour stays consistent and reviewable.
//
// Add only values that meet at least one of:
//   - Shared between two or more packages.
//   - Likely to be tuned by operators.
//   - Has a security/safety implication if wrong (timeout, size cap, path).
//
// Per-package private constants stay where they are used.
package constants

import "time"

// --- Filesystem layout (relative to project root) -------------------------

const (
	HarnessDir = ".harness"

	ConfigSubdir    = "config"
	DBSubdir        = "db"
	LogsSubdir      = "logs"
	CacheSubdir     = "cache"
	ArtifactsSubdir = "artifacts"
	ProductSubdir   = "product"
	ProjectSubdir   = "project"
	SkillsSubdir    = "skills"

	DBFilename     = "harness.sqlite"
	LogsFilename   = "events.jsonl"
	ConfigFilename = "harness.yaml"
	RoutesFilename = "routes.yaml"
	AgentsSubdir   = "agents"
	IgnoreFilename = ".harnessignore"
)

// --- Global (cross-project) home -----------------------------------------

const (
	GlobalHarnessDirName       = "harness"
	GlobalRegistryFilename     = "registry.sqlite"
	GlobalRegistryLockFilename = "registry.lock"
	EnvHarnessHome             = "HARNESS_HOME"
	EnvProjectOverride         = "HARNESS_PROJECT"
)

// --- Slug + identifier defaults -----------------------------------------

const (
	SlugSeparator    = "-"
	SlugFallbackName = "project"
)

// --- Cleanup engine -----------------------------------------------------

const (
	CleanupSubdir              = "cleanup"
	CleanupPolicyFilename      = "policy.yaml"
	EnvCleanupAcknowledgement  = "HARNESS_CLEANUP_I_UNDERSTAND"
	CleanupRiskLow             = "low"
	CleanupRiskMedium          = "medium"
	CleanupRiskHigh            = "high"
	CleanupLargeFileThresholdB = int64(50 * 1024 * 1024)
	CleanupStaleThresholdHours = 24 * 30
)

const (
	KindCleanupWorktree       = "worktree"
	KindCleanupCache          = "cache"
	KindCleanupAbandonedHX    = "abandoned_harness"
	KindCleanupVMLeftover     = "vm_leftover"
	KindCleanupClaudeLeftover = "claude_leftover"
	KindCleanupLargeFile      = "large_file"
	KindCleanupContainer      = "container"
)

const (
	EnvDockerBinary           = "HARNESS_DOCKER"
	DefaultDockerBinary       = "docker"
	DefaultContainerUpTimeout = 90 * time.Second
)

// --- Autonomy levels ----------------------------------------------------

const (
	AutonomyManual               = "manual"
	AutonomyPlanAndAsk           = "plan_and_ask"
	AutonomySafeExecute          = "safe_execute"
	AutonomyFullProjectLoop      = "full_project_loop"
	AutonomyScheduledMaintenance = "scheduled_maintenance"
)

const (
	AutonomyOpRead            = "read"
	AutonomyOpPlan            = "plan"
	AutonomyOpExecuteLowRisk  = "execute_low_risk"
	AutonomyOpExecuteHighRisk = "execute_high_risk"
	AutonomyOpClean           = "clean"
	AutonomyOpSchedule        = "schedule"
)

// --- Health score -------------------------------------------------------

const (
	HealthMaxScore          = 100
	HealthDefaultScore      = 50
	HealthSubsystemTests    = "tests"
	HealthSubsystemSensors  = "sensors"
	HealthSubsystemSecurity = "security"
	HealthSubsystemPerf     = "perf"
	HealthSubsystemDeps     = "deps"
	HealthSubsystemDocs     = "docs"
	HealthSubsystemParity   = "design_parity"
	HealthSubsystemRoadmap  = "roadmap_readiness"
	HealthSubsystemMemory   = "memory_freshness"
	HealthSubsystemConfigs  = "configs"
)

// --- Network --------------------------------------------------------------

const (
	DefaultDashboardAddr = "127.0.0.1:7373"
	DefaultDashboardHost = "127.0.0.1"
	DefaultDashboardPort = 7373
)

// --- Limits + timeouts ----------------------------------------------------

const (
	DefaultProbeTimeout        = 2 * time.Second
	DefaultLSPHandshakeTimeout = 15 * time.Second
	DefaultLSPDiagnosticsWait  = 2 * time.Second
	DefaultLSPShutdownTimeout  = 2 * time.Second

	DefaultAgentTimeout             = 5 * time.Minute
	DefaultSensorTimeout            = 5 * time.Minute
	DefaultDockerStatsTimeout       = 5 * time.Second
	DefaultDashboardShutdownTimeout = 3 * time.Second
	DefaultGitTimeout               = 10 * time.Second

	DefaultLogRotateBytes    int64 = 10 * 1024 * 1024
	MaxZipExtractBytes       int64 = 200 * 1024 * 1024
	MaxContextFileBytes            = 256 * 1024
	MaxSecretsScanBytes      int64 = 2 * 1024 * 1024
	MaxNoisyLogCallSites           = 200
	MaxRipgrepKeywords             = 6
	MaxRipgrepHitsPerKeyword       = 8
	MaxSpawnedJSONResponse         = 4 * 1024 * 1024 // bufio buffer cap

	MemoryConfidenceFloor      = 0.4
	TokenEstimateCharsPerToken = 4

	DefaultListLimit        = 50
	DefaultSessionListLimit = 100
	DefaultMaxMemoryEntries = 25
)

// --- Telemetry / metric names --------------------------------------------

const (
	MetricInputTokens      = "input_tokens"
	MetricOutputTokens     = "output_tokens"
	MetricEstimatedCostUSD = "estimated_cost_usd"
	MetricLatencyMs        = "latency_ms"
	MetricSensorPassRate   = "sensor_pass_rate"
)

// --- Cost defaults (USD per 1M tokens) -----------------------------------
//
// Mirrors templates/agents/*.yaml. Adapters can override per-adapter; the
// values here are the "use this when nothing else is configured" fallback.
const (
	DefaultInputTokenPricePer1M  = 1.0
	DefaultOutputTokenPricePer1M = 5.0
)

// --- Exit codes ----------------------------------------------------------

const (
	ExitOK       = 0
	ExitFailure  = 1
	ExitNotImpl  = 2 // historical; new commands should not use this
	ExitUserDeny = 3
)
