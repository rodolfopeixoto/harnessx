package index

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func writeFixture(t *testing.T, root string, files map[string]string) {
	t.Helper()
	for p, content := range files {
		full := filepath.Join(root, p)
		require.NoError(t, os.MkdirAll(filepath.Dir(full), 0o755))
		require.NoError(t, os.WriteFile(full, []byte(content), 0o644))
	}
}

func TestBuild_GoProject(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, map[string]string{
		"go.mod":       "module sample\n\ngo 1.23\n\nrequire (\n\tgithub.com/stretchr/testify v1.9.0\n\tgithub.com/foo/bar v0.1.0 // indirect\n)\n",
		"main.go":      "package main\n\nfunc main() {}\n",
		"main_test.go": "package main\n\nimport \"testing\"\n\nfunc TestX(t *testing.T) {}\n",
	})

	res, err := Build(Options{Root: root})
	require.NoError(t, err)
	require.Contains(t, res.Updated, MapProfile)

	var p Profile
	require.NoError(t, ReadMap(root, MapProfile, &p))
	require.Equal(t, root, p.Root)
	stackNames := func() []string {
		out := []string{}
		for _, s := range p.Stacks {
			out = append(out, s.Name)
		}
		return out
	}()
	require.Contains(t, stackNames, "go")

	var d Dependencies
	require.NoError(t, ReadMap(root, MapDependencies, &d))
	require.Contains(t, d.Ecosystems, "go")
	require.NotEmpty(t, d.Ecosystems["go"].Runtime)
	require.NotEmpty(t, d.Ecosystems["go"].Dev) // indirect

	var c Commands
	require.NoError(t, ReadMap(root, MapCommands, &c))
	require.NotEmpty(t, c.Test)

	var tm TestMap
	require.NoError(t, ReadMap(root, MapTests, &tm))
	require.Equal(t, 1, tm.TotalFiles)
}

func TestBuild_ReactProject(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, map[string]string{
		"package.json":     `{"name":"x","scripts":{"build":"vite build","test":"vitest","lint":"eslint .","dev":"vite"},"dependencies":{"react":"18","react-dom":"18"},"devDependencies":{"vite":"5","vitest":"2"}}`,
		"vite.config.ts":   "export default {}\n",
		"src/App.test.tsx": "test('x', () => {})\n",
	})
	_, err := Build(Options{Root: root})
	require.NoError(t, err)

	var p Profile
	require.NoError(t, ReadMap(root, MapProfile, &p))
	stackNames := []string{}
	for _, s := range p.Stacks {
		stackNames = append(stackNames, s.Name)
	}
	require.Contains(t, stackNames, "react")
	require.Contains(t, stackNames, "vite")

	var c Commands
	require.NoError(t, ReadMap(root, MapCommands, &c))
	require.Len(t, c.Build, 1)
	require.Len(t, c.Test, 1)
	require.Len(t, c.Lint, 1)
	require.Len(t, c.Run, 1)

	var tm TestMap
	require.NoError(t, ReadMap(root, MapTests, &tm))
	require.Equal(t, 1, tm.TotalFiles)
}

func TestBuild_RailsRoutes(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, map[string]string{
		"Gemfile":               "source 'https://rubygems.org'\ngem 'rails', '7.1.0'\ngem 'puma'\ngroup :test do\n  gem 'rspec-rails'\nend\n",
		"config/routes.rb":      "Rails.application.routes.draw do\n  root 'home#index'\n  get '/about', to: 'pages#about'\n  resources :products\nend\n",
		"spec/models/x_spec.rb": "require 'rails_helper'\n",
	})
	_, err := Build(Options{Root: root})
	require.NoError(t, err)

	var p Profile
	require.NoError(t, ReadMap(root, MapProfile, &p))
	names := []string{}
	for _, s := range p.Stacks {
		names = append(names, s.Name)
	}
	require.Contains(t, names, "rails")

	var a APIMap
	require.NoError(t, ReadMap(root, MapAPI, &a))
	require.GreaterOrEqual(t, len(a.Routes), 3)

	var d Dependencies
	require.NoError(t, ReadMap(root, MapDependencies, &d))
	require.Contains(t, d.Ecosystems, "ruby")
	require.NotEmpty(t, d.Ecosystems["ruby"].Runtime)
	require.NotEmpty(t, d.Ecosystems["ruby"].Dev)

	var tm TestMap
	require.NoError(t, ReadMap(root, MapTests, &tm))
	require.Equal(t, 1, tm.TotalFiles)
	require.Equal(t, "rspec", tm.Suites[0].Framework)
}

func TestBuild_Incremental_SkipsUnchanged(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, map[string]string{
		"go.mod":  "module x\n\ngo 1.23\n",
		"main.go": "package main\n",
	})
	res1, err := Build(Options{Root: root})
	require.NoError(t, err)
	require.NotEmpty(t, res1.Updated)

	res2, err := Build(Options{Root: root})
	require.NoError(t, err)
	require.Empty(t, res2.Updated, "second pass with no input change should skip everything")
	require.Equal(t, len(AllMaps()), len(res2.Skipped))

	// Force always rebuilds.
	res3, err := Build(Options{Root: root, Force: true})
	require.NoError(t, err)
	require.Equal(t, len(AllMaps()), len(res3.Updated))
}

func TestBuild_AlwaysWritesPerformanceBudget(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, "go.mod"), []byte("module x\n\ngo 1.23\n"), 0o644))
	_, err := Build(Options{Root: root})
	require.NoError(t, err)

	var pb PerformanceBudget
	require.NoError(t, ReadMap(root, MapPerformance, &pb))
	require.NotEmpty(t, pb.Budgets)
}
