package interpreter

import "fmt"
import "time"
import "math/rand"
import . "github.com/conlang-software-dev/Logopoeist/parser"

const ( // Node Types
	Production = iota
	Definition
	Condition
	Exclusion
	SVar
	CVar
	Class
	Phoneme
	Seq
	Freq
	Num
	Boundary
)

func toString(n *Node) string {
	if n == nil {
		return ""
	}
	switch n.Type {
	case Production:
		return fmt.Sprintf("%s -> %s\n", toString(n.Left), toString(n.Right))
	case Definition:
		return fmt.Sprintf("%s = %s\n", toString(n.Left), toString(n.Right))
	case Condition:
		return fmt.Sprintf("%s -> %s\n", toString(n.Left), toString(n.Right))
	case Exclusion:
		return fmt.Sprintf("%s !> %s\n", toString(n.Left), toString(n.Right))
	case SVar:
		return fmt.Sprintf("$%s", n.Value)
	case CVar:
		return fmt.Sprintf("#%s", n.Value)
	case Class:
		return fmt.Sprintf("<%s>", toString(n.Left))
	case Seq:
		if n.Right == nil {
			return toString(n.Left)
		}
		return fmt.Sprintf("%s %s", toString(n.Left), toString(n.Right))
	case Freq:
		if n.Right.Value == "1" {
			return toString(n.Left)
		}
		return fmt.Sprintf("%s *%s", toString(n.Left), toString(n.Right))
	case Num:
		return n.Value
	case Phoneme:
		return n.Value
	case Boundary:
		return "_"
	default:
		return "{unknown}"
	}
}

func Interpreter() *model {
	return &model{
		start:    "",
		cvars:    make(map[string]*charSet),
		synmodel: make(grammar),
		chrmodel: make(ngrams),
		excmodel: make(ngrams),
		rnd:      rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}
