// SPDX-License-Identifier: MIT

package specflow

import (
	"strings"
	"testing"
)

func TestLookupTemplateKnownAndUnknown(t *testing.T) {
	if got := LookupTemplate("auth"); got.ID != "auth" {
		t.Errorf("known lookup wrong: %+v", got)
	}
	if got := LookupTemplate("nonexistent"); got.ID != "none" {
		t.Errorf("unknown lookup must fall back to none, got %+v", got)
	}
	if got := LookupTemplate(""); got.ID != "none" {
		t.Errorf("empty lookup must fall back to none, got %+v", got)
	}
}

func TestQuestionsForMergesBaselineAndExtras(t *testing.T) {
	got := QuestionsFor("auth")
	if len(got) != len(BaselineQuestions)+len(LookupTemplate("auth").ExtraQuestions) {
		t.Errorf("count mismatch: %d", len(got))
	}
	first := got[0].Key
	if first != BaselineQuestions[0].Key {
		t.Errorf("baseline must come first, got %s", first)
	}
}

func TestSkeletonForReturnsNonEmptyForKnownTemplates(t *testing.T) {
	for _, id := range []string{"auth", "authz", "pagination", "rate-limit", "caching", "audit-log", "algorithm-custom"} {
		if SkeletonFor(id) == "" {
			t.Errorf("%s skeleton must not be empty", id)
		}
	}
	if SkeletonFor("none") != "" {
		t.Error("none skeleton must be empty")
	}
}

func TestModesIncludesFeatureFirst(t *testing.T) {
	if len(Modes) == 0 || Modes[0] != "feature" {
		t.Errorf("first mode must be feature, got %v", Modes)
	}
}

func TestTemplatesContainsExpectedSet(t *testing.T) {
	want := []string{"none", "auth", "authz", "pagination", "rate-limit", "caching", "audit-log", "algorithm-custom"}
	got := map[string]bool{}
	for _, t := range Templates {
		got[t.ID] = true
	}
	for _, w := range want {
		if !got[w] {
			t.Errorf("missing template %q", w)
		}
	}
}

func TestAuthTemplateRequiresLifetimeQuestion(t *testing.T) {
	auth := LookupTemplate("auth")
	found := false
	for _, q := range auth.ExtraQuestions {
		if q.Key == "auth_token_lifetime" && q.Required {
			found = true
		}
	}
	if !found {
		t.Error("auth template must require auth_token_lifetime")
	}
}

func TestSkeletonAuthMentionsTokenLifetime(t *testing.T) {
	if !strings.Contains(SkeletonFor("auth"), "Token lifetime") {
		t.Error("auth skeleton must mention Token lifetime")
	}
}
