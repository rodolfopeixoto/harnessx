// SPDX-License-Identifier: MIT

package multimodal

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

const SchemaVersion = 1

type Region struct {
	X int `json:"x"`
	Y int `json:"y"`
	W int `json:"w"`
	H int `json:"h"`
}

type Annotation struct {
	Region Region `json:"region"`
	Label  string `json:"label"`
}

type Sidecar struct {
	SchemaVersion int          `json:"schema_version"`
	Annotations   []Annotation `json:"annotations"`
}

type GroundingResult struct {
	Hits    []string
	Missing []string
}

func LoadAnnotations(path string) ([]Annotation, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var s Sidecar
	if err := json.Unmarshal(body, &s); err != nil {
		return nil, fmt.Errorf("multimodal: parse %s: %w", path, err)
	}
	if s.SchemaVersion != SchemaVersion {
		return nil, fmt.Errorf("multimodal: schema_version=%d not supported (want %d)", s.SchemaVersion, SchemaVersion)
	}
	return s.Annotations, nil
}

func CheckGrounding(text string, anns []Annotation) GroundingResult {
	low := strings.ToLower(text)
	var res GroundingResult
	for _, a := range anns {
		label := strings.TrimSpace(strings.ToLower(a.Label))
		if label == "" {
			continue
		}
		if strings.Contains(low, label) {
			res.Hits = append(res.Hits, a.Label)
		} else {
			res.Missing = append(res.Missing, a.Label)
		}
	}
	return res
}
