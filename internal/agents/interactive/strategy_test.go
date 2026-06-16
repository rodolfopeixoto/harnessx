// SPDX-License-Identifier: MIT

package interactive

import (
	"testing"
	"time"
)

func TestPickStrategyDefaults(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"", "pty"},
		{"pty", "pty"},
		{"tmux", "tmux"},
		{"iterm2", "iterm2"},
		{"unknown", "pty"},
	}
	for _, c := range cases {
		got := pickStrategy(c.in).ID()
		if got != c.want {
			t.Errorf("pickStrategy(%q)=%q, want %q", c.in, got, c.want)
		}
	}
}

func TestDurationOrDefault(t *testing.T) {
	if got := durationOrDefault(0, 180); got != 180*time.Second {
		t.Errorf("zero falls back to default: got %v", got)
	}
	if got := durationOrDefault(42, 180); got != 42*time.Second {
		t.Errorf("non-zero uses seconds: got %v", got)
	}
	if got := durationOrDefault(-1, 180); got != 180*time.Second {
		t.Errorf("negative falls back: got %v", got)
	}
}

func TestIdleThreshold(t *testing.T) {
	if got := idleThreshold(0); got != defaultIdle {
		t.Errorf("zero idle should fall back: got %v", got)
	}
	if got := idleThreshold(500); got != 500*time.Millisecond {
		t.Errorf("non-zero idle: got %v", got)
	}
}

func TestEstimateTokens(t *testing.T) {
	if estimateTokens("") != 0 {
		t.Error("empty should be 0")
	}
	if got := estimateTokens("abcdefgh"); got != 2 {
		t.Errorf("8 chars / 4 = 2, got %d", got)
	}
}
