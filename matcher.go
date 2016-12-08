/*

This regular expression package is inspired from Russ Cox's
article "Regular Expression Matching Can Be Simple And Fast".
https://swtch.com/~rsc/regexp/regexp1.html

Uses Thompson construction for the NFA from
Communications of the ACM
Volume 11 Issue 6, June 1968
"Programming Techniques: Regular expression search algorithm"

DFA construction is based off of powerset construction while
caching the result after each state.

Lexer is inspired by Rob Pike's lecture: "Lexical Scanning in Go"
https://youtu.be/HxaD_trXwRE

		c	matches any literal character c
		.	matches any single character
		*	matches zero or more occurrences of the previous character
		|	matches the previous character or the next character
*/
package matcher

import "fmt"

//go:generate stringer -type=styp
type styp int

const (
	match styp = iota
	split
	single
)

// A state represents a single node in the nondeterministic finite
// automaton (NFA).
type state struct {
	typ      styp     // the type of the state
	c        char     // token that the state represents
	out      []*state // pointers to the next state(s)
	lastlist int      // allows partial scanning of the state lists
}

func (s state) String() string {
	return fmt.Sprintf("{typ: %s, c: %s, out: %p}", s.typ, s.c, s.out)
}

// A frag represents an NFA fragment, used to compile the postfix
// expression with a stack.
type frag struct {
	start *state
	out   []ptr
}

func (f frag) String() string {
	return fmt.Sprintf("{start: %v, out: %v}", *f.start, f.out)
}

// A ptr represents an arrow that connects an NFA fragment to another state.
type ptr struct {
	s *state
	i int
}

func (p ptr) String() string {
	return fmt.Sprintf("{s: %v, i: %d}", *p.s, p.i)
}

// A dfastate is a cached list of states, containing pointers to
// dfastates for each possible character
type dfastate struct {
	list *[]*state
	next [256]*dfastate
}

// cacheddfa is a map that is keyed by pointers to
// a list of states, with the corresponding value of
// a dfastate. Avoids recomputation of each powerset.
var cacheddfa map[*[]*state]*dfastate

// patch connects the arrows from out to the state start.
func patch(out []ptr, start *state) {
	var p ptr
	for i := len(out) - 1; i >= 0; i-- {
		p, out = out[len(out)-1], out[:len(out)-1]
		index := p.i
		p.s.out[index] = start
	}
}

// Post2nfa loops over the postfix expression,
// and uses a stack of fragments to construct a
// single nfa fragment representing the state machine.
func Post2nfa(postfix []char) (start *state) {
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
		case charOr:
			e2 := pop()
			e1 := pop()
			s := &state{typ: split, out: []*state{e1.start, e2.start}}
			out := append(e1.out, e2.out...)
			push(frag{s, out})
		}
	}
	e := pop()
	patch(e.out, &state{typ: match})
	return e.start
}

// addstate a new unique state to the list, following
// any unlabeled arrows along the way.
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

// Match loops through the source input, and
// steps through the state machine. Returns true
// if there is a match, false if not.
func Match(start *state, source string) bool {
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

// step computes the next list of states for a single character
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

// getdfastate returns a dfa state for the corresponding list
// of states. It first checks the cache, allocating a
// new one if necessary.
func getdfastate(list *[]*state) *dfastate {
	d := cacheddfa[list]
	if d != nil {
		return d
	}
	d = &dfastate{list: list}
	cacheddfa[list] = d
	return d
}

// ismatch iterates through a list of
// states to check for a match state
func ismatch(list []*state) bool {
	for _, s := range list {
		if s.typ == match {
			return true
		}
	}
	return false
}

// recursively prints state machine, for debugging purposes
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
