// SPDX-License-Identifier: MIT

package mcppkg

import "testing"

func TestListReturnsBundledMCPs(t *testing.T) {
	got, err := List()
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]bool{
		"filesystem":   false,
		"github":       false,
		"postgres":     false,
		"sqlite":       false,
		"brave-search": false,
		"fetch":        false,
		"memory":       false,
	}
	for _, name := range got {
		if _, ok := want[name]; ok {
			want[name] = true
		}
	}
	for name, ok := range want {
		if !ok {
			t.Errorf("missing bundled mcp: %s", name)
		}
	}
}

func TestLoadParsesYAML(t *testing.T) {
	tpl, err := Load("filesystem")
	if err != nil {
		t.Fatal(err)
	}
	if tpl.Name == "" {
		t.Error("name empty")
	}
}

func TestLoadUnknown(t *testing.T) {
	_, err := Load("does-not-exist")
	if err == nil {
		t.Error("expected error")
	}
}
