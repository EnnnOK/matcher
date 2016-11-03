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
*/

const (
	eof charType = iota
	charEscapeLiteral
	charLiteral
	charDot
	charBegin
	charEnd
	charStar
)

type charType int

func (t charType) String() string {
	switch t {
	case eof:
		return "EOF"
	case charEscapeLiteral:
		return "ESCAPE"
	case charLiteral:
		return "LITERAL"
	case charDot:
		return "DOT"
	case charBegin:
		return "BEGIN"
	case charEnd:
		return "END"
	case charStar:
		return "STAR"
	}
	return ""
}

type char struct {
	typ charType
	val byte
}

func (c char) String() string {
	return fmt.Sprintf("{%s %q}", c.typ, c.val)
}

type lexer struct {
	expression string
	pos        int
	chars      []char
}

func (l *lexer) run() {
	for {
		switch l.expression[l.pos] {
		case '\\':
			l.emit(charEscapeLiteral)
		case '.':
			l.emit(charDot)
		case '^':
			l.emit(charBegin)
		case '$':
			l.emit(charEnd)
		case '*':
			l.emit(charStar)
		default:
			l.emit(charLiteral)
		}
		if !l.next() {
			break
		}
	}
}

func (l *lexer) top() *char {
	if len(l.chars) > 0 {
		return &l.chars[len(l.chars)-1]
	}
	return nil
}

func (l *lexer) emit(t charType) {
	c := l.expression[l.pos]
	if t == charEscapeLiteral {
		if l.next() {
			c = escape(l.expression[l.pos])
		} else {
			log.Fatalln("cannot have a trailing backslash in regular expression")
		}
	}
	if t == charStar {
		top := l.top()
		if top == nil || (top.typ != charLiteral && top.typ != charDot) {
			log.Fatalln("Preceding token to star is not quantifiable")
		}
	}
	l.chars = append(l.chars, char{t, c})
}

func escape(c byte) byte {
	switch c {
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
		return c
	}
}

func (l *lexer) next() bool {
	l.pos++
	if l.pos < len(l.expression) {
		return true
	}
	return false
}

func lex(expression string) []char {
	l := &lexer{
		expression: expression,
		chars:      make([]char, 0, len(expression)),
	}
	l.run()
	return l.chars
}

// /* match: search for regexp anywhere in text */
// func match(regexp, text []byte) bool {
// 	if regexp[0] == '^' {
// 		return matchhere(regexp[1:], text)
// 	}
// 	for i := range text {
// 		if matchhere(regexp, text[i:]) {
// 			return true
// 		}
// 	}
// 	return false
// }

// /* matchhere: search for regexp at beginning of text */
// func matchhere(regexp, text []byte) bool {
// 	if len(regexp) == 0 {
// 		return true
// 	}
// 	if len(regexp) > 1 && regexp[1] == '*' {
// 		return matchstar(regexp[0], regexp[2:], text)
// 	}
// 	if regexp[0] == '$' && len(regexp) == 1 {
// 		return len(text) == 0
// 	}
// 	if len(text) != 0 && (regexp[0] == '.' || regexp[0] == text[0]) {
// 		return matchhere(regexp[1:], text[1:])
// 	}
// 	return false
// }

// /* matchstar: search for c*regexp at beginning of text */
// func matchstar(c byte, regexp, text []byte) bool {
// 	i := 0
// 	for {
// 		if matchhere(regexp, text[i:]) {
// 			return true
// 		}
// 		if i == len(text) || (text[i] != c && c != '.') {
// 			break
// 		}
// 		i++
// 	}
// 	return false
// }

// assumes input is correct
func main() {
	regexp := `aa*`
	// text := "abcb"
	// fmt.Println(match([]byte(regexp), []byte(text)))
	chars := lex(regexp)
	fmt.Println(chars)
}
