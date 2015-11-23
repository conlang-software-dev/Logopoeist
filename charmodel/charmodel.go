package charmodel

import "strings"
import . "github.com/conlang-software-dev/Logopoeist/types"

type ngrams map[string]*CharSet

type CharModel struct {
	conds ngrams
	excls ngrams
}

func NewModel() *CharModel {
	return &CharModel{
		conds: make(ngrams),
		excls: make(ngrams),
	}
}

func (m *CharModel) AddCondition(ngram string, dist *CharSet) {
	if ndist, ok := m.conds[ngram]; ok {
		// copy the old map in case it was shared,
		union := make(CharSet, len(*ndist))
		m.conds[ngram] = &union
		for k, v := range *ndist {
			union[k] = v
		}

		// then union with the current distribution
		for k, v := range *dist {
			if _, ok := union[k]; ok {
				union[k] += v
			} else {
				union[k] = v
			}
		}
	} else {
		m.conds[ngram] = dist
	}
}

func (m *CharModel) AddExclusion(ngram string, dist *CharSet) {
	if edist, ok := m.excls[ngram]; ok {

		// create a new map in case the original was shared
		union := make(CharSet, len(*edist))
		m.excls[ngram] = &union

		for k := range *edist {
			union[k] = 0
		}
		for k := range *dist {
			union[k] = 0
		}
	} else {
		// reference a single common object as much as possible
		m.excls[ngram] = dist
	}
}

func (m *CharModel) CalcDistribution(base *CharSet, context []string) CharSet {
	ndist := make(CharSet, len(*base))
	for k, v := range *base {
		ndist[k] = v
	} 

	// iterate over conditioning ngrams
	order := len(context)
	for j := order; j > 0; j-- {
		ngram := strings.Join(context[order-j:order], "")

		// remove any exclusions
		if edist, ok := m.excls[ngram]; ok {
			for char, _ := range *edist {
				delete(ndist, char)
			}
		}

		// intersect with conditional distributions
		if cdist, ok := m.conds[ngram]; ok {
			for char, nweight := range ndist {
				if cweight, ok := (*cdist)[char]; ok {
					ndist[char] = nweight * cweight
				} else {
					delete(ndist, char)
				}
			}
		}
	}
	return ndist
}