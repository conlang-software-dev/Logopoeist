package parser

import "fmt"
import . "github.com/conlang-software-dev/Logopoeist/lexer"

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

type Node struct {
	Type  int
	Value string
	Left  *Node
	Right *Node
}

func (n * Node) ToString() string {
	if n == nil {
		return ""
	}
	switch n.Type {
	case Production:
		return fmt.Sprintf("%s -> %s\n", n.Left.ToString(), n.Right.ToString())
	case Definition:
		return fmt.Sprintf("%s = %s\n", n.Left.ToString(), n.Right.ToString())
	case Condition:
		return fmt.Sprintf("%s -> %s\n", n.Left.ToString(), n.Right.ToString())
	case Exclusion:
		return fmt.Sprintf("%s !> %s\n", n.Left.ToString(), n.Right.ToString())
	case SVar:
		return fmt.Sprintf("$%s", n.Value)
	case CVar:
		return fmt.Sprintf("#%s", n.Value)
	case Class:
		return fmt.Sprintf("<%s>", n.Left.ToString())
	case Seq:
		if n.Right == nil {
			return n.Left.ToString()
		}
		return fmt.Sprintf("%s %s", n.Left.ToString(), n.Right.ToString())
	case Freq:
		if n.Right.Value == "1" {
			return n.Left.ToString()
		}
		return fmt.Sprintf("%s *%s", n.Left.ToString(), n.Right.ToString())
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

func Parse(lex *Lexer) chan *Node {
	nodes := make(chan *Node)

	go func() {

		defer func() {
			if err := recover(); err != nil {
				close(nodes)
				fmt.Printf("Error: %s\n", err)
			}
		}()

		for {
			item, ok := lex.Peek()
			if !ok || item.Type == "EOF" {
				close(nodes)
				return
			}
			nodes <- parseCommand(lex)
		}
	}()

	return nodes
}
