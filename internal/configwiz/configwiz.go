// SPDX-License-Identifier: MIT

package configwiz

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/ropeixoto/harnessx/internal/router"
)

const (
	routesRel   = ".harness/config/routes.yaml"
	auditLogRel = ".harness/logs/config-mutations.jsonl"
)

func routesPath(root string) string { return filepath.Join(root, routesRel) }
func auditPath(root string) string  { return filepath.Join(root, auditLogRel) }

type Snapshot struct {
	Routes map[string]router.RouteConfig `yaml:"routes" json:"routes"`
}

func Load(root string) (Snapshot, error) {
	routes, err := router.LoadConfig(routesPath(root))
	if err != nil {
		return Snapshot{}, err
	}
	return Snapshot{Routes: routes}, nil
}

func Save(root string, snap Snapshot) error {
	if err := os.MkdirAll(filepath.Dir(routesPath(root)), 0o755); err != nil {
		return err
	}
	body, err := yaml.Marshal(Snapshot{Routes: snap.Routes})
	if err != nil {
		return err
	}
	return os.WriteFile(routesPath(root), body, 0o644)
}

type Mutation struct {
	Time   time.Time          `json:"time"`
	Action string             `json:"action"`
	Task   string             `json:"task,omitempty"`
	Before router.RouteConfig `json:"before,omitempty"`
	After  router.RouteConfig `json:"after,omitempty"`
	Notes  string             `json:"notes,omitempty"`
}

func appendAudit(root string, m Mutation) error {
	if err := os.MkdirAll(filepath.Dir(auditPath(root)), 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(auditPath(root), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	m.Time = time.Now().UTC()
	return json.NewEncoder(f).Encode(m)
}

func SetTaskPrimary(root, task, primary string, fallback []string, budget float64, model string) error {
	snap, err := Load(root)
	if err != nil {
		return err
	}
	if snap.Routes == nil {
		snap.Routes = map[string]router.RouteConfig{}
	}
	before := snap.Routes[task]
	after := router.RouteConfig{Primary: primary, Fallback: fallback, BudgetUSD: budget, Model: model}
	snap.Routes[task] = after
	if err := Save(root, snap); err != nil {
		return err
	}
	return appendAudit(root, Mutation{Action: "set", Task: task, Before: before, After: after})
}

func DeleteTask(root, task string) error {
	snap, err := Load(root)
	if err != nil {
		return err
	}
	before, ok := snap.Routes[task]
	if !ok {
		return fmt.Errorf("configwiz: task %q not present", task)
	}
	delete(snap.Routes, task)
	if err := Save(root, snap); err != nil {
		return err
	}
	return appendAudit(root, Mutation{Action: "delete", Task: task, Before: before})
}

func Diff(before, after Snapshot) []string {
	keys := map[string]struct{}{}
	for k := range before.Routes {
		keys[k] = struct{}{}
	}
	for k := range after.Routes {
		keys[k] = struct{}{}
	}
	sorted := make([]string, 0, len(keys))
	for k := range keys {
		sorted = append(sorted, k)
	}
	sort.Strings(sorted)
	var out []string
	for _, k := range sorted {
		b, hasB := before.Routes[k]
		a, hasA := after.Routes[k]
		switch {
		case hasB && !hasA:
			out = append(out, "- "+k+": "+formatRoute(b))
		case !hasB && hasA:
			out = append(out, "+ "+k+": "+formatRoute(a))
		case formatRoute(b) != formatRoute(a):
			out = append(out, "~ "+k+": "+formatRoute(b)+" → "+formatRoute(a))
		}
	}
	return out
}

func formatRoute(r router.RouteConfig) string {
	parts := []string{"primary=" + r.Primary}
	if len(r.Fallback) > 0 {
		parts = append(parts, "fallback="+strings.Join(r.Fallback, ","))
	}
	if r.BudgetUSD > 0 {
		parts = append(parts, fmt.Sprintf("budget=$%.2f", r.BudgetUSD))
	}
	if r.Model != "" {
		parts = append(parts, "model="+r.Model)
	}
	return strings.Join(parts, " ")
}

type WizardOptions struct {
	Root         string
	AvailableIDs []string
	Tasks        []string
	In           io.Reader
	Out          io.Writer
}

func RunWizard(opts WizardOptions) error {
	if len(opts.AvailableIDs) == 0 {
		return errors.New("configwiz: no adapters available")
	}
	tasks := opts.Tasks
	if len(tasks) == 0 {
		tasks = []string{"planning", "implementation", "security_review", "cheap_review"}
	}
	r := bufio.NewReader(opts.In)
	snap, _ := Load(opts.Root)
	if snap.Routes == nil {
		snap.Routes = map[string]router.RouteConfig{}
	}
	for _, task := range tasks {
		current := snap.Routes[task]
		fmt.Fprintf(opts.Out, "\n[%s] current=%s\n", task, formatRoute(current))
		fmt.Fprintf(opts.Out, "available adapters: %s\n", strings.Join(opts.AvailableIDs, ", "))
		primary, err := readNonEmpty(r, opts.Out, "primary adapter", current.Primary)
		if err != nil {
			return err
		}
		fallback, err := readLine(r, opts.Out, "fallback chain (csv)", strings.Join(current.Fallback, ","))
		if err != nil {
			return err
		}
		budgetStr, err := readLine(r, opts.Out, "budget USD (0 = no cap)", fmt.Sprintf("%.2f", current.BudgetUSD))
		if err != nil {
			return err
		}
		var budget float64
		_, _ = fmt.Sscanf(budgetStr, "%f", &budget)
		fbList := splitCSV(fallback)
		if err := SetTaskPrimary(opts.Root, task, primary, fbList, budget, current.Model); err != nil {
			return err
		}
		fmt.Fprintf(opts.Out, "✓ %s set\n", task)
	}
	return nil
}

func readNonEmpty(r *bufio.Reader, out io.Writer, label, def string) (string, error) {
	for {
		s, err := readLine(r, out, label, def)
		if err != nil {
			return "", err
		}
		if s != "" {
			return s, nil
		}
		fmt.Fprintln(out, "(required)")
	}
}

func readLine(r *bufio.Reader, out io.Writer, label, def string) (string, error) {
	fmt.Fprintf(out, "%s [%s]: ", label, def)
	line, err := r.ReadString('\n')
	if err != nil && line == "" {
		return "", err
	}
	line = strings.TrimSpace(line)
	if line == "" {
		return def, nil
	}
	return line, nil
}

func SplitCSVForCLI(s string) []string { return splitCSV(s) }

func splitCSV(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}
