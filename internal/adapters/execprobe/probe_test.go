package execprobe

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func runner(stdout, stderr string, err error) func(context.Context, string, ...string) ([]byte, []byte, error) {
	return func(context.Context, string, ...string) ([]byte, []byte, error) {
		return []byte(stdout), []byte(stderr), err
	}
}

func TestProbe_Absent(t *testing.T) {
	p := &Probe{Lookup: func(string) (string, error) { return "", errors.New("not found") }}
	r := p.Run(context.Background(), "missing", nil, time.Second)
	require.False(t, r.Present)
	require.Empty(t, r.Version)
	require.NoError(t, r.Err)
}

func TestProbe_PresentNoVersion(t *testing.T) {
	p := &Probe{Lookup: func(string) (string, error) { return "/usr/bin/x", nil }}
	r := p.Run(context.Background(), "x", nil, time.Second)
	require.True(t, r.Present)
	require.Empty(t, r.Version)
}

func TestProbe_PresentWithVersion(t *testing.T) {
	p := &Probe{
		Lookup: func(string) (string, error) { return "/usr/bin/x", nil },
		Runner: runner("x version 1.2.3\nextra line\n", "", nil),
	}
	r := p.Run(context.Background(), "x", []string{"--version"}, time.Second)
	require.True(t, r.Present)
	require.Equal(t, "x version 1.2.3", r.Version)
	require.NoError(t, r.Err)
}

func TestProbe_NonZeroExitButVersionExtracted(t *testing.T) {
	p := &Probe{
		Lookup: func(string) (string, error) { return "/usr/bin/gemini", nil },
		Runner: runner("Gemini CLI 0.9.1\n", "", errors.New("exit status 1")),
	}
	r := p.Run(context.Background(), "gemini", []string{"--version"}, time.Second)
	require.True(t, r.Present)
	require.Equal(t, "Gemini CLI 0.9.1", r.Version)
	require.NoError(t, r.Err, "version present should clear Err even on non-zero exit")
}

func TestProbe_CustomRegexExtractsSemverFromGoOutput(t *testing.T) {
	p := &Probe{
		Lookup: func(string) (string, error) { return "/usr/bin/go", nil },
		Runner: runner("go version go1.21.0 darwin/amd64\n", "", nil),
	}
	r := p.RunSpec(context.Background(), Spec{
		Binary:       "go",
		Args:         []string{"version"},
		VersionRegex: `go version go(\d+\.\d+(?:\.\d+)?)`,
	})
	require.Equal(t, "1.21.0", r.Version)
}

func TestProbe_VersionFromStderrWhenStdoutEmpty(t *testing.T) {
	p := &Probe{
		Lookup: func(string) (string, error) { return "/usr/bin/x", nil },
		Runner: runner("", "x 2.0.0\n", errors.New("exit status 2")),
	}
	r := p.Run(context.Background(), "x", []string{"--version"}, time.Second)
	require.True(t, r.Present)
	require.Equal(t, "x 2.0.0", r.Version)
	require.NoError(t, r.Err)
}

func TestProbe_RealFailureNoVersionKeepsErr(t *testing.T) {
	p := &Probe{
		Lookup: func(string) (string, error) { return "/usr/bin/x", nil },
		Runner: runner("usage info, no version anywhere", "permission denied", errors.New("exit status 1")),
	}
	r := p.Run(context.Background(), "x", []string{"--version"}, time.Second)
	require.True(t, r.Present)
	require.Error(t, r.Err)
}

func TestProbe_RuntimeErrorMessageWithSemverIsStillError(t *testing.T) {
	p := &Probe{
		Lookup: func(string) (string, error) { return "/usr/local/bin/go", nil },
		Runner: runner("go: cannot find GOROOT directory: /Users/x/.gvm/gos/go1.19.2", "", errors.New("exit status 2")),
	}
	r := p.Run(context.Background(), "go", []string{"version"}, time.Second)
	require.True(t, r.Present)
	require.Error(t, r.Err, "runtime error must not be masked by an incidental semver match")
}

func TestProbe_CommandNotFoundFlagsAsError(t *testing.T) {
	p := &Probe{
		Lookup: func(string) (string, error) { return "/bin/x", nil },
		Runner: runner("zsh: command not found: x 1.0.0", "", errors.New("exit status 127")),
	}
	r := p.Run(context.Background(), "x", []string{"--version"}, time.Second)
	require.Error(t, r.Err)
}
