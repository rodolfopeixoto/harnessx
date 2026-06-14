// SPDX-License-Identifier: MIT

package design

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/ropeixoto/harnessx/internal/platform/paths"
)

// Build is the top-level entry point for the design-to-product workflow.
// It returns paths to every product map written under .harness/product/.
type BuildOptions struct {
	Root   string
	Source string // path to ZIP or folder
}

type BuildResult struct {
	ManifestPath     string
	FeatureMapPath   string
	ToggleMapPath    string
	RoadmapPath      string
	APIContractsPath string
	FlowMapPath      string
	ImagesAnalyzed   int
	Manifest         *Manifest
}

func Build(opts BuildOptions) (*BuildResult, error) {
	if opts.Root == "" || opts.Source == "" {
		return nil, fmt.Errorf("design.Build: missing root or source")
	}
	src, err := Resolve(opts.Source)
	if err != nil {
		return nil, err
	}
	defer src.Cleanup()

	manifest, err := Inventory(src)
	if err != nil {
		return nil, err
	}

	fm := BuildFeatureMap(manifest)
	tm := PromoteToggleMap(fm)
	roadmap := BuildRoadmap(fm)
	api := BuildAPIContracts(fm)
	flowMap := BuildFlowMap(manifest)

	cache := ImageCache{Root: opts.Root}
	images, _ := cache.AnalyseAll(src)

	dir := filepath.Join(paths.HarnessDir(opts.Root), "product")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}

	res := &BuildResult{
		ImagesAnalyzed: len(images),
		Manifest:       manifest,
	}
	if res.ManifestPath, err = writeJSON(dir, "design-manifest.json", manifest); err != nil {
		return nil, err
	}
	if res.FeatureMapPath, err = writeJSON(dir, "feature-map.json", fm); err != nil {
		return nil, err
	}
	if res.ToggleMapPath, err = writeJSON(dir, "toggle-map.json", tm); err != nil {
		return nil, err
	}
	if res.RoadmapPath, err = writeJSON(dir, "roadmap.json", roadmap); err != nil {
		return nil, err
	}
	if res.APIContractsPath, err = writeJSON(dir, "api-contracts.json", api); err != nil {
		return nil, err
	}
	if res.FlowMapPath, err = writeJSON(dir, "flow-map.json", flowMap); err != nil {
		return nil, err
	}
	return res, nil
}

func writeJSON(dir, name string, v any) (string, error) {
	path := filepath.Join(dir, name)
	tmp := path + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return "", err
	}
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		_ = f.Close()
		_ = os.Remove(tmp)
		return "", err
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(tmp)
		return "", err
	}
	if err := os.Rename(tmp, path); err != nil {
		return "", err
	}
	return path, nil
}

var _ = time.Now // keep import in case future builders need it
