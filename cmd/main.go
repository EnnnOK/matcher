package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/smasher164/matcher"
)

func main() {
	flag.Parse()
	regexp := flag.Arg(0)
	source := flag.Arg(1)
	if source == "" || regexp == "" {
		fmt.Println("usage: matcher [REGEXP] [SOURCE]")
		os.Exit(1)
	}

	chars := matcher.Lex(regexp)
	chars = matcher.Postfix(chars)
	nfa := matcher.Post2nfa(chars)
	fmt.Println(matcher.Match(nfa, source))
}
