// SPDX-License-Identifier: MIT

package optimize

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const ContractSchemaVersion = 1

type ComponentKind string

const (
	ComponentAgentSpec      ComponentKind = "agent_spec"
	ComponentRouterStrength ComponentKind = "router_strength"
	ComponentAutonomyPolicy ComponentKind = "autonomy_policy"
	ComponentSensorSet      ComponentKind = "sensor_set"
)

type ChangeContract struct {
	SchemaVersion        int           `json:"schema_version"`
	ID                   string        `json:"id"`
	Component            ComponentKind `json:"component"`
	Target               string        `json:"target"`
	TargetFailure        string        `json:"target_failure"`
	PredictedImprovement string        `json:"predicted_improvement"`
	Invariants           []string      `json:"invariants"`
	FalsifierTest        string        `json:"falsifier_test"`
	RollbackCmd          string        `json:"rollback_cmd"`
	Patch                string        `json:"patch"`
	CreatedAt            time.Time     `json:"created_at"`
}

func (c ChangeContract) Validate() error {
	if c.ID == "" {
		return errors.New("contract: missing id")
	}
	if c.Component == "" {
		return errors.New("contract: missing component")
	}
	switch c.Component {
	case ComponentAgentSpec, ComponentRouterStrength, ComponentAutonomyPolicy, ComponentSensorSet:
	default:
		return fmt.Errorf("contract: unknown component %q", c.Component)
	}
	if c.RollbackCmd == "" {
		return errors.New("contract: missing rollback_cmd (paper requirement)")
	}
	if c.FalsifierTest == "" {
		return errors.New("contract: missing falsifier_test (paper requirement)")
	}
	if c.TargetFailure == "" {
		return errors.New("contract: missing target_failure")
	}
	return nil
}

func WriteContract(path string, c ChangeContract) error {
	if c.SchemaVersion == 0 {
		c.SchemaVersion = ContractSchemaVersion
	}
	if c.CreatedAt.IsZero() {
		c.CreatedAt = time.Now().UTC()
	}
	if err := c.Validate(); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	body, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, body, 0o644)
}

func ReadContract(path string) (ChangeContract, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return ChangeContract{}, err
	}
	var c ChangeContract
	if err := json.Unmarshal(body, &c); err != nil {
		return ChangeContract{}, fmt.Errorf("contract: parse %s: %w", path, err)
	}
	if c.SchemaVersion != ContractSchemaVersion {
		return ChangeContract{}, fmt.Errorf("contract: schema_version=%d not supported (want %d)", c.SchemaVersion, ContractSchemaVersion)
	}
	return c, nil
}
