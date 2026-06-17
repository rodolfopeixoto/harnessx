package configwiz

import (
	"bufio"
	"bytes"
	"strings"
	"testing"

	"github.com/ropeixoto/harnessx/internal/router"
)

func TestFormatRouteAllFields(t *testing.T) {
	r := router.RouteConfig{Primary: "p", Fallback: []string{"a", "b"}, BudgetUSD: 1.5, Model: "x"}
	got := formatRoute(r)
	for _, want := range []string{"primary=p", "fallback=a,b", "budget=$1.50", "model=x"} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q: %q", want, got)
		}
	}
}

func TestFormatRouteOnlyPrimary(t *testing.T) {
	got := formatRoute(router.RouteConfig{Primary: "only"})
	if got != "primary=only" {
		t.Errorf("got %q", got)
	}
}

func TestSplitCSVEmptyReturnsNil(t *testing.T) {
	if splitCSV("   ") != nil {
		t.Error("blank must yield nil")
	}
}

func TestRunWizardSkipsBlankCSVFallback(t *testing.T) {
	dir := t.TempDir()
	in := strings.NewReader("kimi\n\n0\n")
	var out bytes.Buffer
	err := RunWizard(WizardOptions{
		Root:         dir,
		AvailableIDs: []string{"kimi"},
		Tasks:        []string{"cheap_review"},
		In:           in,
		Out:          &out,
	})
	if err != nil {
		t.Fatal(err)
	}
	snap, _ := Load(dir)
	if len(snap.Routes["cheap_review"].Fallback) != 0 {
		t.Errorf("fallback should be empty, got %v", snap.Routes["cheap_review"].Fallback)
	}
}

func TestRunWizardKeepsDefaultsOnBlankPrimary(t *testing.T) {
	dir := t.TempDir()
	_ = SetTaskPrimary(dir, "planning", "claude", []string{"kimi"}, 0.7, "")

	in := strings.NewReader("\n\n\n")
	var out bytes.Buffer
	if err := RunWizard(WizardOptions{
		Root:         dir,
		AvailableIDs: []string{"claude", "kimi"},
		Tasks:        []string{"planning"},
		In:           in,
		Out:          &out,
	}); err != nil {
		t.Fatal(err)
	}
	snap, _ := Load(dir)
	r := snap.Routes["planning"]
	if r.Primary != "claude" || r.BudgetUSD != 0.7 {
		t.Errorf("defaults lost: %+v", r)
	}
}

func TestReadLineFallsBackOnBlank(t *testing.T) {
	br := bufio.NewReader(strings.NewReader("\n"))
	got, err := readLine(br, &bytes.Buffer{}, "label", "default")
	if err != nil {
		t.Fatal(err)
	}
	if got != "default" {
		t.Errorf("got %q", got)
	}
}

func TestReadNonEmptyRepromptsOnBlankNoDefault(t *testing.T) {
	br := bufio.NewReader(strings.NewReader("\nfinal\n"))
	got, err := readNonEmpty(br, &bytes.Buffer{}, "label", "")
	if err != nil {
		t.Fatal(err)
	}
	if got != "final" {
		t.Errorf("got %q", got)
	}
}
