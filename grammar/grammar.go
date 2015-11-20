package grammar

import "math/rand"
import . "github.com/conlang-software-dev/Logopoeist/parser"

type RuleSet struct {
	total   float64
	weights []float64
	rules   [][]*Node
}

type Grammar map[string]*RuleSet

func (g Grammar) AddRule(v string, rule []*Node, weight float64) {
	if rset, ok := g[v]; ok {
		rset.total += weight
		rset.rules = append(rset.rules, rule)
		rset.weights = append(rset.weights, weight)
	} else {
		g[v] = &RuleSet{
			total:   weight,
			weights: []float64{weight},
			rules:   [][]*Node{rule},
		}
	}
}

func (g Grammar) choose(v string, rnd *rand.Rand) []*Node {
	ruleset := g[v]
	s := rnd.Float64() * ruleset.total
	var rule ([]*Node)
	for i, rule := range ruleset.rules {
		s -= ruleset.weights[i]
		if s <= 0 {
			return rule
		}
	}
	//We should never get here, but just in case...
	//Floating point math might cause problems
	return rule
}

func (g Grammar) Rules(v string) ([][]*Node, bool) {
	if ruleset, ok := g[v]; ok {
		return ruleset.rules, true
	}
	return [][]*Node{}, false
}

func (g Grammar) Generate(start string, rnd *rand.Rand) []string {
	symbols := g.choose(start, rnd)
	slots := make([]string, 0, 10)
	for len(symbols) > 0 {
		sym := symbols[0]
		switch sym.Type {
		case SVar:
			replace := g.choose(sym.Value, rnd)
			symbols = append(replace, symbols[1:]...)
		case CVar:
			slots = append(slots, sym.Value)
			symbols = symbols[1:]
		default:
			panic("Invalid Node Type in Syntax Rule")
		}
	}
	return slots
}
