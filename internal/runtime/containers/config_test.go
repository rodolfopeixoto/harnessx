// SPDX-License-Identifier: MIT

package containers

import (
	"context"
	"testing"
	"time"
)

func TestLoadConfigMissingReturnsZero(t *testing.T) {
	root := t.TempDir()
	c, err := LoadConfig(root)
	if err != nil {
		t.Fatal(err)
	}
	if c.Runtime != "" {
		t.Fatalf("runtime=%q want empty", c.Runtime)
	}
}

func TestSaveConfigRoundTrip(t *testing.T) {
	root := t.TempDir()
	in := Config{Runtime: "docker", Version: "28.0.0"}
	if err := SaveConfig(root, in); err != nil {
		t.Fatal(err)
	}
	out, err := LoadConfig(root)
	if err != nil {
		t.Fatal(err)
	}
	if out.Runtime != "docker" || out.Version != "28.0.0" {
		t.Fatalf("round-trip mismatch: %+v", out)
	}
	if out.SelectedAt.IsZero() {
		t.Fatal("SelectedAt should be auto-populated")
	}
	if time.Since(out.SelectedAt) > time.Hour {
		t.Fatalf("SelectedAt too old: %v", out.SelectedAt)
	}
}

func TestResolveEnvOverride(t *testing.T) {
	t.Setenv("HARNESS_RUNTIME", "does-not-exist")
	_, source, err := Resolve(context.Background(), t.TempDir())
	if source != "env" {
		t.Fatalf("source=%q want env", source)
	}
	if err == nil {
		t.Fatal("expected error for unknown runtime id")
	}
}

func TestResolveFromConfig(t *testing.T) {
	t.Setenv("HARNESS_RUNTIME", "")
	root := t.TempDir()
	if err := SaveConfig(root, Config{Runtime: "docker"}); err != nil {
		t.Fatal(err)
	}
	_, source, _ := Resolve(context.Background(), root)
	if source != "config" && source != "auto" {
		t.Fatalf("source=%q want config or auto", source)
	}
}
