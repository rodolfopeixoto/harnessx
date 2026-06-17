package coverage

import (
	"strings"
	"testing"
)

func TestParseGoCoverHappy(t *testing.T) {
	in := `?   	github.com/x/empty	[no test files]
ok  	github.com/x/a	2.0s	coverage: 95.0% of statements
ok  	github.com/x/b	2.0s	coverage: 92.5% of statements
`
	r, err := ParseGoCoverString(in, 0.9)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Packages) != 2 {
		t.Fatalf("want 2 pkgs, got %d", len(r.Packages))
	}
	if !r.Pass() {
		t.Errorf("should pass: %+v", r)
	}
}

func TestParseGoCoverFailsBelowThreshold(t *testing.T) {
	in := `ok  	github.com/x/a	2.0s	coverage: 80.0% of statements
ok  	github.com/x/b	2.0s	coverage: 99.0% of statements
`
	r, err := ParseGoCoverString(in, 0.9)
	if err != nil {
		t.Fatal(err)
	}
	if r.Pass() {
		t.Errorf("should fail: avg=%.2f failed=%v", r.Average, r.Failed)
	}
	if len(r.Failed) != 1 {
		t.Errorf("want 1 failed, got %d", len(r.Failed))
	}
}

func TestParseGoCoverNoLinesErrors(t *testing.T) {
	_, err := ParseGoCoverString("nothing here", 0.9)
	if err == nil {
		t.Fatal("want error")
	}
}

func TestParseGoCoverInvalidThreshold(t *testing.T) {
	if _, err := ParseGoCoverString("anything", 0); err == nil {
		t.Fatal("want error")
	}
	if _, err := ParseGoCoverString("anything", 1.5); err == nil {
		t.Fatal("want error")
	}
}

func TestFormatResultRendersVerdict(t *testing.T) {
	in := `ok  	github.com/x/a	2.0s	coverage: 95.0% of statements
`
	r, _ := ParseGoCoverString(in, 0.9)
	out := FormatResult(r)
	if !strings.Contains(out, "verdict: PASS") {
		t.Errorf("want PASS: %s", out)
	}
}

func TestFormatResultRendersFailureDetail(t *testing.T) {
	in := `ok  	github.com/x/a	2.0s	coverage: 70.0% of statements
`
	r, _ := ParseGoCoverString(in, 0.9)
	out := FormatResult(r)
	if !strings.Contains(out, "below threshold: github.com/x/a") {
		t.Errorf("missing detail: %s", out)
	}
}
