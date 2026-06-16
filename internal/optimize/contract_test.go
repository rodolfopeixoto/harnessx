// SPDX-License-Identifier: MIT

package optimize

import (
	"os"
	"path/filepath"
	"testing"
)

func validContract() ChangeContract {
	return ChangeContract{
		ID:                   "01ABC",
		Component:            ComponentRouterStrength,
		Target:               "claude",
		TargetFailure:        "router picks wrong adapter for image tasks",
		PredictedImprovement: "+15% correct route on image-tagged prompts",
		Invariants:           []string{"existing strengths unchanged for non-image tasks"},
		FalsifierTest:        "go test ./internal/router/... -run TestImageRoute",
		RollbackCmd:          "git checkout HEAD -- internal/app/agentcmd/bundled/claude.yaml",
		Patch:                "+image\n",
	}
}

func TestContractValidateAccepts(t *testing.T) {
	if err := validContract().Validate(); err != nil {
		t.Errorf("valid contract rejected: %v", err)
	}
}

func TestContractValidateMissingID(t *testing.T) {
	c := validContract()
	c.ID = ""
	if err := c.Validate(); err == nil {
		t.Error("expected missing id error")
	}
}

func TestContractValidateUnknownComponent(t *testing.T) {
	c := validContract()
	c.Component = "bogus"
	if err := c.Validate(); err == nil {
		t.Error("expected unknown component error")
	}
}

func TestContractValidateRequiresRollback(t *testing.T) {
	c := validContract()
	c.RollbackCmd = ""
	if err := c.Validate(); err == nil {
		t.Error("expected missing rollback_cmd error")
	}
}

func TestContractValidateRequiresFalsifier(t *testing.T) {
	c := validContract()
	c.FalsifierTest = ""
	if err := c.Validate(); err == nil {
		t.Error("expected missing falsifier_test error")
	}
}

func TestContractValidateRequiresTargetFailure(t *testing.T) {
	c := validContract()
	c.TargetFailure = ""
	if err := c.Validate(); err == nil {
		t.Error("expected missing target_failure error")
	}
}

func TestWriteAndReadContractRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "x", "contract.json")
	c := validContract()
	if err := WriteContract(path, c); err != nil {
		t.Fatal(err)
	}
	got, err := ReadContract(path)
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != c.ID || got.Component != c.Component || got.RollbackCmd != c.RollbackCmd {
		t.Errorf("round-trip mismatch: %+v", got)
	}
	if got.SchemaVersion != ContractSchemaVersion {
		t.Errorf("schema_version: got %d", got.SchemaVersion)
	}
	if got.CreatedAt.IsZero() {
		t.Error("CreatedAt should be set on write")
	}
}

func TestWriteContractRefusesInvalid(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "x", "contract.json")
	c := validContract()
	c.RollbackCmd = ""
	if err := WriteContract(path, c); err == nil {
		t.Error("expected write to reject invalid contract")
	}
}

func TestReadContractMissingFile(t *testing.T) {
	if _, err := ReadContract("/tmp/__nope_contract__.json"); err == nil {
		t.Error("expected error for missing file")
	}
}

func TestReadContractInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.json")
	if err := writeFile(path, "{not json"); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadContract(path); err == nil {
		t.Error("expected parse error")
	}
}

func writeFile(p, body string) error {
	return os.WriteFile(p, []byte(body), 0o644)
}
