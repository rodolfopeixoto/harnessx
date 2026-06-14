package certify

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/ropeixoto/harnessx/internal/agents"
	"github.com/ropeixoto/harnessx/internal/agents/fake"
)

func TestRun_HealthyFake_Passes(t *testing.T) {
	a := fake.New("ok")
	a.FinalMessage = "OK"
	r := Run(context.Background(), a, Options{PerCheckTimeout: 500 * time.Millisecond})
	require.Equal(t, "passed", r.Status)
	require.GreaterOrEqual(t, r.Score, 80)
}

func TestRun_BrokenHealth_SkipsRunChecks(t *testing.T) {
	a := fake.New("bad")
	a.HealthOK = false
	a.HealthDetail = "binary missing"
	r := Run(context.Background(), a, Options{})
	require.Contains(t, []string{"partial", "failed"}, r.Status)
	skipped := 0
	for _, c := range r.Checks {
		if c.Status == StatusSkipped {
			skipped++
		}
	}
	require.GreaterOrEqual(t, skipped, 3)
}

func TestRun_DetailsJSON_RoundTrip(t *testing.T) {
	a := fake.New("ok")
	r := Run(context.Background(), a, Options{})
	j := r.DetailsJSON()
	require.Contains(t, j, "\"agent_id\":\"ok\"")
}

// Sanity check that ForceFailure surfaces through ClassifyFailure for the
// certification self-test.
func TestFakeForceFailure(t *testing.T) {
	a := fake.New("rate")
	a.ForceFailure = agents.FailureRateLimit
	out := agents.AgentOutput{}
	require.Equal(t, agents.FailureRateLimit, a.ClassifyFailure(out, 1, nil))
}
