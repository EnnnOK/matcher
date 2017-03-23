package main

import (
	"fmt"

	"github.com/smasher164/matcher"
)

// START OMIT
func main() {
	ch := matcher.Lex("abc")
	fmt.Println(matcher.Postfix(ch))
}

// END OMIT
