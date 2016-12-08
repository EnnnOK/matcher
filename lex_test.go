package matcher

import (
	"reflect"
	"testing"
)

func TestLex(t *testing.T) {
	cases := []struct {
		expression string
		tokens     []char
	}{
		{`abc`, []char{
			{charLiteral, 'a'},
			{charConcat, '.'},
			{charLiteral, 'b'},
			{charConcat, '.'},
			{charLiteral, 'c'},
		}},
		{`.*c`, []char{
			{charDot, '.'},
			{charStar, '*'},
			{charConcat, '.'},
			{charLiteral, 'c'},
		}},
		{`a|b`, []char{
			{charLiteral, 'a'},
			{charOr, '|'},
			{charLiteral, 'b'},
		}},
		{`\.ac`, []char{
			{charEscapeLiteral, '.'},
			{charConcat, '.'},
			{charLiteral, 'a'},
			{charConcat, '.'},
			{charLiteral, 'c'},
		}},
	}

	for i, c := range cases {
		chars := Lex(c.expression)
		if !reflect.DeepEqual(c.tokens, chars) {
			t.Fatalf("case: %v, got: %v, wanted: %v", i, chars, c)
		}
	}
}

func TestPostfix(t *testing.T) {
	cases := []struct {
		tokens  []char
		postfix []char
	}{
		{[]char{
			{charLiteral, 'a'},
			{charConcat, '.'},
			{charLiteral, 'b'},
			{charConcat, '.'},
			{charLiteral, 'c'},
		}, []char{
			{charLiteral, 'a'},
			{charLiteral, 'b'},
			{charConcat, '.'},
			{charLiteral, 'c'},
			{charConcat, '.'},
		}},

		{[]char{
			{charDot, '.'},
			{charStar, '*'},
			{charConcat, '.'},
			{charLiteral, 'c'},
		}, []char{
			{charDot, '.'},
			{charStar, '*'},
			{charLiteral, 'c'},
			{charConcat, '.'},
		}},

		{[]char{
			{charLiteral, 'a'},
			{charOr, '|'},
			{charLiteral, 'b'},
		}, []char{
			{charLiteral, 'a'},
			{charLiteral, 'b'},
			{charOr, '|'},
		}},

		{[]char{
			{charEscapeLiteral, '.'},
			{charConcat, '.'},
			{charLiteral, 'a'},
			{charConcat, '.'},
			{charLiteral, 'c'},
		}, []char{
			{charEscapeLiteral, '.'},
			{charLiteral, 'a'},
			{charConcat, '.'},
			{charLiteral, 'c'},
			{charConcat, '.'},
		}},
	}

	for i, c := range cases {
		chars := Postfix(c.tokens)
		if !reflect.DeepEqual(c.postfix, chars) {
			t.Fatalf("case: %v, got: %v, wanted: %v", i, chars, c)
		}
	}
}
