// SPDX-License-Identifier: MIT

// Package kinds bundles the per-CapabilityKind plug-ins. Each Kind is a tiny
// declarative struct (globs + install destination); the heavy lifting lives
// in the catalog package's shared Discover / Apply / HashOps helpers.
package kinds

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ropeixoto/harnessx/internal/catalog"
	"github.com/ropeixoto/harnessx/internal/domain"
	"github.com/ropeixoto/harnessx/internal/platform/constants"
)

// spec is the per-kind blueprint shared by every implementation.
type spec struct {
	kind       domain.CapabilityKind
	globs      []string
	installExt string
}

func (s spec) Kind() domain.CapabilityKind { return s.kind }

func (s spec) Discover(_ context.Context, root string) ([]domain.Capability, error) {
	caps, err := catalog.DiscoverByGlobs(root, s.kind, s.globs)
	if err != nil {
		return nil, err
	}
	// Layer installed-status from the project capability dir.
	installed := s.scanInstalled(root)
	for i := range caps {
		if _, ok := installed[caps[i].Name]; ok {
			caps[i].Status = domain.CapInstalled
			caps[i].ConfigPath = s.configPath(root, caps[i].Name)
		}
	}
	for name, path := range installed {
		if !contains(caps, name) {
			caps = append(caps, domain.Capability{
				Kind:        s.kind,
				Name:        name,
				Status:      domain.CapInstalled,
				Source:      "user",
				ConfigPath:  path,
				Description: "(user-installed; no bundled manifest)",
			})
		}
	}
	return caps, nil
}

func (s spec) Plan(_ context.Context, root, name string) ([]domain.FileOp, error) {
	if name == "" {
		return nil, fmt.Errorf("catalog/%s: empty name", s.kind)
	}
	manifestPath, body, err := s.locateManifest(root, name)
	if err != nil {
		return nil, err
	}
	dst := s.configPath(root, name)
	ops := []domain.FileOp{
		{Op: domain.FileMkdir, Path: filepath.Dir(dst)},
		{Op: domain.FileCreate, Path: dst, Body: body},
	}
	_ = manifestPath
	return ops, nil
}

func (s spec) configPath(root, name string) string {
	return filepath.Join(root, constants.HarnessDir, "capabilities", string(s.kind), name+s.installExt)
}

func (s spec) scanInstalled(root string) map[string]string {
	out := map[string]string{}
	dir := filepath.Join(root, constants.HarnessDir, "capabilities", string(s.kind))
	entries, err := os.ReadDir(dir)
	if err != nil {
		return out
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if ext := filepath.Ext(name); ext != "" {
			name = name[:len(name)-len(ext)]
		}
		out[name] = filepath.Join(dir, e.Name())
	}
	return out
}

func (s spec) locateManifest(root, name string) (string, []byte, error) {
	for _, g := range s.globs {
		full := filepath.Join(root, g)
		matches, _ := filepath.Glob(full)
		for _, m := range matches {
			base := filepath.Base(filepath.Dir(m))
			if base == name {
				b, err := os.ReadFile(m)
				return m, b, err
			}
			// Single-file manifest matches by basename.
			plain := filepath.Base(m)
			if ext := filepath.Ext(plain); ext != "" {
				plain = plain[:len(plain)-len(ext)]
			}
			if plain == name {
				b, err := os.ReadFile(m)
				return m, b, err
			}
		}
	}
	return "", nil, fmt.Errorf("catalog/%s: no manifest for %q (looked under %v)", s.kind, name, s.globs)
}

func contains(caps []domain.Capability, name string) bool {
	for _, c := range caps {
		if c.Name == name {
			return true
		}
	}
	return false
}

// All returns one Kind impl per CapabilityKind, suitable for catalog.Register.
func All() []catalog.Kind {
	return []catalog.Kind{
		spec{kind: domain.KindAgent, globs: []string{
			"templates/agents/*.yaml",
			"templates/capabilities/agent/*/manifest.yaml",
			".harness/capabilities/agent/*.yaml",
		}, installExt: ".yaml"},
		spec{kind: domain.KindMCP, globs: []string{
			"templates/capabilities/mcp/*/manifest.yaml",
			".harness/mcp/**/*.yml", ".harness/mcp/**/*.json",
		}, installExt: ".yaml"},
		spec{kind: domain.KindHook, globs: []string{
			"templates/capabilities/hook/*/manifest.yaml",
			"scripts/git-hooks/*",
		}, installExt: ".yaml"},
		spec{kind: domain.KindSensor, globs: []string{
			"templates/capabilities/sensor/*/manifest.yaml",
		}, installExt: ".yaml"},
		spec{kind: domain.KindSkill, globs: []string{
			"templates/capabilities/skill/*/manifest.yaml",
			".harness/skills/**/*.yaml",
		}, installExt: ".yaml"},
		spec{kind: domain.KindContext, globs: []string{
			"templates/capabilities/context/*/manifest.yaml",
		}, installExt: ".yaml"},
		spec{kind: domain.KindResource, globs: []string{
			"templates/capabilities/resource/*/manifest.yaml",
		}, installExt: ".yaml"},
		spec{kind: domain.KindPlugin, globs: []string{
			"templates/capabilities/plugin/*/manifest.yaml",
		}, installExt: ".yaml"},
	}
}
