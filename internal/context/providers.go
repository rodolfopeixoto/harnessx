// SPDX-License-Identifier: MIT

package context

import (
	"context"
	"os/exec"
)

// Provider extends a Pack with information from one source. Providers must
// be idempotent and side-effect-free with respect to the project (read-only).
type Provider interface {
	Name() string
	// Apply enriches the pack in place. Implementations should set fields
	// they own (e.g. GitProvider sets GitStatus/GitDiff) and leave others
	// untouched. Return non-nil error only for unexpected I/O failures —
	// missing tools should be a no-op + Stats.ProvidersSkipped++.
	Apply(ctx context.Context, root string, pack *Pack) error
}

// hasBinary is a tiny helper to keep provider files clean.
func hasBinary(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}
