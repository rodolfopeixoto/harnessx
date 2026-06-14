// SPDX-License-Identifier: MIT

// Package i18n provides a tiny message-bundle system so user-facing
// strings can be translated by the community without touching Go code.
//
// Design:
//   - English (`en`) is the canonical bundle, loaded at init via embed.FS.
//     New messages MUST be added to en first; missing translations fall
//     back to en silently.
//   - Other locales live alongside as JSON files under `locales/<lang>.json`.
//   - Lookup: i18n.T("agent.list.header")  → "ID  NAME  CERT  SOURCE"
//   - Locale resolution order: $HARNESS_LANG → $LANG (first two chars) → en.
//
// Plugin model: third-party translations land via PR adding a new JSON
// file under internal/platform/i18n/locales/. No code changes required.
package i18n

import (
	"embed"
	"encoding/json"
	"os"
	"strings"
	"sync"
)

//go:embed locales/*.json
var bundleFS embed.FS

type bundle map[string]string

var (
	mu       sync.RWMutex
	current  string = "en"
	loaded          = map[string]bundle{}
	fallback        = "en"
)

func init() {
	_ = loadAll()
	if l := detectLocale(); l != "" {
		_ = SetLocale(l)
	}
}

func loadAll() error {
	entries, err := bundleFS.ReadDir("locales")
	if err != nil {
		return err
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		b, err := bundleFS.ReadFile("locales/" + e.Name())
		if err != nil {
			continue
		}
		var m bundle
		if err := json.Unmarshal(b, &m); err != nil {
			continue
		}
		lang := strings.TrimSuffix(e.Name(), ".json")
		loaded[lang] = m
	}
	return nil
}

func detectLocale() string {
	if v := os.Getenv("HARNESS_LANG"); v != "" {
		return normalise(v)
	}
	if v := os.Getenv("LANG"); v != "" {
		return normalise(v)
	}
	return ""
}

func normalise(s string) string {
	s = strings.ToLower(s)
	if i := strings.IndexAny(s, "_.@"); i > 0 {
		s = s[:i]
	}
	return s
}

// SetLocale switches the active locale; returns false when the bundle is
// not loaded. Callers can still use T and get fallback English.
func SetLocale(lang string) bool {
	mu.Lock()
	defer mu.Unlock()
	if _, ok := loaded[lang]; !ok {
		return false
	}
	current = lang
	return true
}

// Locale returns the active locale id.
func Locale() string {
	mu.RLock()
	defer mu.RUnlock()
	return current
}

// AvailableLocales returns the loaded locale ids in stable order.
func AvailableLocales() []string {
	mu.RLock()
	defer mu.RUnlock()
	out := make([]string, 0, len(loaded))
	for k := range loaded {
		out = append(out, k)
	}
	return out
}

// T returns the translated string for key. Missing key in the active
// locale falls back to en; missing in en returns the key itself so the
// bug is obvious in the UI.
func T(key string) string {
	mu.RLock()
	defer mu.RUnlock()
	if b, ok := loaded[current]; ok {
		if v, ok := b[key]; ok {
			return v
		}
	}
	if b, ok := loaded[fallback]; ok {
		if v, ok := b[key]; ok {
			return v
		}
	}
	return key
}
