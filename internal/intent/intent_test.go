package intent

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ropeixoto/harnessx/internal/domain"
)

func TestClassify(t *testing.T) {
	cases := []struct {
		in   string
		want domain.Mode
	}{
		{"What does the auth middleware do?", domain.ModeQuestion},
		{"explain context engineering pipeline", domain.ModeQuestion},
		{"add product search with filters", domain.ModeFeature},
		{"implement OAuth2 callback handler", domain.ModeFeature},
		{"fix N+1 query in OrdersController#index", domain.ModeBugfix},
		{"use this Claude Design zip and convert to React parity", domain.ModeDesignToProduct},
		{"optimize Docker image size", domain.ModeOptimization},
		{"audit dependency tree", domain.ModeAudit},
		{"review the diff on this branch", domain.ModeReview},
		{"scaffold a new Go service", domain.ModeSetup},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			got := Classify(tc.in)
			require.Equal(t, tc.want, got.Mode, "reasons=%v", got.Reasons)
			require.Greater(t, got.Confidence, 0.0)
		})
	}
}

func TestClassify_NoRule_DefaultsToFeatureLowConfidence(t *testing.T) {
	got := Classify("xyzzy frob")
	require.Equal(t, domain.ModeFeature, got.Mode)
	require.Less(t, got.Confidence, 0.5)
}
