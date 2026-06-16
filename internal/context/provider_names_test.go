// SPDX-License-Identifier: MIT

package context

import "testing"

func TestProviderNames(t *testing.T) {
	cases := []struct {
		p    Provider
		want string
	}{
		{GitProvider{}, "git"},
		{MemoryProvider{}, "memory"},
		{RipgrepProvider{}, "ripgrep"},
	}
	for _, c := range cases {
		if got := c.p.Name(); got != c.want {
			t.Errorf("%T.Name()=%q, want %q", c.p, got, c.want)
		}
	}
}

func TestLastSegment(t *testing.T) {
	cases := map[string]string{
		"a/b/c":      "c",
		"single":     "single",
		"":           "",
		"trailing/":  "",
		"./relative": "relative",
	}
	for in, want := range cases {
		if got := lastSegment(in); got != want {
			t.Errorf("lastSegment(%q)=%q, want %q", in, got, want)
		}
	}
}

func TestAppendUniqueDeduplicates(t *testing.T) {
	out := []string{"a", "b"}
	for _, s := range []string{"a", "c", "b", "d"} {
		out = appendUnique(out, s)
	}
	want := []string{"a", "b", "c", "d"}
	if len(out) != len(want) {
		t.Fatalf("len: want %d, got %d (%v)", len(want), len(out), out)
	}
	for i, w := range want {
		if out[i] != w {
			t.Errorf("idx %d: want %q, got %q", i, w, out[i])
		}
	}
}
