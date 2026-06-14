// SPDX-License-Identifier: MIT

package optimize

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// scanDockerfile performs static analysis on the project's root Dockerfile.
// Returns nil when no Dockerfile is present.
func scanDockerfile(root string) *DockerfileMetrics {
	candidates := []string{"Dockerfile", "docker/Dockerfile"}
	var path string
	for _, c := range candidates {
		full := filepath.Join(root, c)
		if _, err := os.Stat(full); err == nil {
			path = full
			break
		}
	}
	if path == "" {
		return nil
	}
	rel, _ := filepath.Rel(root, path)
	d := &DockerfileMetrics{Path: rel}
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1024*1024), 4*1024*1024)
	var firstFrom string
	hasCleanup := false
	for scanner.Scan() {
		raw := strings.TrimSpace(scanner.Text())
		if raw == "" || strings.HasPrefix(raw, "#") {
			continue
		}
		upper := strings.ToUpper(raw)
		switch {
		case strings.HasPrefix(upper, "FROM "):
			d.Stages++
			if firstFrom == "" {
				firstFrom = strings.TrimSpace(raw[len("FROM "):])
				d.BaseImage = firstFrom
			}
		case strings.HasPrefix(upper, "RUN "):
			d.RunSteps++
			body := strings.ToLower(raw)
			if strings.Contains(body, "rm -rf /var/lib/apt/lists/*") ||
				strings.Contains(body, "yum clean all") ||
				strings.Contains(body, "apk --no-cache") ||
				strings.Contains(body, "npm cache clean") {
				hasCleanup = true
			}
		case strings.HasPrefix(upper, "COPY "), strings.HasPrefix(upper, "ADD "):
			d.CopySteps++
		case strings.HasPrefix(upper, "USER "):
			d.HasUSER = true
		case strings.HasPrefix(upper, "HEALTHCHECK"):
			d.HasHealthcheck = true
		}
	}
	d.HasCacheCleanup = hasCleanup
	d.UsesLatestTag = strings.HasSuffix(firstFrom, ":latest") || (firstFrom != "" && !strings.Contains(firstFrom, ":"))

	d.Findings = dockerfileFindings(d)
	return d
}

func dockerfileFindings(d *DockerfileMetrics) []Finding {
	var out []Finding
	if d.UsesLatestTag {
		out = append(out, Finding{
			ID: "docker.latest_tag", Severity: SeverityWarn,
			Message: "base image uses `:latest` (or no tag) — pin a digest for reproducible builds",
			Detail:  d.BaseImage,
		})
	}
	if !d.HasUSER {
		out = append(out, Finding{
			ID: "docker.no_user", Severity: SeverityWarn,
			Message: "no USER directive — container will run as root",
		})
	}
	if !d.HasHealthcheck {
		out = append(out, Finding{
			ID: "docker.no_healthcheck", Severity: SeverityInfo,
			Message: "no HEALTHCHECK declared — orchestrator can't probe readiness",
		})
	}
	if d.RunSteps >= 3 && !d.HasCacheCleanup {
		out = append(out, Finding{
			ID: "docker.no_cache_cleanup", Severity: SeverityWarn,
			Message: "RUN steps install packages but never clear apt/yum/npm caches — image will be larger than necessary",
		})
	}
	if d.Stages == 1 && d.CopySteps > 4 {
		out = append(out, Finding{
			ID: "docker.single_stage_heavy", Severity: SeverityInfo,
			Message: "single-stage build with many COPYs — consider a multi-stage build to keep build deps out of the runtime image",
		})
	}
	return out
}
