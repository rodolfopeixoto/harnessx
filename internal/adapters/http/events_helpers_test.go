// SPDX-License-Identifier: MIT

package http

import (
	"bytes"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestWriteSSEFormatsEvent(t *testing.T) {
	w := httptest.NewRecorder()
	writeSSE(w, "open", []byte("hello"))
	body := w.Body.String()
	if !bytes.Contains([]byte(body), []byte("event: open")) {
		t.Errorf("missing event line: %q", body)
	}
	if !bytes.Contains([]byte(body), []byte("data: hello")) {
		t.Errorf("missing data line: %q", body)
	}
}

func TestEnsureExistsReturnsErrWhenAbsent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "missing.jsonl")
	if err := ensureExists(path); err == nil {
		t.Error("expected error for absent file (function only polls, does not create)")
	}
}

func TestEnsureExistsReturnsNilWhenPresent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "existing.jsonl")
	if err := os.WriteFile(path, []byte("ok"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := ensureExists(path); err != nil {
		t.Errorf("present file should not error: %v", err)
	}
}
