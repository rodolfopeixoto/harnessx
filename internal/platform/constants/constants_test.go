package constants

import "testing"

// Compile-time sanity: every constant must keep its expected category
// invariant. If a value changes accidentally during refactor, this test
// is the canary.
func TestInvariants(t *testing.T) {
	if DefaultProbeTimeout <= 0 {
		t.Fatal("probe timeout must be positive")
	}
	if DefaultDashboardPort < 1024 || DefaultDashboardPort > 65535 {
		t.Fatalf("dashboard port out of range: %d", DefaultDashboardPort)
	}
	if MaxZipExtractBytes < MaxContextFileBytes {
		t.Fatal("zip cap must exceed per-file cap")
	}
	if MemoryConfidenceFloor <= 0 || MemoryConfidenceFloor > 1 {
		t.Fatalf("confidence floor out of range: %v", MemoryConfidenceFloor)
	}
}
