package installcmd

import (
	"bytes"
	"context"
	"sort"
	"strings"
	"testing"
)

func TestToolsForStackKnown(t *testing.T) {
	cases := map[string][]string{
		"swift":   {"swift-format", "swiftlint"},
		"kotlin":  {"detekt", "ktlint"},
		"php":     {"php-cs-fixer", "phpstan", "psalm"},
		"dotnet":  {"dotnet"},
		"flutter": {"dart"},
	}
	for stack, want := range cases {
		got := append([]string{}, ToolsForStack(stack)...)
		sort.Strings(got)
		sort.Strings(want)
		if len(got) != len(want) {
			t.Fatalf("stack %s: want %v, got %v", stack, want, got)
		}
		for i := range want {
			if got[i] != want[i] {
				t.Errorf("stack %s pos %d: want %s, got %s", stack, i, want[i], got[i])
			}
		}
	}
}

func TestToolsForStackUnknownNil(t *testing.T) {
	if got := ToolsForStack("nope"); got != nil {
		t.Fatalf("want nil, got %v", got)
	}
}

func TestInstallReportsAlreadyInstalled(t *testing.T) {
	var buf bytes.Buffer
	err := Install(context.Background(), &buf, InstallOptions{
		Stack: "swift",
		Probe: func(_ string) bool { return true },
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "already installed") {
		t.Fatalf("expected already-installed line, got:\n%s", buf.String())
	}
}

func TestInstallNoPackForStack(t *testing.T) {
	err := Install(context.Background(), &bytes.Buffer{}, InstallOptions{Stack: "nope"})
	if err == nil {
		t.Fatal("expected error for unknown stack")
	}
}
