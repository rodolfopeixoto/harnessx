// SPDX-License-Identifier: MIT

package router

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config is the YAML shape of .harness/config/routes.yaml.
type Config struct {
	Routes map[string]RouteConfig `yaml:"routes"`
}

// LoadConfig reads a YAML file and returns its routes map. Missing file
// returns nil + nil so callers can fall back to defaults.
func LoadConfig(path string) (map[string]RouteConfig, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("router: read %s: %w", path, err)
	}
	var c Config
	if err := yaml.Unmarshal(b, &c); err != nil {
		return nil, fmt.Errorf("router: parse %s: %w", path, err)
	}
	return c.Routes, nil
}
