// SPDX-License-Identifier: MIT

package taskgraph

type ReasonKind string

const (
	ReasonMissingDep         ReasonKind = "missing_dep"
	ReasonPreconditionFailed ReasonKind = "precondition_failed"
	ReasonOutOfScope         ReasonKind = "out_of_scope"
	ReasonSplit              ReasonKind = "split"
)

type Reason struct {
	Kind    ReasonKind
	Detail  string
	TaskIdx int
	Inject  Task
	Into    []Task
}

func Replan(graph []Task, r Reason) []Task {
	switch r.Kind {
	case ReasonMissingDep, ReasonPreconditionFailed:
		return InjectBefore(graph, r.TaskIdx, r.Inject)
	case ReasonOutOfScope:
		return Skip(graph, r.TaskIdx)
	case ReasonSplit:
		return Split(graph, r.TaskIdx, r.Into)
	}
	return graph
}

func InjectBefore(graph []Task, idx int, t Task) []Task {
	if idx < 0 || idx > len(graph) {
		return graph
	}
	out := make([]Task, 0, len(graph)+1)
	out = append(out, graph[:idx]...)
	out = append(out, t)
	out = append(out, graph[idx:]...)
	return out
}

func Skip(graph []Task, idx int) []Task {
	if idx < 0 || idx >= len(graph) {
		return graph
	}
	out := make([]Task, 0, len(graph)-1)
	out = append(out, graph[:idx]...)
	out = append(out, graph[idx+1:]...)
	return out
}

func Split(graph []Task, idx int, into []Task) []Task {
	if idx < 0 || idx >= len(graph) || len(into) == 0 {
		return graph
	}
	out := make([]Task, 0, len(graph)-1+len(into))
	out = append(out, graph[:idx]...)
	out = append(out, into...)
	out = append(out, graph[idx+1:]...)
	return out
}
