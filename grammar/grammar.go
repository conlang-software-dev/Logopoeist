package grammar

import . "github.com/conlang-software-dev/Logopoeist/parser"

type RuleSet struct {
	total   float64
	Weights []float64
	Rules   [][]*Node
}

type Grammar map[string]*RuleSet

func (g Grammar) AddRule(v string, rule []*Node, weight float64) {
	if rset, ok := g[v]; ok {
		rset.total += weight
		rset.Rules = append(rset.Rules, rule)
		rset.Weights = append(rset.Weights, weight)
	} else {
		g[v] = &RuleSet{
			total:   weight,
			Weights: []float64{weight},
			Rules:   [][]*Node{rule},
		}
	}
}

func (g Grammar) Rules(v string) (*RuleSet, bool) {
	if ruleset, ok := g[v]; ok {
		return ruleset, true
	}
	return &RuleSet{}, false
}