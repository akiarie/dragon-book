package main

import (
	"fmt"
	"unicode"
)

type token string

const (
	// punctuation
	tkLbrac token = "("
	tkRbrac       = ")"
	tkColon       = ";"

	// control
	tkIf  = "if"
	tkFor = "for"

	// placeholders
	tkExpr  = "expr"
	tkOther = "other"
)

func parseToken(s string) (token, bool) {
	for _, tk := range []token{tkLbrac, tkRbrac, tkColon, tkIf, tkFor, tkExpr, tkOther} {
		if s == string(tk) {
			return tk, true
		}
	}
	return token("Unknown"), false
}

type lexer struct {
	pos    int
	input  string
	tokens []token
}

type stateFn func(*lexer) stateFn

func tokenize(lex *lexer) stateFn {
	if lex.pos >= len(lex.input) {
		return nil
	}
	errstream := ""
	stream := ""
	for i, c := range lex.input[lex.pos:] {
		errstream += fmt.Sprintf("%c", c)
		if !unicode.IsSpace(c) {
			stream += fmt.Sprintf("%c", c)
		}
		if tk, ok := parseToken(stream); ok {
			lex.tokens = append(lex.tokens, tk)
			lex.pos += i + 1
			return tokenize
		}
	}
	panic(fmt.Sprintf("Unknown sequence '%s'", errstream))
}

type production []string
type nonterminal []production

var grammar = map[string]nonterminal{
	"stmt": nonterminal{
		production{"expr", ";"},
		production{"if", "(", "expr", ")", "stmt"},
		production{"for", "(", "optexpr", ";", "optexpr", ";", "optexpr", ")", "stmt"},
		production{"other"},
	},
	"optexpr": nonterminal{
		production{},
		production{"expr"},
	},
}

/*
type node struct {
	symbol   string
	children []node
}

func parsetree(tokens []token, start string) (*node, int, error) {
	prods, ok := grammar[start]
	if !ok {
		return nil, -1, fmt.Errorf("Unknown symbol '%s'", start)
	}
	pos := 0
	children := []node{}
	for _, prod := range prods {
		if prod[0] == string(tokens[pos]) {
			child, lookahead, err := parsetree(tokens, prod[0])
			if err != nil {
				continue
			}
			children = append(children, *child)
		}
	}
}
*/

func main() {
	lex := &lexer{
		input:  "for (; expr; expr) other",
		tokens: []token{},
	}
	for state := stateFn(tokenize); state != nil; state = state(lex) {
	}
	fmt.Println(lex.tokens)
}
