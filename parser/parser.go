package parser

import "fmt"
import . "github.com/conlang-software-dev/Logopoeist/lexer"

type Node struct {
	Type  uint8
	Value string
	Left  *Node
	Right *Node
}

func Parse(lex *Lexer, start func(*Lexer) *Node) chan *Node {
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
			nodes <- start(lex)
		}
	}()

	return nodes
}
