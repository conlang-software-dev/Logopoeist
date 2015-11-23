package interpreter

import "time"
import "math/rand"
import . "github.com/conlang-software-dev/Logopoeist/parser"
import . "github.com/conlang-software-dev/Logopoeist/grammar"
import . "github.com/conlang-software-dev/Logopoeist/types"

func Interpreter() *model {
	return &model{
		start:    "",
		cvars:    make(Environment),
		synmodel: make(Grammar),
		chrmodel: make(ngrams),
		excmodel: make(ngrams),
		rnd:      rand.New(rand.NewSource(time.Now().UnixNano())),
		words:    make(map[string]struct{}),
	}
}
