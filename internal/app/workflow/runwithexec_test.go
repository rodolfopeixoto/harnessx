// SPDX-License-Identifier: MIT

package workflow

import (
	"testing"

	"github.com/ropeixoto/harnessx/internal/domain"
	"github.com/ropeixoto/harnessx/internal/execution"
)

func TestMapMode(t *testing.T) {
	cases := []struct {
		in   domain.Mode
		want execution.Mode
	}{
		{domain.ModeFeature, execution.ModeFeature},
		{domain.ModeBugfix, execution.ModeBugfix},
		{domain.ModeQuestion, execution.ModeAsk},
		{domain.ModeBootstrap, execution.ModeFeature},
		{domain.Mode("random"), execution.ModeFeature},
	}
	for _, c := range cases {
		if got := mapMode(c.in); got != c.want {
			t.Errorf("mapMode(%q)=%v, want %v", c.in, got, c.want)
		}
	}
}

func TestDefaultSensorsRespectsFlag(t *testing.T) {
	if got := defaultSensors(true); got != nil {
		t.Errorf("disabled true should return nil, got %v", got)
	}
	if got := defaultSensors(false); got != nil {
		t.Errorf("current impl returns nil regardless; got %v", got)
	}
}
