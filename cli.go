package main

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/xlab/treeprint"
)

type token string

const (
	tkEmpty token = "ε"

	// punctuation
	tkLbrac = "("
	tkRbrac = ")"
	tkColon = ";"

	// control
	tkIf  = "if"
	tkFor = "for"

	// placeholders
	tkExpr  = "expr"
	tkOther = "other"
)

func parsetoken(s string) (token, bool) {
	for _, tk := range []token{tkEmpty, tkLbrac, tkRbrac, tkColon, tkIf, tkFor, tkExpr, tkOther} {
		if s == string(tk) {
			return tk, true
		}
	}
	return token("Unknown"), false
}

func (tk token) parse(tokens []token, G grammar) (*node, int, error) {
	if tokens[0] == tk {
		return &node{symbol: tk}, 1, nil
	}
	return nil, -1, fmt.Errorf("Unknown token %v", tokens[0])
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
		if unicode.IsSpace(c) {
			continue
		}
		stream += fmt.Sprintf("%c", c)
		if tk, ok := parsetoken(stream); ok {
			lex.tokens = append(lex.tokens, tk)
			lex.pos += i + 1
			return tokenize
		}
	}
	panic(fmt.Sprintf("Unknown sequence '%s'", errstream))
}

type symbol interface {
	parse([]token, grammar) (*node, int, error)
}

type production string
type nonterminal struct {
	name  string
	prods []production
}

func (nt nonterminal) parse(tokens []token, G grammar) (*node, int, error) {
	optional := false
	pos := 0
	children := []node{}
	for _, prod := range nt.prods {
		if prod == "ε" {
			optional = true
			continue
		}
		var sym symbol
		for _, field := range strings.Fields(string(prod)) {
			if tk, ok := parsetoken(field); ok {
				sym = tk
			} else {
				for _, subnt := range G {
					if field == subnt.name {
						sym = subnt
					}
				}
			}
			if sym == nil { // should be impossible, but in case
				panic(fmt.Sprintf("Unknown field: %s", field))
			}
			if child, shift, err := sym.parse(tokens[pos:], G); err == nil {
				children = append(children, *child)
				pos += shift
			} else {
				goto nextprod
			}
		}
		return &node{symbol: nt, children: children}, pos, nil
	nextprod:
	}
	if optional {
		return &node{symbol: tkEmpty}, 0, nil
	}
	return nil, -1, fmt.Errorf("Cannot identify symbol '%s' at %d in %v", tokens[pos], pos, tokens)
}

type grammar []nonterminal

func (G grammar) parsetree(tokens []token) (*node, error) {
	tree, _, err := G[0].parse(tokens, G)
	if err != nil {
		return nil, err
	}
	return tree, nil
}

func (G grammar) String() string {
	padlen := 0
	for _, nt := range G {
		if padlen < len(nt.name) {
			padlen = len(nt.name)
		}
	}
	s := ""
	for _, nt := range G {
		s += fmt.Sprintf("%-*s → %v\n", padlen, nt.name, nt.prods[0])
		for _, prod := range nt.prods[1:] {
			s += fmt.Sprintf("%*s | %v\n", padlen, "", prod)
		}
		s += fmt.Sprintln()
	}
	return strings.TrimRightFunc(s, unicode.IsSpace)
}

type node struct {
	symbol   symbol
	children []node
}

func (n node) String() string {
	tree := treeprint.NewWithRoot(n.symbol)
	for _, c := range n.children {
		tree.AddNode(c.String())
	}
	return tree.String()
}

func main() {
	G := grammar{
		nonterminal{
			"stmt",
			[]production{
				"expr ;",
				"if ( expr ) stmt",
				"for ( optexpr ; optexpr ; optexpr ) stmt",
				"other",
			},
		},
		nonterminal{
			"optexpr",
			[]production{
				"ε",
				"expr",
			},
		},
	}
	fmt.Println(G)
	input := `for ( ; expr ; expr ) other`
	lex := &lexer{
		input:  strings.TrimSpace(input),
		tokens: []token{},
	}
	for state := stateFn(tokenize); state != nil; state = state(lex) {
	}
	/*
		tree, err := G.parsetree(lex.tokens)
		if err != nil {
			log.Fatalln(err)
		}
		fmt.Println(tree)
	*/
}
