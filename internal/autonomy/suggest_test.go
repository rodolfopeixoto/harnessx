// SPDX-License-Identifier: MIT

package autonomy

import "testing"

func mk(path, decision string, n int) []Event {
	out := make([]Event, n)
	for i := range out {
		out[i] = Event{Path: path, Decision: decision}
	}
	return out
}

func TestSuggestConsistentDenyProposesDeny(t *testing.T) {
	events := mk(".env", "deny", 5)
	got := Suggest(events)
	if len(got) != 1 {
		t.Fatalf("want 1 proposal, got %d", len(got))
	}
	if got[0].To != "deny" || got[0].EvidenceCount != 5 {
		t.Errorf("got %+v", got[0])
	}
}

func TestSuggestConsistentApprovalProposesAllow(t *testing.T) {
	events := append(mk("src/", "approve", 5), mk("src/", "require_approval", 5)...)
	got := Suggest(events)
	if len(got) != 1 || got[0].To != "allow" {
		t.Errorf("want allow proposal, got %+v", got)
	}
}

func TestSuggestMixedProposesRevisit(t *testing.T) {
	events := append(mk("scripts/", "approve", 5), mk("scripts/", "deny", 1)...)
	got := Suggest(events)
	if len(got) != 1 || got[0].To != "require_approval" {
		t.Errorf("want revisit proposal, got %+v", got)
	}
}

func TestSuggestBelowThresholdEmits(t *testing.T) {
	events := mk("file", "deny", 3)
	got := Suggest(events)
	if len(got) != 0 {
		t.Errorf("below threshold should not propose, got %+v", got)
	}
}

func TestSuggestEmpty(t *testing.T) {
	if got := Suggest(nil); len(got) != 0 {
		t.Errorf("nil events should yield no proposals, got %+v", got)
	}
}
