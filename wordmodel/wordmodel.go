package wordmodel

import "time"
import "strings"
import "math/rand"
import . "github.com/conlang-software-dev/Logopoeist/parser"
import . "github.com/conlang-software-dev/Logopoeist/types"
import . "github.com/conlang-software-dev/Logopoeist/interpreter"
import . "github.com/conlang-software-dev/Logopoeist/environment"
import . "github.com/conlang-software-dev/Logopoeist/grammar"
import . "github.com/conlang-software-dev/Logopoeist/charmodel"
import . "github.com/conlang-software-dev/Logopoeist/earley"

type model struct {
	start    string
	nextvar  int
	env      Environment
	synmodel Grammar
	chrmodel *CharModel
	rnd      *rand.Rand
	words    map[string]struct{}
}

func (m *model) addRule(svar string, n *Node) {
	freq := InterpretNumber(n.Right)
	rule := make([]*Node, 0, 10)
	for sn := n.Left; sn != nil; sn = sn.Right {
		subst := sn.Left
		switch subst.Type {
		case SVar, CVar:
			rule = append(rule, subst)
		case Class:
			cvar := m.env.AssignNew(subst)
			rule = append(rule, &Node{
				Type:  CVar,
				Value: cvar,
			})
		default:
			panic("Invalid Node Type in Syntax Rule")
		}
	}

	m.synmodel.AddRule(svar, rule, freq)
}

func (m *model) generateNgrams(cond_n *Node) [][]string {
	var last_ngrams [][]string

	sn := cond_n
	if sn.Left.Type == Boundary {
		last_ngrams = append(last_ngrams, []string{"_"})
		sn = sn.Right
	} else {
		last_ngrams = append(last_ngrams, []string{})
	}

	for ; sn != nil; sn = sn.Right {
		cclass := m.env.GetClass(sn.Left)
		next_ngrams := make([][]string, 0, cap(last_ngrams)*len(cclass.List))
		for _, ngram := range last_ngrams {
			for _, chr := range cclass.List {
				new_ngram := append(ngram, chr)
				next_ngrams = append(next_ngrams, new_ngram)
			}
		}
		last_ngrams = next_ngrams
	}

	return last_ngrams
}

func (m *model) addCondition(cond_n *Node, dist_n *Node) {
	dist := m.env.GetClass(dist_n).Weights
	for _, ngchars := range m.generateNgrams(cond_n) {
		ngram := strings.Join(ngchars, "")
		m.chrmodel.AddCondition(ngram, &dist)
	}
}

func (m *model) addExclusion(cond_n *Node, dist_n *Node) {
	dist := m.env.GetClass(dist_n).Weights
	for _, ngchars := range m.generateNgrams(cond_n) {
		ngram := strings.Join(ngchars, "")
		m.chrmodel.AddExclusion(ngram, &dist)
	}
}

func (m *model) Execute(n *Node) {
	if n == nil || m == nil {
		return
	}
	switch n.Type {
	case Production:
		m.addRule(n.Left.Value, n.Right)
		if m.start == "" {
			m.start = n.Left.Value
		}
	case Definition:
		m.env.Assign(n.Left.Value, n.Right)
	case Condition:
		m.addCondition(n.Left, n.Right)
	case Exclusion:
		m.addExclusion(n.Left, n.Right)
	}
}

func (m *model) gen_rec(ep *EarleyParser, clist []string, min int, max int) ([]string, bool) {

	finalize := func() ([]string, bool) {
		final := clist[1:]
		word := strings.Join(final, "")
		if _, ok := m.words[word]; !ok {
			m.words[word] = struct{}{}
			return final, true
		}
		return nil, false
	}

	recurse := func() ([]string, bool) {
		if max > 0 && len(clist) > max {
			return nil, false
		}

		base := ep.AllowedTokens()
		dist := m.chrmodel.CalcDistribution(base, clist)

		total := 0.0
		for _, w := range dist {
			total += w
		}

		for len(dist) > 0 {
			r := m.rnd.Float64() * total
			for c, w := range dist {
				r -= w
				if r <= 0 {
					total -= w
					delete(dist, c)

					if np, ok := ep.Next(c); ok {
						if nclist, ok := m.gen_rec(np, append(clist, c), min, max); ok {
							return nclist, true
						}
					}
					break
				}
			}
		}
		return nil, false
	}

	if ep.IsFinished() && len(clist) > min {
		var attempt func() ([]string, bool)
		var fallback func() ([]string, bool)
		if m.rnd.Float64() < ep.TerminationProbability() {
			attempt = finalize
			fallback = recurse
		} else {
			attempt = recurse
			fallback = finalize
		}
		if nclist, ok := attempt(); ok {
			return nclist, true
		}
		return fallback()
	}

	return recurse()
}

func (m *model) Generate(min int, max int) ([]string, bool) {
	clist := make([]string, 1, 10)
	clist[0] = "_"

	ep := NewParser(m.env, m.synmodel, m.start)
	return m.gen_rec(ep, clist, min, max)
}

func WordModel() *model {
	return &model{
		start:    "",
		env:      make(Environment),
		synmodel: make(Grammar),
		chrmodel: NewModel(),
		rnd:      rand.New(rand.NewSource(time.Now().UnixNano())),
		words:    make(map[string]struct{}),
	}
}
