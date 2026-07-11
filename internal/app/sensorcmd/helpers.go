// SPDX-License-Identifier: MIT

package sensorcmd

import "github.com/ropeixoto/harnessx/internal/sensors"

func filterByIDs(in []sensors.Sensor, ids []string) []sensors.Sensor {
	if len(ids) == 0 {
		return in
	}
	want := map[string]bool{}
	for _, id := range ids {
		want[id] = true
	}
	var out []sensors.Sensor
	for _, s := range in {
		if want[s.ID()] {
			out = append(out, s)
		}
	}
	return out
}

func icon(s sensors.Status) string {
	switch s {
	case sensors.StatusPassed:
		return "✓"
	case sensors.StatusFailed:
		return "✗"
	default:
		return "·"
	}
}

func detail(res sensors.Result) string {
	if res.Detail == "" {
		return ""
	}
	return "— " + res.Detail
}

func ifFail(b bool) int {
	if b {
		return 1
	}
	return 0
}
