// SPDX-License-Identifier: MIT

package importwizcmd

import (
	"context"
	"fmt"
	"io"

	"github.com/ropeixoto/harnessx/internal/importwiz"
	"github.com/ropeixoto/harnessx/internal/workspace"
)

func Run(ctx context.Context, opts importwiz.Options, registryPath string, out io.Writer) error {
	reg, err := workspace.Open(registryPath)
	if err != nil {
		return err
	}
	defer reg.Close()
	res, err := importwiz.Run(ctx, reg, opts)
	for _, step := range res.Steps {
		fmt.Fprintf(out, "%-9s %-32s %s\n", step.Status, step.Title, step.Detail)
	}
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "imported %s (%s) — stack: %v\n", res.Project.DisplayName, res.Project.Slug, res.Stack)
	return nil
}
