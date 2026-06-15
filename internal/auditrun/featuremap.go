// SPDX-License-Identifier: MIT

package auditrun

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/ropeixoto/harnessx/internal/platform/constants"
)

const (
	RoleAnonymous Role = "anonymous"
	RoleOperator  Role = "operator"
	RoleAdmin     Role = "admin"
)

func DefaultViewports() []Viewport {
	return []Viewport{
		{Name: constants.AuditViewportDesk, Width: constants.AuditViewportDeskW, Height: constants.AuditViewportDeskH},
		{Name: constants.AuditViewportTab, Width: constants.AuditViewportTabW, Height: constants.AuditViewportTabH},
		{Name: constants.AuditViewportMob, Width: constants.AuditViewportMobW, Height: constants.AuditViewportMobH},
	}
}

func uiFeature(id, name, route string, role Role, priority string, viewports []string, extra ...string) Feature {
	selectors := []string{"[data-testid='shell']", "[data-testid='page-" + id + "']"}
	selectors = append(selectors, extra...)
	return Feature{
		ID: id, Name: name, Route: route, Role: role,
		Category: "ui", Priority: priority, ExpectedHTTPStatus: 200,
		ExpectedSelectors: selectors,
		Viewports:         viewports,
	}
}

func DefaultFeatures(_ string) []Feature {
	uiDesk := []string{constants.AuditViewportDesk}
	uiAll := []string{constants.AuditViewportDesk, constants.AuditViewportTab, constants.AuditViewportMob}
	uiMobile := []string{constants.AuditViewportDesk, constants.AuditViewportMob}
	apiCheck := func(id, name, route, priority string) Feature {
		return Feature{
			ID: id, Name: name, Route: route, Role: RoleOperator,
			Category: "api", Priority: priority, ExpectedHTTPStatus: 200, Viewports: uiDesk,
		}
	}
	return []Feature{
		{
			ID: "home", Name: "Home", Route: "/", Role: RoleOperator,
			Category: "ui", Priority: constants.AuditSeverityP0, ExpectedHTTPStatus: 200,
			ExpectedSelectors: []string{
				"[data-testid='shell']",
				"[data-testid='page-home']",
				"[data-testid='workspace-summary']",
				"[data-testid='health-score-card']",
				"[data-testid='next-action-card']",
				"[data-testid='recent-runs']",
				"[data-testid='terminal-reflection']",
			},
			Viewports: uiAll,
		},
		uiFeature("projects", "Projects hub", "/projects", RoleOperator, constants.AuditSeverityP0, uiAll,
			"[data-testid='project-switcher']", "[data-testid='projects-explorer']"),
		uiFeature("command", "Command", "/command", RoleOperator, constants.AuditSeverityP0, uiAll,
			"[data-testid='command-prompt']", "[data-testid='command-detection']", "[data-testid='command-actions']"),
		uiFeature("plan", "Plan", "/plan", RoleOperator, constants.AuditSeverityP0, uiDesk,
			"[data-testid='plan-summary']", "[data-testid='plan-files']", "[data-testid='plan-actions']"),
		uiFeature("activerun", "Active run", "/run", RoleOperator, constants.AuditSeverityP0, uiDesk,
			"[data-testid='activerun-summary']", "[data-testid='run-timeline']", "[data-testid='run-actions']"),
		{
			ID: "design", Name: "Design to product", Route: "/design", Role: RoleOperator,
			Category: "ui", Priority: constants.AuditSeverityP1, ExpectedHTTPStatus: 200,
			ExpectedSelectors: []string{"[data-testid='shell']", "[data-testid='nav-design']"},
			Viewports:         uiDesk,
		},
		{
			ID: "roadmap", Name: "Roadmap", Route: "/roadmap", Role: RoleOperator,
			Category: "ui", Priority: constants.AuditSeverityP1, ExpectedHTTPStatus: 200,
			ExpectedSelectors: []string{"[data-testid='shell']", "[data-testid='nav-roadmap']"},
			Viewports:         uiDesk,
		},
		{
			ID: "agents", Name: "Agents", Route: "/agents", Role: RoleOperator,
			Category: "ui", Priority: constants.AuditSeverityP1, ExpectedHTTPStatus: 200,
			ExpectedSelectors: []string{"[data-testid='shell']", "[data-testid='nav-agents']"},
			Viewports:         uiDesk,
		},
		uiFeature("catalog", "Capabilities", "/catalog", RoleOperator, constants.AuditSeverityP0, uiAll,
			"[data-testid='capabilities-tabs']"),
		{
			ID: "sensors", Name: "Sensors", Route: "/sensors", Role: RoleOperator,
			Category: "ui", Priority: constants.AuditSeverityP1, ExpectedHTTPStatus: 200,
			ExpectedSelectors: []string{"[data-testid='shell']", "[data-testid='nav-sensors']"},
			Viewports:         uiMobile,
		},
		uiFeature("context", "Context pack", "/context", RoleOperator, constants.AuditSeverityP1, uiDesk,
			"[data-testid='context-summary']", "[data-testid='context-explorer']"),
		{
			ID: "memory", Name: "Memory", Route: "/memory", Role: RoleOperator,
			Category: "ui", Priority: constants.AuditSeverityP1, ExpectedHTTPStatus: 200,
			ExpectedSelectors: []string{"[data-testid='shell']", "[data-testid='nav-memory']"},
			Viewports:         uiDesk,
		},
		uiFeature("resources", "Resources", "/resources", RoleOperator, constants.AuditSeverityP1, uiDesk,
			"[data-testid='resources-summary']", "[data-testid='resources-categories']"),
		{
			ID: "cleanup", Name: "Cleanup", Route: "/cleanup", Role: RoleOperator,
			Category: "ui", Priority: constants.AuditSeverityP1, ExpectedHTTPStatus: 200,
			ExpectedSelectors: []string{
				"[data-testid='shell']",
				"[data-testid='page-cleanup']",
				"[data-testid='cleanup-plan-banner']",
				"[data-testid='cleanup-summary']",
				"[data-testid='cleanup-explorer']",
			},
			Viewports: uiDesk,
		},
		uiFeature("reports", "Reports", "/reports", RoleOperator, constants.AuditSeverityP1, uiDesk,
			"[data-testid='reports-summary']", "[data-testid='reports-explorer']"),
		uiFeature("stakeholder", "Stakeholder", "/stakeholder", RoleOperator, constants.AuditSeverityP2, uiDesk,
			"[data-testid='stakeholder-summary']", "[data-testid='stakeholder-next-decision']"),
		uiFeature("onboarding", "Onboarding", "/onboarding", RoleOperator, constants.AuditSeverityP1, uiDesk,
			"[data-testid='onboarding-steps']", "[data-testid='onboarding-next-action']"),
		{
			ID: "settings", Name: "Settings", Route: "/settings", Role: RoleAdmin,
			Category: "ui", Priority: constants.AuditSeverityP1, ExpectedHTTPStatus: 200,
			ExpectedSelectors: []string{"[data-testid='shell']", "[data-testid='nav-settings']"},
			Viewports:         uiDesk,
		},
		apiCheck("api-workspace", "Workspace projects API", "/api/workspace/projects", constants.AuditSeverityP0),
		apiCheck("api-catalog", "Catalog kinds API", "/api/catalog/kinds", constants.AuditSeverityP0),
		apiCheck("api-mcps", "MCP scan API", "/api/mcps", constants.AuditSeverityP1),
		apiCheck("api-hooks", "Hook scan API", "/api/hooks", constants.AuditSeverityP1),
		apiCheck("api-cleanup", "Cleanup scan API", "/api/cleanup/scan", constants.AuditSeverityP1),
		apiCheck("api-autonomy", "Autonomy matrix API", "/api/autonomy", constants.AuditSeverityP1),
		apiCheck("api-health-score", "Health score API", "/api/health/score", constants.AuditSeverityP1),
	}
}

func WriteFeatureMap(dir, baseURL string, features []Feature, viewports []Viewport) (string, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	fm := FeatureMap{
		GeneratedAt: time.Now().UTC(),
		BaseURL:     baseURL,
		Features:    features,
		Viewports:   viewports,
	}
	out := filepath.Join(dir, constants.AuditFeatureMapFile)
	body, err := json.MarshalIndent(fm, "", "  ")
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(out, body, 0o644); err != nil {
		return "", err
	}
	return out, nil
}
