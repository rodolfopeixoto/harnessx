// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ropeixoto/harnessx/internal/agents"
	"github.com/ropeixoto/harnessx/internal/app/agentcmd"
	"github.com/ropeixoto/harnessx/internal/platform/constants"
	"github.com/ropeixoto/harnessx/internal/router"
	"github.com/ropeixoto/harnessx/internal/ui"
)

func driveTestEmit(ctx context.Context, out io.Writer, opts driveOpts) (string, error) {
	driveHeader(out, "2/5", "test-emit", "— cheap chain writes failing tests")
	reg, _, err := agentcmd.LoadAll(opts.root)
	if err != nil {
		fmt.Fprintln(out, "  "+ui.MarkWarn()+" "+ui.Warn.Render("no adapter registry: "+err.Error()+" — falling back to placeholder"))
		return writePlaceholderTest(opts)
	}
	rtr := router.New(reg, router.Defaults(reg))
	dec, derr := rtr.Select(constants.DriveTaskTestEmit)
	if derr != nil || len(dec.Chain) == 0 {
		fmt.Fprintln(out, "  "+ui.MarkWarn()+" "+ui.Warn.Render("no cheap_review chain — placeholder test"))
		return writePlaceholderTest(opts)
	}
	adapter := dec.Chain[0]
	if wrapped, werr := wrapWithVCR(adapter, opts.vcrDir, opts.vcrMode); werr != nil {
		fmt.Fprintln(out, "  "+ui.MarkWarn()+" "+ui.Warn.Render("vcr disabled: "+werr.Error()))
	} else {
		adapter = wrapped
	}
	fmt.Fprintln(out, "  "+ui.Info.Render("routing through "+adapter.ID())+
		ui.Muted.Render(" ("+constants.DriveTaskTestEmit+")"))
	path, body, err := emitTestsViaAdapter(ctx, adapter, opts)
	if err != nil {
		fmt.Fprintln(out, "  "+ui.MarkWarn()+" "+ui.Warn.Render("adapter test-emit failed: "+err.Error()+" — placeholder"))
		return writePlaceholderTest(opts)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", err
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		return "", err
	}
	fmt.Fprintln(out, "  "+ui.MarkSuccess()+" "+ui.Muted.Render(fmt.Sprintf("wrote %d bytes via %s", len(body), adapter.ID())))
	return path, nil
}

func emitTestsViaAdapter(ctx context.Context, adapter agents.AgentAdapter, opts driveOpts) (string, string, error) {
	path := filepath.Join(opts.root, "tests", constants.DriveTestFilePrefix+opts.slug+constants.DriveTestFileSuffix)
	prompt := renderTestEmitPrompt(opts.prompt, opts.slug, opts.root)
	timeout := 2 * time.Minute
	rctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	res := adapter.Run(rctx, agents.AgentRequest{
		Prompt:     prompt,
		WorkingDir: opts.root,
		Timeout:    timeout,
		Extra:      map[string]string{"task": constants.DriveTaskTestEmit},
	})
	if res.Err != nil {
		return path, "", res.Err
	}
	body := extractPythonBody(res.Output.FinalMessage)
	if body == "" {
		body = extractPythonBody(string(res.Output.Stdout))
	}
	if body == "" {
		return path, "", fmt.Errorf("no python block found in adapter output")
	}
	if !strings.Contains(body, "def test_") {
		return path, "", fmt.Errorf("emitted file has no pytest test function")
	}
	return path, body, nil
}

func renderTestEmitPrompt(featurePrompt, slug, root string) string {
	return fmt.Sprintf(`You are HarnessX in spec-driven test-first mode.

A new feature is about to be implemented:

  %s

Write a pytest module to live at tests/test_drive_%s.py inside the
project at %s. The tests MUST fail today because the implementation
is not yet written. Cover the happy path plus one error case.

Constraints:
  - Output ONLY the Python file body inside a single
    triple-backtick code fence (with or without the "python" tag).
  - Do not include any prose outside the fence.
  - Do not edit any other file.
  - Use pytest function-style tests.
  - If the project uses FastAPI, assume a client fixture is already
    provided in tests/conftest.py (TestClient(app)).

Return only the file body.`, featurePrompt, slug, root)
}

func extractPythonBody(s string) string {
	if s == "" {
		return ""
	}
	for _, fence := range []string{"```python", "```py", "```"} {
		if idx := strings.Index(s, fence); idx >= 0 {
			rest := s[idx+len(fence):]
			if end := strings.Index(rest, "```"); end > 0 {
				body := strings.TrimSpace(rest[:end])
				body = strings.TrimPrefix(body, "\n")
				return body + "\n"
			}
		}
	}
	if strings.Contains(s, "def test_") {
		return strings.TrimSpace(s) + "\n"
	}
	return ""
}

func writePlaceholderTest(opts driveOpts) (string, error) {
	testDir := filepath.Join(opts.root, "tests")
	if err := os.MkdirAll(testDir, 0o755); err != nil {
		return "", err
	}
	path := filepath.Join(testDir, constants.DriveTestFilePrefix+opts.slug+constants.DriveTestFileSuffix)
	body := renderPlaceholderTest(opts.prompt, opts.slug)
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		return "", err
	}
	return path, nil
}

func renderPlaceholderTest(prompt, slug string) string {
	return fmt.Sprintf(`# harness drive — failing placeholder test for: %s
# Replaced once the implementation lands and the gate passes.

import pytest


def test_drive_placeholder_for_%s() -> None:
    pytest.fail("harness drive scaffold — implementation pending")
`, prompt, sanitisePyIdent(slug))
}

func sanitisePyIdent(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
		case r >= 'A' && r <= 'Z':
			b.WriteRune(r + 32)
		default:
			b.WriteRune('_')
		}
	}
	out := b.String()
	if out == "" || (out[0] >= '0' && out[0] <= '9') {
		out = "x_" + out
	}
	return out
}
