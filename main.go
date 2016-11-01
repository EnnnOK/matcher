package main

import "fmt"

/*
	c	matches any literal character c
	.	matches any single character
	^	matches the beginning of the input string
	$	matches the end of the input string
	*	matches zero or more occurrences of the previous character
*/

/* match: search for regexp anywhere in text */
func match(regexp, text []byte) bool {
	if regexp[0] == '^' {
		return matchhere(regexp[1:], text)
	}
	for i := range text {
		if matchhere(regexp, text[i:]) {
			return true
		}
	}
	return false
}

/* matchhere: search for regexp at beginning of text */
func matchhere(regexp, text []byte) bool {
	if len(regexp) == 0 {
		return true
	}
	if len(regexp) > 1 && regexp[1] == '*' {
		return matchstar(regexp[0], regexp[2:], text)
	}
	if regexp[0] == '$' && len(regexp) == 1 {
		return len(text) == 0
	}
	if len(text) != 0 && (regexp[0] == '.' || regexp[0] == text[0]) {
		return matchhere(regexp[1:], text[1:])
	}
	return false
}

/* matchstar: search for c*regexp at beginning of text */
func matchstar(c byte, regexp, text []byte) bool {
	i := 0
	for {
		if matchhere(regexp, text[i:]) {
			return true
		}
		if i == len(text) || (text[i] != c && c != '.') {
			break
		}
		i++
	}
	return false
}

// assumes input is correct
func main() {
	regexp := "abc$"
	text := "abcb"
	fmt.Println(match([]byte(regexp), []byte(text)))
}
