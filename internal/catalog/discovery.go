// SPDX-License-Identifier: MIT

package catalog

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/ropeixoto/harnessx/internal/domain"
)

// Manifest is the YAML shape every bundled capability ships under
// templates/capabilities/<kind>/<name>/manifest.yaml. User-installed
// capabilities reuse the same shape under .harness/capabilities/<kind>/<name>.yaml.
type Manifest struct {
	Kind        string `yaml:"kind"`
	Name        string `yaml:"name"`
	Version     string `yaml:"version"`
	Description string `yaml:"description"`
	Source      string `yaml:"source,omitempty"`
	Transport   string `yaml:"transport,omitempty"`
	Tools       int    `yaml:"tools,omitempty"`
	Scope       string `yaml:"scope,omitempty"`
	Body        string `yaml:"body,omitempty"`
}

// DiscoverByGlobs scans every glob under root, parses any manifest.yaml
// files it finds, and returns the matching capabilities. Each glob is
// resolved relative to the working tree; missing dirs are silently ignored
// so a project without an .harness/capabilities/ surface still passes.
func DiscoverByGlobs(root string, kind domain.CapabilityKind, globs []string) ([]domain.Capability, error) {
	seen := map[string]struct{}{}
	var out []domain.Capability
	for _, g := range globs {
		full := filepath.Join(root, g)
		matches, err := filepath.Glob(full)
		if err != nil && !errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("catalog: glob %s: %w", g, err)
		}
		for _, m := range matches {
			cap, err := loadCapability(kind, m)
			if err != nil {
				continue
			}
			key := string(cap.Kind) + "\x00" + cap.Name
			if _, dup := seen[key]; dup {
				continue
			}
			seen[key] = struct{}{}
			out = append(out, cap)
		}
	}
	return out, nil
}

func loadCapability(kind domain.CapabilityKind, manifestPath string) (domain.Capability, error) {
	b, err := os.ReadFile(manifestPath)
	if err != nil {
		return domain.Capability{}, err
	}
	var m Manifest
	if err := yaml.Unmarshal(b, &m); err != nil {
		return domain.Capability{}, err
	}
	if m.Name == "" {
		m.Name = strings.TrimSuffix(filepath.Base(manifestPath), filepath.Ext(manifestPath))
	}
	if m.Source == "" {
		m.Source = inferSource(manifestPath)
	}
	return domain.Capability{
		Kind:         kind,
		Name:         m.Name,
		Version:      m.Version,
		Source:       m.Source,
		Status:       domain.CapDetected,
		Description:  m.Description,
		ManifestPath: manifestPath,
		Transport:    m.Transport,
		Tools:        m.Tools,
		Scope:        m.Scope,
	}, nil
}

func inferSource(p string) string {
	switch {
	case strings.Contains(p, string(os.PathSeparator)+".harness"+string(os.PathSeparator)):
		return "user"
	case strings.Contains(p, string(os.PathSeparator)+"templates"+string(os.PathSeparator)):
		return "bundled"
	default:
		return "external"
	}
}
