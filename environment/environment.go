package environment

import "fmt"
import "strconv"
import . "github.com/conlang-software-dev/Logopoeist/parser"
import . "github.com/conlang-software-dev/Logopoeist/types"
import . "github.com/conlang-software-dev/Logopoeist/interpreter"

type Environment map[string]*CharClass

func (e Environment) Assign(varname string, n *Node) {
	e[varname] = e.GetClass(n)
}

var nextvar = 0

func (e Environment) AssignNew(n *Node) string {
	nextvar += 1
	varname := strconv.Itoa(nextvar)
	e[varname] = e.GetClass(n)
	return varname
}

func (e Environment) Lookup(varname string) (*CharClass, bool) {
	if cclass, ok := e[varname]; ok {
		return cclass, true
	}
	return nil, false
}

func (e Environment) GetClass(n *Node) *CharClass {
	switch n.Type {
	case CVar:
		if cclass, ok := e.Lookup(n.Value); ok {
			return cclass
		}
		panic(fmt.Sprintf("Variable #%s referenced before definition", n.Value))
	case Class:
		return InterpretClass(n)
	default:
		panic(fmt.Sprintf("Invalid Node Type for Character Class: %s", n.ToString()))
	}
}