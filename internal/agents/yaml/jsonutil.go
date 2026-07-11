// SPDX-License-Identifier: MIT

package yaml

import (
	"bytes"
	"encoding/json"
	"strings"
)

// extractJSONPath implements a deliberately small subset of JSONPath used
// by adapter outputs: dotted paths starting with $., e.g. "$.usage.input_tokens".
// Returns the value's JSON encoding (string-encoded JSON) so callers can
// re-unmarshal as needed.
func extractJSONPath(data []byte, format, path string) string {
	if len(data) == 0 || path == "" {
		return ""
	}
	path = strings.TrimPrefix(path, "$.")
	parts := strings.Split(path, ".")

	// For JSONL output, scan lines and try each as a JSON object until one
	// resolves the path. Last writer wins (mimics the final-message convention).
	if format == "jsonl" {
		var last string
		for _, line := range bytes.Split(data, []byte("\n")) {
			if v := walkJSON(line, parts); v != "" {
				last = v
			}
		}
		return last
	}
	return walkJSON(data, parts)
}

func walkJSON(b []byte, parts []string) string {
	if len(b) == 0 {
		return ""
	}
	var cur any
	if err := json.Unmarshal(b, &cur); err != nil {
		return ""
	}
	for _, p := range parts {
		m, ok := cur.(map[string]any)
		if !ok {
			return ""
		}
		cur, ok = m[p]
		if !ok {
			return ""
		}
	}
	if s, ok := cur.(string); ok {
		return s
	}
	out, err := json.Marshal(cur)
	if err != nil {
		return ""
	}
	return string(out)
}

func readInt(m map[string]any, keys ...string) int {
	for _, k := range keys {
		if v, ok := m[k]; ok {
			switch n := v.(type) {
			case float64:
				return int(n)
			case int:
				return n
			case int64:
				return int(n)
			}
		}
	}
	return 0
}

func trimLine(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		s = s[:i]
	}
	return s
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
