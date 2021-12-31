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
	if tk.string == "" {
		tk.string = "[space]"
	}
	if tk.preimage != "" {
		return fmt.Sprintf("%s", tk.string)
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

type production string

var prodsymsre = regexp.MustCompile(`(\w+)|(\/[^/]+\/)|([^ |])|(?:'([^']+)')`)
var prodsymsrebar = regexp.MustCompile(`((?:\w|\|{2})+)|(\/[^/]+\/)|([^ |])|(?:'([^']+)')`)

func (prod production) symbols() []string {
	fields := []string{}
	for _, m := range prodsymsre.FindAllStringSubmatch(string(prod), -1) {
		for _, sym := range m[1:] {
			if len(sym) > 0 {
				fields = append(fields, strings.TrimSpace(sym))
			}
		}
	}
	return fields
}

type prodsymbol struct {
	string
	canspace bool
}

func (prod production) symbolsconcat(split bool) []prodsymbol {
	symbols := []string{}
	for _, m := range prodsymsrebar.FindAllStringSubmatch(string(prod), -1) {
		for _, sym := range m[1:] {
			if len(sym) > 0 {
				symbols = append(symbols, strings.TrimSpace(sym))
			}
		}
	}
	fields := []prodsymbol{}
	for i, sym := range symbols {
		canspace := i > 0 // can space only if there are other elems
		if parts := strings.Split(sym, "||"); split && len(parts) > 0 {
			for j, part := range parts {
				fields = append(fields, prodsymbol{strings.TrimSpace(part), canspace && j == 0}) // only last concat can
			}
		} else {
			fields = append(fields, prodsymbol{strings.TrimSpace(sym), canspace})
		}
	}
	return fields
}

func productionstostrings(prods []production) []string {
	fields := make([]string, len(prods))
	for i := range prods {
		fields[i] = string(prods[i])
	}
	return fields
}

func (prod production) parse() {
	G := Grammar{
		Nonterminal{"production",
			[]production{
				"symbol production", "symbol",
				"command production", "command",
			},
		},
		Nonterminal{"command",
			[]production{
				"{ print(string) }",
			},
		},
		Nonterminal{"string", []production{"quote||literal||quote"}},
		Nonterminal{"literal", []production{"symbol", "symbol literal"}},
		Nonterminal{"name", []production{"alpha||name", "alpha||digit", "alpha"}},
		Nonterminal{"symbol", []production{"char||symbol", "char"}},
		Nonterminal{"char",
			[]production{
				"alpha",
				"digit",
				"specialchar",
				"quote||bnfchar||quote",
				"quote",
				"'ε'",
			},
		},
		Nonterminal{"alpha", []production{`/[a-zA-Z]/`}},
		Nonterminal{"digit", []production{`/[0-9]/`}},
		Nonterminal{"specialchar", []production{`/[-!@#%_—=;:,<>"]/`, `/[\?\$\^&\*\(\)\+\[\]\(\)\.\\]/`, "`", "escquote"}},
		Nonterminal{"bnfchar", []production{"'|'", "'||'", "'→'"}},
		Nonterminal{"quote", []production{"'"}},
	}
	fmt.Println(G)
	//lex := &lexer{G: G, input: string(prod)}
	//for state := stateFn(tokenize); state != nil; state = state(lex) {
	//}
	//fmt.Println(G.terminals())
	//fmt.Println(prod)
	//fmt.Println(lex.tokens)

	/*
		tree, n, err := G[0].parse(lex.tokens, G)
		if err != nil {
			panic(err)
		}
		if n < len(lex.tokens) {
			panic(fmt.Errorf("Unable to parse '%s' at %d", preimage(lex.tokens[n:]), n))
		}
		fmt.Println(tree)
	*/
}

// Nonterminal represents a nonterminal in a context-free grammar.
type Nonterminal struct {
	Head        string
	Productions []production
}

// AntiLeftRecurse eliminates left-recursion, if possible, by re-writing the
// Nonterminal
//     A → A α | A β | γ | δ
// as the pair
//     A → γ R | δ R
//     R → α R | β R | ε
func (nt Nonterminal) AntiLeftRecurse() ([]Nonterminal, error) {
	var Rsym string = "R"
	static := []production{}
	tails := []production{}
	for _, prod := range nt.Productions {
		symbols := prod.symbols()
		if len(symbols) < 1 {
			return nil, fmt.Errorf("Empty production %s → %s", nt.Head, prod)
		}
		if sym := prod.symbols()[0]; sym == nt.Head {
			if len(symbols) < 2 {
				return nil, fmt.Errorf("Cannot anti recurse %s → %s, too few symbols", nt.Head, prod)
			}
			α := strings.Join(symbols[1:], " ")
			tails = append(tails, production(fmt.Sprintf("%s %s", α, Rsym)))
		} else {
			γ := strings.TrimSpace(string(prod))
			static = append(static, production(fmt.Sprintf("%s %s", γ, Rsym)))
		}
	}
	if len(static) == len(nt.Productions) {
		return []Nonterminal{nt}, nil
	}
	if len(tails) == len(nt.Productions) {
		return nil, fmt.Errorf("Sinister Nonterminal %s left-recursion cannot be eliminated", nt)
	}
	return []Nonterminal{
		Nonterminal{nt.Head, static},
		Nonterminal{Rsym, append(tails, "ε")},
	}, nil
}

func (nt Nonterminal) String() string {
	if len(nt.Productions) == 0 {
		panic("Cannot display empty Nonterminal")
	}
	return fmt.Sprintf("{%s → %v}", nt.Head, strings.Join(productionstostrings(nt.Productions), " | "))
}

func (nt Nonterminal) parse(tokens []Token, G Grammar) (*node, int, error) {
	for _, prod := range nt.Productions {
		children := []node{}
		pos := 0
		var parser func(int) (*node, int, error)
		for _, sym := range prod.symbolsconcat(true) {
			if sym.canspace {
				for len(tokens[pos].string) == 0 {
					pos += 1
				}
			}
			if tk, ok := G.parsetoken(sym.string); ok {
				parser = func(i int) (*node, int, error) {
					if i >= len(tokens) {
						return nil, -1, fmt.Errorf("Empty token list %v", tokens)
					}
					if tokens[i].string == tk.string {
						return &node{symbol: tk.string}, 1, nil
					} else if m := regexptk.FindStringSubmatch(tk.string); len(m) > 1 {
						re := regexp.MustCompile(m[1])
						if len(tokens[i].string) > 0 && re.FindString(tokens[i].string) == tokens[i].string {
							return &node{symbol: tokens[i].string}, 1, nil
						}
					}
					return nil, -1, fmt.Errorf("Unknown Token %v", tokens[0])
				}
			} else {
				for _, subnt := range G {
					if sym.string == subnt.Head {
						parser = func(i int) (*node, int, error) {
							return subnt.parse(tokens[i:], G)
						}
						break // crucial to prevent subnt in the function above being re-written
					}
				}
			}
			if parser == nil { // should be impossible, but in case
				panic(fmt.Sprintf("Unknown symbol: %s", sym.string))
			}
			fmt.Printf("Parse %q with [%s, %t] in [%s] from %s\n", tokens[pos], sym.string, sym.canspace, prod, nt)
			if child, shift, err := parser(pos); err == nil {
				children = append(children, *child)
				fmt.Printf("%v\nParsed '%s' with [%s, %t] in [%s] from %s\n", children, tokens[pos:pos+shift], sym.string, sym.canspace, prod, nt)
				pos += shift
			} else {
				fmt.Println(err)
				fmt.Printf("%v\nCannot parse %q with [%s, %t] in [%s] from %s\n", children, tokens[pos], sym.string, sym.canspace, prod, nt)
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

	for _, nt := range G {
		for _, prod := range nt.Productions {
			for _, sym := range prod.symbols() {
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

var regexptk = regexp.MustCompile(`\/([^/]+)\/`)

func (G Grammar) parsetoken(s string) (Token, bool) {
	trim := strings.TrimSpace(s)
	for _, tk := range G.terminals() {
		if m := regexptk.FindStringSubmatch(tk.string); len(m) > 1 {
			if re := regexp.MustCompile(m[1]); re.FindString(trim) == trim {
				return Token{trim, s}, true
			}
		}
		if trim == tk.string {
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
	prettyprod := func(prod production) string {
		pieces := []string{}
		for _, sym := range prod.symbolsconcat(false) {
			if sym.string == "||" {
				pieces = append(pieces, sym.string)
			} else {
				parts := []string{}
				for _, p := range strings.Split(sym.string, "||") {
					if _, ok := ntmap[p]; ok {
						parts = append(parts, chalk.Blue.NewStyle().Style(p))
					} else if p[0] == '/' && p[len(p)-1] == '/' {
						parts = append(parts, "/"+chalk.Magenta.NewStyle().Style(p[1:len(p)-1])+"/")
					} else {
						parts = append(parts, p)
					}
				}
				pieces = append(pieces, strings.Join(parts, "||"))
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
		if len(nt.Productions) == 0 {
			panic(fmt.Sprintf("Nonterminal %v has no productions", nt))
		}
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

// AntiLeftRecurse eliminates left-recursion in the Grammar by calling
// (Nonterminal).AntiLeftRecurse() on every Nonterminal and replacing as
// appropriate.
func (G Grammar) AntiLeftRecurse() (Grammar, error) {
	newG := Grammar{}
	for _, nt := range G {
		repnts, err := nt.AntiLeftRecurse()
		if err != nil {
			return nil, err
		}
		newG = append(newG, repnts...)
	}
	return newG, nil
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
