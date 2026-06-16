// SPDX-License-Identifier: MIT

package sharedstate

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const SchemaVersion = 1

type Task struct {
	Index       int            `json:"index"`
	Kind        string         `json:"kind"`
	Adapter     string         `json:"adapter,omitempty"`
	ReadSet     []string       `json:"read_set"`
	WriteSet    []string       `json:"write_set"`
	Assumptions map[string]int `json:"assumptions"`
	Version     int            `json:"version"`
	Status      string         `json:"status,omitempty"`
}

type Snapshot struct {
	SchemaVersion int    `json:"schema_version"`
	RunID         string `json:"run_id"`
	Tasks         []Task `json:"tasks"`
}

type Conflict struct {
	LaterIdx      int
	EarlierIdx    int
	OverlappingOn string
	Reason        string
}

func Detect(s Snapshot) []Conflict {
	var conflicts []Conflict
	for i, later := range s.Tasks {
		for j := 0; j < i; j++ {
			earlier := s.Tasks[j]
			for _, w := range earlier.WriteSet {
				if !contains(later.ReadSet, w) {
					continue
				}
				assumed, ok := later.Assumptions[w]
				if !ok || assumed < earlier.Version {
					conflicts = append(conflicts, Conflict{
						LaterIdx: later.Index, EarlierIdx: earlier.Index,
						OverlappingOn: w,
						Reason:        fmt.Sprintf("task %d reads %q (written by task %d at version %d) with stale assumption %d", later.Index, w, earlier.Index, earlier.Version, assumed),
					})
				}
			}
		}
	}
	return conflicts
}

func contains(haystack []string, needle string) bool {
	for _, h := range haystack {
		if h == needle {
			return true
		}
	}
	return false
}

func Path(root, runID string) string {
	return filepath.Join(root, ".harness", "runs", runID, "shared.json")
}

func Write(root string, s Snapshot) error {
	if s.RunID == "" {
		return fmt.Errorf("sharedstate: missing RunID")
	}
	s.SchemaVersion = SchemaVersion
	p := Path(root, s.RunID)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	body, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, body, 0o644)
}

func Read(root, runID string) (Snapshot, error) {
	body, err := os.ReadFile(Path(root, runID))
	if err != nil {
		return Snapshot{}, err
	}
	var s Snapshot
	if err := json.Unmarshal(body, &s); err != nil {
		return Snapshot{}, fmt.Errorf("sharedstate: parse: %w", err)
	}
	if s.SchemaVersion != SchemaVersion {
		return Snapshot{}, fmt.Errorf("sharedstate: schema_version=%d not supported (want %d)", s.SchemaVersion, SchemaVersion)
	}
	return s, nil
}
