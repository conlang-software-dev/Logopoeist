package main

import "bytes"
import "strings"
import . "github.com/conlang-software-dev/Logopoeist/lexer"

func commentState(in *RuneBuffer, out chan *Item) StateFn {
	for {
		r, ok := in.Next()
		if !ok || r == '\n' {
			break
		}
	}
	return switchState
}

func numberState(in *RuneBuffer, out chan *Item) StateFn {
	buf := new(bytes.Buffer)
	for {
		r, ok, more := in.Accept("0123456789.")
		if !(more && ok) {
			break
		}
		buf.WriteRune(r)
	}
	out <- &Item{Type: "number", Token: buf.String()}
	return switchState
}

func symbolState(in *RuneBuffer, out chan *Item) StateFn {
	buf := new(bytes.Buffer)
	for {
		r, ok, _ := in.AcceptNot(" \t\r\n;*<>-!=")
		if !ok {
			break
		}
		buf.WriteRune(r)
	}

	out <- &Item{Type: "symbol", Token: buf.String()}
	return switchState
}

func arrowState(in *RuneBuffer, out chan *Item) StateFn {
	first, _ := in.Next()
	second, _ := in.Next()

	out <- &Item{Type: "arrow", Token: string([]rune{first, second})}
	return switchState
}

func phonemeState(in *RuneBuffer, out chan *Item) StateFn {
	buf := new(bytes.Buffer)
	for {
		r, ok, _ := in.AcceptNot(" \t\r\n*>")
		if !ok {
			break
		}
		if r == '\\' { // escape character
			in.Next()
			r, ok = in.Next()
			if !ok {
				break
			}
		}
		buf.WriteRune(r)
	}

	out <- &Item{Type: "phoneme", Token: buf.String()}
	return setState
}

func setState(in *RuneBuffer, out chan *Item) StateFn {
	if r, ok := in.Peek(); ok {
		switch {
		case strings.IndexRune(" \t\r\n", r) >= 0:
			for ok { //skip whitespace
				_, ok, _ = in.Accept(" \t\r\n")
			}
			return setState
		case strings.IndexRune("*/", r) >= 0:
			in.Next()
			out <- &Item{Type: string(r), Token: string(r)}
			return setState
		case strings.IndexRune("0123456789", r) >= 0:
			numberState(in, out)
			return setState
		case r == ';':
			in.Next()
			out <- &Item{Type: "EOL", Token: "EOL"}
			commentState(in, out)
			return setState
		case r == '>':
			in.Next()
			out <- &Item{Type: ">", Token: ">"}
			return switchState
		default:
			return phonemeState
		}
	} else {
		return nil
	}
}

func switchState(in *RuneBuffer, out chan *Item) StateFn {
	if r, ok := in.Peek(); ok {
		for ok { // skip spaces
			_, ok, _ = in.Accept(" \t\r")
		}
		switch {
		case strings.IndexRune(" \t\r", r) >= 0:
			for ok { // skip whitespace
				_, ok, _ = in.Accept(" \t\r")
			}
			return switchState
		case r == '\n':
			in.Next()
			out <- &Item{Type: "EOL", Token: "EOL"}
			return switchState
		case r == ';':
			in.Next()
			out <- &Item{Type: "EOL", Token: "EOL"}
			return commentState
		case strings.IndexRune("#$_*/=", r) >= 0:
			in.Next()
			out <- &Item{Type: string(r), Token: string(r)}
			return switchState
		case r == '<':
			in.Next()
			out <- &Item{Type: "<", Token: "<"}
			return setState
		case strings.IndexRune("-!", r) >= 0:
			return arrowState
		case strings.IndexRune("0123456789", r) >= 0:
			return numberState
		default:
			return symbolState
		}
	}
	out <- &Item{Type: "EOF", Token: "EOF"}
	return nil
}
