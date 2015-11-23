package main

import "fmt"
import "bufio"
import "os"
import "flag"
import "strings"

import "github.com/conlang-software-dev/Logopoeist/lexer"
import "github.com/conlang-software-dev/Logopoeist/parser"
import "github.com/conlang-software-dev/Logopoeist/interpreter"

func main() {
	var file *os.File
	var fname string
	var wcount int

	flag.StringVar(&fname, "file", "", "The name of the configuration file; defaults to standard input.")
	flag.IntVar(&wcount, "n", 10, "The number of words to generate; defaults to 10.")

	flag.Parse()

	if fname != "" {
		var err error
		file, err = os.Open(fname)
		if err != nil {
			fmt.Printf("Error opening source file.\n")
			return
		}
		defer file.Close()
	} else {
		file = os.Stdin
	}

	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("Error: %s\n", err)
		}
	}()

	lex := lexer.Lex(bufio.NewReader(file), switchState)
	interp := interpreter.Interpreter()
	for command := range parser.Parse(lex) {
		interp.Execute(command)
	}

	for i := 0; i < wcount; i++ {
		clist, ok := interp.Generate()
		if !ok {
			fmt.Printf("Could Not Generate Additional Unique Words")
			break
		}
		word := strings.Join(clist, "")
		fmt.Printf("%s\n", word)
	}
}
