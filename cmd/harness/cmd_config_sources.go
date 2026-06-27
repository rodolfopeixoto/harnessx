package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func sourcesPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "harness", "sources.yaml"), nil
}

func readSourcesFile() ([]byte, error) {
	p, err := sourcesPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(p)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	return data, nil
}

func writeSourceKey(key, value string) error {
	if key == "" {
		return errors.New("config sources set: empty key")
	}
	allowed := map[string]bool{
		"update_repo":        true,
		"adapter_index_url":  true,
		"install_index_url":  true,
		"scaffold_index_url": true,
	}
	if !allowed[key] {
		return fmt.Errorf("config sources set: unknown key %q (allowed: update_repo, adapter_index_url, install_index_url, scaffold_index_url)", key)
	}
	p, err := sourcesPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	existing, _ := readSourcesFile()
	var out []string
	found := false
	for _, line := range strings.Split(string(existing), "\n") {
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, key+":") {
			out = append(out, fmt.Sprintf("%s: %s", key, value))
			found = true
			continue
		}
		out = append(out, line)
	}
	if !found {
		out = append(out, fmt.Sprintf("%s: %s", key, value))
	}
	return os.WriteFile(p, []byte(strings.Join(out, "\n")+"\n"), 0o644)
}

func removeSourceKey(key string) error {
	p, err := sourcesPath()
	if err != nil {
		return err
	}
	existing, err := readSourcesFile()
	if err != nil {
		return err
	}
	if len(existing) == 0 {
		return nil
	}
	var kept []string
	for _, line := range strings.Split(string(existing), "\n") {
		if line == "" || strings.HasPrefix(line, key+":") {
			continue
		}
		kept = append(kept, line)
	}
	if len(kept) == 0 {
		return os.Remove(p)
	}
	return os.WriteFile(p, []byte(strings.Join(kept, "\n")+"\n"), 0o644)
}

func removeSourcesFile() error {
	p, err := sourcesPath()
	if err != nil {
		return err
	}
	if err := os.Remove(p); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}
