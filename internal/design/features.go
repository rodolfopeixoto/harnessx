// SPDX-License-Identifier: MIT

package design

import (
	"sort"
	"strings"
	"time"
)

// BuildFeatureMap derives feature-map.json from a manifest. Heuristic
// classification:
//   - "auth"-shaped routes require backend → status=mock, priority=mvp.
//   - "marketplace", "admin", "settings" routes default to disabled / post_mvp.
//   - everything else gets status=mock when interactions exist, static otherwise.
func BuildFeatureMap(m *Manifest) FeatureMap {
	fm := FeatureMap{
		GeneratedAt: time.Now().UTC(),
		Features:    map[string]FeatureSpec{},
	}
	for _, page := range m.Pages {
		id := featureIDFromPage(page)
		spec := FeatureSpec{
			Routes:     []string{page.Path},
			Components: page.Components,
			Priority:   priorityFor(page),
		}
		switch {
		case requiresBackend(page):
			spec.Status = StatusMock
			spec.BackendRequired = true
			spec.APIContract = apiContractFor(page)
		case len(page.Interactions) > 0:
			spec.Status = StatusMock
		default:
			spec.Status = StatusStatic
		}
		fm.Features[id] = spec
	}
	return fm
}

// PromoteToggleMap projects a FeatureMap into the runtime ToggleMap
// surface used by the React app at runtime.
func PromoteToggleMap(fm FeatureMap) ToggleMap {
	tm := ToggleMap{GeneratedAt: time.Now().UTC(), Toggles: map[string]Toggle{}}
	for id, f := range fm.Features {
		tm.Toggles[id] = Toggle{
			Status: f.Status, Routes: f.Routes,
			Description: descFor(id, f),
		}
	}
	return tm
}

// BuildRoadmap emits the canonical MVP 0–4 phases (spec §12). Features
// land in a phase based on (status, priority).
func BuildRoadmap(fm FeatureMap) Roadmap {
	r := Roadmap{
		GeneratedAt: time.Now().UTC(),
		Phases: []RoadmapPhase{
			{Name: "MVP 0", Goal: "React parity — visual + navigable with mock data"},
			{Name: "MVP 1", Goal: "Core flows backed by real API"},
			{Name: "MVP 2", Goal: "Business rules, validations, permissions, persistence"},
			{Name: "MVP 3", Goal: "Operational readiness — logs, security, perf, real E2E"},
			{Name: "MVP 4", Goal: "Growth features — secondary, marketplace, gamification"},
		},
	}
	for id, f := range fm.Features {
		switch {
		case f.Status == StatusStatic, f.Status == StatusMock && !f.BackendRequired:
			r.Phases[0].Features = append(r.Phases[0].Features, id)
		case f.Status == StatusMock && f.BackendRequired && f.Priority == PriorityMVP:
			r.Phases[1].Features = append(r.Phases[1].Features, id)
		case f.Status == StatusAPIContract:
			r.Phases[2].Features = append(r.Phases[2].Features, id)
		case f.Status == StatusBackendReady:
			r.Phases[3].Features = append(r.Phases[3].Features, id)
		default:
			r.Phases[4].Features = append(r.Phases[4].Features, id)
		}
	}
	for i := range r.Phases {
		sort.Strings(r.Phases[i].Features)
	}
	return r
}

// BuildAPIContracts emits api-contracts.json from features that declare a
// backend contract. Methods + paths are inferred from feature ids.
func BuildAPIContracts(fm FeatureMap) APIContracts {
	a := APIContracts{GeneratedAt: time.Now().UTC()}
	ids := make([]string, 0, len(fm.Features))
	for id := range fm.Features {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		f := fm.Features[id]
		if !f.BackendRequired {
			continue
		}
		ep := APIEndpointSpec{
			Method: inferMethod(id), Path: f.APIContract,
			Feature: id, Status: "proposed",
			Notes: []string{"contract drafted from design inventory — confirm with product before backend work"},
		}
		a.Endpoints = append(a.Endpoints, ep)
	}
	return a
}

// BuildFlowMap walks the manifest's DetectedFlows and groups them by
// origin page. Each flow is two-step (page → target) by construction;
// callers can graft multi-step flows on top.
func BuildFlowMap(m *Manifest) FlowMap {
	fm := FlowMap{GeneratedAt: time.Now().UTC()}
	for i, f := range m.DetectedFlows {
		fm.Flows = append(fm.Flows, Flow{
			ID:    flowID(i),
			Steps: []string{f},
			Pages: []string{strings.SplitN(f, " → ", 2)[0]},
		})
	}
	return fm
}

// --- heuristics -------------------------------------------------------------

func featureIDFromPage(p Page) string {
	id := strings.TrimPrefix(p.Path, "/")
	if id == "" {
		id = "home"
	}
	id = strings.ReplaceAll(id, "/", ".")
	id = strings.ReplaceAll(id, "-", "_")
	return "feature." + id
}

func requiresBackend(p Page) bool {
	hot := []string{"auth", "signup", "login", "checkout", "payment", "settings", "profile", "admin", "dashboard", "api"}
	body := strings.ToLower(p.Path + " " + p.Title + " " + strings.Join(p.Interactions, " "))
	for _, h := range hot {
		if strings.Contains(body, h) {
			return true
		}
	}
	for _, a := range p.Interactions {
		if strings.HasPrefix(a, "submit") {
			return true
		}
	}
	return false
}

func priorityFor(p Page) Priority {
	low := strings.ToLower(p.Path + " " + p.Title)
	switch {
	case strings.Contains(low, "marketplace"), strings.Contains(low, "gamif"), strings.Contains(low, "rewards"):
		return PriorityPostMVP
	case strings.Contains(low, "admin"):
		return PriorityBacklog
	default:
		return PriorityMVP
	}
}

func apiContractFor(p Page) string {
	method := "POST"
	if strings.Contains(strings.ToLower(p.Path), "list") || strings.Contains(strings.ToLower(p.Path), "/index") {
		method = "GET"
	}
	path := "/api" + p.Path
	if path == "/api/" {
		path = "/api/root"
	}
	return method + " " + path
}

func inferMethod(featureID string) string {
	low := strings.ToLower(featureID)
	switch {
	case strings.Contains(low, "list"), strings.Contains(low, "index"), strings.Contains(low, "home"):
		return "GET"
	case strings.Contains(low, "signup"), strings.Contains(low, "submit"), strings.Contains(low, "create"):
		return "POST"
	case strings.Contains(low, "update"), strings.Contains(low, "edit"):
		return "PUT"
	case strings.Contains(low, "delete"):
		return "DELETE"
	default:
		return "POST"
	}
}

func descFor(id string, f FeatureSpec) string {
	if f.BackendRequired {
		return "needs backend: " + f.APIContract
	}
	return "static / mock content"
}

func flowID(i int) string {
	return "flow_" + leftpad(i, 3)
}

func leftpad(n, width int) string {
	s := ""
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	for len(s) < width {
		s = "0" + s
	}
	if s == "" {
		s = strings.Repeat("0", width)
	}
	return s
}
