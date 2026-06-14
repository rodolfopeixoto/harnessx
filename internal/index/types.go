// SPDX-License-Identifier: MIT

// Package index produces .harness/project/*.json maps describing a project's
// stack, dependencies, architecture, tests and APIs. Phase 2.
package index

import "time"

// Confidence captures how trustworthy a detected fact is. Maps that infer
// information (api map, design system) must set this honestly.
type Confidence string

const (
	ConfidenceHigh   Confidence = "high"
	ConfidenceMedium Confidence = "medium"
	ConfidenceLow    Confidence = "low"
)

type Profile struct {
	GeneratedAt time.Time  `json:"generated_at"`
	Root        string     `json:"root"`
	Stacks      []Stack    `json:"stacks"`
	Languages   []string   `json:"languages"`
	Markers     []string   `json:"markers"`
	Confidence  Confidence `json:"confidence"`
}

type Stack struct {
	Name       string     `json:"name"`     // rails | react | nextjs | go | python | rust | docker
	Evidence   []string   `json:"evidence"` // files that led to detection
	Confidence Confidence `json:"confidence"`
}

type Commands struct {
	GeneratedAt time.Time     `json:"generated_at"`
	Build       []CommandHint `json:"build,omitempty"`
	Test        []CommandHint `json:"test,omitempty"`
	Lint        []CommandHint `json:"lint,omitempty"`
	Typecheck   []CommandHint `json:"typecheck,omitempty"`
	Format      []CommandHint `json:"format,omitempty"`
	Run         []CommandHint `json:"run,omitempty"`
}

type CommandHint struct {
	Stack      string     `json:"stack"`
	Command    string     `json:"command"`
	Source     string     `json:"source"` // package.json#scripts.test, Makefile, etc.
	Confidence Confidence `json:"confidence"`
}

type Dependencies struct {
	GeneratedAt time.Time            `json:"generated_at"`
	Ecosystems  map[string]Ecosystem `json:"ecosystems"`
}

type Ecosystem struct {
	Manifest string            `json:"manifest"`
	Runtime  []DependencyEntry `json:"runtime,omitempty"`
	Dev      []DependencyEntry `json:"dev,omitempty"`
	Count    int               `json:"count"`
}

type DependencyEntry struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
}

type Architecture struct {
	GeneratedAt time.Time      `json:"generated_at"`
	TopLevel    []ArchDirEntry `json:"top_level"`
}

type ArchDirEntry struct {
	Path    string `json:"path"`
	Purpose string `json:"purpose"` // src | tests | docs | infra | assets | config | unknown
	Files   int    `json:"files"`
}

type TestMap struct {
	GeneratedAt time.Time   `json:"generated_at"`
	Suites      []TestSuite `json:"suites"`
	TotalFiles  int         `json:"total_files"`
}

type TestSuite struct {
	Framework string   `json:"framework"` // go-test | rspec | vitest | jest | pytest | cargo-test
	Files     []string `json:"files"`
}

type APIMap struct {
	GeneratedAt time.Time    `json:"generated_at"`
	Routes      []RouteEntry `json:"routes"`
	Confidence  Confidence   `json:"confidence"`
}

type RouteEntry struct {
	Method string `json:"method,omitempty"`
	Path   string `json:"path"`
	Source string `json:"source"`
	Stack  string `json:"stack"`
}

type DesignSystem struct {
	GeneratedAt time.Time `json:"generated_at"`
	Detected    bool      `json:"detected"`
	Sources     []string  `json:"sources,omitempty"`
	Note        string    `json:"note,omitempty"`
	Colors      []string  `json:"colors,omitempty"`
	Spacing     []string  `json:"spacing,omitempty"`
	Fonts       []string  `json:"fonts,omitempty"`
	Breakpoints []string  `json:"breakpoints,omitempty"`
}

type PerformanceBudget struct {
	GeneratedAt time.Time      `json:"generated_at"`
	Budgets     map[string]any `json:"budgets"`
	Note        string         `json:"note,omitempty"`
}

func defaultBudget() PerformanceBudget {
	return PerformanceBudget{
		Budgets: map[string]any{
			"bundle_main_kb":        300,
			"bundle_vendor_kb":      500,
			"image_size_mb":         200,
			"build_time_s":          120,
			"test_time_s":           120,
			"container_memory_mb":   512,
			"container_cpu_percent": 80,
			"request_p95_ms":        500,
		},
		Note: "edit .harness/project/performance-budget.json to override",
	}
}
