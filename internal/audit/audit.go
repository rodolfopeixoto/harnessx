// SPDX-License-Identifier: MIT

package audit

import (
	"context"
	"encoding/json"
	"os"
	"sort"
	"sync"
	"time"
)

type Event struct {
	ID         string            `json:"id"`
	Kind       string            `json:"kind"`
	Source     string            `json:"source"`
	Project    string            `json:"project,omitempty"`
	Subject    string            `json:"subject"`
	Metadata   map[string]string `json:"metadata,omitempty"`
	OccurredAt time.Time         `json:"occurred_at"`
}

// UnmarshalJSON tolerates the two on-disk shapes harness uses today:
// (1) the native audit schema (kind/source/subject/occurred_at) written by
// FileSink.Write, and (2) the run-event log written by internal/eventlog
// where the timestamp lives in `ts`, the category in `stage`, and the
// subject is synthesised from `sensor`+`status`. Without this fallback
// `harness audit tail` rendered every row as `01-01 00:00:00` because the
// run-event log lines decoded into a zero-value OccurredAt.
func (e *Event) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	get := func(keys ...string) string {
		for _, k := range keys {
			if v, ok := raw[k]; ok {
				var s string
				if err := json.Unmarshal(v, &s); err == nil && s != "" {
					return s
				}
			}
		}
		return ""
	}
	e.ID = get("id", "run_id", "session_id")
	e.Kind = get("kind", "stage", "level")
	e.Source = get("source", "sensor", "agent")
	e.Project = get("project", "root")
	e.Subject = get("subject")
	if e.Subject == "" {
		if sensor := get("sensor"); sensor != "" {
			if status := get("status"); status != "" {
				e.Subject = sensor + "=" + status
			} else {
				e.Subject = sensor
			}
		}
	}
	if v, ok := raw["metadata"]; ok {
		_ = json.Unmarshal(v, &e.Metadata)
	}
	ts := get("occurred_at", "ts", "timestamp", "time")
	if ts != "" {
		if t, err := time.Parse(time.RFC3339Nano, ts); err == nil {
			e.OccurredAt = t
		} else if t, err := time.Parse(time.RFC3339, ts); err == nil {
			e.OccurredAt = t
		}
	}
	return nil
}

type Sink interface {
	Write(context.Context, Event) error
	List(context.Context) ([]Event, error)
}

type MemorySink struct {
	mu     sync.RWMutex
	events []Event
}

func NewMemorySink() *MemorySink { return &MemorySink{} }

func (s *MemorySink) Write(_ context.Context, ev Event) error {
	if ev.OccurredAt.IsZero() {
		ev.OccurredAt = time.Now().UTC()
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.events = append(s.events, ev)
	return nil
}

func (s *MemorySink) List(_ context.Context) ([]Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Event, len(s.events))
	copy(out, s.events)
	sort.SliceStable(out, func(i, j int) bool { return out[i].OccurredAt.After(out[j].OccurredAt) })
	return out, nil
}

type FileSink struct {
	Path string
	mu   sync.Mutex
}

func (f *FileSink) Write(_ context.Context, ev Event) error {
	if ev.OccurredAt.IsZero() {
		ev.OccurredAt = time.Now().UTC()
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	fh, err := os.OpenFile(f.Path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer fh.Close()
	enc := json.NewEncoder(fh)
	return enc.Encode(ev)
}

func (f *FileSink) List(_ context.Context) ([]Event, error) {
	b, err := os.ReadFile(f.Path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var out []Event
	dec := json.NewDecoder(jsonlReader(b))
	for dec.More() {
		var ev Event
		if err := dec.Decode(&ev); err != nil {
			return nil, err
		}
		out = append(out, ev)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].OccurredAt.After(out[j].OccurredAt) })
	return out, nil
}

func jsonlReader(b []byte) *jsonlBuffer { return &jsonlBuffer{b: b} }

type jsonlBuffer struct {
	b []byte
	i int
}

func (j *jsonlBuffer) Read(p []byte) (int, error) {
	if j.i >= len(j.b) {
		return 0, errEOF
	}
	n := copy(p, j.b[j.i:])
	j.i += n
	return n, nil
}

var errEOF = &eof{}

type eof struct{}

func (*eof) Error() string { return "EOF" }
