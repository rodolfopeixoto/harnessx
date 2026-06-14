// SPDX-License-Identifier: MIT

package cleanup

import (
	"context"
	"sort"
	"time"
)

type Risk string

type Finding struct {
	Kind        string
	Path        string
	Risk        Risk
	Reason      string
	SizeBytes   int64
	LastTouched time.Time
}

type Detector interface {
	Name() string
	Detect(ctx context.Context, root string) ([]Finding, error)
}

type Scanner struct {
	detectors []Detector
}

func New(detectors ...Detector) *Scanner {
	return &Scanner{detectors: detectors}
}

func (s *Scanner) Scan(ctx context.Context, root string) ([]Finding, error) {
	var out []Finding
	for _, d := range s.detectors {
		findings, err := d.Detect(ctx, root)
		if err != nil {
			return nil, &DetectorError{Name: d.Name(), Err: err}
		}
		out = append(out, findings...)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Risk != out[j].Risk {
			return riskOrder(out[i].Risk) > riskOrder(out[j].Risk)
		}
		if out[i].Kind != out[j].Kind {
			return out[i].Kind < out[j].Kind
		}
		return out[i].Path < out[j].Path
	})
	return out, nil
}

type DetectorError struct {
	Name string
	Err  error
}

func (e *DetectorError) Error() string {
	return "cleanup detector " + e.Name + ": " + e.Err.Error()
}

func (e *DetectorError) Unwrap() error { return e.Err }

func riskOrder(r Risk) int {
	switch r {
	case RiskHigh:
		return 3
	case RiskMedium:
		return 2
	case RiskLow:
		return 1
	}
	return 0
}
