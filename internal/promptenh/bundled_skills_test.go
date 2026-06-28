package promptenh

import (
	"strings"
	"testing"

	"github.com/ropeixoto/harnessx/internal/domain"
)

func TestBundledSkillsAutoPinHumanizeForAllModes(t *testing.T) {
	src := BundledSkills{}
	list, err := src.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(list) == 0 {
		t.Fatal("expected bundled skills")
	}
	enhanced := Enhance("add a /healthz", domain.ModeFeature, nil, src)
	if !strings.Contains(enhanced.Enhanced, "humanize") {
		t.Errorf("humanize must be auto-applied to feature mode, got:\n%s", enhanced.Enhanced)
	}
	if !strings.Contains(enhanced.Enhanced, "frontend") {
		t.Errorf("frontend (mode *) must be auto-applied, got:\n%s", enhanced.Enhanced)
	}
}

func TestBundledSkillsDesignOnlyForDesignToProductMode(t *testing.T) {
	src := BundledSkills{}
	featureEnh := Enhance("x", domain.ModeFeature, nil, src)
	designEnh := Enhance("x", domain.ModeDesignToProduct, nil, src)
	if strings.Contains(featureEnh.Enhanced, "Skill: design") {
		t.Error("design skill must NOT apply to feature mode")
	}
	if !strings.Contains(designEnh.Enhanced, "Skill: design") {
		t.Errorf("design skill must apply to design_to_product mode, got:\n%s", designEnh.Enhanced)
	}
}
