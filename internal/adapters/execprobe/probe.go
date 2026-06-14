// SPDX-License-Identifier: MIT

package execprobe

import (
	"context"
	"errors"
	"os/exec"
	"strings"
	"time"
)

// Result reports the outcome of probing a binary.
type Result struct {
	Binary  string
	Present bool   // resolved on PATH
	Version string // first line of version output, trimmed
	Err     error  // non-nil when running --version failed
	Took    time.Duration
}

// Probe checks whether binary is on PATH and, if versionArgs is non-empty,
// invokes it with that arg list capturing the first line of stdout. The
// probe is bounded by timeout; on timeout Err is set but Present reflects
// the PATH lookup.
type Probe struct {
	Lookup func(string) (string, error) // injectable for tests
	Runner func(ctx context.Context, name string, args ...string) ([]byte, error)
}

func Default() *Probe {
	return &Probe{
		Lookup: exec.LookPath,
		Runner: func(ctx context.Context, name string, args ...string) ([]byte, error) {
			return exec.CommandContext(ctx, name, args...).Output()
		},
	}
}

func (p *Probe) Run(parent context.Context, binary string, versionArgs []string, timeout time.Duration) Result {
	if timeout <= 0 {
		timeout = 2 * time.Second
	}
	start := time.Now()
	res := Result{Binary: binary}
	if _, err := p.Lookup(binary); err != nil {
		res.Took = time.Since(start)
		return res
	}
	res.Present = true
	if len(versionArgs) == 0 {
		res.Took = time.Since(start)
		return res
	}
	ctx, cancel := context.WithTimeout(parent, timeout)
	defer cancel()
	out, err := p.Runner(ctx, binary, versionArgs...)
	if err != nil {
		// Some tools print the banner to stderr; still try to surface it.
		if exitErr := new(exec.ExitError); errors.As(err, &exitErr) && len(exitErr.Stderr) > 0 {
			res.Version = firstLine(string(exitErr.Stderr))
		}
		res.Err = err
		res.Took = time.Since(start)
		return res
	}
	res.Version = firstLine(string(out))
	res.Took = time.Since(start)
	return res
}

func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		s = s[:i]
	}
	return strings.TrimSpace(s)
}
