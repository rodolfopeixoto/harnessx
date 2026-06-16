// SPDX-License-Identifier: MIT

package taskgraph

import "testing"

func TestDecompose_SingleClause(t *testing.T) {
	tasks := Decompose("scaffold python", Options{})
	if len(tasks) != 1 {
		t.Fatalf("want 1 task, got %d", len(tasks))
	}
	if tasks[0].Kind != KindScaffold {
		t.Fatalf("want scaffold, got %s", tasks[0].Kind)
	}
	if tasks[0].Lang != "python" {
		t.Fatalf("want python, got %q", tasks[0].Lang)
	}
}

func TestDecompose_MultipleClauses(t *testing.T) {
	tasks := Decompose("scaffold python and add a /healthz endpoint then generate a hero image", Options{})
	if len(tasks) != 3 {
		t.Fatalf("want 3 tasks, got %d: %+v", len(tasks), tasks)
	}
	want := []Kind{KindScaffold, KindCode, KindImage}
	for i, k := range want {
		if tasks[i].Kind != k {
			t.Fatalf("task %d: want %s, got %s", i, k, tasks[i].Kind)
		}
	}
}

func TestDecompose_Generic(t *testing.T) {
	tasks := Decompose("ponder the existential nature of unit tests", Options{})
	if tasks[0].Kind != KindGeneric {
		t.Fatalf("want generic, got %s", tasks[0].Kind)
	}
}

func TestDecompose_LintAndTest(t *testing.T) {
	tasks := Decompose("run the lint and run the tests", Options{})
	if len(tasks) != 2 {
		t.Fatalf("want 2, got %d", len(tasks))
	}
	if tasks[0].Kind != KindLint || tasks[1].Kind != KindTest {
		t.Fatalf("got %s, %s", tasks[0].Kind, tasks[1].Kind)
	}
}
