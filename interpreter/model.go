package interpreter

import "fmt"
import "math/rand"
import "strconv"
import "strings"
import . "github.com/conlang-software-dev/Logopoeist/parser"
import . "github.com/conlang-software-dev/Logopoeist/grammar"
import . "github.com/conlang-software-dev/Logopoeist/types"


type ngrams map[string]*CharSet

type model struct {
	start    string
	nextvar  int
	cvars    Environment
	synmodel Grammar
	chrmodel ngrams
	excmodel ngrams
	rnd      *rand.Rand
}

func interpretNumber(n *Node) float64 {
	freq, err := strconv.ParseFloat(n.Value, 64)
	if err != nil {
		panic(fmt.Sprintf("Invalid numeric literal: %s", n.Value))
	}
	return freq
}

func interpretClass(n *Node) *CharSet {
	chars := make(CharSet)
	for sn := n.Left; sn != nil; sn = sn.Right {
		fnode := sn.Left
		phoneme := fnode.Left.Value
		freq := interpretNumber(fnode.Right)

		if _, ok := chars[phoneme]; ok {
			chars[phoneme] += freq
		} else {
			chars[phoneme] = freq
		}
	}
	return &chars
}

func (m *model) getClass(n *Node) *CharSet {
	switch n.Type {
	case CVar:
		if cset, ok := m.cvars[n.Value]; ok {
			return cset
		}
		panic(fmt.Sprintf("Variable #%s referenced before definition", n.Value))
	case Class:
		return interpretClass(n)
	default:
		panic(fmt.Sprintf("Invalid Node Type for Character Class: %s", n.ToString()))
	}
}

func (m *model) assign(cvar string, n *Node) {
	m.cvars[cvar] = m.getClass(n)
}

func (m *model) genvar() string {
	m.nextvar += 1
	return strconv.Itoa(m.nextvar)
}

func (m *model) addRule(svar string, n *Node) {
	freq := interpretNumber(n.Right)
	rule := make([]*Node, 0, 10)
	for sn := n.Left; sn != nil; sn = sn.Right {
		subst := sn.Left
		switch subst.Type {
		case SVar, CVar:
			rule = append(rule, subst)
		case Class:
			cset := interpretClass(subst)
			cvar := m.genvar()
			m.cvars[cvar] = cset
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

func (m *model) generateNgrams(cond_n *Node) []string {
	last_ngrams := make([]string, 1)

	sn := cond_n
	if sn.Left.Type == Boundary {
		last_ngrams[0] = "_"
		sn = sn.Right
	} else {
		last_ngrams[0] = ""
	}

	for ; sn != nil; sn = sn.Right {
		cset := m.getClass(sn.Left)
		next_ngrams := make([]string, 0, len(last_ngrams))
		for _, ngram := range last_ngrams {
			for chr := range *cset {
				new_ngram := ngram + chr
				next_ngrams = append(next_ngrams, new_ngram)
			}
		}
		last_ngrams = next_ngrams
	}

	return last_ngrams
}

func (m *model) addCondition(cond_n *Node, dist_n *Node) {
	dist := m.getClass(dist_n)

	for _, ngram := range m.generateNgrams(cond_n) {
		if ndist, ok := m.chrmodel[ngram]; ok {

			// copy the old map in case it was shared,
			union := make(CharSet, len(*ndist))
			m.chrmodel[ngram] = &union
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
			// reference a single common object as much as possible
			m.chrmodel[ngram] = dist
		}
	}
}

func (m *model) addExclusion(cond_n *Node, dist_n *Node) {
	eset := m.getClass(dist_n)

	for _, ngram := range m.generateNgrams(cond_n) {
		if cset, ok := m.excmodel[ngram]; ok {

			// create a new map in case the original was shared
			union := make(CharSet, len(*cset))
			m.chrmodel[ngram] = &union

			for k := range *cset {
				union[k] = 0
			}
			for k := range *eset {
				union[k] = 0
			}
		} else {
			// reference a single common object as much as possible
			m.excmodel[ngram] = eset
		}
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
		m.assign(n.Left.Value, n.Right)
	case Condition:
		m.addCondition(n.Left, n.Right)
	case Exclusion:
		m.addExclusion(n.Left, n.Right)
	}
}

func (m *model) lookup(cvar string) *CharSet {
	if cset, ok := m.cvars[cvar]; ok {
		return cset
	}
	panic(fmt.Sprintf("#%s is undefined", cvar))
}

func (m *model) slotDist(cvar string, clist []string) CharSet {
	sdist := m.lookup(cvar)
	ndist := make(CharSet, len(*sdist))
	for k, v := range *sdist {
		ndist[k] = v
	}

	// iterate over conditioning ngrams
	order := len(clist)
	for j := 1; j <= order; j++ {
		ngram := strings.Join(clist[order-j:order+1], "")

		// remove any exclusions
		if edist, ok := m.excmodel[ngram]; ok {
			for char, _ := range *edist {
				delete(ndist, char)
			}
		}

		// intersect with conditional distributions
		if cdist, ok := m.chrmodel[ngram]; ok {
			//fmt.Printf("Intersecting ngram %s\n", ngram)
			for char, nweight := range ndist {
				if cweight, ok := (*cdist)[char]; ok {
					//fmt.Printf("\tNew Weight for %s\n", char)
					ndist[char] = nweight * cweight
				} else {
					//fmt.Printf("\tRemoving %s\n", char)
					delete(ndist, char)
				}
			}
		}
	}
	return ndist
}

func (m *model) choose(dist CharSet) string {
	var total float64 = 0
	for _, weight := range dist {
		total += weight
	}

	s := m.rnd.Float64() * total
	var chr string
	for chr, weight := range dist {
		s -= weight
		if s <= 0 {
			return chr
		}
	}
	//We should never get here, but just in case...
	//Floating point math might cause problems
	return chr
}

func (m *model) Generate() string {
	slots := m.synmodel.Generate(m.start, m.rnd)
	clist := make([]string, 1, len(slots)+1)
	clist[0] = "_"

	for _, cvar := range slots {
		ndist := m.slotDist(cvar, clist)
		nchar := m.choose(ndist)
		clist = append(clist, nchar)
	}

	return strings.Join(clist[1:], "")
}
