package palette

import (
	"context"
	"strings"
	"testing"
)

func TestBuiltinCommandsIncludeAuditedCLIVerbs(t *testing.T) {
	must := []string{"ci", "version", "doctor", "runs list", "runs inspect", "dependency-audit", "ask", "plan", "metrics", "images list", "stack status"}
	have := map[string]bool{}
	for _, c := range BuiltinCommands {
		have[c.Name] = true
	}
	for _, want := range must {
		if !have[want] {
			t.Errorf("BuiltinCommands missing %q", want)
		}
	}
}

func TestCommandsSourceMatchesPartialQuery(t *testing.T) {
	src := CommandsSource{Commands: BuiltinCommands}
	hits, err := src.Search(context.Background(), "ci")
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) < 3 {
		t.Fatalf("expected ≥3 hits for 'ci', got %d", len(hits))
	}
	titles := make([]string, 0, len(hits))
	for _, h := range hits {
		titles = append(titles, h.Title)
	}
	joined := strings.Join(titles, ",")
	if !strings.Contains(joined, "ci") {
		t.Fatalf("expected 'ci' among titles, got %s", joined)
	}
}
