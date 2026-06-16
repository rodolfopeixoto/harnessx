// SPDX-License-Identifier: MIT

package devloop

import "testing"

func TestCheckVerificationEmptySamplesPasses(t *testing.T) {
	v := checkVerification(nil)
	if !v.Passed {
		t.Errorf("empty samples should pass, got %+v", v)
	}
}

func TestCheckVerificationHighConfidencePasses(t *testing.T) {
	v := checkVerification([]VerificationSample{
		{SensorID: "secrets_scan", Confidence: 0.9, UnverifiedCount: 0},
	})
	if !v.Passed {
		t.Errorf("high conf + zero unverified should pass: %+v", v)
	}
}

func TestCheckVerificationRefusesLowConfWithUnverified(t *testing.T) {
	v := checkVerification([]VerificationSample{
		{SensorID: "secrets_scan", Confidence: 0.3, UnverifiedCount: 2},
	})
	if v.Passed {
		t.Errorf("low conf + unverified should fail")
	}
	if v.Reason == "" {
		t.Error("reason should be populated on fail")
	}
}

func TestCheckVerificationPassesLowConfNoUnverified(t *testing.T) {
	v := checkVerification([]VerificationSample{
		{SensorID: "x", Confidence: 0.2, UnverifiedCount: 0},
	})
	if !v.Passed {
		t.Errorf("low conf alone should pass: %+v", v)
	}
}

func TestCheckVerificationPassesHighConfManyUnverified(t *testing.T) {
	v := checkVerification([]VerificationSample{
		{SensorID: "x", Confidence: 0.95, UnverifiedCount: 10},
	})
	if !v.Passed {
		t.Errorf("high conf with unverified should pass: %+v", v)
	}
}

func TestCheckVerificationPicksFirstFailure(t *testing.T) {
	v := checkVerification([]VerificationSample{
		{SensorID: "ok", Confidence: 0.9},
		{SensorID: "bad", Confidence: 0.2, UnverifiedCount: 3},
	})
	if v.Passed {
		t.Error("expected failure")
	}
	if v.Reason == "" || v.Reason[len(v.Reason)-1] == '0' {
		t.Errorf("reason should name bad sensor: %q", v.Reason)
	}
}
