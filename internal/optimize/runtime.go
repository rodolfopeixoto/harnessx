// SPDX-License-Identifier: MIT

package optimize

import (
	"bytes"
	"context"
	"encoding/json"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// RuntimeMetrics covers spec §21 Cycle E. Populated best-effort from
// `docker stats` and `runtime` package introspection; zero values mean
// "not measured" rather than "zero".
type RuntimeMetrics struct {
	ProcessRSSMB         float64          `json:"process_rss_mb,omitempty"`
	ProcessNumGoroutines int              `json:"process_num_goroutines,omitempty"`
	Containers           []ContainerStats `json:"containers,omitempty"`
	DockerAvailable      bool             `json:"docker_available"`
}

type ContainerStats struct {
	Name       string  `json:"name"`
	CPUPercent float64 `json:"cpu_percent"`
	MemUsageMB float64 `json:"mem_usage_mb"`
	MemLimitMB float64 `json:"mem_limit_mb,omitempty"`
}

// captureRuntime collects host process + container stats. Always returns
// a struct; check DockerAvailable to know whether container rows are real.
func captureRuntime() RuntimeMetrics {
	m := RuntimeMetrics{
		ProcessNumGoroutines: runtime.NumGoroutine(),
		ProcessRSSMB:         rssMB(),
	}
	if _, err := exec.LookPath("docker"); err == nil {
		m.DockerAvailable = true
		m.Containers = dockerStats()
	}
	return m
}

func rssMB() float64 {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	return float64(ms.Sys) / 1024.0 / 1024.0
}

// dockerStats invokes `docker stats --no-stream --format '{{json .}}'`.
// 5 s timeout so a hung daemon never blocks `harness perf-snapshot`.
func dockerStats() []ContainerStats {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "docker", "stats", "--no-stream", "--format", "{{json .}}")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return nil
	}
	var rows []ContainerStats
	for _, line := range strings.Split(strings.TrimRight(out.String(), "\n"), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var raw struct {
			Name   string `json:"Name"`
			CPU    string `json:"CPUPerc"`
			MemPct string `json:"MemPerc"`
			Mem    string `json:"MemUsage"`
		}
		if err := json.Unmarshal([]byte(line), &raw); err != nil {
			continue
		}
		used, limit := splitMemUsage(raw.Mem)
		rows = append(rows, ContainerStats{
			Name:       raw.Name,
			CPUPercent: parsePercent(raw.CPU),
			MemUsageMB: used,
			MemLimitMB: limit,
		})
	}
	return rows
}

func parsePercent(s string) float64 {
	s = strings.TrimSuffix(strings.TrimSpace(s), "%")
	v, _ := strconv.ParseFloat(s, 64)
	return v
}

// splitMemUsage parses strings like "12.34MiB / 1.95GiB" into MB.
func splitMemUsage(s string) (used, limit float64) {
	parts := strings.SplitN(s, "/", 2)
	used = parseBytes(strings.TrimSpace(parts[0]))
	if len(parts) == 2 {
		limit = parseBytes(strings.TrimSpace(parts[1]))
	}
	return
}

func parseBytes(s string) float64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	var unit string
	for i := len(s) - 1; i >= 0; i-- {
		c := s[i]
		if c >= '0' && c <= '9' || c == '.' {
			unit = s[i+1:]
			s = s[:i+1]
			break
		}
	}
	n, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	switch strings.ToUpper(unit) {
	case "B":
		return n / 1024.0 / 1024.0
	case "KIB", "KB":
		return n / 1024.0
	case "MIB", "MB":
		return n
	case "GIB", "GB":
		return n * 1024.0
	case "TIB", "TB":
		return n * 1024.0 * 1024.0
	}
	return n
}
