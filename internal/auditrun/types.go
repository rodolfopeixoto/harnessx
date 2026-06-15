// SPDX-License-Identifier: MIT

package auditrun

import "time"

type Role string

type Viewport struct {
	Name   string `json:"name"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

type Feature struct {
	ID                 string   `json:"id"`
	Name               string   `json:"name"`
	Route              string   `json:"route"`
	Role               Role     `json:"role"`
	Category           string   `json:"category"`
	Priority           string   `json:"priority"`
	ReferenceScreen    string   `json:"reference_screen,omitempty"`
	DataNeeded         []string `json:"data_needed,omitempty"`
	Actions            []string `json:"actions,omitempty"`
	APIsUsed           []string `json:"apis_used,omitempty"`
	ExpectedHTTPStatus int      `json:"expected_http_status"`
	ExpectedSelectors  []string `json:"expected_selectors"`
	ExpectedContent    []string `json:"expected_content,omitempty"`
	Viewports          []string `json:"viewports,omitempty"`
}

type FeatureMap struct {
	GeneratedAt time.Time  `json:"generated_at"`
	BaseURL     string     `json:"base_url"`
	Features    []Feature  `json:"features"`
	Viewports   []Viewport `json:"viewports"`
}

type VisualDiff struct {
	FeatureID    string  `json:"feature_id"`
	Viewport     string  `json:"viewport"`
	Reference    string  `json:"reference,omitempty"`
	Current      string  `json:"current,omitempty"`
	DiffImage    string  `json:"diff_image,omitempty"`
	DiffPct      float64 `json:"diff_pct"`
	VisualStatus string  `json:"visual_status"`
	Reason       string  `json:"reason,omitempty"`
}

type ConsoleError struct {
	FeatureID string `json:"feature_id"`
	Viewport  string `json:"viewport"`
	Severity  string `json:"severity"`
	Message   string `json:"message"`
}

type NetworkError struct {
	FeatureID string `json:"feature_id"`
	Viewport  string `json:"viewport"`
	URL       string `json:"url"`
	Status    int    `json:"status"`
	Method    string `json:"method"`
}

type MissingSelector struct {
	FeatureID string `json:"feature_id"`
	Viewport  string `json:"viewport"`
	Selector  string `json:"selector"`
}

type LayoutMetric struct {
	FeatureID           string `json:"feature_id"`
	Viewport            string `json:"viewport"`
	HasHorizontalScroll bool   `json:"has_horizontal_scroll"`
	BodyWidth           int    `json:"body_width"`
	ViewportWidth       int    `json:"viewport_width"`
}

type Result struct {
	FeatureID        string         `json:"feature_id"`
	Viewport         string         `json:"viewport"`
	Status           string         `json:"status"`
	Reason           string         `json:"reason,omitempty"`
	URLFinal         string         `json:"url_final,omitempty"`
	HTTPStatus       int            `json:"http_status,omitempty"`
	Screenshot       string         `json:"screenshot,omitempty"`
	ConsoleErrors    []ConsoleError `json:"console_errors,omitempty"`
	NetworkErrors    []NetworkError `json:"network_errors,omitempty"`
	MissingSelectors []string       `json:"missing_selectors,omitempty"`
	Visual           *VisualDiff    `json:"visual,omitempty"`
	Layout           *LayoutMetric  `json:"layout,omitempty"`
	DurationMs       int            `json:"duration_ms"`
	RecordedAt       time.Time      `json:"recorded_at"`
}

type Results struct {
	GeneratedAt time.Time `json:"generated_at"`
	BaseURL     string    `json:"base_url"`
	Results     []Result  `json:"results"`
}

type Summary struct {
	GeneratedAt   time.Time      `json:"generated_at"`
	BaseURL       string         `json:"base_url"`
	Counts        map[string]int `json:"counts"`
	Visual        map[string]int `json:"visual_counts"`
	Severity      map[string]int `json:"severity_counts"`
	PassRate      float64        `json:"pass_rate"`
	TotalFeatures int            `json:"total_features"`
	TotalResults  int            `json:"total_results"`
}

type BacklogItem struct {
	ID             string `json:"id"`
	Severity       string `json:"severity"`
	Feature        string `json:"feature"`
	Route          string `json:"route"`
	Role           string `json:"role"`
	Viewport       string `json:"viewport"`
	FailureType    string `json:"failure_type"`
	Reproduce      string `json:"reproduce"`
	Expected       string `json:"expected"`
	Actual         string `json:"actual"`
	Screenshot     string `json:"screenshot,omitempty"`
	DiffImage      string `json:"diff_image,omitempty"`
	Suggestion     string `json:"suggestion,omitempty"`
	AcceptCriteria string `json:"accept_criteria,omitempty"`
}
