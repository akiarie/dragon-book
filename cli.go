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

func parsetoken(s string) (token, bool) {
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
		if tk, ok := parsetoken(stream); ok {
			lex.tokens = append(lex.tokens, tk)
			lex.pos += i + 1
			return tokenize
		}
	}
	panic(fmt.Sprintf("Unknown sequence '%s'", errstream))
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
		"",
		"expr",
	},
}

type node struct {
	symbol   string
	children []node
}

func (n node) String() string {
	tree := treeprint.NewWithRoot(n.symbol)
	for _, c := range n.children {
		tree.AddNode(c.String())
	}
	return tree.String()
}

// parsetree parses token list into tree based on start, returning the number
// of consumed tokens
func parsetree(tokens []token, startnt string) (*node, int, error) {
	optional := false
	pos := 0
	children := []node{}
	for _, prod := range grammar[startnt] {
		if prod == "" {
			optional = true
			continue
		}
		fields := strings.Fields(string(prod))
		for _, field := range fields {
			// if field terminal
			if tk, ok := parsetoken(field); ok {
				if field == string(tk) {
					if tk == tokens[pos] {
						children = append(children, node{symbol: field})
						pos++
						goto nextfield
					}
				}
				// with terminals, we must have an exact match, so if we don't
				// get one we know the production won't work
				goto nextprod
			} else { // otherwise, perhaps nonterminal?
				for subnt := range grammar {
					if field == subnt {
						child, shift, err := parsetree(tokens[pos:], subnt)
						if err == nil {
							children = append(children, *child)
							pos += shift
							goto nextfield
						}
					}
				}
			}
			// production doesn't match
			goto nextprod
		nextfield:
		}
		// if these match then some production parsed
		if len(children) == len(fields) {
			return &node{symbol: startnt, children: children}, pos, nil
		}
	nextprod:
	}
	if optional {
		return &node{symbol: "ε"}, 0, nil
	}
	return nil, -1, fmt.Errorf("Cannot identify symbol '%s' at %d in %v with start %s", tokens[pos], pos, tokens, startnt)
}

func main() {
	lex := &lexer{
		input:  "for (; expr; expr) other",
		tokens: []token{},
	}
	for state := stateFn(tokenize); state != nil; state = state(lex) {
	}
	tree, _, err := parsetree(lex.tokens, "stmt")
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(tree)
}
