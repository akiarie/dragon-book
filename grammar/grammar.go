package grammar

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/ttacon/chalk"
	"github.com/xlab/treeprint"
)

type Token struct {
	string
	preimage string
}

func (tk Token) String() string {
	if tk.preimage != "" {
		return fmt.Sprintf("{%s %q}", tk.string, tk.preimage)
	}
	return tk.string
}

func preimage(tokens []Token) string {
	images := []string{}
	for _, tk := range tokens {
		images = append(images, tk.preimage)
	}
	return strings.Join(images, "")
}

var tkEmpty Token = Token{"ε", ""}

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
	stream := ""
	for i, c := range lex.input[lex.pos:] {
		stream += fmt.Sprintf("%c", c)
		if tk, ok := lex.G.parsetoken(stream); ok {
			lex.tokens = append(lex.tokens, tk)
			lex.pos += i + utf8.RuneLen(c)
			return tokenize
		}
	}
	if strings.TrimSpace(stream) == "" {
		return nil
	}
	panic(fmt.Sprintf("Unknown sequence '%s', tokens: %v", stream, lex.tokens))
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
	for _, prod := range nt.Productions {
		children := []node{}
		pos := 0
		var parser func(int) (*node, int, error)
		for _, field := range strings.Fields(prod) {
			if tk, ok := G.parsetoken(field); ok {
				parser = func(i int) (*node, int, error) {
					if i >= len(tokens) {
						return nil, -1, fmt.Errorf("Empty token list %v", tokens)
					}
					if tokens[i].string == tk.string {
						return &node{symbol: tk.string}, 1, nil
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
	for _, prod := range nt.Productions {
		if prod == "ε" {
			return &node{symbol: fmt.Sprintf("%s → %s", nt.Head, tkEmpty)}, 0, nil
		}
	}
	return nil, -1, fmt.Errorf("Syntax error in '%s' using %s", preimage(tokens), nt)
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
	quoted := regexp.MustCompile("`([^`]+)`")
	for _, nt := range G {
		for _, prod := range nt.Productions {
			for _, rawsym := range strings.Fields(prod) {
				sym := quoted.ReplaceAllString(rawsym, "$1")
				if _, ok := ntmap[sym]; !ok {
					if _, ok := tokenmap[sym]; !ok {
						tokens = append(tokens, Token{sym, ""})
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
		if strings.TrimSpace(s) == tk.string {
			return Token{tk.string, s}, true
		}
	}
	return Token{"Unknown", ""}, false
}

func (G Grammar) String() string {
	ntmap := map[string]bool{}
	for _, nt := range G {
		ntmap[nt.Head] = true
	}
	prettyprod := func(prod string) string {
		pieces := []string{}
		for _, field := range strings.Fields(prod) {
			if _, ok := ntmap[field]; ok {
				pieces = append(pieces, chalk.Blue.NewStyle().Style(field))

			} else {
				pieces = append(pieces, field)
			}
		}
		return strings.Join(pieces, " ")
	}
	padlen := 0
	for _, nt := range G {
		if padlen < len(nt.Head) {
			padlen = len(nt.Head)
		}
	}
	s := ""
	for _, nt := range G {
		s += fmt.Sprintf("%-*s → %v\n", padlen, nt.Head, prettyprod(nt.Productions[0]))
		for _, prod := range nt.Productions[1:] {
			s += fmt.Sprintf("%*s | %v\n", padlen, "", prettyprod(prod))
		}
		s += fmt.Sprintln()
	}
	return strings.TrimRightFunc(s, unicode.IsSpace)
}

// Validate ensures that every Nonterminal has at least one production.
func (G Grammar) Validate() error {
	for _, nt := range G {
		if len(nt.Productions) == 0 {
			return fmt.Errorf("Nonterminal %s with no productions", nt.Head)
		}
	}
	return nil
}

// ParseAST parses the input string according to the Grammar, returning an
// error if this is not possible.
func (G Grammar) ParseAST(input []byte) (*node, error) {
	lex := &lexer{G: G, input: string(input)}
	for state := stateFn(tokenize); state != nil; state = state(lex) {
	}
	tree, n, err := G[0].parse(lex.tokens, G)
	if err != nil {
		return nil, err
	}
	if n < len(lex.tokens) {
		return nil, fmt.Errorf("Unable to parse '%s' at %d", preimage(lex.tokens[n:]), n)
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
