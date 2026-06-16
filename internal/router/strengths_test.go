// SPDX-License-Identifier: MIT

package router

import "testing"

func TestScoreOverlap(t *testing.T) {
	cases := []struct {
		task, adapter []string
		want          float64
	}{
		{[]string{"code"}, []string{"code", "reasoning"}, 1.0},
		{[]string{"image", "vision"}, []string{"vision"}, 0.5},
		{[]string{"code"}, []string{"docs"}, 0.0},
		{[]string{"code", "review"}, []string{"code", "review"}, 1.0},
	}
	for _, c := range cases {
		got := scoreOverlap(c.task, c.adapter)
		if got != c.want {
			t.Errorf("task=%v adapter=%v: want %v, got %v", c.task, c.adapter, c.want, got)
		}
	}
}
