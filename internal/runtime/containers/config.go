// SPDX-License-Identifier: MIT

package containers

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Runtime    string    `yaml:"runtime"`
	Version    string    `yaml:"version,omitempty"`
	SelectedAt time.Time `yaml:"selected_at,omitempty"`
}

const configRelPath = ".harness/config/runtime.yaml"

func LoadConfig(projectRoot string) (Config, error) {
	path := filepath.Join(projectRoot, configRelPath)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Config{}, nil
		}
		return Config{}, err
	}
	var c Config
	if err := yaml.Unmarshal(data, &c); err != nil {
		return Config{}, fmt.Errorf("runtime config: %w", err)
	}
	return c, nil
}

func SaveConfig(projectRoot string, c Config) error {
	dir := filepath.Join(projectRoot, ".harness", "config")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	if c.SelectedAt.IsZero() {
		c.SelectedAt = time.Now().UTC()
	}
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "runtime.yaml"), data, 0o644)
}

// Resolve returns the runtime picked for projectRoot. Precedence: env
// HARNESS_RUNTIME > config file > first available.
func Resolve(ctx context.Context, projectRoot string) (Runtime, string, error) {
	if env := os.Getenv("HARNESS_RUNTIME"); env != "" {
		rt, err := ByID(env)
		if err != nil {
			return nil, "env", err
		}
		return rt, "env", nil
	}
	cfg, err := LoadConfig(projectRoot)
	if err == nil && cfg.Runtime != "" {
		rt, err := ByID(cfg.Runtime)
		if err == nil {
			return rt, "config", nil
		}
	}
	detected := Detect(ctx)
	if len(detected) == 0 {
		return nil, "", errors.New("containers: no runtime detected on host")
	}
	return detected[0], "auto", nil
}
