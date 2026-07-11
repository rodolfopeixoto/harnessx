// SPDX-License-Identifier: MIT

package agentcmd

import (
	"embed"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ropeixoto/harnessx/internal/agents"
	"github.com/ropeixoto/harnessx/internal/agents/fake"
	httpadapter "github.com/ropeixoto/harnessx/internal/agents/http"
	"github.com/ropeixoto/harnessx/internal/agents/interactive"
	yamladapter "github.com/ropeixoto/harnessx/internal/agents/yaml"
	"github.com/ropeixoto/harnessx/internal/platform/paths"
)

//go:embed bundled/*.yaml
var bundledFS embed.FS

// AvailableAdapterIDs enumerates every registrable adapter id (bundled +
// project overrides) sourced from the caller's cwd.
func AvailableAdapterIDs() ([]string, error) {
	reg, _, err := LoadAll(".")
	if err != nil {
		return nil, err
	}
	return reg.IDs(), nil
}

// LoadAll resolves adapters from (1) project .harness/config/agents/*.yaml
// and (2) bundled templates. Project entries override bundled entries by ID.
// Returns a registry plus the per-id source path for `agent list`.
func LoadAll(root string) (*agents.Registry, map[string]string, error) {
	reg := agents.NewRegistry()
	sources := map[string]string{}

	type discovered struct {
		spec   yamladapter.Spec
		source string
	}
	picks := map[string]discovered{}

	// Bundled first.
	bundled, err := bundledFS.ReadDir("bundled")
	if err == nil {
		for _, e := range bundled {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".yaml") {
				continue
			}
			b, err := bundledFS.ReadFile("bundled/" + e.Name())
			if err != nil {
				continue
			}
			s, err := parseSpec(b)
			if err != nil {
				return nil, nil, fmt.Errorf("bundled %s: %w", e.Name(), err)
			}
			picks[s.ID] = discovered{spec: s, source: "bundled:" + e.Name()}
		}
	}

	// Project YAMLs override.
	projDir := filepath.Join(root, ".harness", "config", "agents")
	if entries, err := os.ReadDir(projDir); err == nil {
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".yaml") {
				continue
			}
			path := filepath.Join(projDir, e.Name())
			s, err := yamladapter.Load(path)
			if err != nil {
				return nil, nil, err
			}
			picks[s.ID] = discovered{spec: s, source: path}
		}
	}

	ids := make([]string, 0, len(picks))
	for id := range picks {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		d := picks[id]
		ad := buildAdapter(d.spec)
		if err := reg.Register(ad); err != nil {
			return nil, nil, err
		}
		sources[id] = d.source
	}
	return reg, sources, nil
}

func buildAdapter(s yamladapter.Spec) agents.AgentAdapter {
	switch s.Type {
	case "api":
		return httpadapter.New(s)
	case "interactive":
		return interactive.New(s)
	default:
		return yamladapter.New(s)
	}
}

func parseSpec(b []byte) (yamladapter.Spec, error) {
	tmp, err := os.CreateTemp("", "harnessx-spec-*.yaml")
	if err != nil {
		return yamladapter.Spec{}, err
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(b); err != nil {
		_ = tmp.Close()
		return yamladapter.Spec{}, err
	}
	_ = tmp.Close()
	return yamladapter.Load(tmp.Name())
}

// Add copies a bundled adapter template into the project overrides dir.
func Add(out io.Writer, startDir, id string) error {
	if id == "" {
		return errors.New("agent add: missing id")
	}
	root, err := paths.FindProjectRoot(startDir)
	if err != nil {
		return err
	}
	dst := filepath.Join(root, ".harness", "config", "agents")
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return err
	}
	dstFile := filepath.Join(dst, id+".yaml")
	if _, err := os.Stat(dstFile); err == nil {
		return fmt.Errorf("agent add: %s already exists", dstFile)
	}
	b, err := bundledFS.ReadFile("bundled/" + id + ".yaml")
	if err != nil {
		return fmt.Errorf("agent add: no bundled adapter for %q (available: %s)", id, strings.Join(listBundled(), ", "))
	}
	if err := os.WriteFile(dstFile, b, 0o644); err != nil {
		return err
	}
	fmt.Fprintf(out, "wrote %s\n", dstFile)
	return nil
}

func listBundled() []string {
	entries, _ := bundledFS.ReadDir("bundled")
	var out []string
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".yaml") {
			out = append(out, strings.TrimSuffix(e.Name(), ".yaml"))
		}
	}
	sort.Strings(out)
	return out
}

func loadExperimentalFlags(root string) map[string]bool {
	out := map[string]bool{}
	bundled, _ := bundledFS.ReadDir("bundled")
	for _, e := range bundled {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".yaml") {
			continue
		}
		b, err := bundledFS.ReadFile("bundled/" + e.Name())
		if err != nil {
			continue
		}
		if s, err := parseSpec(b); err == nil {
			out[s.ID] = s.Experimental
		}
	}
	projDir := filepath.Join(root, ".harness", "config", "agents")
	if entries, err := os.ReadDir(projDir); err == nil {
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".yaml") {
				continue
			}
			if s, err := yamladapter.Load(filepath.Join(projDir, e.Name())); err == nil {
				out[s.ID] = s.Experimental
			}
		}
	}
	return out
}

// Ensure fake adapter package stays referenced when no caller imports it
// (the test e2e exercises it indirectly).
var _ fs.FS = bundledFS
var _ = fake.New
