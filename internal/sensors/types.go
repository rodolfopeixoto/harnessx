// SPDX-License-Identifier: MIT

// Package sensors implements deterministic gates that decide whether work
// produced by an agent (or a human) is acceptable. No sensor calls an LLM —
// inferential judgement is a Phase 6+ concern handled outside this package.
package sensors

import (
	"context"
	"time"

	"github.com/ropeixoto/harnessx/internal/index"
)

type Status string

const (
	StatusPassed  Status = "passed"
	StatusFailed  Status = "failed"
	StatusSkipped Status = "skipped"
)

type Category string

const (
	CatSpec        Category = "spec"
	CatForbidden   Category = "forbidden"
	CatSecrets     Category = "secrets"
	CatChangedFile Category = "changed_files"
	CatFormat      Category = "format"
	CatLint        Category = "lint"
	CatTypecheck   Category = "typecheck"
	CatTest        Category = "test"
	CatSecurity    Category = "security"
	CatPerf        Category = "performance"
	CatDeps        Category = "dependencies"
	CatLogs        Category = "logs"
	CatImage       Category = "image"
	CatRuntime     Category = "runtime"
	CatDocs        Category = "docs"
	CatAPI         Category = "api"
	CatDesign      Category = "design_system"
	CatOther       Category = "other"
)

// Kind separates cheap deterministic sensors (computational) from those
// that need semantic judgement (inferential). The runner always executes
// computational sensors first; inferential ones may need an LLM (Phase 6+).
type Kind string

const (
	KindComputational Kind = "computational"
	KindInferential   Kind = "inferential"
)

type RunCtx struct {
	Ctx       context.Context
	Root      string // project root
	OutputDir string // <root>/.harness/artifacts/sensors/<run_id>
}

type Sensor interface {
	ID() string
	Category() Category
	Kind() Kind
	// AppliesTo decides whether the sensor is relevant for the project.
	// A go-test sensor returns false for a Rails-only project.
	AppliesTo(p index.Profile) bool
	Run(rc RunCtx) Result
}

type Result struct {
	ID         string
	Status     Status
	Category   Category
	Kind       Kind
	Duration   time.Duration
	OutputPath string
	Detail     string
	ExitCode   int
}
