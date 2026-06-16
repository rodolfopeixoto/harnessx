// SPDX-License-Identifier: MIT

package autonomy

const minEvidence = 5

type Proposal struct {
	Path          string
	From          string
	To            string
	Reason        string
	EvidenceCount int
}

func Suggest(events []Event) []Proposal {
	byPath := map[string]map[string]int{}
	for _, e := range events {
		if _, ok := byPath[e.Path]; !ok {
			byPath[e.Path] = map[string]int{}
		}
		byPath[e.Path][e.Decision]++
	}
	var out []Proposal
	for path, counts := range byPath {
		denied := counts["deny"]
		approved := counts["approve"]
		askPending := counts["require_approval"]
		switch {
		case denied >= minEvidence && approved == 0:
			out = append(out, Proposal{
				Path: path, From: "require_approval", To: "deny",
				Reason:        "consistently denied",
				EvidenceCount: denied,
			})
		case approved >= minEvidence && denied == 0 && askPending >= minEvidence:
			out = append(out, Proposal{
				Path: path, From: "require_approval", To: "allow",
				Reason:        "consistently approved",
				EvidenceCount: approved,
			})
		case approved >= minEvidence && denied >= 1:
			out = append(out, Proposal{
				Path: path, From: "deny", To: "require_approval",
				Reason:        "mixed approvals after deny — revisit",
				EvidenceCount: approved,
			})
		}
	}
	return out
}
