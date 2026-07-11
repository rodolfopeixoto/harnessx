// SPDX-License-Identifier: MIT

package specflow

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// persist.go owns filesystem persistence + the $EDITOR shell-out. Split
// from specflow.go so the planning/refine flow stays focused on the LLM
// interaction.

// Save writes the spec markdown to `.harness/artifacts/specs/<id>.md`
// and the revision history alongside as `<id>.history.jsonl`. Returns
// the spec path.
func (s *Session) Save() (string, error) {
	if strings.TrimSpace(s.Body) == "" {
		return "", fmt.Errorf("specflow save: empty draft")
	}
	dir := filepath.Join(s.Root, ".harness", "artifacts", "specs")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	specPath := filepath.Join(dir, s.ID+".md")
	header := s.metadataHeader()
	if err := os.WriteFile(specPath, []byte(header+s.Body+"\n"), 0o644); err != nil {
		return "", err
	}
	histPath := filepath.Join(dir, s.ID+".history.jsonl")
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	for _, r := range s.Revisions {
		_ = enc.Encode(r)
	}
	if err := os.WriteFile(histPath, buf.Bytes(), 0o644); err != nil {
		return "", err
	}
	return specPath, nil
}

func (s *Session) metadataHeader() string {
	var b strings.Builder
	b.WriteString("<!--\n")
	b.WriteString("harness-spec-id: " + s.ID + "\n")
	if s.Mode != "" {
		b.WriteString("mode: " + s.Mode + "\n")
	}
	if s.Template != "" {
		b.WriteString("template: " + s.Template + "\n")
	}
	b.WriteString("prompt: " + strings.ReplaceAll(s.Prompt, "\n", " ") + "\n")
	if len(s.Answers) > 0 {
		b.WriteString("answers:\n")
		for _, a := range s.Answers {
			b.WriteString("  " + a.Key + ": " + strings.ReplaceAll(a.Value, "\n", " ") + "\n")
		}
	}
	b.WriteString("-->\n\n")
	return b.String()
}

// EditViaEditor writes the current draft to a temp file, shells out
// to $EDITOR (fallback $VISUAL, then `vi`), and re-reads the file.
// The returned body is the user-edited content; caller invokes
// ApplyEdit to record it on the session.
func EditViaEditor(initial string) (string, error) {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = os.Getenv("VISUAL")
	}
	if editor == "" {
		editor = "vi"
	}
	f, err := os.CreateTemp("", "harness-spec-*.md")
	if err != nil {
		return "", err
	}
	tmp := f.Name()
	if _, err := f.WriteString(initial); err != nil {
		_ = f.Close()
		return "", err
	}
	_ = f.Close()
	defer os.Remove(tmp)
	if err := runEditor(editor, tmp); err != nil {
		return "", err
	}
	body, err := os.ReadFile(tmp)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

// runEditor is split into its own var so tests can stub it without
// spawning a real editor process.
var runEditor = defaultRunEditor
