// SPDX-License-Identifier: MIT

package sensors

import (
	"bufio"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/ropeixoto/harnessx/internal/index"
	"github.com/ropeixoto/harnessx/internal/platform/ignore"
)

// forbiddenPatterns are file paths/globs that must never enter the
// repository. Matched against project-relative paths.
var forbiddenPatterns = []string{
	".env",
	".envrc",
	"*.pem",
	"*.key",
	"id_rsa",
	"id_dsa",
	"id_ed25519",
	"secrets.yml",
	"secrets.yaml",
	"credentials.json",
	".aws/credentials",
	".npmrc.local",
}

// forbiddenCommandRe captures patterns that suggest a destructive or
// unsafe shell snippet entering tracked code (scripts, Makefiles, CI
// configs). Tested against changed-file content, not status.
// Go's regexp engine (RE2) doesn't support lookaround. We accept
// `git push --force-with-lease` as a hit alongside `--force`; callers
// who legitimately need force-with-lease can suppress via a future
// `.harness/config/sensors.yaml` allowlist.
var forbiddenCommandRe = regexp.MustCompile(
	`(?i)(chmod\s+-R\s+777|rm\s+-rf\s+/|curl\s+[^\n]*\s\|\s*bash|wget\s+[^\n]*\s\|\s*bash|git\s+push\s+--force|--no-verify|sudo\s+rm\s+-rf\s+/)`,
)

// secretPatterns are deliberately conservative; false positives are worse
// than false negatives here because every hit triggers a sensor failure.
var secretPatterns = []*regexp.Regexp{
	regexp.MustCompile(`AKIA[0-9A-Z]{16}`), // AWS access key
	regexp.MustCompile(`(?i)aws_secret_access_key\s*=\s*[A-Za-z0-9/+=]{40}`),
	regexp.MustCompile(`(?i)slack_(?:bot|user|app)_token\s*=\s*xox[abprs]-[A-Za-z0-9-]{20,}`),
	regexp.MustCompile(`xox[abprs]-[A-Za-z0-9-]{20,}`),
	regexp.MustCompile(`-----BEGIN (?:RSA |EC |DSA |OPENSSH )?PRIVATE KEY-----`),
	regexp.MustCompile(`gh[opusr]_[A-Za-z0-9]{36,}`), // GitHub tokens
	regexp.MustCompile(`(?i)api[_-]?key\s*[:=]\s*['"][A-Za-z0-9_\-]{24,}['"]`),
}

// scanExcludedDirs are skipped during forbidden/secrets walks. Matches the
// dirs we already exclude from test discovery.
var scanExcludedDirs = map[string]bool{
	".git":         true,
	".harness":     true,
	"node_modules": true,
	"vendor":       true,
	"target":       true,
	"dist":         true,
	"build":        true,
	"bin":          true,
}

// ForbiddenFilesSensor scans the working tree for files that must not be
// tracked. Pure Go, no shelling out — runs on every project.
type ForbiddenFilesSensor struct{}

func (ForbiddenFilesSensor) ID() string                     { return "forbidden_files" }
func (ForbiddenFilesSensor) Category() Category             { return CatForbidden }
func (ForbiddenFilesSensor) Kind() Kind                     { return KindComputational }
func (ForbiddenFilesSensor) AppliesTo(p index.Profile) bool { return true }

func (s ForbiddenFilesSensor) Run(rc RunCtx) Result {
	start := time.Now()
	res := Result{ID: s.ID(), Category: s.Category(), Kind: s.Kind()}
	ig, _ := ignore.Load(rc.Root)
	var hits []string
	_ = filepath.WalkDir(rc.Root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		rel, _ := filepath.Rel(rc.Root, p)
		if d.IsDir() {
			if scanExcludedDirs[d.Name()] || ig.Match(rel, true) {
				return filepath.SkipDir
			}
			return nil
		}
		if ig.Match(rel, false) {
			return nil
		}
		for _, pat := range forbiddenPatterns {
			matched, _ := filepath.Match(pat, filepath.Base(rel))
			if matched || rel == pat || strings.HasSuffix(rel, "/"+pat) {
				hits = append(hits, rel)
				break
			}
		}
		return nil
	})
	sort.Strings(hits)
	res.Duration = time.Since(start)
	if len(hits) == 0 {
		res.Status = StatusPassed
		return res
	}
	res.Status = StatusFailed
	res.Detail = "forbidden files present: " + strings.Join(hits, ", ")
	res.OutputPath = writeOutput(rc.OutputDir, s.ID(), []byte(strings.Join(hits, "\n")+"\n"), nil)
	return res
}

// ForbiddenCommandsSensor scans tracked text files (scripts, Makefile, CI
// configs, Dockerfiles, .sh) for dangerous shell invocations.
type ForbiddenCommandsSensor struct{}

func (ForbiddenCommandsSensor) ID() string                     { return "forbidden_commands" }
func (ForbiddenCommandsSensor) Category() Category             { return CatForbidden }
func (ForbiddenCommandsSensor) Kind() Kind                     { return KindComputational }
func (ForbiddenCommandsSensor) AppliesTo(p index.Profile) bool { return true }

var scanCommandExts = map[string]bool{
	".sh":   true,
	".bash": true,
	".zsh":  true,
	".mk":   true,
	".yml":  true,
	".yaml": true,
}

var scanCommandNames = map[string]bool{
	"Makefile":   true,
	"Dockerfile": true,
}

func (s ForbiddenCommandsSensor) Run(rc RunCtx) Result {
	start := time.Now()
	res := Result{ID: s.ID(), Category: s.Category(), Kind: s.Kind()}
	ig, _ := ignore.Load(rc.Root)
	var hits []string
	_ = filepath.WalkDir(rc.Root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		rel, _ := filepath.Rel(rc.Root, p)
		if d.IsDir() {
			if scanExcludedDirs[d.Name()] || ig.Match(rel, true) {
				return filepath.SkipDir
			}
			return nil
		}
		if ig.Match(rel, false) {
			return nil
		}
		base := d.Name()
		if !scanCommandNames[base] && !scanCommandExts[filepath.Ext(base)] {
			return nil
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return nil
		}
		for _, m := range forbiddenCommandRe.FindAll(b, -1) {
			hits = append(hits, rel+": "+string(m))
		}
		return nil
	})
	sort.Strings(hits)
	res.Duration = time.Since(start)
	if len(hits) == 0 {
		res.Status = StatusPassed
		return res
	}
	res.Status = StatusFailed
	res.Detail = "forbidden command patterns detected"
	res.OutputPath = writeOutput(rc.OutputDir, s.ID(), []byte(strings.Join(hits, "\n")+"\n"), nil)
	return res
}

// SecretsScanSensor walks text files looking for high-confidence secret
// shapes. Conservative on purpose — a hit fails the sensor and forces a
// human review.
type SecretsScanSensor struct{}

func (SecretsScanSensor) ID() string                     { return "secrets_scan" }
func (SecretsScanSensor) Category() Category             { return CatSecrets }
func (SecretsScanSensor) Kind() Kind                     { return KindComputational }
func (SecretsScanSensor) AppliesTo(p index.Profile) bool { return true }

var scanSecretExtIgnored = map[string]bool{
	".png":   true,
	".jpg":   true,
	".jpeg":  true,
	".gif":   true,
	".webp":  true,
	".pdf":   true,
	".zip":   true,
	".gz":    true,
	".bz2":   true,
	".tgz":   true,
	".xz":    true,
	".woff":  true,
	".woff2": true,
	".ttf":   true,
	".ico":   true,
	".mp4":   true,
}

func (s SecretsScanSensor) Run(rc RunCtx) Result {
	start := time.Now()
	res := Result{ID: s.ID(), Category: s.Category(), Kind: s.Kind()}
	ig, _ := ignore.Load(rc.Root)
	var hits []string
	_ = filepath.WalkDir(rc.Root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		rel, _ := filepath.Rel(rc.Root, p)
		if d.IsDir() {
			if scanExcludedDirs[d.Name()] || ig.Match(rel, true) {
				return filepath.SkipDir
			}
			return nil
		}
		if ig.Match(rel, false) {
			return nil
		}
		name := d.Name()
		if strings.HasSuffix(name, "_test.go") ||
			strings.HasSuffix(name, ".test.ts") || strings.HasSuffix(name, ".test.tsx") ||
			strings.HasSuffix(name, "_spec.rb") {
			return nil
		}
		if scanSecretExtIgnored[strings.ToLower(filepath.Ext(d.Name()))] {
			return nil
		}
		info, err := d.Info()
		if err != nil || info.Size() > 2*1024*1024 {
			return nil
		}
		f, err := os.Open(p)
		if err != nil {
			return nil
		}
		defer f.Close()
		scanner := bufio.NewScanner(f)
		scanner.Buffer(make([]byte, 1024*1024), 4*1024*1024)
		for lineNo := 1; scanner.Scan(); lineNo++ {
			line := scanner.Bytes()
			for _, pat := range secretPatterns {
				if pat.Match(line) {
					hits = append(hits, rel+":"+itoa(lineNo)+": match "+pat.String())
					break
				}
			}
		}
		return nil
	})
	sort.Strings(hits)
	res.Duration = time.Since(start)
	if len(hits) == 0 {
		res.Status = StatusPassed
		return res
	}
	res.Status = StatusFailed
	res.Detail = "potential secrets detected"
	res.OutputPath = writeOutput(rc.OutputDir, s.ID(), []byte(strings.Join(hits, "\n")+"\n"), nil)
	return res
}

// ChangedFilesSensor records `git diff --name-only HEAD` for downstream
// sensors and reports. Always passes (informational).
type ChangedFilesSensor struct{}

func (ChangedFilesSensor) ID() string                     { return "changed_files" }
func (ChangedFilesSensor) Category() Category             { return CatChangedFile }
func (ChangedFilesSensor) Kind() Kind                     { return KindComputational }
func (ChangedFilesSensor) AppliesTo(p index.Profile) bool { return true }

func (s ChangedFilesSensor) Run(rc RunCtx) Result {
	start := time.Now()
	res := Result{ID: s.ID(), Category: s.Category(), Kind: s.Kind(), Status: StatusPassed}
	out := runGitCapture(rc.Root, "diff", "--name-only", "HEAD")
	res.OutputPath = writeOutput(rc.OutputDir, s.ID(), []byte(out), nil)
	if out = strings.TrimSpace(out); out == "" {
		res.Detail = "no changes vs HEAD"
	} else {
		res.Detail = "changes detected"
	}
	res.Duration = time.Since(start)
	return res
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}
