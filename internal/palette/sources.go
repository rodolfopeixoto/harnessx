// SPDX-License-Identifier: MIT

package palette

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/ropeixoto/harnessx/internal/catalog"
	"github.com/ropeixoto/harnessx/internal/domain"
	"github.com/ropeixoto/harnessx/internal/workspace"
)

const (
	SourceProjects     = "projects"
	SourceCapabilities = "capabilities"
	SourceCommands     = "commands"
)

type ProjectsSource struct {
	Registry *workspace.Registry
}

func (p ProjectsSource) Name() string { return SourceProjects }

func (p ProjectsSource) Search(ctx context.Context, q string) ([]Hit, error) {
	if p.Registry == nil {
		return nil, nil
	}
	projects, err := p.Registry.List(ctx, true)
	if err != nil {
		return nil, err
	}
	var out []Hit
	for _, proj := range projects {
		score := bestScore(q, proj.Slug, proj.DisplayName, proj.RootPath)
		if score == 0 {
			continue
		}
		out = append(out, Hit{
			Source:     SourceProjects,
			Kind:       "project",
			Title:      proj.DisplayName,
			Subtitle:   proj.RootPath,
			RouterPath: "/workspace/" + proj.Slug,
			Score:      score,
		})
	}
	return out, nil
}

type CapabilitiesSource struct {
	Catalog *catalog.Catalog
	Root    string
}

func (c CapabilitiesSource) Name() string { return SourceCapabilities }

func (c CapabilitiesSource) Search(ctx context.Context, q string) ([]Hit, error) {
	if c.Catalog == nil {
		return nil, nil
	}
	caps, err := c.Catalog.Discover(ctx, c.Root)
	if err != nil {
		return nil, err
	}
	var out []Hit
	for _, cap := range caps {
		score := bestScore(q, cap.Name, string(cap.Kind), cap.Description)
		if score == 0 {
			continue
		}
		out = append(out, Hit{
			Source:     SourceCapabilities,
			Kind:       string(cap.Kind),
			Title:      string(cap.Kind) + "/" + cap.Name,
			Subtitle:   cap.Description,
			RouterPath: "/catalog/" + string(cap.Kind) + "/" + cap.Name,
			Score:      score,
		})
	}
	_ = domain.AllCapabilityKinds()
	return out, nil
}

type CommandsSource struct {
	Commands []Command
}

type Command struct {
	Name        string
	Description string
	RouterPath  string
}

func (c CommandsSource) Name() string { return SourceCommands }

func (c CommandsSource) Search(_ context.Context, q string) ([]Hit, error) {
	var out []Hit
	for _, cmd := range c.Commands {
		score := bestScore(q, cmd.Name, cmd.Description)
		if score == 0 {
			continue
		}
		out = append(out, Hit{
			Source:     SourceCommands,
			Kind:       "command",
			Title:      cmd.Name,
			Subtitle:   cmd.Description,
			RouterPath: cmd.RouterPath,
			Score:      score,
		})
	}
	return out, nil
}

var BuiltinCommands = []Command{
	{Name: "Open workspace", Description: "Show all registered projects", RouterPath: "/workspace"},
	{Name: "Open catalog", Description: "Browse capabilities (agents, MCPs, hooks, sensors, …)", RouterPath: "/catalog"},
	{Name: "Open sessions", Description: "List recent runs", RouterPath: "/"},
	{Name: "Open settings", Description: "Project + global settings", RouterPath: "/settings"},
}

func bestScore(q string, candidates ...string) int {
	best := 0
	for _, c := range candidates {
		if c == "" {
			continue
		}
		if s := Score(q, c); s > best {
			best = s
		}
		if s := Score(q, filepath.Base(c)); s > best {
			best = s
		}
	}
	if best == 0 && strings.TrimSpace(q) == "" {
		best = 1
	}
	return best
}
