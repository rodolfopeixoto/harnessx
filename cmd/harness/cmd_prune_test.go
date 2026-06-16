// SPDX-License-Identifier: MIT

package main

import (
	"testing"
	"time"
)

func TestParseDurationDays(t *testing.T) {
	cases := []struct {
		in   string
		want time.Duration
	}{
		{"", 0},
		{"7d", 7 * 24 * time.Hour},
		{"30D", 30 * 24 * time.Hour},
		{"2h", 2 * time.Hour},
		{"45m", 45 * time.Minute},
	}
	for _, c := range cases {
		got, err := parseDuration(c.in)
		if err != nil {
			t.Errorf("parseDuration(%q): %v", c.in, err)
			continue
		}
		if got != c.want {
			t.Errorf("parseDuration(%q): want %v, got %v", c.in, c.want, got)
		}
	}
}

func TestParseDurationInvalid(t *testing.T) {
	if _, err := parseDuration("garbage"); err == nil {
		t.Error("garbage should error")
	}
	if _, err := parseDuration("xd"); err == nil {
		t.Error("xd should error")
	}
}

func TestRunsPruneCmdFlags(t *testing.T) {
	c := newRunsPruneCmd()
	for _, want := range []string{"older-than", "keep-last", "apply"} {
		if c.Flags().Lookup(want) == nil {
			t.Errorf("flag missing: %s", want)
		}
	}
}

func TestProjectPruneCmdFlags(t *testing.T) {
	c := newProjectPruneCmd()
	for _, want := range []string{"older-than", "apply"} {
		if c.Flags().Lookup(want) == nil {
			t.Errorf("flag missing: %s", want)
		}
	}
}
