package index

import (
	"os"
	"path/filepath"
	"testing"
)

func writeFile(t *testing.T, root, rel, body string) {
	t.Helper()
	full := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestDetectStacksJavaMaven(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "pom.xml", "<project/>")
	got := DetectStacks(root)
	if len(got) != 1 || got[0].Name != "java" {
		t.Fatalf("want [java], got %+v", got)
	}
}

func TestDetectStacksKotlinGradle(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "build.gradle.kts", "kotlin(\"jvm\")")
	got := DetectStacks(root)
	if len(got) != 1 || got[0].Name != "kotlin" {
		t.Fatalf("want [kotlin], got %+v", got)
	}
}

func TestDetectStacksJavaGradle(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "build.gradle", "apply plugin: 'java'")
	got := DetectStacks(root)
	if len(got) != 1 || got[0].Name != "java" {
		t.Fatalf("want [java], got %+v", got)
	}
}

func TestDetectStacksSwift(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "Package.swift", "// swift-tools-version:5.9")
	got := DetectStacks(root)
	if len(got) != 1 || got[0].Name != "swift" {
		t.Fatalf("want [swift], got %+v", got)
	}
}

func TestDetectStacksElixir(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "mix.exs", "defmodule X.MixProject do")
	got := DetectStacks(root)
	if len(got) != 1 || got[0].Name != "elixir" {
		t.Fatalf("want [elixir], got %+v", got)
	}
}

func TestDetectStacksLaravel(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "composer.json", `{"require":{"laravel/framework":"^11.0"}}`)
	got := DetectStacks(root)
	if len(got) != 1 || got[0].Name != "laravel" {
		t.Fatalf("want [laravel], got %+v", got)
	}
}

func TestDetectStacksSymfony(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "composer.json", `{"require":{"symfony/framework-bundle":"^7.0"}}`)
	got := DetectStacks(root)
	if len(got) != 1 || got[0].Name != "symfony" {
		t.Fatalf("want [symfony], got %+v", got)
	}
}

func TestDetectStacksPlainPHP(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "composer.json", `{"require":{"monolog/monolog":"^3"}}`)
	got := DetectStacks(root)
	if len(got) != 1 || got[0].Name != "php" {
		t.Fatalf("want [php], got %+v", got)
	}
}

func TestDetectStacksDotnet(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "App.csproj", "<Project Sdk=\"Microsoft.NET.Sdk\"></Project>")
	got := DetectStacks(root)
	if len(got) != 1 || got[0].Name != "dotnet" {
		t.Fatalf("want [dotnet], got %+v", got)
	}
}

func TestDetectStacksFlutter(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "pubspec.yaml", "name: x\nflutter:\n  uses-material-design: true\n")
	got := DetectStacks(root)
	if len(got) != 1 || got[0].Name != "flutter" {
		t.Fatalf("want [flutter], got %+v", got)
	}
}

func TestDetectStacksDart(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "pubspec.yaml", "name: x\ndependencies:\n  http: ^1\n")
	got := DetectStacks(root)
	if len(got) != 1 || got[0].Name != "dart" {
		t.Fatalf("want [dart], got %+v", got)
	}
}
