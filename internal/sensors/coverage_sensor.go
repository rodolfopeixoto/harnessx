// SPDX-License-Identifier: MIT

package sensors

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"time"

	"github.com/ropeixoto/harnessx/internal/index"
	"github.com/ropeixoto/harnessx/internal/sensors/coverage"
)

type CoverageSensor struct {
	IDValue     string
	Threshold   float64
	PackageGlob string
	Stacks      []string
	Timeout     time.Duration
	Runner      CoverageRunner
}

type CoverageRunner func(ctx context.Context, root, pkg string) ([]byte, error)

func (s CoverageSensor) ID() string         { return s.IDValue }
func (s CoverageSensor) Category() Category { return CatTest }
func (s CoverageSensor) Kind() Kind         { return KindComputational }

func (s CoverageSensor) AppliesTo(p index.Profile) bool {
	if len(s.Stacks) == 0 {
		return true
	}
	for _, st := range p.Stacks {
		for _, want := range s.Stacks {
			if st.Name == want {
				return true
			}
		}
	}
	return false
}

func (s CoverageSensor) Run(rc RunCtx) Result {
	start := time.Now()
	threshold := s.Threshold
	if threshold == 0 {
		threshold = coverage.DefaultThreshold
	}
	pkg := s.PackageGlob
	if pkg == "" {
		pkg = "./..."
	}
	timeout := s.Timeout
	if timeout == 0 {
		timeout = 5 * time.Minute
	}
	ctx, cancel := context.WithTimeout(rc.Ctx, timeout)
	defer cancel()

	runner := s.Runner
	if runner == nil {
		runner = defaultCoverageRunner
	}
	out, err := runner(ctx, rc.Root, pkg)
	res := Result{ID: s.IDValue, Category: CatTest, Kind: KindComputational}
	if err != nil {
		res.Status = StatusFailed
		res.Detail = fmt.Sprintf("go test failed: %v\n%s", err, truncate(string(out), 2_000))
		res.Duration = time.Since(start)
		return res
	}
	parsed, perr := coverage.ParseGoCoverString(string(out), threshold)
	if perr != nil {
		res.Status = StatusFailed
		res.Detail = perr.Error()
		res.Duration = time.Since(start)
		return res
	}
	res.Detail = coverage.FormatResult(parsed)
	res.Confidence = parsed.Average
	if parsed.Pass() {
		res.Status = StatusPassed
		res.Verified = packageNames(parsed)
	} else {
		res.Status = StatusFailed
		res.Unverified = parsed.Failed
	}
	res.Duration = time.Since(start)
	return res
}

func packageNames(r coverage.Result) []string {
	out := make([]string, 0, len(r.Packages))
	for _, p := range r.Packages {
		out = append(out, p.Package)
	}
	return out
}

func defaultCoverageRunner(ctx context.Context, root, pkg string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "go", "test", "-cover", pkg)
	cmd.Dir = root
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	if err := cmd.Run(); err != nil {
		return buf.Bytes(), err
	}
	return buf.Bytes(), nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "...[truncated]"
}

func goCoverageSensorDefault() CoverageSensor {
	return CoverageSensor{
		IDValue:     "go_coverage_gate",
		Threshold:   coverage.DefaultThreshold,
		PackageGlob: "./...",
		Stacks:      []string{"go"},
		Timeout:     5 * time.Minute,
	}
}
