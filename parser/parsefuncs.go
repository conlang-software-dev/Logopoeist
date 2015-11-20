package parser

import "fmt"
import . "github.com/conlang-software-dev/Logopoeist/lexer"

func parseSVar(lex *Lexer) *Node {
	lex.Next() // skip $ sigil
	symbol, ok := lex.Next()
	if !ok || symbol.Type != "symbol" {
		panic("Parse error: Missing Syntax Variable")
	}
	return &Node{
		Type:  SVar,
		Value: symbol.Token,
	}
}

func parseCVar(lex *Lexer) *Node {
	lex.Next() // skip # sigil
	symbol, ok := lex.Next()
	if !ok || symbol.Type != "symbol" {
		panic("Parse error: Missing Class Variable")
	}
	return &Node{
		Type:  CVar,
		Value: symbol.Token,
	}
}

func parsePhoneme(lex *Lexer) *Node {
loop:
	symbol, ok := lex.Next()
	if !ok || symbol.Type == ">" {
		return nil
	}
	if symbol.Type == "EOL" {
		goto loop
	}

	frequency := parseFrequency(lex)
	rest := parsePhoneme(lex)
	phoneme := &Node{
		Type:  Phoneme,
		Value: symbol.Token,
	}

	return &Node{
		Type:  Seq,
		Value: "",
		Right: rest,
		Left: &Node{
			Type:  Freq,
			Left:  phoneme,
			Right: frequency,
		},
	}
}

func parseClass(lex *Lexer) *Node {
	lex.Next() // skip < token
	phonemes := parsePhoneme(lex)
	return &Node{
		Type: Class,
		Left: phonemes,
	}
}

func parseClassOrCVar(lex *Lexer) *Node {
	item, ok := lex.Peek()
	if !ok {
		panic("Parse error: Expected Character Class or Variable")
	}
	switch item.Type {
	case "#":
		return parseCVar(lex)
	case "<":
		return parseClass(lex)
	default:
		panic(fmt.Sprintf("Parse error: Expected Character Class or Variable; saw %s", item.Token))
	}
}

func parseSubstitutions(lex *Lexer) *Node {
	item, ok := lex.Peek()
	if !ok {
		return nil
	}

	var left *Node
	switch item.Type {
	case "*", "EOL":
		return nil
	case "$":
		left = parseSVar(lex)
	case "#":
		left = parseCVar(lex)
	case "<":
		left = parseClass(lex)
	default:
		panic(fmt.Sprintf("Parse error: Unexpected Token %s in Syntax Rule", item.Token))
	}

	right := parseSubstitutions(lex)
	return &Node{
		Type:  Seq,
		Left:  left,
		Right: right,
	}
}

func parseFrequency(lex *Lexer) *Node {
	item, ok := lex.Peek()
	if !ok || item.Type != "*" {
		return &Node{
			Type:  Num,
			Value: "1",
		}
	}

	lex.Next() // skip * token
	item, ok = lex.Next()
	if !ok || item.Type != "number" {
		panic("Parse error: Missing Number")
	}

	return &Node{
		Type:  Num,
		Value: item.Token,
	}
}

func parseSyntax(lex *Lexer) *Node {
	left := parseSVar(lex)

	arrow, ok := lex.Next()
	if !ok || arrow.Token != "->" {
		panic("Parse error: Expected -> in syntax definition")
	}

	substitutions := parseSubstitutions(lex)
	frequency := parseFrequency(lex)

	return &Node{
		Type:  Production,
		Value: "",
		Left:  left,
		Right: &Node{
			Type:  Freq,
			Left:  substitutions,
			Right: frequency,
		},
	}
}

func parseCondList(lex *Lexer) *Node {
	item, ok := lex.Peek()
	if !ok {
		return nil
	}

	var left *Node
	switch item.Type {
	case "EOL", "EOF", "arrow":
		return nil
	case "#":
		left = parseCVar(lex)
	case "<":
		left = parseClass(lex)
	default:
		panic(fmt.Sprintf("Parse error: Unexpected Token %s in Condition Expression", item.Token))
	}

	right := parseCondList(lex)
	return &Node{
		Type:  Seq,
		Left:  left,
		Right: right,
	}
}

func parseCondOrDef(lex *Lexer) *Node {
	var first *Node

	item, ok := lex.Peek()
	if !ok {
		panic("Invalid call to parseCondOrDef")
	}

	switch item.Type {
	case "#":
		first = parseCVar(lex)
	case "<":
		first = parseClass(lex)
	case "_":
		lex.Next()
		first = &Node{Type: Boundary}
	default:
		panic("Invalid call to parseCondOrDef")
	}

	item, ok = lex.Peek()
	if !ok || item.Type == "EOL" {
		if first.Type == CVar {
			panic("Parse error: Incomplete Variable Definition")
		}
		panic("Parse error: Incomplete Condition Expression")
	}

	if item.Type == "=" {
		if first.Type != CVar {
			panic("Parse error: Unexpected _")
		}

		lex.Next() // skip = token
		second := parseClassOrCVar(lex)
		return &Node{
			Type:  Definition,
			Left:  first,
			Right: second,
		}
	} else {
		rest := parseCondList(lex)

		arrow, ok := lex.Next()
		if !ok {
			panic("Parse error: Missing Arrow in Condition Expression")
		}

		right := parseClassOrCVar(lex)
		left := &Node{
			Type:  Seq,
			Left:  first,
			Right: rest,
		}

		switch arrow.Token {
		case "->":
			return &Node{
				Type:  Condition,
				Left:  left,
				Right: right,
			}
		case "!>":
			return &Node{
				Type:  Exclusion,
				Left:  left,
				Right: right,
			}
		default:
			panic("Parse error: Invalid Arrow in Condition Expression")
		}
	}
}

func parseCommand(lex *Lexer) *Node {
	item, ok := lex.Peek()
	for ok && item.Type != "EOF" {
		switch item.Type {
		case "#", "_", "<":
			return parseCondOrDef(lex)
		case "$":
			return parseSyntax(lex)
		case "EOL":
			lex.Next()
			item, ok = lex.Peek()
		default:
			panic(fmt.Sprintf("Parse error: Unexpected Token %s", item.Token))
		}
	}
	return nil
}
