// SPDX-License-Identifier: MIT

package secrets

import (
	"errors"
	"os"
	"strings"
)

type EnvBackend struct{}

func (EnvBackend) Name() string    { return "env" }
func (EnvBackend) Available() bool { return true }

func (EnvBackend) Get(name string) (string, error) {
	for _, candidate := range envCandidates(name) {
		if v := os.Getenv(candidate); v != "" {
			return v, nil
		}
	}
	return "", ErrNotFound
}

func (EnvBackend) Set(string, string) error { return errors.New("env backend is read-only") }
func (EnvBackend) Delete(string) error      { return errors.New("env backend is read-only") }
func (EnvBackend) List() ([]string, error) {
	var out []string
	for _, kv := range os.Environ() {
		if strings.HasPrefix(kv, "HARNESS_SECRET_") {
			i := strings.IndexByte(kv, '=')
			if i > 0 {
				out = append(out, strings.TrimPrefix(kv[:i], "HARNESS_SECRET_"))
			}
		}
	}
	return out, nil
}

func envCandidates(name string) []string {
	upper := strings.ToUpper(name)
	return []string{
		"HARNESS_SECRET_" + upper,
		upper,
	}
}
