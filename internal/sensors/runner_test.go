package sensors

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/ropeixoto/harnessx/internal/index"
)

type fakeSensor struct {
	id     string
	kind   Kind
	status Status
	cat    Category
}

func (f fakeSensor) ID() string                     { return f.id }
func (f fakeSensor) Category() Category             { return f.cat }
func (f fakeSensor) Kind() Kind                     { return f.kind }
func (f fakeSensor) AppliesTo(p index.Profile) bool { return true }
func (f fakeSensor) Run(rc RunCtx) Result {
	return Result{ID: f.id, Status: f.status, Category: f.cat, Kind: f.kind, Duration: time.Millisecond}
}

func TestRunner_OrderComputationalFirst(t *testing.T) {
	in := []Sensor{
		fakeSensor{id: "b_infer", kind: KindInferential, status: StatusPassed},
		fakeSensor{id: "a_comp", kind: KindComputational, status: StatusPassed},
		fakeSensor{id: "c_comp", kind: KindComputational, status: StatusFailed},
	}
	var got []string
	r := &Runner{OnResult: func(res Result) { got = append(got, res.ID) }}
	results := r.Run(context.Background(), in, RunCtx{})
	require.Equal(t, []string{"a_comp", "c_comp", "b_infer"}, got)
	s := Summarize(results)
	require.Equal(t, 3, s.Total)
	require.Equal(t, 2, s.Passed)
	require.Equal(t, 1, s.Failed)
}

func TestCatalog_AlwaysIncludesUniversal(t *testing.T) {
	cs := Catalog(index.Profile{Stacks: nil})
	ids := map[string]bool{}
	for _, s := range cs {
		ids[s.ID()] = true
	}
	for _, want := range []string{"forbidden_files", "forbidden_commands", "secrets_scan", "changed_files"} {
		require.Truef(t, ids[want], "missing universal %s", want)
	}
}

func TestCatalog_GoStackAddsGoSensors(t *testing.T) {
	cs := Catalog(index.Profile{Stacks: []index.Stack{{Name: "go"}}})
	have := map[string]bool{}
	for _, s := range cs {
		have[s.ID()] = true
	}
	require.True(t, have["go_vet"])
	require.True(t, have["go_test"])
}
