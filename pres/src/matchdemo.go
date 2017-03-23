package main

import (
	"fmt"

	"github.com/smasher164/matcher"
)

func match(regexp, source string) {
	chars := matcher.Lex(regexp)
	chars = matcher.Postfix(chars)
	nfa := matcher.Post2nfa(chars)
	fmt.Println(matcher.Match(nfa, source))
}

// START OMIT
func main() {
	// c   matches any literal character c
	// .   matches any single character
	// *   matches zero or more occurrences of the previous character
	// |   matches the previous character or the next character
	match("abc", "abc")
}

// END OMIT
