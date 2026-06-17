package memory

import (
	"context"
	"errors"
	"testing"

	"github.com/ropeixoto/harnessx/internal/adapters/sqlite"
)

func TestPromoteAcceptsEveryPaperKind(t *testing.T) {
	repo, _ := sqlite.Open(":memory:")
	defer repo.Close()
	adapter := sqlAdapter{db: repo.DB()}
	for _, k := range KnownKinds() {
		_, err := Promote(context.Background(), repo, Candidate{
			Scope: "project", Kind: k,
			Content:       "kind " + k,
			EvidenceRunID: "run-1", Confidence: 0.8,
		}, adapter)
		if err != nil {
			t.Errorf("kind %s rejected: %v", k, err)
		}
	}
}

func TestPromoteRejectsUnknownKind(t *testing.T) {
	repo, _ := sqlite.Open(":memory:")
	defer repo.Close()
	_, err := Promote(context.Background(), repo, Candidate{
		Scope: "project", Kind: "not-a-kind",
		Content: "x", EvidenceRunID: "r", Confidence: 0.7,
	}, sqlAdapter{db: repo.DB()})
	if !errors.Is(err, ErrUnknownKind) {
		t.Fatalf("want ErrUnknownKind, got %v", err)
	}
}

func TestPromoteDefaultsEmptyKindToSemantic(t *testing.T) {
	repo, _ := sqlite.Open(":memory:")
	defer repo.Close()
	m, err := Promote(context.Background(), repo, Candidate{
		Scope: "project", Kind: "",
		Content: "x", EvidenceRunID: "r", Confidence: 0.7,
	}, sqlAdapter{db: repo.DB()})
	if err != nil {
		t.Fatal(err)
	}
	if m.Kind != KindSemantic {
		t.Errorf("default kind: want semantic got %q", m.Kind)
	}
}
