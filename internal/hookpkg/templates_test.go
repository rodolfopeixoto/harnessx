// SPDX-License-Identifier: MIT

package hookpkg

import (
	"strings"
	"testing"
)

func TestListReturnsBundled(t *testing.T) {
	got, err := List()
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]bool{
		"pre-tool-use-lint":    false,
		"pre-tool-use-secrets": false,
		"pre-tool-use-noforce": false,
		"post-tool-use-audit":  false,
		"post-tool-use-test":   false,
	}
	for _, tpl := range got {
		want[tpl.Name] = true
	}
	for name, ok := range want {
		if !ok {
			t.Errorf("missing bundled template: %s", name)
		}
	}
}

func TestLoadByName(t *testing.T) {
	tpl, err := Load("pre-tool-use-lint")
	if err != nil {
		t.Fatal(err)
	}
	if tpl.Name != "pre-tool-use-lint" {
		t.Errorf("name: got %q", tpl.Name)
	}
	if len(tpl.Body) == 0 {
		t.Error("body empty")
	}
	if tpl.Event == "" {
		t.Error("event missing")
	}
}

func TestLoadUnknownErrors(t *testing.T) {
	_, err := Load("does-not-exist")
	if err == nil {
		t.Error("expected error for unknown template")
	}
}

func TestInferEventFromName(t *testing.T) {
	cases := map[string]string{
		"pre-tool-use-lint":   "pre-tool-use",
		"post-tool-use-audit": "post-tool-use",
		"unknown-event":       "",
	}
	for name, want := range cases {
		got := inferEvent(name, nil)
		if got != want {
			t.Errorf("inferEvent(%q)=%q, want %q", name, got, want)
		}
	}
}

func TestInferDescriptionExtractsFromComment(t *testing.T) {
	body := []byte("#!/bin/sh\n# description: protect against force-push\nexit 0\n")
	if got := inferDescription(body); !strings.Contains(got, "force-push") {
		t.Errorf("description: got %q", got)
	}
}

func TestInferDescriptionEmptyWhenAbsent(t *testing.T) {
	if got := inferDescription([]byte("#!/bin/sh\nexit 0\n")); got != "" {
		t.Errorf("empty body should yield empty description: got %q", got)
	}
}

func TestParseHeaderFindsPrefix(t *testing.T) {
	body := []byte("#!/bin/sh\n# event: pre-tool-use\n# description: x\nexit 0\n")
	if got := parseHeader(body, "# event:"); got != "pre-tool-use" {
		t.Errorf("parseHeader: got %q", got)
	}
}

func TestParseHeaderMissing(t *testing.T) {
	if got := parseHeader([]byte("nothing"), "# event:"); got != "" {
		t.Errorf("missing should be empty, got %q", got)
	}
}
