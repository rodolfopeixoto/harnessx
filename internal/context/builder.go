// SPDX-License-Identifier: MIT

package context

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	stdsort "sort"
	"time"

	"github.com/ropeixoto/harnessx/internal/index"
	"github.com/ropeixoto/harnessx/internal/platform/clock"
	"github.com/ropeixoto/harnessx/internal/platform/tokens"
)

// Options drives a build pass.
type Options struct {
	Root      string
	Task      string
	Clock     clock.Clock
	Tokens    tokens.Estimator
	Providers []Provider // nil = DefaultProviders()
	// CacheDir is .harness/cache/context by default.
	CacheDir string
	// Force ignores the cache and rebuilds.
	Force bool
}

// DefaultProviders returns the canonical provider order from spec §14.
// LSP is included only when at least one client is registered (callers
// pass an LSPProvider with their clients via Options.Providers when ready).
func DefaultProviders() []Provider {
	return []Provider{
		MemoryProvider{},
		GitProvider{},
		RipgrepProvider{},
		TestMapProvider{},
	}
}

// Build assembles a Pack for the given task. It honours the on-disk
// cache keyed on the canonical inputs: task + project profile + git HEAD
// + provider order. A cache hit returns a Pack with Stats.CacheHit=true.
func Build(ctx context.Context, opts Options) (*Pack, error) {
	if opts.Root == "" {
		return nil, fmt.Errorf("context: empty root")
	}
	if opts.Clock == nil {
		opts.Clock = clock.Real{}
	}
	if opts.Tokens == nil {
		opts.Tokens = tokens.Heuristic4{}
	}
	if opts.Providers == nil {
		opts.Providers = DefaultProviders()
	}
	if opts.CacheDir == "" {
		opts.CacheDir = filepath.Join(opts.Root, ".harness", "cache", "context")
	}
	if err := os.MkdirAll(opts.CacheDir, 0o755); err != nil {
		return nil, err
	}

	start := time.Now()
	pack := &Pack{
		Task:        opts.Task,
		GeneratedAt: opts.Clock.Now(),
	}
	if err := attachProfile(opts.Root, pack); err != nil {
		return nil, err
	}

	cacheHash := canonicalHash(opts.Task, pack.ProjectProfile, opts.Providers, opts.Root)
	cachePath := filepath.Join(opts.CacheDir, cacheHash+".json")
	if !opts.Force {
		if cached, ok := readCache(cachePath); ok {
			cached.Stats.CacheHit = true
			cached.Stats.BuildDurationMs = time.Since(start).Milliseconds()
			return cached, nil
		}
	}

	for _, p := range opts.Providers {
		if err := p.Apply(ctx, opts.Root, pack); err != nil {
			return nil, fmt.Errorf("context: provider %s: %w", p.Name(), err)
		}
	}

	enrichRelevantFiles(opts.Root, pack, opts.Tokens)
	pack.Stats.FilesCount = len(pack.RelevantFiles)
	for _, f := range pack.RelevantFiles {
		pack.Stats.BytesCount += f.Bytes
		pack.Stats.EstimatedTokens += f.EstimatedTokens
	}
	pack.Stats.EstimatedTokens += opts.Tokens.Estimate(pack.GitDiff)
	pack.Stats.EstimatedTokens += opts.Tokens.Estimate(pack.GitStatus)
	pack.Stats.BuildDurationMs = time.Since(start).Milliseconds()
	pack.Stats.ContextPackHash = cacheHash
	pack.Hash = cacheHash
	if pack.Stats.LSPQueries > 0 {
		pack.Stats.CacheHitRate = float64(pack.Stats.LSPCacheHits) / float64(pack.Stats.LSPQueries)
	}

	if err := writeCache(cachePath, pack); err != nil {
		return pack, err
	}
	return pack, nil
}

func attachProfile(root string, pack *Pack) error {
	var prof index.Profile
	if err := index.ReadMap(root, index.MapProfile, &prof); err != nil {
		// No profile yet — that's allowed, just leave it empty.
		return nil
	}
	b, err := json.Marshal(prof)
	if err != nil {
		return err
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}
	pack.ProjectProfile = m
	return nil
}

// enrichRelevantFiles fills in Bytes + EstimatedTokens + SHA for each
// file. Files larger than 256 KiB are recorded with their size but their
// bytes are not loaded — the agent should request them on demand.
func enrichRelevantFiles(root string, pack *Pack, est tokens.Estimator) {
	const maxBytes = 256 * 1024
	for i := range pack.RelevantFiles {
		f := &pack.RelevantFiles[i]
		abs := filepath.Join(root, f.Path)
		info, err := os.Stat(abs)
		if err != nil {
			continue
		}
		f.Bytes = int(info.Size())
		if info.Size() > maxBytes {
			continue
		}
		b, err := os.ReadFile(abs)
		if err != nil {
			continue
		}
		sum := sha256.Sum256(b)
		f.SHA256 = hex.EncodeToString(sum[:])
		f.EstimatedTokens = est.Estimate(string(b))
	}
}

// canonicalHash hashes the inputs that determine a Pack's content. Two
// builds with the same task + profile + provider set + git HEAD must
// produce the same hash so the cache hit is deterministic.
func canonicalHash(task string, profile map[string]any, providers []Provider, root string) string {
	names := make([]string, 0, len(providers))
	for _, p := range providers {
		names = append(names, p.Name())
	}
	stdsort.Strings(names)
	head := gitHead(root)
	payload := map[string]any{
		"task":      task,
		"providers": names,
		"profile":   profile,
		"git_head":  head,
	}
	b, _ := json.Marshal(payload)
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

func gitHead(root string) string {
	b, err := os.ReadFile(filepath.Join(root, ".git", "HEAD"))
	if err != nil {
		return ""
	}
	return string(b)
}

func readCache(path string) (*Pack, bool) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}
	var p Pack
	if err := json.Unmarshal(b, &p); err != nil {
		return nil, false
	}
	return &p, true
}

func writeCache(path string, p *Pack) error {
	tmp := path + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(p); err != nil {
		_ = f.Close()
		_ = os.Remove(tmp)
		return err
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return os.Rename(tmp, path)
}
