package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

/*
	c	matches any literal character c
	.	matches any single character
	^	matches the beginning of the input string
	$	matches the end of the input string
	*	matches zero or more occurrences of the previous character
*/

const (
	charNil charType = iota
	charEscapeLiteral
	charLiteral
	charDot
	charBegin
	charEnd
	charStar
	charConcat
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
	if t != charStar {
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

func postfix(chars []char) []char {
	output := []char{}
	operator := []char{}
	pop := func() *char {
		if len(operator) > 0 {
			c := &operator[len(operator)-1]
			operator = operator[:len(operator)-1]
			return c
		}
		return nil
	}
	for _, c := range chars {
		switch c.typ {
		case charConcat:
			if t := pop(); t != nil {
				output = append(output, *t)
			}
			operator = append(operator, c)
		default:
			output = append(output, c)
		}
	}
	output = append(output, operator...)
	return output
}

//go:generate stringer -type=styp
type styp int

const (
	match styp = iota
	split
	single
)

type state struct {
	typ      styp
	c        char
	out      []*state
	lastlist int // list generation number
}

func (s state) String() string {
	return fmt.Sprintf("{typ: %s, c: %s, out: %p}", s.typ, s.c, s.out)
}

type frag struct {
	start *state
	out   []ptr
}

type ptr struct {
	s *state
	i int
}

type dfastate struct {
	list *[]*state
	next [256]*dfastate
}

var cacheddfa map[*[]*state]*dfastate

func patch(out []ptr, start *state) {
	var p ptr
	for i := len(out) - 1; i >= 0; i-- {
		p, out = out[len(out)-1], out[:len(out)-1]
		index := p.i
		p.s.out[index] = start
	}
}

func post2nfa(postfix []char) (start *state) {
	stack := []frag{}
	push := func(f frag) {
		stack = append(stack, f)
	}
	pop := func() (f frag) {
		f, stack = stack[len(stack)-1], stack[:len(stack)-1]
		return
	}
	for _, p := range postfix {
		switch p.typ {
		default:
			s := &state{typ: single, c: p, out: []*state{nil}}
			out := []ptr{{s, 0}}
			push(frag{s, out})
		case charConcat:
			e2 := pop()
			e1 := pop()
			patch(e1.out, e2.start)
			push(frag{e1.start, e2.out})
		case charStar:
			e := pop()
			s := &state{typ: split, out: []*state{e.start, nil}}
			patch(e.out, s)
			out := []ptr{{s, 1}}
			push(frag{s, out})
		}
	}
	e := pop()
	patch(e.out, &state{typ: match})
	return e.start
}

func addstate(list *[]*state, s *state, listid int) {
	if s.lastlist == listid {
		return
	}
	s.lastlist = listid
	if s.typ == split {
		addstate(list, s.out[0], listid)
		addstate(list, s.out[1], listid)
		return
	}
	*list = append(*list, s)
}

func matchregex(start *state, source string) bool {
	cacheddfa = make(map[*[]*state]*dfastate)

	listid := 1
	list := []*state{}
	addstate(&list, start, listid)
	d := getdfastate(&list)
	var next *dfastate

	for i := range source {
		c := source[i]
		next = d.next[c]
		if next == nil {
			list, listid = step(list, c, listid)
			d.next[c] = getdfastate(&list)
			next = d.next[c]
		}
		d = next
	}
	return ismatch(*d.list)
}

func step(list []*state, c byte, listid int) ([]*state, int) {
	nlist := []*state{}
	listid++
	for _, s := range list {
		if s.typ == single && s.c.val == c || s.c.typ == charDot {
			addstate(&nlist, s.out[0], listid)
		}
	}
	return nlist, listid
}

func getdfastate(list *[]*state) *dfastate {
	d := cacheddfa[list]
	if d != nil {
		return d
	}
	d = &dfastate{list: list}
	cacheddfa[list] = d
	return d
}

func ismatch(list []*state) bool {
	for _, s := range list {
		if s.typ == match {
			return true
		}
	}
	return false
}

func printnfa(s *state) {
	fmt.Println(s)
	switch s.typ {
	case single:
		printnfa(s.out[0])
	case split:
		printnfa(s.out[1])
	case match:
		return
	}
}
func main() {
	inSrc, _ := ioutil.ReadAll(os.Stdin)
	source := string(inSrc)
	flag.Parse()
	regexp := flag.Arg(0)

	chars := lex(regexp)
	chars = postfix(chars)
	nfa := post2nfa(chars)
	fmt.Println(matchregex(nfa, source))
}
