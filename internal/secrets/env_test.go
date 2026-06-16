// SPDX-License-Identifier: MIT

package secrets

import (
	"errors"
	"testing"
)

func TestEnvBackendName(t *testing.T) {
	b := EnvBackend{}
	if b.Name() != "env" {
		t.Error("name should be env")
	}
}

func TestEnvBackendAlwaysAvailable(t *testing.T) {
	b := EnvBackend{}
	if !b.Available() {
		t.Error("env backend should always be available")
	}
}

func TestEnvBackendGetFromHarnessSecretPrefix(t *testing.T) {
	t.Setenv("HARNESS_SECRET_API_KEY", "value-xyz")
	b := EnvBackend{}
	got, err := b.Get("api_key")
	if err != nil {
		t.Fatal(err)
	}
	if got != "value-xyz" {
		t.Errorf("got %q", got)
	}
}

func TestEnvBackendGetFromBareUpper(t *testing.T) {
	t.Setenv("MY_TOKEN", "raw-value")
	b := EnvBackend{}
	got, err := b.Get("my_token")
	if err != nil {
		t.Fatal(err)
	}
	if got != "raw-value" {
		t.Errorf("got %q", got)
	}
}

func TestEnvBackendGetMissing(t *testing.T) {
	b := EnvBackend{}
	_, err := b.Get("definitely-not-set-xyz")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("want ErrNotFound, got %v", err)
	}
}

func TestEnvBackendSetIsReadOnly(t *testing.T) {
	b := EnvBackend{}
	if err := b.Set("x", "y"); err == nil {
		t.Error("expected error from read-only Set")
	}
}

func TestEnvBackendDeleteIsReadOnly(t *testing.T) {
	b := EnvBackend{}
	if err := b.Delete("x"); err == nil {
		t.Error("expected error from read-only Delete")
	}
}

func TestEnvBackendListReturnsPrefixed(t *testing.T) {
	t.Setenv("HARNESS_SECRET_FOO", "x")
	t.Setenv("HARNESS_SECRET_BAR", "y")
	b := EnvBackend{}
	got, err := b.List()
	if err != nil {
		t.Fatal(err)
	}
	found := map[string]bool{}
	for _, n := range got {
		found[n] = true
	}
	if !found["FOO"] || !found["BAR"] {
		t.Errorf("missing entries: %+v", got)
	}
}

func TestEnvCandidatesOrder(t *testing.T) {
	got := envCandidates("api_key")
	if len(got) != 2 || got[0] != "HARNESS_SECRET_API_KEY" || got[1] != "API_KEY" {
		t.Errorf("got %+v", got)
	}
}
