package grammar

import "math/rand"
import . "github.com/conlang-software-dev/Logopoeist/parser"

type ruleSet struct {
	total   float64
	weights []float64
	rules   [][]*Node
}

type Grammar map[string]*ruleSet

func (g Grammar) AddRule(v string, rule []*Node, weight float64) {
	if rset, ok := g[v]; ok {
		rset.total += weight
		rset.rules = append(rset.rules, rule)
		rset.weights = append(rset.weights, weight)
	} else {
		g[v] = &ruleSet{
			total:   weight,
			weights: []float64{weight},
			rules:   [][]*Node{rule},
		}
	}
}

func (g Grammar) Rules(v string) (*ruleSet, bool) {
	if ruleset, ok := g[v]; ok {
		return ruleset, true
	}
	return &ruleSet{}, false
}

type Generator struct {
	g       Grammar
	rnd     *rand.Rand
	symbols []*Node
}

func (g Grammar) Generator(start string, rnd *rand.Rand) *Generator {
	return &Generator{
		g:       g,
		rnd:     rnd,
		symbols: []*Node{&Node{Type: SVar, Value: start}},
	}
}

func (gen *Generator) IsFinished() bool {
	return len(gen.symbols) == 0
}

func (gen *Generator) Terminal() (string, bool) {
	if len(gen.symbols) > 0 && gen.symbols[0].Type == CVar {
		return gen.symbols[0].Value, true
	}
	return "", false
}

type Selector struct {
	g       Grammar
	rnd     *rand.Rand
	total   float64
	rules   [][]*Node
	wmap    map[int]float64
	symbols []*Node
}

func (gen *Generator) Selector() (*Selector, bool) {
	if gen.IsFinished() {
		return nil, false
	}

	first := gen.symbols[0]
	rest := gen.symbols[1:]

	if first.Type == CVar {
		return &Selector{
			g:       gen.g,
			rnd:     gen.rnd,
			total:   1.0,
			rules:   [][]*Node{rest},
			wmap:    map[int]float64{0: 1.0},
			symbols: []*Node{},
		}, true
	} else {
		ruleset, _ := gen.g.Rules(first.Value)
		weights := ruleset.weights
		wmap := make(map[int]float64, len(weights))
		for i, w := range weights {
			wmap[i] = w
		}

		return &Selector{
			g:       gen.g,
			rnd:     gen.rnd,
			total:   ruleset.total,
			rules:   ruleset.rules,
			wmap:    wmap,
			symbols: rest,
		}, true
	}
}

func (sel *Selector) Next() (*Generator, bool) {
	if len(sel.wmap) == 0 {
		return nil, false
	}

	index := 0
	r := sel.rnd.Float64() * sel.total
	for i, w := range sel.wmap {
		r -= w
		if r <= 0 {
			index = i
			break
		}
	}
	sel.total -= sel.wmap[index]
	delete(sel.wmap, index)

	symbols := append(sel.rules[index], sel.symbols...)

	return &Generator{
		g:       sel.g,
		rnd:     sel.rnd,
		symbols: symbols,
	}, true
}
