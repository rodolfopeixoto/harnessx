// SPDX-License-Identifier: MIT

package limits

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestForAdapterReturnsKnownLimits(t *testing.T) {
	if got := ForAdapter("codex"); got.MaxSkillDescriptionChars != 1024 {
		t.Errorf("codex limit wrong: %+v", got)
	}
	if got := ForAdapter("claude"); got.MaxSkillDescriptionChars != 8192 {
		t.Errorf("claude limit wrong: %+v", got)
	}
	if got := ForAdapter("unknown"); got.MaxSkillDescriptionChars != 0 {
		t.Errorf("unknown adapter must return zero limits, got %+v", got)
	}
}

func writeSkill(t *testing.T, dir, name, desc string) {
	t.Helper()
	skillDir := filepath.Join(dir, name)
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	body := "name: " + name + "\ndescription: " + desc + "\n---\nbody here\n"
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestPrepareSkillsTruncatesOversizedDescription(t *testing.T) {
	root := t.TempDir()
	out := t.TempDir()
	writeSkill(t, root, "remotion-to-hyperframes", strings.Repeat("x", 4096))
	writeSkill(t, root, "short", "small one")

	_, reports, _, err := PrepareSkills("codex", []string{root}, out)
	if err != nil {
		t.Fatal(err)
	}
	if len(reports) != 1 || !reports[0].Truncated {
		t.Fatalf("want 1 truncated report, got %+v", reports)
	}
	body, _ := os.ReadFile(reports[0].SanitisedPath)
	if !strings.Contains(string(body), "…") {
		t.Errorf("truncated file must end description with ellipsis: %s", body)
	}
	if len(body) > 1024+200 {
		t.Errorf("sanitised file should be near the cap, got %d bytes", len(body))
	}
}

func TestPrepareSkillsLeavesShortDescriptionAlone(t *testing.T) {
	root := t.TempDir()
	out := t.TempDir()
	writeSkill(t, root, "tiny", "ok")
	_, reports, _, err := PrepareSkills("codex", []string{root}, out)
	if err != nil {
		t.Fatal(err)
	}
	if len(reports) != 0 {
		t.Errorf("short description must not be reported, got %+v", reports)
	}
}

func TestPrepareSkillsUnknownAdapterIsNoOp(t *testing.T) {
	dir, reports, skip, err := PrepareSkills("nonexistent", []string{t.TempDir()}, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if dir != "" || reports != nil || skip != nil {
		t.Errorf("unknown adapter must return empties, got %q %v %v", dir, reports, skip)
	}
}

func TestSkipEnvJoinsPaths(t *testing.T) {
	if got := SkipEnv([]string{"/a", "/b"}); got != "/a,/b" {
		t.Errorf("got %q", got)
	}
	if got := SkipEnv(nil); got != "" {
		t.Errorf("empty list must yield empty string, got %q", got)
	}
}

func TestWriteReportDeduplicatesPerSession(t *testing.T) {
	var buf bytes.Buffer
	reports := []SkillReport{{Path: "/x/SKILL.md", DescLen: 2048, Truncated: true}}
	WriteReport(&buf, "session-A", reports)
	once := buf.Len()
	WriteReport(&buf, "session-A", reports)
	if buf.Len() != once {
		t.Errorf("duplicate WARN must not write more bytes: first=%d second=%d", once, buf.Len())
	}
	WriteReport(&buf, "session-B", reports)
	if buf.Len() == once {
		t.Errorf("different session should warn again")
	}
}
