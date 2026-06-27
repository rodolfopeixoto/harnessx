package main

import (
	"sort"
	"testing"
)

func TestToolsForStackKnown(t *testing.T) {
	cases := []struct {
		stack string
		want  []string
	}{
		{"swift", []string{"swift-format", "swiftlint"}},
		{"kotlin", []string{"detekt", "ktlint"}},
		{"php", []string{"php-cs-fixer", "phpstan", "psalm"}},
		{"laravel", []string{"php-cs-fixer", "phpstan", "psalm"}},
		{"dotnet", []string{"dotnet"}},
	}
	for _, c := range cases {
		got := append([]string{}, toolsForStack(c.stack)...)
		sort.Strings(got)
		sort.Strings(c.want)
		if len(got) != len(c.want) {
			t.Fatalf("stack %s: want %v, got %v", c.stack, c.want, got)
		}
		for i := range got {
			if got[i] != c.want[i] {
				t.Errorf("stack %s: pos %d want %s got %s", c.stack, i, c.want[i], got[i])
			}
		}
	}
}

func TestToolsForStackUnknown(t *testing.T) {
	if got := toolsForStack("unknownstack"); got != nil {
		t.Fatalf("unknown stack should return nil, got %v", got)
	}
}
