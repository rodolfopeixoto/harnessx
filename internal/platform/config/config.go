// SPDX-License-Identifier: MIT

package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config mirrors templates/.harness/config/harness.yaml. Only fields read
// by Phase 1 are typed; future phases extend this struct.
type Config struct {
	Version  int            `yaml:"version"`
	Project  ProjectConfig  `yaml:"project"`
	Database DatabaseConfig `yaml:"database"`
	Logging  LoggingConfig  `yaml:"logging"`
}

type ProjectConfig struct {
	Name string `yaml:"name"`
	Root string `yaml:"root"`
}

type DatabaseConfig struct {
	Path string `yaml:"path"`
}

type LoggingConfig struct {
	Path           string `yaml:"path"`
	RotateMaxBytes int64  `yaml:"rotate_max_bytes"`
}

// Default returns a Config populated with the defaults a fresh project
// would receive from `harness init`. root is the absolute project root.
func Default(root string) Config {
	return Config{
		Version: 1,
		Project: ProjectConfig{
			Name: filepath.Base(root),
			Root: root,
		},
		Database: DatabaseConfig{
			Path: filepath.Join(".harness", "db", "harness.sqlite"),
		},
		Logging: LoggingConfig{
			Path:           filepath.Join(".harness", "logs", "events.jsonl"),
			RotateMaxBytes: 10 * 1024 * 1024,
		},
	}
}

// Load reads the YAML config at path. Missing scalars fall back to Default.
func Load(path, root string) (Config, error) {
	def := Default(root)
	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return def, nil
		}
		return Config{}, fmt.Errorf("config: read %s: %w", path, err)
	}
	cfg := def
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return Config{}, fmt.Errorf("config: parse %s: %w", path, err)
	}
	merge(&cfg, def)
	return cfg, nil
}

func merge(c *Config, d Config) {
	if c.Version == 0 {
		c.Version = d.Version
	}
	if c.Project.Name == "" {
		c.Project.Name = d.Project.Name
	}
	if c.Project.Root == "" {
		c.Project.Root = d.Project.Root
	}
	if c.Database.Path == "" {
		c.Database.Path = d.Database.Path
	}
	if c.Logging.Path == "" {
		c.Logging.Path = d.Logging.Path
	}
	if c.Logging.RotateMaxBytes == 0 {
		c.Logging.RotateMaxBytes = d.Logging.RotateMaxBytes
	}
}

// Resolve returns an absolute version of p, joined against root if relative.
func Resolve(root, p string) string {
	if filepath.IsAbs(p) {
		return p
	}
	return filepath.Join(root, p)
}
