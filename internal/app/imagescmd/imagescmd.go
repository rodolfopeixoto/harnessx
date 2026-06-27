package imagescmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"text/tabwriter"
	"time"

	"github.com/ropeixoto/harnessx/internal/runtime/containers"
)

type Options struct {
	Root string
	JSON bool
}

func List(ctx context.Context, out io.Writer, opts Options) error {
	rt, _, err := containers.Resolve(ctx, opts.Root)
	if err != nil {
		return err
	}
	images, err := rt.ListImages(ctx)
	if err != nil || len(images) == 0 {
		if fallback, ok := dockerFallback(ctx); ok {
			images = fallback
			err = nil
		}
	}
	if err != nil {
		return err
	}
	if opts.JSON {
		return json.NewEncoder(out).Encode(images)
	}
	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "REPOSITORY\tTAG\tID\tCREATED")
	for _, img := range images {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", img.Repository, img.Tag, short(img.ID, 12), img.CreatedAt.Format("2006-01-02"))
	}
	return w.Flush()
}

func dockerFallback(ctx context.Context) ([]containers.Image, bool) {
	if _, err := exec.LookPath("docker"); err != nil {
		return nil, false
	}
	imgs, err := (containers.Docker{}).ListImages(ctx)
	if err != nil {
		return nil, false
	}
	return imgs, true
}

func short(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

type PruneOptions struct {
	Root      string
	OlderThan time.Duration
	Confirmed bool
}

func Prune(ctx context.Context, out io.Writer, opts PruneOptions) error {
	rt, _, err := containers.Resolve(ctx, opts.Root)
	if err != nil {
		return err
	}
	popts := containers.ImagePruneOptions{OlderThan: opts.OlderThan, IUnderstand: opts.Confirmed}
	if !popts.IUnderstand {
		return fmt.Errorf("images prune: confirmation required (use --confirm or HARNESS_CONTAINERS_I_UNDERSTAND=1)")
	}
	if _, err := rt.PruneImages(ctx, popts); err != nil {
		return err
	}
	fmt.Fprintln(out, "images pruned")
	return nil
}
