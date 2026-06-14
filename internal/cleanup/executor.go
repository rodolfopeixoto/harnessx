// SPDX-License-Identifier: MIT

package cleanup

import (
	"context"
	"fmt"
	"os"

	"github.com/ropeixoto/harnessx/internal/platform/constants"
)

type Executor struct {
	Policy          Policy
	Audit           AuditSink
	Interactive     Approver
	Acknowledgement string
}

type Approver func(Finding) (bool, error)

func NewExecutor(policy Policy, audit AuditSink) *Executor {
	if audit == nil {
		audit = NoopAudit{}
	}
	return &Executor{Policy: policy, Audit: audit, Acknowledgement: os.Getenv(constants.EnvCleanupAcknowledgement)}
}

func (e *Executor) Apply(ctx context.Context, f Finding) (Outcome, error) {
	if err := e.gate(f); err != nil {
		return Outcome{}, err
	}
	hash, size, err := hashPath(f.Path)
	if err != nil && !os.IsNotExist(err) {
		return Outcome{}, fmt.Errorf("cleanup: stat %s: %w", f.Path, err)
	}
	if err := os.RemoveAll(f.Path); err != nil {
		return Outcome{}, fmt.Errorf("cleanup: remove %s: %w", f.Path, err)
	}
	outcome := Outcome{
		Path:        f.Path,
		SizeBytes:   size,
		ContentHash: hash,
		AppliedAt:   nowFn(),
	}
	_ = e.Audit.Write(AuditEvent{
		Kind:        f.Kind,
		Path:        f.Path,
		SizeBytes:   size,
		ContentHash: hash,
		Reason:      f.Reason,
		PolicyPath:  e.Policy.Source,
		AppliedAt:   outcome.AppliedAt,
	})
	_ = ctx
	return outcome, nil
}

func (e *Executor) gate(f Finding) error {
	if _, matched := e.Policy.Match(f); matched {
		if e.requireAck() {
			return nil
		}
		return ErrAcknowledgementMissing
	}
	if e.Interactive != nil {
		ok, err := e.Interactive(f)
		if err != nil {
			return err
		}
		if !ok {
			return ErrUserDenied
		}
		return nil
	}
	return ErrPolicyMissing
}

func (e *Executor) requireAck() bool {
	if !e.Policy.Globals.RequireAcknowledgement {
		return true
	}
	return e.Acknowledgement == "1"
}

var ErrUserDenied = fmt.Errorf("cleanup: approval declined by operator")
