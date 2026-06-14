package i18n

import "testing"

func TestT_FallbackToEnglish(t *testing.T) {
	if got := T("doctor.title"); got != "HarnessX Doctor" {
		t.Fatalf("expected English fallback, got %q", got)
	}
}

func TestT_MissingKeyReturnsKey(t *testing.T) {
	if got := T("does.not.exist"); got != "does.not.exist" {
		t.Fatalf("expected key passthrough, got %q", got)
	}
}

func TestSetLocale_PortugueseRoundTrip(t *testing.T) {
	if !SetLocale("pt") {
		t.Fatal("pt bundle should be loaded")
	}
	defer SetLocale("en")
	if got := T("doctor.section.tools"); got != "Ferramentas" {
		t.Fatalf("expected pt translation, got %q", got)
	}
}

func TestSetLocale_Unknown(t *testing.T) {
	if SetLocale("xx") {
		t.Fatal("unknown locale must return false")
	}
}

func TestAvailableLocales(t *testing.T) {
	got := AvailableLocales()
	if len(got) < 2 {
		t.Fatalf("expected at least en + pt, got %v", got)
	}
}

func TestNormalise(t *testing.T) {
	if normalise("pt_BR.UTF-8") != "pt" {
		t.Fatal("normalise must trim locale modifiers")
	}
}
