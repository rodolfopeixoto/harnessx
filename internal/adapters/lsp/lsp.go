// SPDX-License-Identifier: MIT

// Package lsp defines the language-server client contract used by the
// context engineering layer. Phase 5 ships the interface + cache key
// layout. Real per-server implementations land in a later sub-phase.
package lsp

import (
	"context"
	"path/filepath"
)

// Client is the minimal surface needed by the context Builder. Each
// method may consult an on-disk cache to satisfy spec §15's caching
// requirement. A bool `cacheHit` is returned so the Builder can record
// cache_hit_rate metrics.
type Client interface {
	Language() string
	DocumentSymbols(ctx context.Context, root, path string) (syms []Symbol, cacheHit bool, err error)
	Diagnostics(ctx context.Context, root, path string) (diags []Diagnostic, cacheHit bool, err error)
	Definitions(ctx context.Context, root, path string, line, col int) (defs []Symbol, cacheHit bool, err error)
	References(ctx context.Context, root, path string, line, col int) (refs []Symbol, cacheHit bool, err error)
}

type Symbol struct {
	Name string
	Path string
	Line int
}

type Diagnostic struct {
	Path     string
	Line     int
	Severity string
	Message  string
}

// CacheDir returns the spec §15 cache layout:
//
//	.harness/cache/lsp/<repo-hash>/<language>/
func CacheDir(root, repoHash, language string) string {
	return filepath.Join(root, ".harness", "cache", "lsp", repoHash, language)
}

// CacheKey returns a per-query path: .../<repo-hash>/<language>/<query-hash>.json
func CacheKey(root, repoHash, language, queryHash string) string {
	return filepath.Join(CacheDir(root, repoHash, language), queryHash+".json")
}
