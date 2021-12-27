package main

import (
	"fmt"
	"log"
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
	for _, tk := range []token{tkLbrac, tkRbrac, tkColon, tkIf, tkFor, tkExpr, tkOther} {
		if s == string(tk) {
			return tk, true
		}
	}
	return token("Unknown"), false
}

func (tk token) parse(tokens []token) (*node, int, error) {
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
	parse([]token) (*node, int, error)
}

type production string
type nonterminal []production

var grammar = map[string]nonterminal{
	"stmt": nonterminal{
		"expr ;",
		"if ( expr ) stmt",
		"for ( optexpr ; optexpr ; optexpr ) stmt",
		"other",
	},
	"optexpr": nonterminal{
		"ε",
		"expr",
	},
}

func (nt nonterminal) parse(tokens []token) (*node, int, error) {
	optional := false
	pos := 0
	children := []node{}
	for _, prod := range nt {
		if prod == "ε" {
			optional = true
			continue
		}
		var sym symbol
		for _, field := range strings.Fields(string(prod)) {
			if tk, ok := parsetoken(field); ok {
				sym = tk
			} else {
				for subnt := range grammar {
					if field == subnt {
						sym = grammar[subnt]
					}
				}
			}
			if sym == nil { // should be impossible, but in case
				panic(fmt.Sprintf("Unknown field: %s", field))
			}
			if child, shift, err := sym.parse(tokens[pos:]); err == nil {
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
	input := `for ( ; expr ; expr ) other`
	lex := &lexer{
		input:  strings.TrimSpace(input),
		tokens: []token{},
	}
	for state := stateFn(tokenize); state != nil; state = state(lex) {
	}
	tree, _, err := grammar["stmt"].parse(lex.tokens)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(tree)
}
