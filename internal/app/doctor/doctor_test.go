package doctor

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/ropeixoto/harnessx/internal/adapters/execprobe"
)

func TestRun_PartitionsAndDetectsRequired(t *testing.T) {
	specs := []ProbeSpec{
		{Binary: "git", Label: "Git", Required: true, Category: "tool"},
		{Binary: "docker", Label: "Docker", Required: false, Category: "tool"},
		{Binary: "claude", Label: "Claude", Required: false, Category: "agent"},
	}
	present := map[string]bool{"git": true, "claude": true}
	probe := &execprobe.Probe{
		Lookup: func(name string) (string, error) {
			if present[name] {
				return "/usr/bin/" + name, nil
			}
			return "", errors.New("not found")
		},
	}
	r := Run(context.Background(), probe, specs, ProjectInfo{}, time.Second)
	require.Len(t, r.Tools, 2)
	require.Len(t, r.Agents, 1)
	require.True(t, r.AllRequiredPresent())
}

func TestAllRequiredPresent_False(t *testing.T) {
	specs := []ProbeSpec{
		{Binary: "git", Required: true, Category: "tool"},
	}
	probe := &execprobe.Probe{
		Lookup: func(string) (string, error) { return "", errors.New("nope") },
	}
	r := Run(context.Background(), probe, specs, ProjectInfo{}, time.Second)
	require.False(t, r.AllRequiredPresent())
}
