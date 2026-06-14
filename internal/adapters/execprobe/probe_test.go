package execprobe

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestProbe_Absent(t *testing.T) {
	p := &Probe{
		Lookup: func(string) (string, error) { return "", errors.New("not found") },
	}
	r := p.Run(context.Background(), "missing", nil, time.Second)
	require.False(t, r.Present)
	require.Empty(t, r.Version)
	require.NoError(t, r.Err)
}

func TestProbe_PresentNoVersion(t *testing.T) {
	p := &Probe{
		Lookup: func(string) (string, error) { return "/usr/bin/x", nil },
	}
	r := p.Run(context.Background(), "x", nil, time.Second)
	require.True(t, r.Present)
	require.Empty(t, r.Version)
}

func TestProbe_PresentWithVersion(t *testing.T) {
	p := &Probe{
		Lookup: func(string) (string, error) { return "/usr/bin/x", nil },
		Runner: func(ctx context.Context, name string, args ...string) ([]byte, error) {
			return []byte("x version 1.2.3\nextra line\n"), nil
		},
	}
	r := p.Run(context.Background(), "x", []string{"--version"}, time.Second)
	require.True(t, r.Present)
	require.Equal(t, "x version 1.2.3", r.Version)
}
