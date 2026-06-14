// SPDX-License-Identifier: MIT

package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// JSONL writes one JSON object per line, rotating when the file exceeds
// MaxBytes. Safe for concurrent use.
type JSONL struct {
	mu       sync.Mutex
	path     string
	maxBytes int64
	f        *os.File
}

func Open(path string, maxBytes int64) (*JSONL, error) {
	if maxBytes <= 0 {
		maxBytes = 10 * 1024 * 1024
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("logger: mkdir: %w", err)
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, fmt.Errorf("logger: open: %w", err)
	}
	return &JSONL{path: path, maxBytes: maxBytes, f: f}, nil
}

func (l *JSONL) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.f == nil {
		return nil
	}
	err := l.f.Close()
	l.f = nil
	return err
}

// Write serialises fields plus standard envelope keys (ts, level) as one
// JSON line.
func (l *JSONL) Write(level string, fields map[string]any) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.f == nil {
		return fmt.Errorf("logger: closed")
	}
	if err := l.maybeRotateLocked(); err != nil {
		return err
	}
	envelope := make(map[string]any, len(fields)+2)
	envelope["ts"] = time.Now().UTC().Format(time.RFC3339Nano)
	envelope["level"] = level
	for k, v := range fields {
		envelope[k] = v
	}
	b, err := json.Marshal(envelope)
	if err != nil {
		return fmt.Errorf("logger: marshal: %w", err)
	}
	b = append(b, '\n')
	_, err = l.f.Write(b)
	return err
}

func (l *JSONL) maybeRotateLocked() error {
	info, err := l.f.Stat()
	if err != nil {
		return err
	}
	if info.Size() < l.maxBytes {
		return nil
	}
	if err := l.f.Close(); err != nil {
		return err
	}
	rotated := l.path + "." + time.Now().UTC().Format("20060102T150405") + ".jsonl"
	if err := os.Rename(l.path, rotated); err != nil {
		return err
	}
	f, err := os.OpenFile(l.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	l.f = f
	return nil
}
