// SPDX-License-Identifier: MIT

package recall

import "testing"

func TestBM25ZeroOnNoOverlap(t *testing.T) {
	b := NewBM25()
	if got := b.Score("alpha", "beta gamma delta"); got != 0 {
		t.Errorf("zero overlap should score 0, got %v", got)
	}
}

func TestBM25NonZeroOnMatch(t *testing.T) {
	b := NewBM25()
	if got := b.Score("healthz", "implement healthz endpoint"); got <= 0 {
		t.Errorf("match should score >0, got %v", got)
	}
}

func TestBM25LengthNormalization(t *testing.T) {
	b := NewBM25()
	short := "healthz endpoint added"
	long := short + " " + repeatStr("filler text words", 50)
	if b.Score("healthz", short) <= b.Score("healthz", long) {
		t.Errorf("shorter doc with same matches should score higher")
	}
}

func TestBM25EmptyQueryIsZero(t *testing.T) {
	b := NewBM25()
	if got := b.Score("", "anything"); got != 0 {
		t.Errorf("empty query should be 0, got %v", got)
	}
}

func TestBM25EmptyDocIsZero(t *testing.T) {
	b := NewBM25()
	if got := b.Score("x", ""); got != 0 {
		t.Errorf("empty doc should be 0, got %v", got)
	}
}

func TestBM25TuneAdjustsParams(t *testing.T) {
	b := NewBM25().Tune(2.0, 0.5)
	if b.K1 != 2.0 || b.B != 0.5 {
		t.Errorf("Tune did not stick: %+v", b)
	}
}

func TestDefaultScorerIsBagOfWords(t *testing.T) {
	s := DefaultScorer()
	if _, ok := s.(BagOfWordsScorer); !ok {
		t.Errorf("default should be BagOfWordsScorer, got %T", s)
	}
}

func TestBagOfWordsScorerScores(t *testing.T) {
	s := BagOfWordsScorer{}
	if got := s.Score("healthz fix", "fix healthz issue"); got <= 0 {
		t.Errorf("expected >0, got %v", got)
	}
}

func repeatStr(s string, n int) string {
	out := ""
	for i := 0; i < n; i++ {
		out += " " + s
	}
	return out
}
