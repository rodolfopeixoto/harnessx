// SPDX-License-Identifier: MIT

package skillpkg

import "testing"

func TestListReturnsBundledSkills(t *testing.T) {
	got, err := List()
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]bool{
		"bugfix-loop":   false,
		"clean-code":    false,
		"go-feature":    false,
		"security-rule": false,
	}
	for _, tpl := range got {
		want[tpl.Name] = true
	}
	for name, ok := range want {
		if !ok {
			t.Errorf("missing bundled skill: %s", name)
		}
	}
}

func TestLoadByName(t *testing.T) {
	tpl, err := Load("clean-code")
	if err != nil {
		t.Fatal(err)
	}
	if tpl.Name != "clean-code" || len(tpl.Body) == 0 {
		t.Errorf("template malformed: %+v", tpl)
	}
}

func TestLoadUnknown(t *testing.T) {
	_, err := Load("does-not-exist")
	if err == nil {
		t.Error("expected error")
	}
}

func TestParseHeader(t *testing.T) {
	body := []byte("---\ndescription: clean code rules\n---\nbody...")
	if got := parseHeader(body, "description:"); got != "clean code rules" {
		t.Errorf("got %q", got)
	}
}
