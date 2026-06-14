// SPDX-License-Identifier: MIT

// Package contextcmd wires `harness context build|inspect` on top of
// internal/context.
package contextcmd

import (
	stdctx "context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"

	hxcontext "github.com/ropeixoto/harnessx/internal/context"
	"github.com/ropeixoto/harnessx/internal/platform/paths"
)

type BuildOptions struct {
	StartDir string
	Task     string
	Force    bool
}

func Build(ctx stdctx.Context, opts BuildOptions, out io.Writer) (*hxcontext.Pack, error) {
	if opts.Task == "" {
		return nil, errors.New("context build: empty task")
	}
	root, err := paths.FindProjectRoot(opts.StartDir)
	if err != nil {
		return nil, err
	}
	pack, err := hxcontext.Build(ctx, hxcontext.Options{
		Root: root, Task: opts.Task, Force: opts.Force,
		Providers: hxcontext.AutoLSP(root),
	})
	if err != nil {
		return nil, err
	}
	fmt.Fprintf(out, "context: %s (hash=%s)\n", labelFor(pack), pack.Hash)
	fmt.Fprintf(out, "  files=%d bytes=%d est_tokens=%d providers_ran=%d providers_skipped=%d build=%dms\n",
		pack.Stats.FilesCount, pack.Stats.BytesCount, pack.Stats.EstimatedTokens,
		pack.Stats.ProvidersRan, pack.Stats.ProvidersSkipped, pack.Stats.BuildDurationMs)
	if len(pack.RelevantFiles) > 0 {
		fmt.Fprintln(out, "  relevant files:")
		for _, f := range pack.RelevantFiles {
			fmt.Fprintf(out, "    - %s (%s)\n", f.Path, f.Reason)
		}
	}
	return pack, nil
}

func labelFor(p *hxcontext.Pack) string {
	if p.Stats.CacheHit {
		return "cache HIT"
	}
	return "built"
}

type InspectOptions struct {
	StartDir string
	Hash     string // empty = pretty-print most recent
}

func Inspect(opts InspectOptions, out io.Writer) error {
	root, err := paths.FindProjectRoot(opts.StartDir)
	if err != nil {
		return err
	}
	cacheDir := filepath.Join(root, ".harness", "cache", "context")
	if opts.Hash == "" {
		latest, err := newestCache(cacheDir)
		if err != nil {
			return err
		}
		opts.Hash = latest
	}
	if opts.Hash == "" {
		return errors.New("context inspect: no cached packs (run `harness context build \"<task>\"` first)")
	}
	path := filepath.Join(cacheDir, opts.Hash+".json")
	b, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("context inspect: %w", err)
	}
	var v any
	if err := json.Unmarshal(b, &v); err != nil {
		_, _ = out.Write(b)
		return nil
	}
	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func newestCache(dir string) (string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}
	type pick struct {
		name string
		mod  time.Time
	}
	var found []pick
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		name := e.Name()
		if filepath.Ext(name) != ".json" {
			continue
		}
		found = append(found, pick{name: name[:len(name)-5], mod: info.ModTime()})
	}
	if len(found) == 0 {
		return "", nil
	}
	sort.Slice(found, func(i, j int) bool { return found[i].mod.After(found[j].mod) })
	return found[0].name, nil
}
