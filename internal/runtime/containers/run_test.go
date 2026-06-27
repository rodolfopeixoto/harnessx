// SPDX-License-Identifier: MIT

package containers

import (
	"context"
	"testing"
)

func TestDockerLikeRunRequiresImage(t *testing.T) {
	d := Docker{}
	_, err := d.Run(context.Background(), RunSpec{})
	if err == nil {
		t.Fatal("expected error when Image is empty")
	}
}

func TestPruneImagesRequiresAck(t *testing.T) {
	if _, err := (Docker{}).PruneImages(context.Background(), ImagePruneOptions{}); err == nil {
		t.Fatal("docker: expected ack error")
	}
	if _, err := (Podman{}).PruneImages(context.Background(), ImagePruneOptions{}); err == nil {
		t.Fatal("podman: expected ack error")
	}
	if _, err := (OrbStack{}).PruneImages(context.Background(), ImagePruneOptions{}); err == nil {
		t.Fatal("orbstack: expected ack error")
	}
	if _, err := (Colima{}).PruneImages(context.Background(), ImagePruneOptions{}); err == nil {
		t.Fatal("colima: expected ack error")
	}
}

func TestAppleContainerRunAndPruneStillStub(t *testing.T) {
	apple := AppleContainer{}
	if _, err := apple.Run(context.Background(), RunSpec{Image: "busybox"}); err == nil {
		t.Fatal("expected not-wired error from AppleContainer.Run")
	}
	if _, err := apple.PruneImages(context.Background(), ImagePruneOptions{IUnderstand: true}); err == nil {
		t.Fatal("expected not-wired error from AppleContainer.PruneImages")
	}
}

func TestRuntimeAliasesForward(t *testing.T) {
	// Exercises every runtime alias against an empty RunSpec so each
	// wrapper executes its validation path (cannot reach docker in CI).
	ctx := context.Background()
	p := Podman{}
	o := OrbStack{}
	c := Colima{}
	d := Docker{}
	_, _ = p.Run(ctx, RunSpec{})
	_, _ = o.Run(ctx, RunSpec{})
	_, _ = c.Run(ctx, RunSpec{})
	_, _ = p.ListImages(ctx)
	_, _ = o.ListImages(ctx)
	_, _ = c.ListImages(ctx)
	_, _ = d.ListImages(ctx)
}
