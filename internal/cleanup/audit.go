// SPDX-License-Identifier: MIT

package cleanup

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"time"
)

type Outcome struct {
	Path        string
	SizeBytes   int64
	ContentHash string
	AppliedAt   time.Time
}

type AuditEvent struct {
	Kind        string
	Path        string
	SizeBytes   int64
	ContentHash string
	Reason      string
	PolicyPath  string
	AppliedAt   time.Time
}

type AuditSink interface {
	Write(AuditEvent) error
}

type NoopAudit struct{}

func (NoopAudit) Write(AuditEvent) error { return nil }

var nowFn = time.Now

func hashPath(path string) (string, int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", 0, err
	}
	if info.IsDir() {
		return "", 0, nil
	}
	f, err := os.Open(path)
	if err != nil {
		return "", 0, err
	}
	defer f.Close()
	h := sha256.New()
	n, err := io.Copy(h, f)
	if err != nil {
		return "", 0, err
	}
	return hex.EncodeToString(h.Sum(nil)), n, nil
}

var ErrAcknowledgementMissing = errors.New("cleanup: HARNESS_CLEANUP_I_UNDERSTAND=1 required for non-interactive apply")
