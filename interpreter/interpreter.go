package interpreter

import "fmt"
import "strconv"
import . "github.com/conlang-software-dev/Logopoeist/parser"
import . "github.com/conlang-software-dev/Logopoeist/types"

func InterpretNumber(n *Node) float64 {
	freq, err := strconv.ParseFloat(n.Value, 64)
	if err != nil {
		panic(fmt.Sprintf("Invalid numeric literal: %s", n.Value))
	}
	return freq
}

func InterpretClass(n *Node) *CharClass {
	list := make([]string, 0, 10)
	weights := make(CharSet, 10)
	for sn := n.Left; sn != nil; sn = sn.Right {
		fnode := sn.Left
		phoneme := fnode.Left.Value
		freq := InterpretNumber(fnode.Right)

		if _, ok := weights[phoneme]; ok {
			weights[phoneme] += freq
		} else {
			weights[phoneme] = freq
			list = append(list, phoneme)
		}
	}
	return &CharClass{
		List:    list,
		Weights: weights,
	}
}
