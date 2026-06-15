// SPDX-License-Identifier: MIT

package execprobe

import (
	"context"
	"errors"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

type Result struct {
	Binary  string
	Present bool
	Version string
	Err     error
	Took    time.Duration
}

type Probe struct {
	Lookup func(string) (string, error)
	Runner func(ctx context.Context, name string, args ...string) ([]byte, []byte, error)
}

func Default() *Probe {
	return &Probe{
		Lookup: exec.LookPath,
		Runner: func(ctx context.Context, name string, args ...string) ([]byte, []byte, error) {
			cmd := exec.CommandContext(ctx, name, args...)
			var stderr strings.Builder
			cmd.Stderr = &stderr
			stdout, err := cmd.Output()
			return stdout, []byte(stderr.String()), err
		},
	}
}

// Spec drives a single probe run. VersionRegex is optional; when set, the
// first capturing group from a match against stdout+stderr becomes the
// reported version even if the underlying command exited non-zero. This
// is how binaries that print "go version go1.21.0 darwin/amd64" and exit
// 0, OR "Gemini CLI 0.9.1" then exit 1, both surface as ✓ with a clean
// semver.
type Spec struct {
	Binary       string
	Args         []string
	Timeout      time.Duration
	VersionRegex string
}

var defaultSemver = regexp.MustCompile(`(\d+\.\d+(?:\.\d+)?(?:[-+\.][\w.-]+)?)`)

func (p *Probe) Run(parent context.Context, binary string, versionArgs []string, timeout time.Duration) Result {
	return p.RunSpec(parent, Spec{Binary: binary, Args: versionArgs, Timeout: timeout})
}

func (p *Probe) RunSpec(parent context.Context, spec Spec) Result {
	timeout := spec.Timeout
	if timeout <= 0 {
		timeout = 2 * time.Second
	}
	start := time.Now()
	res := Result{Binary: spec.Binary}
	if _, err := p.Lookup(spec.Binary); err != nil {
		res.Took = time.Since(start)
		return res
	}
	res.Present = true
	if len(spec.Args) == 0 {
		res.Took = time.Since(start)
		return res
	}
	ctx, cancel := context.WithTimeout(parent, timeout)
	defer cancel()
	stdout, stderr, err := p.Runner(ctx, spec.Binary, spec.Args...)
	combined := joinNonEmpty(string(stdout), string(stderr))
	if combined == "" && err != nil {
		if exitErr := new(exec.ExitError); errors.As(err, &exitErr) && len(exitErr.Stderr) > 0 {
			combined = string(exitErr.Stderr)
		}
	}
	version, matched := extractVersion(combined, spec.VersionRegex)
	if looksLikeRuntimeError(combined) {
		res.Version = firstLine(combined)
		if err == nil {
			err = errors.New("probe: runtime error in output")
		}
		res.Err = err
		res.Took = time.Since(start)
		return res
	}
	res.Version = version
	if !matched && err != nil {
		res.Err = err
	}
	res.Took = time.Since(start)
	return res
}

var runtimeErrorMarkers = []string{
	"cannot find",
	"command not found",
	"no such file",
	"permission denied",
	"unknown command",
	"error:",
	"fatal:",
	"panic:",
}

func looksLikeRuntimeError(output string) bool {
	low := strings.ToLower(output)
	for _, marker := range runtimeErrorMarkers {
		if strings.Contains(low, marker) {
			return true
		}
	}
	return false
}

func joinNonEmpty(a, b string) string {
	switch {
	case a == "":
		return b
	case b == "":
		return a
	default:
		return a + "\n" + b
	}
}

func extractVersion(output, customRegex string) (string, bool) {
	output = strings.TrimSpace(output)
	if output == "" {
		return "", false
	}
	if customRegex != "" {
		if re, err := regexp.Compile(customRegex); err == nil {
			if m := re.FindStringSubmatch(output); len(m) > 1 {
				return strings.TrimSpace(m[1]), true
			} else if len(m) == 1 {
				return strings.TrimSpace(m[0]), true
			}
		}
	}
	first := firstLine(output)
	if defaultSemver.FindString(first) != "" {
		return first, true
	}
	if m := defaultSemver.FindString(output); m != "" {
		return m, true
	}
	return "", false
}

func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		s = s[:i]
	}
	return strings.TrimSpace(s)
}
