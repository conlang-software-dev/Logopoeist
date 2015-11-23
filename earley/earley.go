package earley

import . "github.com/conlang-software-dev/Logopoeist/parser"
import . "github.com/conlang-software-dev/Logopoeist/grammar"
import . "github.com/conlang-software-dev/Logopoeist/types"
import . "github.com/conlang-software-dev/Logopoeist/environment"

type state struct {
	lhs      string
	rhs      []*Node
	dot      uint
	start    uint
	terminal bool
	weight   float64
}

func (s *state) iscomplete() bool {
	return s.dot >= uint(len(s.rhs))
}

func (s *state) needNonTerminal() bool {
	return s.rhs[s.dot].Type == SVar
}

func (s *state) equals(other *state) bool {
	if s.terminal != other.terminal ||
		s.lhs != other.lhs ||
		s.dot != other.dot ||
		s.start != other.start {
		return false
	}
	if len(s.rhs) != len(other.rhs) {
		return false
	}
	for i, n := range s.rhs {
		o := other.rhs[i]
		if n.Type != o.Type || n.Value != o.Value {
			return false
		}
	}
	return true
}

type EarleyParser struct {
	parent   *EarleyParser
	level    uint
	synmodel Grammar
	env      Environment
	root     string
	column   []*state
	finished bool
}

func NewParser(env Environment, g Grammar, root string) *EarleyParser {
	np := &EarleyParser{
		parent:   nil,
		level:    0,
		env:      env,
		synmodel: g,
		root:     root,
		column:   []*state{},
		finished: false,
	}

	np.init()
	return np
}

func newLevel(p *EarleyParser) *EarleyParser {
	return &EarleyParser{
		parent:   p,
		level:    p.level + 1,
		env:      p.env,
		synmodel: p.synmodel,
		root:     p.root,
		column:   []*state{},
		finished: false,
	}
}

func (p *EarleyParser) init() {
	if rset, ok := p.synmodel.Rules(p.root); ok {
		for i, rhs := range rset.Rules {
			p.addToChart(&state{
				lhs:      p.root,
				rhs:      rhs,
				dot:      0,
				start:    0,
				terminal: false,
				weight:   rset.Weights[i],
			})
		}
	}
	p.process()
}

func (p *EarleyParser) IsFinished() bool {
	return p.finished
}

func (p *EarleyParser) IsEmpty() bool {
	return len(p.column) == 0
}

func (p *EarleyParser) addToChart(s *state) {
	for _, old := range p.column {
		if s.equals(old) {
			old.weight += s.weight
			return
		}
	}
	p.column = append(p.column, s)
}

func (p *EarleyParser) getColumn(index uint) []*state {
	for p.level > index {
		p = p.parent
	}
	return p.column
}

func (chart *EarleyParser) scan(s *state, token string) {
	if s.iscomplete() {
		return
	}

	term := s.rhs[s.dot]
	if term.Type != CVar {
		return
	}

	chars, ok := chart.env.Lookup(term.Value)
	if !ok {
		return
	}

	if chars.Contains(token) {
		chart.addToChart(&state{
			lhs:      term.Value,
			rhs:      []*Node{}, // could store the token here, but it's not necessary for our purposes
			dot:      1,         // 0 would work as well, since rhs is empty; the point is to make this state "finished"
			start:    chart.level - 1,
			terminal: true,
			weight:   s.weight,
		})
	}
}

func (chart *EarleyParser) predict(s *state) {
	g := chart.synmodel
	term := s.rhs[s.dot]
	if term.Type != SVar {
		return
	}
	if rset, ok := g.Rules(term.Value); ok {
		for i, rhs := range rset.Rules {
			chart.addToChart(&state{
				lhs:      term.Value,
				rhs:      rhs,
				dot:      0,
				start:    chart.level,
				terminal: false,
				weight:   s.weight * rset.Weights[i],
			})
		}
	}
}

func (chart *EarleyParser) complete(s *state) {
	for _, old := range chart.getColumn(s.start) {
		if old.iscomplete() {
			continue
		}
		term := old.rhs[old.dot]
		t := SVar
		if s.terminal {
			t = CVar
		}
		if term.Type == t && term.Value == s.lhs {
			chart.addToChart(&state{
				lhs:      old.lhs,
				rhs:      old.rhs,
				dot:      old.dot + 1,
				start:    old.start,
				terminal: false,
				weight:   s.weight,
			})
		}
	}
}

func (p *EarleyParser) process() {
	//can't range because p.column is altered during the loop
	for i := 0; i < len(p.column); i++ {
		s := p.column[i]
		if s.iscomplete() {
			if s.start == 0 && s.lhs == p.root {
				p.finished = true
			}
			p.complete(s)
		} else if s.needNonTerminal() {
			p.predict(s)
		}
	}
	//optional: filter out completed states to save memory
}

func (p *EarleyParser) Next(token string) (*EarleyParser, bool) {
	np := newLevel(p)
	for _, s := range p.column {
		np.scan(s, token)
	}

	np.process()
	return np, len(np.column) > 0
}

func (p *EarleyParser) TerminationProbability() float64 {
	done_weight := 0.0
	cont_weight := 0.0
	for _, s := range p.column {
		if s.iscomplete() {
			if s.start == 0 && s.lhs == p.root {
				done_weight += s.weight
			}
		} else {
			cont_weight += s.weight
		}
	}
	return done_weight / cont_weight
}

func (p *EarleyParser) AllowedTokens() *CharSet {
	cset := make(CharSet)
	for _, s := range p.column {
		if s.iscomplete() {
			continue
		}

		term := s.rhs[s.dot]
		if term.Type != CVar {
			continue
		}

		if sset, ok := p.env.Lookup(term.Value); ok {
			for k, v := range sset.Weights {
				if _, ok := cset[k]; ok {
					cset[k] += v * s.weight
				} else {
					cset[k] = v * s.weight
				}
			}
		}
	}
	return &cset
}
