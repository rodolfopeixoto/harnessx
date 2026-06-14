// SPDX-License-Identifier: MIT

package health

import (
	"sort"

	"github.com/ropeixoto/harnessx/internal/platform/constants"
)

type Subsystem struct {
	Name   string `json:"name"`
	Score  int    `json:"score"`
	Reason string `json:"reason"`
}

type Score struct {
	Total      int         `json:"total"`
	Subsystems []Subsystem `json:"subsystems"`
}

type Inputs struct {
	TestsPassPct       int
	SensorsPassPct     int
	SecurityFindings   int
	PerfBudgetExceeded bool
	OutdatedDeps       int
	DocsCoverage       int
	DesignParityPct    int
	RoadmapClearPct    int
	MemoryFreshDays    int
	InvalidConfigs     int
}

func (i Inputs) Compute() Score {
	subs := []Subsystem{
		{Name: constants.HealthSubsystemTests, Score: percentOrDefault(i.TestsPassPct), Reason: "test pass percentage"},
		{Name: constants.HealthSubsystemSensors, Score: percentOrDefault(i.SensorsPassPct), Reason: "sensors pass percentage"},
		{Name: constants.HealthSubsystemSecurity, Score: clampInverse(i.SecurityFindings, 5), Reason: "open security findings"},
		{Name: constants.HealthSubsystemPerf, Score: boolScore(!i.PerfBudgetExceeded), Reason: "performance budget"},
		{Name: constants.HealthSubsystemDeps, Score: clampInverse(i.OutdatedDeps, 25), Reason: "outdated dependencies"},
		{Name: constants.HealthSubsystemDocs, Score: percentOrDefault(i.DocsCoverage), Reason: "documentation coverage"},
		{Name: constants.HealthSubsystemParity, Score: percentOrDefault(i.DesignParityPct), Reason: "design parity"},
		{Name: constants.HealthSubsystemRoadmap, Score: percentOrDefault(i.RoadmapClearPct), Reason: "roadmap clarity"},
		{Name: constants.HealthSubsystemMemory, Score: freshnessScore(i.MemoryFreshDays), Reason: "memory freshness"},
		{Name: constants.HealthSubsystemConfigs, Score: clampInverse(i.InvalidConfigs, 10), Reason: "invalid configs"},
	}
	sort.SliceStable(subs, func(a, b int) bool { return subs[a].Name < subs[b].Name })
	total := 0
	for _, s := range subs {
		total += s.Score
	}
	if len(subs) > 0 {
		total /= len(subs)
	}
	return Score{Total: total, Subsystems: subs}
}

func percentOrDefault(n int) int {
	if n < 0 {
		return constants.HealthDefaultScore
	}
	if n > constants.HealthMaxScore {
		return constants.HealthMaxScore
	}
	return n
}

func boolScore(ok bool) int {
	if ok {
		return constants.HealthMaxScore
	}
	return 0
}

func clampInverse(n, ceiling int) int {
	if ceiling <= 0 {
		return constants.HealthMaxScore
	}
	if n <= 0 {
		return constants.HealthMaxScore
	}
	if n >= ceiling {
		return 0
	}
	return constants.HealthMaxScore - (n * constants.HealthMaxScore / ceiling)
}

func freshnessScore(days int) int {
	switch {
	case days < 0:
		return constants.HealthDefaultScore
	case days <= 7:
		return constants.HealthMaxScore
	case days <= 30:
		return 70
	case days <= 90:
		return 40
	default:
		return 10
	}
}
