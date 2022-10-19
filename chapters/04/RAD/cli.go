package main

import (
	"fmt"
	"log"
	"unicode"
)

/*
We will parse the grammar
	E   → E + T | T
	T   → T * F | F
	F   → ( E ) | num
	num → [0-9]+
using recursive-descent-ascent.
*/

type parser struct {
	pos    int
	input  []token
	output string
}

func (p *parser) expr() error {
	if err := p.term(); err != nil {
		return err
	}
	if p.pos >= len(p.input) {
		return nil
	}
	tk := p.input[p.pos]
	if tk.lexeme != "+" {
		return nil
	}
	p.output += " + "
	p.pos++
	return p.expr()
}

func (p *parser) term() error {
	if err := p.factor(); err != nil {
		return err
	}
	if p.pos >= len(p.input) {
		return nil
	}
	tk := p.input[p.pos]
	if tk.lexeme != "*" {
		return nil
	}
	p.output += " * "
	p.pos++
	return p.term()
}

func (p *parser) factor() error {
	switch tk := p.input[p.pos]; tk.class {
	case tkBracket:
		if tk.lexeme != "(" {
			return fmt.Errorf("%q cannot start factor", tk.lexeme)
		}
		p.output += "("
		p.pos++
		p.expr()
		tk := p.input[p.pos]
		if tk.lexeme != ")" {
			return fmt.Errorf("%q cannot end factor", tk.lexeme)
		}
		p.output += ")"
		p.pos++
		return nil
	case tkNum:
		p.output += tk.lexeme
		p.pos++
		return nil
	default:
		return fmt.Errorf("unknown class %s in factor", tk.class)
	}
}

type tkclass string

const (
	tkNum     tkclass = "num"
	tkBracket         = "bracket"
	tkOp              = "op"
	tkSpace           = "[space]"
)

type token struct {
	class  tkclass
	lexeme string
	length int
}

func (tk token) String() string {
	return fmt.Sprintf("{%s}", tk.lexeme)
}

func all(input []rune, class tkclass) (*token, error) {
	lexeme := string(input[0])
	if len(input) > 1 {
		n, err := next(input[1:])
		if err != nil {
			return nil, err
		}
		if n.class == class {
			lexeme += n.lexeme
		}
	}
	return &token{class, lexeme, len(lexeme)}, nil
}

func next(input []rune) (*token, error) {
	switch r := input[0]; {
	case r == '+', r == '*':
		return &token{tkOp, string(r), 1}, nil
	case r == '(', r == ')':
		return &token{tkBracket, string(r), 1}, nil
	case unicode.IsSpace(r):
		n, err := next(input[1:])
		if err != nil {
			return nil, err
		}
		return &token{n.class, n.lexeme, n.length + 1}, nil
	case unicode.IsNumber(r):
		return all(input, tkNum)
	default:
		return nil, fmt.Errorf("unknown token %q", r)
	}
}

func tokenise(input []rune) ([]token, error) {
	var tokens []token
	for i := 0; i < len(input); {
		tk, err := next(input[i:])
		if err != nil {
			return nil, fmt.Errorf("error after parsing %v: %s",
				tokens, err)
		}
		tokens = append(tokens, *tk)
		i += tk.length
	}
	return tokens, nil
}

func main() {
	input := "13      + 233     * (	42 +	 2) + 5"
	tokens, err := tokenise([]rune(input))
	if err != nil {
		log.Fatalln("cannot tokenise:", err)
	}
	p := &parser{input: tokens}
	if err := p.expr(); err != nil {
		log.Fatalln("cannot parse:", err)
	}
	fmt.Println(p.output)
}
