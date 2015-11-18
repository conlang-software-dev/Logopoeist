package lexer

import "strings"
import "io"

type Item struct {
	Type  string
	Token string
}

type StateFn func(*RuneBuffer, chan *Item) StateFn

type RuneBuffer struct {
	in   io.RuneReader
	r    rune
	more bool
}

// Peek looks at the next rune but doesn't advance the input.
func (rb *RuneBuffer) Peek() (rune, bool) {
	return rb.r, rb.more
}

// Next returns the next rune in the input.
func (rb *RuneBuffer) Next() (rune, bool) {
	if !rb.more {
		return 0, false
	}
	r, more := rb.r, rb.more
	nr, _, err := rb.in.ReadRune()
	rb.r, rb.more = nr, (err == nil)
	return r, more
}

// accept consumes the next rune if it's from the valid set.
func (rb *RuneBuffer) Accept(valid string) (rune, bool, bool) {
	r, ok := rb.Peek()
	if !ok {
		return r, false, false
	}
	if strings.IndexRune(valid, r) >= 0 {
		rb.Next()
		return r, true, true
	}
	return r, false, true
}

// accept consumes the next rune if it's not from the invalid set.
func (rb *RuneBuffer) AcceptNot(invalid string) (rune, bool, bool) {
	r, ok := rb.Peek()
	if !ok {
		return r, false, false
	}
	if strings.IndexRune(invalid, r) < 0 {
		rb.Next()
		return r, true, true
	}
	return r, false, true
}

type Lexer struct {
	tokens chan *Item
	next   *Item
	more   bool
}

// Peek looks at the next token but doesn't advance the input.
func (l *Lexer) Peek() (*Item, bool) {
	return l.next, l.more
}

// Next returns the next token from the input.
func (l *Lexer) Next() (*Item, bool) {
	item, ok := l.next, l.more
	l.next, l.more = <-l.tokens
	return item, ok
}

func Lex(input io.RuneReader, start StateFn) *Lexer {
	tokens := make(chan *Item)

	go func() {
		r, _, err := input.ReadRune()
		buf := &RuneBuffer{
			in:   input,
			r:    r,
			more: (err == nil),
		}

		for state := start; state != nil; {
			state = state(buf, tokens)
		}

		close(tokens)
	}()

	first, ok := <-tokens
	return &Lexer{
		tokens: tokens,
		next:   first,
		more:   ok,
	}
}
