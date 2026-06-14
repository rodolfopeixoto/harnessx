// SPDX-License-Identifier: MIT

package domain

import "time"

type AgentCertification struct {
	ID             string
	AgentID        string
	CLIVersion     string
	AdapterVersion string
	Score          int
	Status         string
	DetailsJSON    string
	CertifiedAt    time.Time
}

type Memory struct {
	ID            string
	Scope         string
	Kind          string
	Content       string
	EvidenceRunID string
	Confidence    float64
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type Artifact struct {
	ID           string
	SessionID    string
	RunID        string
	Kind         string
	Path         string
	ContentHash  string
	MetadataJSON string
	CreatedAt    time.Time
}
