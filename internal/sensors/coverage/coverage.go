// SPDX-License-Identifier: MIT

package coverage

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
)

const DefaultThreshold = 0.90

type Result struct {
	Threshold float64
	Average   float64
	Packages  []PackageResult
	Failed    []string
}

type PackageResult struct {
	Package string
	Percent float64
	OK      bool
}

func (r Result) Pass() bool { return len(r.Failed) == 0 && r.Average >= r.Threshold }

var goCoverLine = regexp.MustCompile(`ok\s+(\S+)\s+\S+\s+coverage:\s+([0-9.]+)%`)

func ParseGoCover(r io.Reader, threshold float64) (Result, error) {
	if threshold <= 0 || threshold > 1 {
		return Result{}, errors.New("coverage: threshold must be in (0, 1]")
	}
	res := Result{Threshold: threshold}
	sc := bufio.NewScanner(r)
	var sum float64
	var n int
	for sc.Scan() {
		line := sc.Text()
		m := goCoverLine.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		pct, err := strconv.ParseFloat(m[2], 64)
		if err != nil {
			continue
		}
		ratio := pct / 100.0
		pr := PackageResult{Package: m[1], Percent: ratio, OK: ratio >= threshold}
		res.Packages = append(res.Packages, pr)
		sum += ratio
		n++
		if !pr.OK {
			res.Failed = append(res.Failed, fmt.Sprintf("%s (%.1f%%)", m[1], pct))
		}
	}
	if err := sc.Err(); err != nil {
		return res, err
	}
	if n == 0 {
		return res, errors.New("coverage: no package coverage lines found in input")
	}
	res.Average = sum / float64(n)
	return res, nil
}

func ParseGoCoverString(s string, threshold float64) (Result, error) {
	return ParseGoCover(strings.NewReader(s), threshold)
}

func FormatResult(r Result) string {
	var b strings.Builder
	fmt.Fprintf(&b, "coverage: threshold=%.0f%% average=%.1f%% packages=%d\n",
		r.Threshold*100, r.Average*100, len(r.Packages))
	for _, p := range r.Packages {
		mark := "✓"
		if !p.OK {
			mark = "✗"
		}
		fmt.Fprintf(&b, "  %s %s %.1f%%\n", mark, p.Package, p.Percent*100)
	}
	if r.Pass() {
		fmt.Fprintln(&b, "verdict: PASS")
	} else {
		fmt.Fprintln(&b, "verdict: FAIL")
		for _, f := range r.Failed {
			fmt.Fprintf(&b, "  below threshold: %s\n", f)
		}
	}
	return b.String()
}
