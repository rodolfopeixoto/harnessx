// SPDX-License-Identifier: MIT

package repl

import (
	"bytes"
	"strings"
	"testing"
)

func TestSlashSuggesterMatchesByPrefix(t *testing.T) {
	s := newSlashSuggester(&bytes.Buffer{})
	got := s.matches("/dr")
	wantPresent := []string{"/drive"}
	for _, w := range wantPresent {
		found := false
		for _, m := range got {
			if m == w {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected %q in matches for /dr: %v", w, got)
		}
	}
}

func TestSlashSuggesterEmptyForNonSlash(t *testing.T) {
	s := newSlashSuggester(&bytes.Buffer{})
	if got := s.matches("hello"); len(got) != 0 {
		t.Errorf("non-slash should yield 0 matches, got %v", got)
	}
}

func TestSlashSuggesterEmptyForExactMatch(t *testing.T) {
	s := newSlashSuggester(&bytes.Buffer{})
	if got := s.matches("/drive"); len(got) != 0 {
		t.Errorf("exact slash match should suppress popup, got %v", got)
	}
}

func TestSlashSuggesterIgnoresArgsTail(t *testing.T) {
	s := newSlashSuggester(&bytes.Buffer{})
	got := s.matches("/drive add stock")
	for _, m := range got {
		if m == "/drive" {
			return
		}
	}
	if len(got) > 0 {
		t.Errorf("expected /drive in matches for prefix mode: %v", got)
	}
}

func TestSlashSuggesterRendersAndClears(t *testing.T) {
	var buf bytes.Buffer
	s := newSlashSuggester(&buf)
	s.render("/dr")
	first := buf.Len()
	if first == 0 {
		t.Fatal("expected render bytes")
	}
	if !strings.Contains(buf.String(), "/drive") {
		t.Errorf("expected /drive in render output")
	}
	if !strings.Contains(buf.String(), "TAB to complete") {
		t.Errorf("expected TAB hint")
	}
	s.render("hello")
	if s.lastShown != 0 {
		t.Errorf("expected popup cleared after non-slash, lastShown=%d", s.lastShown)
	}
}

func TestSlashSuggesterClearIsIdempotent(t *testing.T) {
	var buf bytes.Buffer
	s := newSlashSuggester(&buf)
	s.clear()
	if buf.Len() != 0 {
		t.Errorf("clear with nothing shown should be no-op")
	}
}

func TestSlashSuggesterNilWriterSafe(t *testing.T) {
	s := newSlashSuggester(nil)
	s.render("/dr")
}

func TestSlashSuggesterMatchesCapped(t *testing.T) {
	s := newSlashSuggester(&bytes.Buffer{})
	got := s.matches("/")
	if len(got) > slashPopupMaxRows {
		t.Errorf("matches exceeded cap: got %d > %d", len(got), slashPopupMaxRows)
	}
}
