// SPDX-License-Identifier: MIT

package input

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAssemble_PromptFileAndPositional(t *testing.T) {
	dir := t.TempDir()
	pf := filepath.Join(dir, "p.md")
	if err := os.WriteFile(pf, []byte("file header"), 0o644); err != nil {
		t.Fatal(err)
	}
	a, err := Assemble(Sources{PromptFile: pf, Positional: "task body"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(a.Prompt, "file header") {
		t.Fatalf("prompt: %q", a.Prompt)
	}
	if !strings.Contains(a.Prompt, "task body") {
		t.Fatalf("missing task body")
	}
}

func TestAssemble_EmptyError(t *testing.T) {
	_, err := Assemble(Sources{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestAssemble_ImageOnlyAttaches(t *testing.T) {
	dir := t.TempDir()
	img := filepath.Join(dir, "x.png")
	if err := os.WriteFile(img, []byte{0x89, 0x50, 0x4e, 0x47}, 0o644); err != nil {
		t.Fatal(err)
	}
	a, err := Assemble(Sources{Image: img, Positional: "look at this"})
	if err != nil {
		t.Fatal(err)
	}
	if len(a.Attachments) != 1 || a.Attachments[0].MimeType != "image/png" {
		t.Fatalf("attachments: %+v", a.Attachments)
	}
	if a.Attachments[0].Base64() == "" {
		t.Fatal("empty base64")
	}
}
