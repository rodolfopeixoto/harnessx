// SPDX-License-Identifier: MIT

package sensors

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/ropeixoto/harnessx/internal/index"
	"github.com/ropeixoto/harnessx/internal/sensors/planscope"
)

type PlanScopeSensor struct {
	IDValue string
	PlanID  string
}

const planPinFile = ".harness/config/plan.yaml"

type planPin struct {
	ActivePlanID string `yaml:"active_plan_id"`
}

func (s PlanScopeSensor) ID() string         { return s.IDValue }
func (s PlanScopeSensor) Category() Category { return CatSpec }
func (s PlanScopeSensor) Kind() Kind         { return KindComputational }
func (s PlanScopeSensor) AppliesTo(p index.Profile) bool {
	return strings.TrimSpace(s.PlanID) != ""
}

func (s PlanScopeSensor) Run(rc RunCtx) Result {
	start := time.Now()
	res := Result{ID: s.IDValue, Category: CatSpec, Kind: KindComputational}
	if s.PlanID == "" {
		res.Status = StatusSkipped
		res.Detail = "no active plan id"
		res.Duration = time.Since(start)
		return res
	}
	out, err := planscope.Check(rc.Ctx, planscope.Options{Root: rc.Root, PlanID: s.PlanID})
	if err != nil {
		res.Status = StatusFailed
		res.Detail = err.Error()
		res.Duration = time.Since(start)
		return res
	}
	res.Detail = planscope.FormatResult(out)
	if out.Pass() {
		res.Status = StatusPassed
		res.Confidence = 1.0
	} else {
		res.Status = StatusFailed
		res.Unverified = collectViolationPaths(out)
	}
	res.Duration = time.Since(start)
	return res
}

func collectViolationPaths(r planscope.Result) []string {
	out := make([]string, 0, len(r.Violations))
	for _, v := range r.Violations {
		out = append(out, v.Path)
	}
	return out
}

func LoadActivePlanID(root string) (string, error) {
	body, err := os.ReadFile(filepath.Join(root, planPinFile))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", nil
		}
		return "", err
	}
	var p planPin
	if err := yaml.Unmarshal(body, &p); err != nil {
		return "", err
	}
	return strings.TrimSpace(p.ActivePlanID), nil
}

func SaveActivePlanID(root, planID string) error {
	dst := filepath.Join(root, planPinFile)
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	body, err := yaml.Marshal(planPin{ActivePlanID: planID})
	if err != nil {
		return err
	}
	return os.WriteFile(dst, body, 0o644)
}
