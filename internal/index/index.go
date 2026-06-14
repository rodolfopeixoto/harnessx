// SPDX-License-Identifier: MIT

package index

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/ropeixoto/harnessx/internal/platform/clock"
	"github.com/ropeixoto/harnessx/internal/platform/hashing"
)

// MapName enumerates the JSON map files this package writes.
type MapName string

const (
	MapProfile      MapName = "profile.json"
	MapCommands     MapName = "commands.json"
	MapDependencies MapName = "dependencies.json"
	MapArchitecture MapName = "architecture.json"
	MapTests        MapName = "test-map.json"
	MapAPI          MapName = "api-map.json"
	MapDesign       MapName = "design-system.json"
	MapPerformance  MapName = "performance-budget.json"
)

// AllMaps returns the deterministic write order.
func AllMaps() []MapName {
	return []MapName{
		MapProfile, MapCommands, MapDependencies, MapArchitecture,
		MapTests, MapAPI, MapDesign, MapPerformance,
	}
}

// Options drives the index pass.
type Options struct {
	Root  string
	Force bool        // rebuild every map even if cache says inputs are unchanged
	Clock clock.Clock // injected for deterministic timestamps in tests
}

// Result reports per-map outcome.
type Result struct {
	OutputDir string
	Updated   []MapName
	Skipped   []MapName // inputs unchanged since last run
}

// Build produces every map under <root>/.harness/project/, using
// .harness/cache/index/inputs.json to skip work when inputs are unchanged.
func Build(opts Options) (Result, error) {
	if opts.Clock == nil {
		opts.Clock = clock.Real{}
	}
	root := opts.Root
	if root == "" {
		return Result{}, fmt.Errorf("index: empty root")
	}
	outDir := filepath.Join(root, ".harness", "project")
	cacheDir := filepath.Join(root, ".harness", "cache", "index")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return Result{}, err
	}
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		return Result{}, err
	}

	stacks := DetectStacks(root)
	markers := DetectMarkers(root)
	languages := DetectLanguages(stacks)
	now := opts.Clock.Now()

	cache, _ := loadCache(filepath.Join(cacheDir, "inputs.json"))
	newCache := map[string]string{}

	type job struct {
		name    MapName
		inputs  []string // files that determine the map's content
		produce func() (any, error)
	}

	jobs := []job{
		{
			name: MapProfile, inputs: markers,
			produce: func() (any, error) {
				return Profile{
					GeneratedAt: now, Root: root, Stacks: stacks,
					Languages: languages, Markers: markers,
					Confidence: aggregateConfidence(stacks),
				}, nil
			},
		},
		{
			name: MapCommands, inputs: filterExisting(root, []string{"package.json", "Makefile", "go.mod", "Gemfile", "pyproject.toml", "requirements.txt", "Cargo.toml"}),
			produce: func() (any, error) {
				c := BuildCommands(root, stacks)
				c.GeneratedAt = now
				return c, nil
			},
		},
		{
			name: MapDependencies, inputs: filterExisting(root, []string{"package.json", "package-lock.json", "pnpm-lock.yaml", "yarn.lock", "go.mod", "go.sum", "Gemfile", "Gemfile.lock", "requirements.txt", "pyproject.toml", "Cargo.toml", "Cargo.lock"}),
			produce: func() (any, error) {
				d := BuildDependencies(root, stacks)
				d.GeneratedAt = now
				return d, nil
			},
		},
		{
			name: MapArchitecture, inputs: topLevelEntriesFingerprint(root),
			produce: func() (any, error) {
				a := BuildArchitecture(root)
				a.GeneratedAt = now
				return a, nil
			},
		},
		{
			name: MapTests, inputs: topLevelEntriesFingerprint(root),
			produce: func() (any, error) {
				t := BuildTestMap(root)
				t.GeneratedAt = now
				return t, nil
			},
		},
		{
			name: MapAPI, inputs: filterExisting(root, []string{"config/routes.rb", "pages", "app", "src"}),
			produce: func() (any, error) {
				a := BuildAPIMap(root, stacks)
				a.GeneratedAt = now
				return a, nil
			},
		},
		{
			name: MapDesign, inputs: filterExisting(root, []string{"tailwind.config.js", "tailwind.config.ts", "design-tokens.json", "tokens.json", ".harness/product/design-manifest.json"}),
			produce: func() (any, error) {
				ds := BuildDesignSystem(root)
				ds.GeneratedAt = now
				return ds, nil
			},
		},
		{
			name: MapPerformance, inputs: []string{"<defaults>"},
			produce: func() (any, error) {
				p := defaultBudget()
				p.GeneratedAt = now
				return p, nil
			},
		},
	}

	res := Result{OutputDir: outDir}
	for _, j := range jobs {
		fp := fingerprint(root, j.inputs)
		newCache[string(j.name)] = fp
		outPath := filepath.Join(outDir, string(j.name))
		if !opts.Force {
			if cache[string(j.name)] == fp {
				if _, err := os.Stat(outPath); err == nil {
					res.Skipped = append(res.Skipped, j.name)
					continue
				}
			}
		}
		v, err := j.produce()
		if err != nil {
			return res, fmt.Errorf("index: build %s: %w", j.name, err)
		}
		if err := writeJSON(outPath, v); err != nil {
			return res, err
		}
		res.Updated = append(res.Updated, j.name)
	}

	if err := saveCache(filepath.Join(cacheDir, "inputs.json"), newCache); err != nil {
		return res, err
	}
	return res, nil
}

// ReadMap loads a previously-written map into v. Returns os.ErrNotExist
// when the file is missing.
func ReadMap(root string, name MapName, v any) error {
	b, err := os.ReadFile(filepath.Join(root, ".harness", "project", string(name)))
	if err != nil {
		return err
	}
	return json.Unmarshal(b, v)
}

func writeJSON(path string, v any) error {
	tmp := path + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
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

func loadCache(path string) (map[string]string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return map[string]string{}, err
	}
	m := map[string]string{}
	_ = json.Unmarshal(b, &m)
	return m, nil
}

func saveCache(path string, m map[string]string) error {
	b, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}

// fingerprint hashes the existence + size + mtime + path of each input.
// Directory inputs hash their immediate entry names. Missing inputs hash
// to a sentinel — so a file going from absent → present is detected.
func fingerprint(root string, inputs []string) string {
	var buf []byte
	for _, in := range inputs {
		buf = append(buf, []byte(in)...)
		buf = append(buf, 0)
		if in == "<defaults>" {
			buf = append(buf, []byte("static")...)
			buf = append(buf, 0)
			continue
		}
		p := filepath.Join(root, in)
		info, err := os.Stat(p)
		if err != nil {
			buf = append(buf, []byte("absent")...)
			buf = append(buf, 0)
			continue
		}
		if info.IsDir() {
			entries, _ := os.ReadDir(p)
			for _, e := range entries {
				buf = append(buf, []byte(e.Name())...)
				buf = append(buf, '|')
			}
			buf = append(buf, 0)
			continue
		}
		buf = fmt.Appendf(buf, "%d|%d", info.Size(), info.ModTime().UnixNano())
		buf = append(buf, 0)
	}
	return hashing.SHA256Bytes(buf)
}

func filterExisting(root string, in []string) []string {
	var out []string
	for _, p := range in {
		if _, err := os.Stat(filepath.Join(root, p)); err == nil {
			out = append(out, p)
		}
	}
	return out
}

func topLevelEntriesFingerprint(root string) []string {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil
	}
	var names []string
	for _, e := range entries {
		names = append(names, e.Name())
	}
	return names
}

func aggregateConfidence(stacks []Stack) Confidence {
	if len(stacks) == 0 {
		return ConfidenceLow
	}
	worst := ConfidenceHigh
	for _, s := range stacks {
		if s.Confidence == ConfidenceLow {
			return ConfidenceLow
		}
		if s.Confidence == ConfidenceMedium {
			worst = ConfidenceMedium
		}
	}
	return worst
}

// IsRecent reports whether the profile.json under root was written within
// the given window. Used by `harness project inspect` to flag staleness.
func IsRecent(root string, window time.Duration) (bool, error) {
	info, err := os.Stat(filepath.Join(root, ".harness", "project", string(MapProfile)))
	if err != nil {
		return false, err
	}
	return time.Since(info.ModTime()) <= window, nil
}
