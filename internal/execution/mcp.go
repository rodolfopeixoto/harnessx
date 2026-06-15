// SPDX-License-Identifier: MIT

package execution

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/ropeixoto/harnessx/internal/mcpscan"
)

// BuildMCPConfig merges every MCP server discovered under the project
// into a single config JSON suitable for Claude / Codex / Gemini CLIs
// (they all share the {"mcpServers": {name: {...}}} shape). Returns the
// written path or empty when no MCP server is configured.
func BuildMCPConfig(projectRoot, runDir string) (string, []string, error) {
	servers, err := mcpscan.Scan(projectRoot)
	if err != nil {
		return "", nil, err
	}
	if len(servers) == 0 {
		return "", nil, nil
	}
	merged := map[string]any{"mcpServers": map[string]any{}}
	mcpServers, _ := merged["mcpServers"].(map[string]any)
	names := make([]string, 0, len(servers))
	for _, s := range servers {
		entry := map[string]any{}
		if s.Command != "" {
			entry["command"] = s.Command
		}
		if s.URL != "" {
			entry["url"] = s.URL
		}
		if s.Transport != "" {
			entry["transport"] = s.Transport
		}
		mcpServers[s.Name] = entry
		names = append(names, s.Source+"/"+s.Name)
	}
	if err := os.MkdirAll(runDir, 0o755); err != nil {
		return "", nil, err
	}
	out := filepath.Join(runDir, "mcp-config.json")
	data, err := json.MarshalIndent(merged, "", "  ")
	if err != nil {
		return "", nil, err
	}
	if err := os.WriteFile(out, data, 0o644); err != nil {
		return "", nil, err
	}
	return out, names, nil
}
