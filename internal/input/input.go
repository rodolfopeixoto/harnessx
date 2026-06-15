// SPDX-License-Identifier: MIT

// Package input assembles the final prompt string from the various
// user-supplied sources: positional argument, --prompt-file, --pdf, and
// --image. Text inputs are concatenated deterministically (no LLM).
// PDFs require `pdftotext` (poppler) on PATH; images are base64-encoded
// for adapters that opt into vision (capabilities.vision = true).
package input

import (
	"encoding/base64"
	"errors"
	"fmt"
	"mime"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Attachment struct {
	Path     string `json:"path"`
	MimeType string `json:"mime_type"`
	Bytes    int    `json:"bytes"`
	Data     []byte `json:"-"`
}

type Sources struct {
	Positional string
	PromptFile string
	PDF        string
	Image      string
}

type Assembled struct {
	Prompt      string
	Attachments []Attachment
	Notes       []string
}

// Assemble validates and merges the sources. Order in the final prompt:
//
//  1. PromptFile contents (if provided)
//  2. PDF text extracted via pdftotext (if provided)
//  3. Positional argument
//
// Image attachment is returned in Attachments; callers must check
// adapter.Capabilities().Vision before passing it on.
func Assemble(s Sources) (Assembled, error) {
	var out Assembled
	var parts []string

	if s.PromptFile != "" {
		data, err := os.ReadFile(s.PromptFile)
		if err != nil {
			return out, fmt.Errorf("input: read prompt file: %w", err)
		}
		parts = append(parts, strings.TrimSpace(string(data)))
		out.Notes = append(out.Notes, "prompt_file:"+s.PromptFile)
	}

	if s.PDF != "" {
		text, err := extractPDF(s.PDF)
		if err != nil {
			return out, err
		}
		parts = append(parts, "## PDF: "+filepath.Base(s.PDF)+"\n"+strings.TrimSpace(text))
		out.Notes = append(out.Notes, "pdf:"+s.PDF)
	}

	if s.Positional != "" {
		parts = append(parts, strings.TrimSpace(s.Positional))
	}

	if len(parts) == 0 && s.Image == "" {
		return out, errors.New("input: empty prompt and no attachments")
	}

	out.Prompt = strings.Join(parts, "\n\n")

	if s.Image != "" {
		att, err := loadImage(s.Image)
		if err != nil {
			return out, err
		}
		out.Attachments = append(out.Attachments, att)
		out.Notes = append(out.Notes, "image:"+s.Image)
	}

	return out, nil
}

func extractPDF(path string) (string, error) {
	bin, err := exec.LookPath("pdftotext")
	if err != nil {
		return "", fmt.Errorf("input: pdftotext not on PATH (install poppler-utils): %w", err)
	}
	cmd := exec.Command(bin, path, "-")
	stdout, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("input: pdftotext failed: %w", err)
	}
	return string(stdout), nil
}

func loadImage(path string) (Attachment, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Attachment{}, fmt.Errorf("input: read image: %w", err)
	}
	mt := mime.TypeByExtension(strings.ToLower(filepath.Ext(path)))
	if mt == "" {
		mt = "application/octet-stream"
	}
	return Attachment{Path: path, MimeType: mt, Bytes: len(data), Data: data}, nil
}

// Base64 helper for adapters that inject images as base64 strings.
func (a Attachment) Base64() string {
	return base64.StdEncoding.EncodeToString(a.Data)
}
