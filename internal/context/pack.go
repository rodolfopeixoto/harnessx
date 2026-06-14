// SPDX-License-Identifier: MIT

// Package context builds minimal, evidence-backed context packs for the
// router and agents. Providers run in a deterministic order (memory → git →
// ripgrep → LSP → AST → deps → tests → design) and never ship a whole repo
// to an LLM.
package context

import "time"

// Pack mirrors the spec §14 "Context Pack" structure. Every field is
// either empty or backed by a provider that ran for this build.
type Pack struct {
	Hash               string         `json:"hash"`
	GeneratedAt        time.Time      `json:"generated_at"`
	Task               string         `json:"task"`
	CurrentSpec        string         `json:"current_spec,omitempty"`
	ProjectProfile     map[string]any `json:"project_profile,omitempty"`
	RelevantFiles      []FileEntry    `json:"relevant_files,omitempty"`
	LSPSymbols         []Symbol       `json:"lsp_symbols,omitempty"`
	LSPDiagnostics     []Diagnostic   `json:"lsp_diagnostics,omitempty"`
	Definitions        []Symbol       `json:"definitions,omitempty"`
	References         []Symbol       `json:"references,omitempty"`
	RelatedTests       []string       `json:"related_tests,omitempty"`
	GitStatus          string         `json:"git_status,omitempty"`
	GitDiff            string         `json:"git_diff,omitempty"`
	DesignContext      map[string]any `json:"design_context,omitempty"`
	PerformanceContext map[string]any `json:"performance_context,omitempty"`
	SecurityContext    map[string]any `json:"security_context,omitempty"`
	AcceptanceCriteria []string       `json:"acceptance_criteria,omitempty"`
	SensorsRequired    []string       `json:"sensors_required,omitempty"`
	Memories           []Memory       `json:"memories,omitempty"`

	Stats Stats `json:"stats"`
}

type FileEntry struct {
	Path            string `json:"path"`
	Bytes           int    `json:"bytes"`
	Reason          string `json:"reason"`
	EstimatedTokens int    `json:"estimated_tokens"`
	SHA256          string `json:"sha256,omitempty"`
}

type Symbol struct {
	Name string `json:"name"`
	Path string `json:"path"`
	Line int    `json:"line,omitempty"`
}

type Diagnostic struct {
	Path     string `json:"path"`
	Line     int    `json:"line"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
}

type Memory struct {
	ID         string  `json:"id"`
	Scope      string  `json:"scope"`
	Kind       string  `json:"kind"`
	Content    string  `json:"content"`
	Confidence float64 `json:"confidence"`
}

// Stats tracks numbers the router and dashboard care about. Caching is
// keyed on Pack.Hash; CacheHit is set by the builder, never by providers.
type Stats struct {
	FilesCount       int     `json:"files_count"`
	BytesCount       int     `json:"bytes_count"`
	EstimatedTokens  int     `json:"estimated_tokens"`
	ProvidersRan     int     `json:"providers_ran"`
	ProvidersSkipped int     `json:"providers_skipped"`
	CacheHitRate     float64 `json:"cache_hit_rate"`
	ContextPackHash  string  `json:"context_pack_hash"`
	BuildDurationMs  int64   `json:"build_duration_ms"`
	CacheHit         bool    `json:"cache_hit"`
	LSPQueries       int     `json:"lsp_queries"`
	LSPCacheHits     int     `json:"lsp_cache_hits"`
}
