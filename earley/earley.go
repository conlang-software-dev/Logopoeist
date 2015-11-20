package earley

import . "github.com/conlang-software-dev/Logopoeist/parser"
import . "github.com/conlang-software-dev/Logopoeist/grammar"
import . "github.com/conlang-software-dev/Logopoeist/types"

type state struct {
	lhs      string
	rhs      []*Node
	dot      uint
	start    uint
	terminal bool
}

func (s *state) iscomplete() bool {
	return s.dot >= uint(len(s.rhs))
}

func (s *state) needNonTerminal(g Grammar) bool {
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

func (s *state) scan(chart *parser, token string) {
	term := s.rhs[s.dot]
	if term.Type != CVar {
		return
	}

	chars := chart.cvars[term.Value]
	if _, ok := (*chars)[token]; ok {
		chart.addToChart(&state{
			lhs:      term.Value,
			rhs:      []*Node{}, // could store the token here, but it's not necessary for our purposes
			dot:      1,         // 0 would work as well, since rhs is empty; the point is to make this state "finished"
			start:    chart.level - 1,
			terminal: true,
		})
	}
}

func (s *state) predict(chart *parser) {
	g := chart.synmodel
	term := s.rhs[s.dot]
	if term.Type != SVar {
		return
	}
	if rules, ok := g.Rules(term.Value); ok {
		for _, rhs := range rules {
			chart.addToChart(&state{
				lhs:      term.Value,
				rhs:      rhs,
				dot:      0,
				start:    chart.level,
				terminal: false,
			})
		}
	}
}

func (s *state) complete(chart *parser) {
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
			})
		}
	}
}

type parser struct {
	parent   *parser
	level    uint
	synmodel Grammar
	cvars    Environment
	root     string
	column   []*state
	finished bool
}

func EarleyParser(cv Environment, g Grammar, root string) *parser {
	np := &parser{
		parent:   nil,
		level:    0,
		cvars:    cv,
		synmodel: g,
		root:     root,
		column:   []*state{},
		finished: false,
	}

	np.init()
	return np
}

func newLevel(p *parser) *parser {
	return &parser{
		parent:   p,
		level:    p.level + 1,
		cvars:    p.cvars,
		synmodel: p.synmodel,
		root:     p.root,
		column:   []*state{},
		finished: false,
	}
}

func (p *parser) init() {
	if rules, ok := p.synmodel.Rules(p.root); ok {
		for _, rhs := range rules {
			p.addToChart(&state{
				lhs:      p.root,
				rhs:      rhs,
				dot:      0,
				start:    0,
				terminal: false,
			})
		}
	}
	p.process()
}

func (p *parser) IsFinished() bool {
	return p.finished
}

func (p *parser) IsEmpty() bool {
	return len(p.column) == 0
}

func (p *parser) addToChart(s *state) {
	for _, old := range p.column {
		if s.equals(old) {
			return
		}
	}
	p.column = append(p.column, s)
}

func (p *parser) getColumn(index uint) []*state {
	for p.level > index {
		p = p.parent
	}
	return p.column
}

func (p *parser) process() {
	//can't range because p.column is altered during the loop
	for i := 0; i < len(p.column); i++ {
		s := p.column[i]
		if s.iscomplete() {
			if s.lhs == p.root {
				p.finished = true
			}
			s.complete(p)
		} else if s.needNonTerminal(p.synmodel) {
			s.predict(p)
		}
	}
	//optional: filter out completed states to save memory
}

func (p *parser) Next(token string) (*parser, bool) {
	np := newLevel(p)
	for _, s := range p.column {
		if !s.iscomplete() && !s.needNonTerminal(p.synmodel) {
			s.scan(np, token)
		}
	}

	if len(np.column) > 0 {
		np.process()
		return np, true
	}

	return np, false
}
