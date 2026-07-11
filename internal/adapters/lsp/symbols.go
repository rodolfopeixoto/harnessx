// SPDX-License-Identifier: MIT

package lsp

import (
	"encoding/json"
	"net/url"
	"path/filepath"
)

func parseDocumentSymbols(raw json.RawMessage, path string) []Symbol {
	var probe []map[string]json.RawMessage
	if err := json.Unmarshal(raw, &probe); err != nil || len(probe) == 0 {
		return nil
	}
	_, isFlat := probe[0]["location"]
	if isFlat {
		var flat []struct {
			Name     string `json:"name"`
			Location struct {
				Range struct {
					Start struct{ Line int } `json:"start"`
				} `json:"range"`
			} `json:"location"`
		}
		if err := json.Unmarshal(raw, &flat); err != nil {
			return nil
		}
		out := make([]Symbol, 0, len(flat))
		for _, s := range flat {
			out = append(out, Symbol{Name: s.Name, Path: path, Line: s.Location.Range.Start.Line + 1})
		}
		return out
	}
	var hier []struct {
		Name  string `json:"name"`
		Range struct {
			Start struct{ Line int } `json:"start"`
		} `json:"range"`
		Children []struct {
			Name  string `json:"name"`
			Range struct {
				Start struct{ Line int } `json:"start"`
			} `json:"range"`
		} `json:"children"`
	}
	if err := json.Unmarshal(raw, &hier); err != nil {
		return nil
	}
	out := make([]Symbol, 0, len(hier))
	for _, s := range hier {
		out = append(out, Symbol{Name: s.Name, Path: path, Line: s.Range.Start.Line + 1})
		for _, c := range s.Children {
			out = append(out, Symbol{Name: s.Name + "." + c.Name, Path: path, Line: c.Range.Start.Line + 1})
		}
	}
	return out
}

func severityName(s int) string {
	switch s {
	case 1:
		return "error"
	case 2:
		return "warning"
	case 3:
		return "information"
	case 4:
		return "hint"
	}
	return "info"
}

func pathToURI(p string) string {
	abs, _ := filepath.Abs(p)
	u := url.URL{Scheme: "file", Path: filepath.ToSlash(abs)}
	return u.String()
}

func uriToPath(u string) string {
	parsed, err := url.Parse(u)
	if err != nil {
		return u
	}
	return parsed.Path
}
