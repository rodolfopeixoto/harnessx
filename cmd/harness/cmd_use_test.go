// SPDX-License-Identifier: MIT

package main

import (
	"path/filepath"
	"strings"
	"testing"
)

// TestResolveTierModel_BUG_FALTA_MOD_2 covers the new `harness use --tier`
// behaviour. The bundled claude adapter ships with cheap/default/deep model
// tiers — selecting `cheap` must resolve to the haiku id documented in
// templates/agents/claude.yaml.
func TestResolveTierModel_BUG_FALTA_MOD_2(t *testing.T) {
	root, err := filepath.Abs("../../")
	if err != nil {
		t.Fatal(err)
	}

	model, err := resolveTierModel(root, "claude", "cheap")
	if err != nil {
		t.Fatalf("resolveTierModel: %v", err)
	}
	if !strings.Contains(model, "haiku") {
		t.Fatalf("cheap tier should resolve to haiku model, got %q", model)
	}

	if _, err := resolveTierModel(root, "claude", "no-such-tier"); err == nil {
		t.Fatal("expected error for unknown tier")
	}
}
