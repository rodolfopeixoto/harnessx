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

func DefaultFeatures(baseURL string) []Feature {
	return []Feature{
		{
			ID: "p01-public-landing", Name: "Sessions list", Route: "/",
			Role: RoleOperator, Category: "core", Priority: constants.AuditSeverityP0,
			ExpectedHTTPStatus: 200,
			ExpectedSelectors:  []string{"[data-testid='shell']", "[data-testid='nav-home']"},
			ExpectedContent:    []string{"Sessions"},
			APIsUsed:           []string{"/api/sessions"},
			Viewports:          []string{constants.AuditViewportDesk, constants.AuditViewportTab, constants.AuditViewportMob},
		},
		{
			ID: "p02-sensors", Name: "Sensors page", Route: "/sensors",
			Role: RoleOperator, Category: "core", Priority: constants.AuditSeverityP1,
			ExpectedHTTPStatus: 200,
			ExpectedSelectors:  []string{"[data-testid='nav-sensors']"},
			APIsUsed:           []string{"/api/sensors"},
			Viewports:          []string{constants.AuditViewportDesk, constants.AuditViewportMob},
		},
		{
			ID: "p03-agents", Name: "Agents page", Route: "/agents",
			Role: RoleOperator, Category: "core", Priority: constants.AuditSeverityP1,
			ExpectedHTTPStatus: 200,
			ExpectedSelectors:  []string{"[data-testid='nav-agents']"},
			APIsUsed:           []string{"/api/agents"},
			Viewports:          []string{constants.AuditViewportDesk},
		},
		{
			ID: "p04-memory", Name: "Memory page", Route: "/memory",
			Role: RoleOperator, Category: "core", Priority: constants.AuditSeverityP1,
			ExpectedHTTPStatus: 200,
			ExpectedSelectors:  []string{"[data-testid='nav-memory']"},
			APIsUsed:           []string{"/api/memory"},
			Viewports:          []string{constants.AuditViewportDesk},
		},
		{
			ID: "p05-design", Name: "Design ingestion view", Route: "/design",
			Role: RoleOperator, Category: "secondary", Priority: constants.AuditSeverityP2,
			ExpectedHTTPStatus: 200,
			ExpectedSelectors:  []string{"[data-testid='nav-design']"},
			APIsUsed:           []string{"/api/design"},
			Viewports:          []string{constants.AuditViewportDesk},
		},
		{
			ID: "p06-roadmap", Name: "Roadmap view", Route: "/roadmap",
			Role: RoleOperator, Category: "secondary", Priority: constants.AuditSeverityP2,
			ExpectedHTTPStatus: 200,
			ExpectedSelectors:  []string{"[data-testid='nav-roadmap']"},
			APIsUsed:           []string{"/api/roadmap"},
			Viewports:          []string{constants.AuditViewportDesk},
		},
		{
			ID: "p07-settings", Name: "Settings", Route: "/settings",
			Role: RoleAdmin, Category: "admin", Priority: constants.AuditSeverityP1,
			ExpectedHTTPStatus: 200,
			ExpectedSelectors:  []string{"[data-testid='nav-settings']"},
			APIsUsed:           []string{"/api/health", "/api/profile"},
			Viewports:          []string{constants.AuditViewportDesk},
		},
		{
			ID: "p08-workspace-api", Name: "Workspace projects endpoint", Route: "/api/workspace/projects",
			Role: RoleOperator, Category: "api", Priority: constants.AuditSeverityP0,
			ExpectedHTTPStatus: 200,
			Viewports:          []string{constants.AuditViewportDesk},
		},
		{
			ID: "p09-catalog-api", Name: "Catalog kinds endpoint", Route: "/api/catalog/kinds",
			Role: RoleOperator, Category: "api", Priority: constants.AuditSeverityP0,
			ExpectedHTTPStatus: 200,
			Viewports:          []string{constants.AuditViewportDesk},
		},
		{
			ID: "p10-autonomy-api", Name: "Autonomy matrix endpoint", Route: "/api/autonomy",
			Role: RoleOperator, Category: "api", Priority: constants.AuditSeverityP1,
			ExpectedHTTPStatus: 200,
			Viewports:          []string{constants.AuditViewportDesk},
		},
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
