// SPDX-License-Identifier: MIT

// Package secrets provides a cross-platform secret store with backends
// in priority order: process env > OS keychain (macOS) / Secret Service
// (Linux) > encrypted file fallback. Resolved values are never logged.
package secrets

import (
	"errors"
	"fmt"
	"runtime"
	"sort"
	"strings"
)

type Backend interface {
	Name() string
	Available() bool
	Get(name string) (string, error)
	Set(name, value string) error
	List() ([]string, error)
	Delete(name string) error
}

type Store struct {
	backends []Backend
}

var ErrNotFound = errors.New("secrets: not found")

func New() *Store {
	return &Store{backends: defaultBackends()}
}

func NewWith(backends ...Backend) *Store {
	return &Store{backends: backends}
}

func defaultBackends() []Backend {
	switch runtime.GOOS {
	case "darwin":
		return []Backend{&EnvBackend{}, &KeychainBackend{}, &EncryptedFileBackend{}}
	case "linux":
		return []Backend{&EnvBackend{}, &SecretServiceBackend{}, &EncryptedFileBackend{}}
	default:
		return []Backend{&EnvBackend{}, &EncryptedFileBackend{}}
	}
}

func (s *Store) Get(name string) (string, error) {
	for _, b := range s.backends {
		if !b.Available() {
			continue
		}
		v, err := b.Get(name)
		if err == nil && v != "" {
			return v, nil
		}
	}
	return "", ErrNotFound
}

func (s *Store) Set(name, value string) (string, error) {
	for _, b := range s.backends {
		if !b.Available() || !writable(b) {
			continue
		}
		if err := b.Set(name, value); err != nil {
			continue
		}
		return b.Name(), nil
	}
	return "", errors.New("secrets: no writable backend available")
}

func (s *Store) List() (map[string][]string, error) {
	out := map[string][]string{}
	for _, b := range s.backends {
		if !b.Available() {
			continue
		}
		names, err := b.List()
		if err != nil {
			continue
		}
		sort.Strings(names)
		out[b.Name()] = names
	}
	return out, nil
}

func (s *Store) Delete(name string) error {
	deleted := false
	for _, b := range s.backends {
		if !b.Available() {
			continue
		}
		if err := b.Delete(name); err == nil {
			deleted = true
		}
	}
	if !deleted {
		return fmt.Errorf("secrets: delete %q: nothing removed", name)
	}
	return nil
}

func (s *Store) Backends() []Backend { return s.backends }

func writable(b Backend) bool {
	return b.Name() != "env"
}

// Resolve handles "secret://<name>" or "${{env.NAME}}" references.
func (s *Store) Resolve(ref string) (string, error) {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return "", ErrNotFound
	}
	if strings.HasPrefix(ref, "secret://") {
		return s.Get(strings.TrimPrefix(ref, "secret://"))
	}
	if strings.HasPrefix(ref, "${{env.") && strings.HasSuffix(ref, "}}") {
		envName := strings.TrimSuffix(strings.TrimPrefix(ref, "${{env."), "}}")
		eb := EnvBackend{}
		return eb.Get(envName)
	}
	return ref, nil
}
