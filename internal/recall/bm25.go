// SPDX-License-Identifier: MIT

package recall

import "math"

type BM25 struct {
	K1     float64
	B      float64
	AvgLen float64
}

func NewBM25() *BM25 {
	return &BM25{K1: 1.5, B: 0.75, AvgLen: 200}
}

func (b *BM25) Tune(k1, bVal float64) *BM25 {
	b.K1 = k1
	b.B = bVal
	return b
}

func (b *BM25) Score(query, doc string) float64 {
	qTerms := tokenise(query)
	if len(qTerms) == 0 {
		return 0
	}
	dTerms := tokenise(doc)
	if len(dTerms) == 0 {
		return 0
	}
	tf := map[string]int{}
	for _, t := range dTerms {
		tf[t]++
	}
	dl := float64(len(dTerms))
	avg := b.AvgLen
	if avg <= 0 {
		avg = dl
	}
	var score float64
	for _, q := range qTerms {
		f := float64(tf[q])
		if f == 0 {
			continue
		}
		idf := math.Log(1 + 1.0/(f+0.5))
		num := f * (b.K1 + 1)
		den := f + b.K1*(1-b.B+b.B*dl/avg)
		score += idf * num / den
	}
	return score
}
