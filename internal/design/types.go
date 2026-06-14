// SPDX-License-Identifier: MIT

// Package design implements the design-to-product workflow (spec §12):
// ingest a Claude Design export (ZIP or folder), inventory its pages/
// components/assets/tokens, and emit the six product maps that drive React
// parity, feature toggles, MVP roadmap, and API contracts.
package design

import "time"

type FeatureStatus string

const (
	StatusDisabled        FeatureStatus = "disabled"
	StatusStatic          FeatureStatus = "static"
	StatusMock            FeatureStatus = "mock"
	StatusAPIContract     FeatureStatus = "api_contract"
	StatusBackendReady    FeatureStatus = "backend_ready"
	StatusProductionReady FeatureStatus = "production_ready"
)

type Priority string

const (
	PriorityMVP     Priority = "mvp"
	PriorityPostMVP Priority = "post_mvp"
	PriorityBacklog Priority = "backlog"
)

// Manifest mirrors spec §12 "Design Manifest".
type Manifest struct {
	Source        string        `json:"source"`
	GeneratedAt   time.Time     `json:"generated_at"`
	Pages         []Page        `json:"pages,omitempty"`
	Components    []Component   `json:"components,omitempty"`
	Assets        []string      `json:"assets,omitempty"`
	Styles        Styles        `json:"styles"`
	DetectedFlows []string      `json:"detected_flows,omitempty"`
	MissingStates []string      `json:"missing_states,omitempty"`
	Responsive    []string      `json:"responsive_notes,omitempty"`
	ExistingDelta *ProjectDelta `json:"existing_project_delta,omitempty"`
}

type Page struct {
	ID           string   `json:"id"`
	Path         string   `json:"path"`
	Title        string   `json:"title"`
	File         string   `json:"file"`
	Components   []string `json:"components,omitempty"`
	Interactions []string `json:"interactions,omitempty"`
}

type Component struct {
	Name     string   `json:"name"`
	File     string   `json:"file"`
	Variants []string `json:"variants,omitempty"`
	States   []string `json:"states,omitempty"`
}

type Styles struct {
	Colors  []string `json:"colors,omitempty"`
	Spacing []string `json:"spacing,omitempty"`
	Fonts   []string `json:"fonts,omitempty"`
}

type ProjectDelta struct {
	ExistingRoutes   []string `json:"existing_routes,omitempty"`
	NewRoutes        []string `json:"new_routes,omitempty"`
	OverlappingPages []string `json:"overlapping_pages,omitempty"`
}

// FeatureMap is .harness/product/feature-map.json
type FeatureMap struct {
	GeneratedAt time.Time              `json:"generated_at"`
	Features    map[string]FeatureSpec `json:"features"`
}

type FeatureSpec struct {
	Status          FeatureStatus `json:"status"`
	Routes          []string      `json:"routes,omitempty"`
	Components      []string      `json:"components,omitempty"`
	BackendRequired bool          `json:"backend_required"`
	APIContract     string        `json:"api_contract,omitempty"`
	Priority        Priority      `json:"priority"`
}

// ToggleMap is the runtime-facing variant emitted to
// .harness/product/toggle-map.json.
type ToggleMap struct {
	GeneratedAt time.Time         `json:"generated_at"`
	Toggles     map[string]Toggle `json:"toggles"`
}

type Toggle struct {
	Status      FeatureStatus `json:"status"`
	Routes      []string      `json:"routes,omitempty"`
	Description string        `json:"description,omitempty"`
}

// Roadmap is .harness/product/roadmap.json
type Roadmap struct {
	GeneratedAt time.Time      `json:"generated_at"`
	Phases      []RoadmapPhase `json:"phases"`
}

type RoadmapPhase struct {
	Name     string   `json:"name"`
	Goal     string   `json:"goal"`
	Features []string `json:"features,omitempty"`
}

// APIContracts is .harness/product/api-contracts.json
type APIContracts struct {
	GeneratedAt time.Time         `json:"generated_at"`
	Endpoints   []APIEndpointSpec `json:"endpoints"`
}

type APIEndpointSpec struct {
	Method      string   `json:"method"`
	Path        string   `json:"path"`
	Feature     string   `json:"feature"`
	Status      string   `json:"status"` // proposed | drafted | ready
	Notes       []string `json:"notes,omitempty"`
	RequestBody string   `json:"request_body,omitempty"`
}

// FlowMap is .harness/product/flow-map.json
type FlowMap struct {
	GeneratedAt time.Time `json:"generated_at"`
	Flows       []Flow    `json:"flows"`
}

type Flow struct {
	ID    string   `json:"id"`
	Steps []string `json:"steps"`
	Pages []string `json:"pages,omitempty"`
}

// ImageAnalysis is the cached result of analysing one image (spec §11).
// Phase 7 ships the cache layout + metadata extraction; vision-model
// analysis is plugged in later.
type ImageAnalysis struct {
	Hash     string         `json:"image_hash"`
	Label    string         `json:"label,omitempty"`
	Format   string         `json:"format,omitempty"`
	Width    int            `json:"width,omitempty"`
	Height   int            `json:"height,omitempty"`
	Bytes    int64          `json:"bytes"`
	Source   string         `json:"source"`
	Detected map[string]any `json:"detected,omitempty"`
	CachedAt time.Time      `json:"cached_at"`
}
