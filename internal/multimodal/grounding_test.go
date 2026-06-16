// SPDX-License-Identifier: MIT

package multimodal

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadAnnotationsRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "mock.image.json")
	body := `{"schema_version":1,"annotations":[{"region":{"x":10,"y":20,"w":100,"h":50},"label":"login button"}]}`
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := LoadAnnotations(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Label != "login button" {
		t.Errorf("got %+v", got)
	}
	if got[0].Region.W != 100 {
		t.Errorf("region: got %+v", got[0].Region)
	}
}

func TestLoadAnnotationsMissingFile(t *testing.T) {
	if _, err := LoadAnnotations("/tmp/__nope__.json"); err == nil {
		t.Error("expected error for missing file")
	}
}

func TestLoadAnnotationsBadJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.json")
	_ = os.WriteFile(path, []byte("{not json"), 0o644)
	if _, err := LoadAnnotations(path); err == nil {
		t.Error("expected parse error")
	}
}

func TestLoadAnnotationsUnsupportedSchema(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.json")
	_ = os.WriteFile(path, []byte(`{"schema_version":99,"annotations":[]}`), 0o644)
	if _, err := LoadAnnotations(path); err == nil {
		t.Error("expected schema error")
	}
}

func TestCheckGroundingAllHit(t *testing.T) {
	anns := []Annotation{{Label: "login button"}, {Label: "logo"}}
	res := CheckGrounding("implements the login button and refreshes the logo", anns)
	if len(res.Hits) != 2 || len(res.Missing) != 0 {
		t.Errorf("got %+v", res)
	}
}

func TestCheckGroundingNoneHit(t *testing.T) {
	anns := []Annotation{{Label: "X"}, {Label: "Y"}}
	res := CheckGrounding("nothing relevant here", anns)
	if len(res.Hits) != 0 || len(res.Missing) != 2 {
		t.Errorf("got %+v", res)
	}
}

func TestCheckGroundingPartial(t *testing.T) {
	anns := []Annotation{{Label: "navbar"}, {Label: "footer"}}
	res := CheckGrounding("adds a navbar but does not touch the bottom", anns)
	if len(res.Hits) != 1 || res.Hits[0] != "navbar" {
		t.Errorf("hits: %+v", res.Hits)
	}
	if len(res.Missing) != 1 || res.Missing[0] != "footer" {
		t.Errorf("missing: %+v", res.Missing)
	}
}

func TestCheckGroundingIgnoresEmptyLabel(t *testing.T) {
	anns := []Annotation{{Label: ""}, {Label: "logo"}}
	res := CheckGrounding("logo here", anns)
	if len(res.Hits) != 1 {
		t.Errorf("got %+v", res)
	}
}

func TestCheckGroundingCaseInsensitive(t *testing.T) {
	anns := []Annotation{{Label: "Login Button"}}
	res := CheckGrounding("the LOGIN BUTTON is wired", anns)
	if len(res.Hits) != 1 {
		t.Errorf("case-insensitive failed: %+v", res)
	}
}
