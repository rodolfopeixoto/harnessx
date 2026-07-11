// SPDX-License-Identifier: MIT

package skills

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestErrNoImprovementSentinel guards the exported error used by CLI
// callers (`harness skill promote`) to decide the exit code and by
// upstream callers (evolve.Promote) to distinguish "gate rejected"
// from "database exploded". Replacing the sentinel breaks both.
func TestErrNoImprovementSentinel(t *testing.T) {
	t.Parallel()
	require.NotNil(t, ErrNoImprovement)
	require.Equal(t,
		"skills: new version did not improve over previous best",
		ErrNoImprovement.Error(),
		"error text is documented in --help output; changing it breaks docs")

	// Callers use errors.Is — wrapping must preserve identity.
	wrapped := errors.Join(errors.New("promote failed"), ErrNoImprovement)
	require.True(t, errors.Is(wrapped, ErrNoImprovement),
		"wrapped errors must satisfy errors.Is against the sentinel")
}
