// SPDX-License-Identifier: MIT

package http

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func (s *Server) registerEvents(mux *http.ServeMux) {
	mux.HandleFunc("/api/events/runs/", s.runEvents)
}

func (s *Server) runEvents(w http.ResponseWriter, r *http.Request) {
	runID := strings.TrimPrefix(r.URL.Path, "/api/events/runs/")
	runID = strings.TrimSuffix(runID, "/")
	if runID == "" {
		http.Error(w, "run id required", http.StatusBadRequest)
		return
	}
	path := filepath.Join(s.root, ".harness", "runs", runID, "events.jsonl")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	flusher.Flush()

	if err := streamRunEvents(r, w, flusher, path); err != nil {
		fmt.Fprintf(w, "event: error\ndata: %q\n\n", err.Error())
		flusher.Flush()
	}
}

func streamRunEvents(r *http.Request, w http.ResponseWriter, flusher http.Flusher, path string) error {
	if err := ensureExists(path); err != nil {
		return err
	}
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		writeSSE(w, "event", scanner.Bytes())
		flusher.Flush()
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	keepalive := time.NewTicker(15 * time.Second)
	defer keepalive.Stop()
	poll := time.NewTicker(500 * time.Millisecond)
	defer poll.Stop()
	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-keepalive.C:
			fmt.Fprint(w, ": keepalive\n\n")
			flusher.Flush()
		case <-poll.C:
			if !tailNew(f, w, flusher) {
				continue
			}
		}
	}
}

func tailNew(f *os.File, w http.ResponseWriter, flusher http.Flusher) bool {
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	produced := false
	for scanner.Scan() {
		writeSSE(w, "event", scanner.Bytes())
		flusher.Flush()
		produced = true
	}
	return produced
}

func writeSSE(w http.ResponseWriter, event string, data []byte) {
	fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, string(data))
}

func ensureExists(path string) error {
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(path); err == nil {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	if _, err := os.Stat(path); err != nil {
		return fmt.Errorf("events file not ready: %s", path)
	}
	return nil
}
