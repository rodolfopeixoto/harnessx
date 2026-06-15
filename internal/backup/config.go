// SPDX-License-Identifier: MIT

// Package backup wraps `rclone` to snapshot and restore harness state
// against any rclone-supported remote (drive, s3, dropbox, onedrive,
// r2, webdav, crypt, ...). Provider credentials live with rclone, not
// inside harness.
package backup

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	DefaultRemote    string   `yaml:"default_remote,omitempty"`
	Include          []string `yaml:"include,omitempty"`
	Exclude          []string `yaml:"exclude,omitempty"`
	Compression      string   `yaml:"compression,omitempty"`
	EncryptionRemote string   `yaml:"encryption_remote,omitempty"`
}

const configRelPath = ".harness/config/backup.yaml"

func DefaultConfig() Config {
	return Config{
		Include: []string{
			".harness/config",
			".harness/artifacts/specs",
			".harness/runs",
		},
		Exclude: []string{
			".harness/db",
			".harness/cache",
			".harness/worktrees",
			".harness/secrets.enc",
			".harness/secret-seed",
		},
		Compression: "gzip",
	}
}

func LoadConfig(projectRoot string) (Config, error) {
	path := filepath.Join(projectRoot, configRelPath)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultConfig(), nil
		}
		return Config{}, err
	}
	var c Config
	if err := yaml.Unmarshal(data, &c); err != nil {
		return Config{}, err
	}
	d := DefaultConfig()
	if len(c.Include) == 0 {
		c.Include = d.Include
	}
	if len(c.Exclude) == 0 {
		c.Exclude = d.Exclude
	}
	if c.Compression == "" {
		c.Compression = d.Compression
	}
	return c, nil
}

func SaveConfig(projectRoot string, c Config) error {
	dir := filepath.Join(projectRoot, ".harness", "config")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "backup.yaml"), data, 0o644)
}
