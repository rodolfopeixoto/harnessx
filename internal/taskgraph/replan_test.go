// SPDX-License-Identifier: MIT

package taskgraph

import "testing"

func mkTasks(kinds ...string) []Task {
	out := make([]Task, len(kinds))
	for i, k := range kinds {
		out[i] = Task{Kind: Kind(k), Prompt: k}
	}
	return out
}

func TestInjectBeforeAtStart(t *testing.T) {
	in := mkTasks("a", "b")
	got := InjectBefore(in, 0, Task{Kind: "pre"})
	if len(got) != 3 || got[0].Kind != "pre" {
		t.Errorf("got %+v", got)
	}
}

func TestInjectBeforeAtMiddle(t *testing.T) {
	in := mkTasks("a", "b", "c")
	got := InjectBefore(in, 1, Task{Kind: "pre"})
	if got[1].Kind != "pre" || got[2].Kind != "b" {
		t.Errorf("got %+v", got)
	}
}

func TestInjectBeforeAtEnd(t *testing.T) {
	in := mkTasks("a")
	got := InjectBefore(in, 1, Task{Kind: "post"})
	if len(got) != 2 || got[1].Kind != "post" {
		t.Errorf("got %+v", got)
	}
}

func TestInjectBeforeOutOfRange(t *testing.T) {
	in := mkTasks("a")
	got := InjectBefore(in, 5, Task{Kind: "x"})
	if len(got) != 1 {
		t.Errorf("OOR should noop, got %+v", got)
	}
}

func TestSkipMiddle(t *testing.T) {
	in := mkTasks("a", "b", "c")
	got := Skip(in, 1)
	if len(got) != 2 || got[0].Kind != "a" || got[1].Kind != "c" {
		t.Errorf("got %+v", got)
	}
}

func TestSkipOutOfRange(t *testing.T) {
	in := mkTasks("a")
	got := Skip(in, 5)
	if len(got) != 1 {
		t.Errorf("OOR should noop, got %+v", got)
	}
}

func TestSplitReplacesIdx(t *testing.T) {
	in := mkTasks("a", "broad", "c")
	got := Split(in, 1, mkTasks("part1", "part2"))
	if len(got) != 4 {
		t.Fatalf("got %+v", got)
	}
	if got[1].Kind != "part1" || got[2].Kind != "part2" || got[3].Kind != "c" {
		t.Errorf("layout: got %+v", got)
	}
}

func TestSplitEmptyIntoNoops(t *testing.T) {
	in := mkTasks("a")
	got := Split(in, 0, nil)
	if len(got) != 1 {
		t.Errorf("empty into should noop, got %+v", got)
	}
}

func TestReplanMissingDepInjects(t *testing.T) {
	in := mkTasks("test")
	got := Replan(in, Reason{Kind: ReasonMissingDep, TaskIdx: 0, Inject: Task{Kind: "install"}})
	if len(got) != 2 || got[0].Kind != "install" {
		t.Errorf("got %+v", got)
	}
}

func TestReplanOutOfScopeSkips(t *testing.T) {
	in := mkTasks("a", "skipme", "c")
	got := Replan(in, Reason{Kind: ReasonOutOfScope, TaskIdx: 1})
	if len(got) != 2 || got[0].Kind != "a" || got[1].Kind != "c" {
		t.Errorf("got %+v", got)
	}
}

func TestReplanSplitDispatches(t *testing.T) {
	in := mkTasks("a", "broad")
	got := Replan(in, Reason{Kind: ReasonSplit, TaskIdx: 1, Into: mkTasks("p1", "p2")})
	if len(got) != 3 || got[1].Kind != "p1" || got[2].Kind != "p2" {
		t.Errorf("got %+v", got)
	}
}

func TestReplanUnknownReturnsUnchanged(t *testing.T) {
	in := mkTasks("a", "b")
	got := Replan(in, Reason{Kind: "bogus"})
	if len(got) != 2 {
		t.Errorf("unknown reason should noop, got %+v", got)
	}
}
