// SPDX-License-Identifier: MIT

package containers

import (
	"context"
	"testing"
)

func TestRuntimeIDsAndBinaries(t *testing.T) {
	cases := []struct {
		rt          Runtime
		wantID, bin string
	}{
		{Docker{}, "docker", "docker"},
		{Podman{}, "podman", "podman"},
		{OrbStack{}, "orbstack", "orbctl"},
		{Colima{}, "colima", "colima"},
		{AppleContainer{}, "apple_container", "container"},
	}
	for _, c := range cases {
		if got := c.rt.ID(); got != c.wantID {
			t.Errorf("ID=%q want %q", got, c.wantID)
		}
		if got := c.rt.Binary(); got != c.bin {
			t.Errorf("Binary=%q want %q", got, c.bin)
		}
	}
}

func TestDetectAndByID(t *testing.T) {
	ctx := context.Background()
	rts := Detect(ctx)
	for _, rt := range rts {
		got, err := ByID(rt.ID())
		if err != nil {
			t.Fatalf("ByID(%q): %v", rt.ID(), err)
		}
		if got.ID() != rt.ID() {
			t.Fatalf("ByID round-trip mismatch: %q vs %q", got.ID(), rt.ID())
		}
	}
	if _, err := ByID("does-not-exist"); err == nil {
		t.Fatal("expected error for unknown runtime id")
	}
	all := DetectIncluding(ctx, true)
	if len(all) < len(rts) {
		t.Fatalf("DetectIncluding should be >= Detect, got %d vs %d", len(all), len(rts))
	}
}

func TestKillEmptyIDErrors(t *testing.T) {
	d := Docker{}
	if err := d.Kill(context.Background(), ""); err == nil {
		t.Fatal("expected empty-id error")
	}
}

func TestPruneRequiresAck(t *testing.T) {
	d := Docker{}
	if _, err := d.Prune(context.Background(), PruneOptions{}); err == nil {
		t.Fatal("expected ack error")
	}
}
