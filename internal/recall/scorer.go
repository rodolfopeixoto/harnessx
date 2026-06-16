// SPDX-License-Identifier: MIT

package recall

type Scorer interface {
	Score(query, doc string) float64
}

type BagOfWordsScorer struct{}

func (BagOfWordsScorer) Score(query, doc string) float64 {
	qTerms := tokenise(query)
	s, _ := scoreReport(doc, qTerms)
	return s
}

func DefaultScorer() Scorer { return BagOfWordsScorer{} }
