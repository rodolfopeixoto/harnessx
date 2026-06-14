// SPDX-License-Identifier: MIT

package kinds

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ropeixoto/harnessx/internal/domain"
)

func setupRoot(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	repo, err := os.Getwd()
	require.NoError(t, err)
	// Copy bundled templates so each kind has at least one manifest to find.
	src := filepath.Join(repo, "..", "..", "..", "templates", "capabilities")
	dst := filepath.Join(root, "templates", "capabilities")
	require.NoError(t, copyDir(src, dst))
	return root
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, path)
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		return os.WriteFile(target, b, 0o644)
	})
}

func TestAll_RegistersEveryKind(t *testing.T) {
	all := All()
	require.Len(t, all, len(domain.AllCapabilityKinds()))
	seen := map[domain.CapabilityKind]bool{}
	for _, k := range all {
		seen[k.Kind()] = true
	}
	for _, expected := range domain.AllCapabilityKinds() {
		require.True(t, seen[expected], "missing kind %s", expected)
	}
}

func TestSpec_DiscoverAndPlan(t *testing.T) {
	root := setupRoot(t)
	cases := []struct {
		kind domain.CapabilityKind
		name string
	}{
		{domain.KindMCP, "filesystem"},
		{domain.KindHook, "pre-commit-gofmt"},
		{domain.KindSensor, "go-vet"},
		{domain.KindSkill, "spec-author"},
		{domain.KindContext, "git-status"},
		{domain.KindResource, "docker-stats"},
		{domain.KindPlugin, "example"},
	}
	impls := map[domain.CapabilityKind]spec{}
	for _, k := range All() {
		if s, ok := k.(spec); ok {
			impls[k.Kind()] = s
		}
	}
	for _, c := range cases {
		t.Run(string(c.kind), func(t *testing.T) {
			s, ok := impls[c.kind]
			require.True(t, ok)
			caps, err := s.Discover(context.Background(), root)
			require.NoError(t, err)
			require.NotEmpty(t, caps)
			ops, err := s.Plan(context.Background(), root, c.name)
			require.NoError(t, err)
			require.NotEmpty(t, ops)
			// Last op must write to .harness/capabilities/<kind>/<name>.yaml
			last := ops[len(ops)-1]
			require.Equal(t, domain.FileCreate, last.Op)
			require.Contains(t, last.Path, filepath.Join(".harness", "capabilities", string(c.kind), c.name))
		})
	}
}

func TestSpec_PlanMissingName(t *testing.T) {
	root := setupRoot(t)
	for _, k := range All() {
		_, err := k.Plan(context.Background(), root, "no-such-name-12345")
		require.Error(t, err, "kind %s should reject unknown name", k.Kind())
	}
}

func TestSpec_PlanEmptyName(t *testing.T) {
	for _, k := range All() {
		_, err := k.Plan(context.Background(), t.TempDir(), "")
		require.Error(t, err)
	}
}

func TestSpec_DiscoverIncludesInstalled(t *testing.T) {
	root := setupRoot(t)
	dir := filepath.Join(root, ".harness", "capabilities", "mcp")
	require.NoError(t, os.MkdirAll(dir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "manual.yaml"), []byte("name: manual\n"), 0o644))
	all := All()
	var mcp spec
	for _, k := range all {
		if s, ok := k.(spec); ok && s.kind == domain.KindMCP {
			mcp = s
		}
	}
	caps, err := mcp.Discover(context.Background(), root)
	require.NoError(t, err)
	found := false
	for _, c := range caps {
		if c.Name == "manual" && c.Status == domain.CapInstalled {
			found = true
		}
	}
	require.True(t, found)
}
