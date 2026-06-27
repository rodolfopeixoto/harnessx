package imagescmd

import "testing"

func TestShortTruncates(t *testing.T) {
	if got := short("0123456789abcdef", 6); got != "012345" {
		t.Errorf("got %s", got)
	}
	if got := short("abc", 6); got != "abc" {
		t.Errorf("got %s", got)
	}
}
