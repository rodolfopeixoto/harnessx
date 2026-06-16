// SPDX-License-Identifier: MIT

package autonomy

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/ropeixoto/harnessx/internal/platform/paths"
)

type Event struct {
	Path     string    `json:"path"`
	Decision string    `json:"decision"`
	Reason   string    `json:"reason,omitempty"`
	At       time.Time `json:"at"`
}

func approvalsPath(root string) string {
	return filepath.Join(paths.HarnessDir(root), "audit", "approvals.jsonl")
}

func AppendApproval(root string, e Event) error {
	if e.At.IsZero() {
		e.At = time.Now().UTC()
	}
	if e.Path == "" || e.Decision == "" {
		return fmt.Errorf("autonomy: Event requires Path + Decision")
	}
	p := approvalsPath(root)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(p, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	body, err := json.Marshal(e)
	if err != nil {
		return err
	}
	body = append(body, '\n')
	_, err = f.Write(body)
	return err
}

func ListApprovals(root string) ([]Event, error) {
	f, err := os.Open(approvalsPath(root))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()
	var out []Event
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var e Event
		if err := json.Unmarshal(line, &e); err != nil {
			continue
		}
		out = append(out, e)
	}
	return out, scanner.Err()
}
