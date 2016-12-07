package main

import (
	"fmt"
	"log"
)

const (
	charNil charType = iota
	charEscapeLiteral
	charLiteral
	charDot
	charStar
	charConcat
	charOr
)

//go:generate stringer -type=charType
type charType int

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
		case '*':
			l.emit(charStar)
		case '|':
			l.emit(charOr)
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

// emit validates and appends the concatenated characters
// to to a slice.
func (l *lexer) emit(t charType) {
	c := l.expression[l.pos]
	if t == charEscapeLiteral {
		if l.next() {
			c = escape(l.expression[l.pos])
		} else {
			log.Fatalln("cannot have a trailing backslash in regular expression")
		}
	}
	top := l.top()
	if t == charStar {
		if top == nil || (top.typ != charLiteral && top.typ != charDot) {
			log.Fatalln("Preceding token to star is not quantifiable")
		}
	}
	if t != charStar && t != charOr && (top == nil || top.typ != charOr) {
		l.chars = append(l.chars, char{charConcat, '.'})
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

// lex parses the input regular expression, and returns
// a sequence of concatenated character tokens.
func lex(expression string) []char {
	if len(expression) == 0 {
		return []char{}
	}
	l := &lexer{
		expression: expression,
		chars:      make([]char, 0, len(expression)),
	}
	l.run()
	return l.chars[1:]
}

// postfix converts a sequence of character tokens
// into postfix format. For instance, in order of
// highest to lowest precedence:
// A.B*		-->		AB*.
// A.B.C	-->		AB.C.
// A.B|C.D	-->		AB.CD.|
func postfix(chars []char) []char {
	output := []char{}
	operator := []char{}
	pop := func() *char {
		c := &operator[len(operator)-1]
		operator = operator[:len(operator)-1]
		return c
	}
	top := func() *char {
		if len(operator) > 0 {
			c := &operator[len(operator)-1]
			return c
		}
		return nil
	}
	for _, c := range chars {
		switch c.typ {
		case charStar:
			if t := top(); t != nil {
				if t.typ == charStar {
					output = append(output, *pop())
				}
			}
			operator = append(operator, c)
		case charConcat:
			if t := top(); t != nil {
				if t.typ == charConcat || t.typ == charStar {
					output = append(output, *pop())
				}
			}
			operator = append(operator, c)
		case charOr:
			if t := top(); t != nil {
				output = append(output, *pop())
			}
			operator = append(operator, c)
		default:
			output = append(output, c)
		}
	}
	oplen := len(operator)
	for i := 0; i < oplen; i++ {
		output = append(output, *pop())
	}
	return output
}
