package main

import (
	"fmt"
	"log"
)

/*
	c	matches any literal character c
	.	matches any single character
	^	matches the beginning of the input string
	$	matches the end of the input string
	*	matches zero or more occurrences of the previous character

	3 concurrent stages
		1. Syntax Sieve
		2. Convert to reverse polish form
		3. Object code producer
*/

const (
	dollar = iota + 128
	star
	dot
	caret
	concat
)

func ascii(b byte) bool {
	return b == '\x00' || (b >= '\x07' && b <= '\x0D') || b == '\x1B' || (b >= '\x20' && b <= '\x7E')
}

func quantifiable(b byte) bool {
	return b != dollar && b != star && b != caret
}

func escape(b byte) byte {
	switch b {
	case '0':
		return '\x00'
	case 'a':
		return '\x07'
	case 'b':
		return '\x08'
	case 't':
		return '\x09'
	case 'n':
		return '\x0A'
	case 'v':
		return '\x0B'
	case 'f':
		return '\x0C'
	case 'r':
		return '\x0D'
	case 'e':
		return '\x1B'
	case '\\':
		return '\x5C'
	default:
		return b
	}

}

func meta(b byte) byte {
	switch b {
	case '$':
		return dollar
	case '*':
		return star
	case '.':
		return dot
	case '^':
		return caret
	default:
		return b
	}
}

type expression []byte

func (e expression) String() string {
	s := ""
	for _, v := range e {
		switch v {
		case dollar:
			s += "$"
		case star:
			s += "*"
		case dot:
			s += "."
		case caret:
			s += "^"
		case concat:
			s += " (concat) "
		case '$':
			s += "\\$"
		case '*':
			s += "\\*"
		case '.':
			s += "\\."
		case '^':
			s += "\\^"
		default:
			s += string(v)
		}
	}
	return s
}

// Syntactically validates the input expression.
// Either exits with a syntax error, or returns
// the expression with the "." to juxtapose expressions.
func validate(input string) expression {
	expr := make([]byte, 0, len(input))
	var prev byte
	var b byte
	for i := range input {
		b = input[i]
		if !ascii(b) {
			// check that b is a printable ascii character, otherwise log and exit
			log.Fatalln(b, "is not a printable ascii character")
		}
		if !quantifiable(prev) && b == '\x2A' {
			// if * is preceded by a non-quantifiable character, log and exit
			log.Fatalln("preceding token", prev, "is not quantifiable")
		}
		if prev == '\x5C' {
			// if the previous character in the expression was a backslash,
			// pop it and replace it with the escaped version of b
			expr = append(expr[:len(expr)-1], escape(b))
		} else if (i == len(input)-1) && (b == '\x5C') {
			// otherwise, if the last character is a backslash, log and exit
			log.Fatalln("cannot end with trailing backslash")
		} else {
			// otherwise, append all characters to the expression, taking into
			// account metacharacters like $, *, ., and ^.
			expr = append(expr, meta(b))
		}
		prev = expr[len(expr)-1]
		if prev == star {
			l := len(expr) - 2
			expr = append(expr, 0)
			copy(expr[l+1:], expr[l:])
			expr[l] = concat
		}
	}
	return expression(expr)
}

func main() {
	fmt.Println(validate(`a\**`))
}
