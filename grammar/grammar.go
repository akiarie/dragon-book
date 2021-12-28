package grammar

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/xlab/treeprint"
)

type Token string

const tkEmpty Token = "ε"

type lexer struct {
	G      Grammar
	pos    int
	input  string
	tokens []Token
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
		if tk, ok := lex.G.parsetoken(stream); ok {
			lex.tokens = append(lex.tokens, tk)
			lex.pos += i + 1
			return tokenize
		}
	}
	panic(fmt.Sprintf("Unknown sequence '%s'", errstream))
}

// Nonterminal represents a nonterminal in a context-free grammar.
type Nonterminal struct {
	Head        string
	Productions []string
}

func (nt Nonterminal) String() string {
	if len(nt.Productions) == 0 {
		panic("Cannot display empty Nonterminal")
	}
	return fmt.Sprintf("{%s → %v}", nt.Head, strings.Join(nt.Productions, " | "))
}

func (nt Nonterminal) parse(tokens []Token, G Grammar) (*node, int, error) {
	optional := false
	pos := 0
	children := []node{}
	for _, prod := range nt.Productions {
		if prod == "ε" {
			optional = true
			continue
		}
		var parser func(int) (*node, int, error)
		for _, field := range strings.Fields(prod) {
			if tk, ok := G.parsetoken(field); ok {
				parser = func(i int) (*node, int, error) {
					if tokens[i] == tk {
						return &node{symbol: string(tk)}, 1, nil
					}
					return nil, -1, fmt.Errorf("Unknown Token %v", tokens[0])
				}
			} else {
				for _, subnt := range G {
					if field == subnt.Head {
						parser = func(i int) (*node, int, error) {
							return subnt.parse(tokens[i:], G)
						}
					}
				}
			}
			if parser == nil { // should be impossible, but in case
				panic(fmt.Sprintf("Unknown field: %s", field))
			}
			if child, shift, err := parser(pos); err == nil {
				children = append(children, *child)
				pos += shift
			} else {
				goto nextprod
			}
		}
		return &node{symbol: fmt.Sprintf("%s → %s", nt.Head, prod), children: children}, pos, nil
	nextprod:
	}
	if optional {
		return &node{symbol: string(tkEmpty)}, 0, nil
	}
	return nil, -1, fmt.Errorf("Cannot identify symbol '%s' at %d in %v", tokens[pos], pos, tokens)
}

// Grammar is the representation of a context-free grammar. By convention, the
// first element is taken to be the start symbol; terminal symbols are
// understood implicitly to be represented by any symbol which cannot be traced
// to a Nonterminal in the grammar.
type Grammar []Nonterminal

func (G Grammar) terminals() []Token {
	ntmap := map[string]bool{}
	for _, nt := range G {
		ntmap[nt.Head] = true
	}
	tokens := []Token{}
	tokenmap := map[string]bool{} // map for uniqueness
	for _, nt := range G {
		for _, prod := range nt.Productions {
			for _, sym := range strings.Fields(prod) {
				if _, ok := ntmap[sym]; !ok {
					if _, ok := tokenmap[sym]; !ok {
						tokens = append(tokens, Token(sym))
						tokenmap[sym] = true
					}
				}
			}
		}
	}
	return tokens
}

func (G Grammar) parsetoken(s string) (Token, bool) {
	for _, tk := range G.terminals() {
		if s == string(tk) {
			return tk, true
		}
	}
	return Token("Unknown"), false
}

func (G Grammar) String() string {
	padlen := 0
	for _, nt := range G {
		if padlen < len(nt.Head) {
			padlen = len(nt.Head)
		}
	}
	s := ""
	for _, nt := range G {
		s += fmt.Sprintf("%-*s → %v\n", padlen, nt.Head, nt.Productions[0])
		for _, prod := range nt.Productions[1:] {
			s += fmt.Sprintf("%*s | %v\n", padlen, "", prod)
		}
		s += fmt.Sprintln()
	}
	return strings.TrimRightFunc(s, unicode.IsSpace)
}

// ParseAST parses the input string according to the Grammar, returning an
// error if this is not possible.
func (G Grammar) ParseAST(input string) (*node, error) {
	lex := &lexer{G: G, input: input}
	for state := stateFn(tokenize); state != nil; state = state(lex) {
	}
	tree, _, err := G[0].parse(lex.tokens, G)
	if err != nil {
		return nil, err
	}
	return tree, nil
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
