// SPDX-License-Identifier: MIT

package activeagent

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const relPath = ".harness/config/active.yaml"

type Pin struct {
	AgentID string `yaml:"agent_id"`
	Model   string `yaml:"model,omitempty"`
}

func path(root string) string { return filepath.Join(root, relPath) }

func Load(root string) (Pin, error) {
	body, err := os.ReadFile(path(root))
	if err != nil {
		if os.IsNotExist(err) {
			return Pin{}, nil
		}
		return Pin{}, err
	}
	var p Pin
	if err := yaml.Unmarshal(body, &p); err != nil {
		return Pin{}, err
	}
	p.AgentID = strings.TrimSpace(p.AgentID)
	return p, nil
}

func Save(root string, p Pin) error {
	if strings.TrimSpace(p.AgentID) == "" {
		return errors.New("activeagent: agent_id required")
	}
	if err := os.MkdirAll(filepath.Dir(path(root)), 0o755); err != nil {
		return err
	}
	body, err := yaml.Marshal(p)
	if err != nil {
		return err
	}
	return os.WriteFile(path(root), body, 0o644)
}

func Clear(root string) error {
	err := os.Remove(path(root))
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func ResolveAgentID(root, override string) string {
	if strings.TrimSpace(override) != "" {
		return override
	}
	p, err := Load(root)
	if err != nil {
		return ""
	}
	return p.AgentID
}
