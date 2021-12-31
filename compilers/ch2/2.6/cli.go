package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"unicode"
)

/*
	In this case we begin with the grammar
		expr   → expr + term   { print('+') }
			   | expr - term   { print('-') }
			   | term

		term   → term * factor { print('*') }
		       | term / factor { print('/') }
		       | factor

		factor → ( expr )
			   | num 		   { print(num.value) }
			   | id 		   { print(id.lexeme) }

	Reasoning as previously, we re-write expr and term:
		expr   → term rest

		rest   → + term { print('+') } rest
			   | - term { print('-') } rest
			   | ε

		term   → factor rise

		rise   → * factor { print('*') } rise
			   | / factor { print('/') } rise
			   | ε

	The productions for factor remain unchanged.
*/

type stream []token

func (tokens stream) String() string {
	var s string
	for _, tk := range tokens {
		s += tk.lexeme
	}
	return s
}

type parser struct {
	pos    int
	input  stream
	output string
}

func (p *parser) expr() {
	p.term()
	for p.pos < len(p.input) {
		var op byte
		switch p.input[p.pos].value[0] {
		case '+':
			op = '+'
			break
		case '-':
			op = '-'
			break
		default:
			fmt.Printf("%q is not an arithmetic expression\n", p.input)
			return
		}
		p.pos++
		p.term()
		p.output += fmt.Sprintf("%c", op)
	}
}

func (p *parser) term() {
	p.factor()
	for p.pos < len(p.input) {
		var op byte
		switch p.input[p.pos].value[0] {
		case '*':
			op = '*'
			break
		case '/':
			op = '/'
			break
		default:
			return
		}
		p.pos++
		p.factor()
		p.output += fmt.Sprintf("%c", op)
	}
}

func (p *parser) factor() {
	switch p.input[p.pos].class {
	case tkNum, tkId:
		p.output += fmt.Sprintf("(%s)", p.input[p.pos].value)
	case tkExpr:
		tokens, err := tokenize(p.input[p.pos].value)
		if err != nil {
			panic(err)
		}
		subp := &parser{input: tokens}
		subp.expr()
		p.output += fmt.Sprintf("%s", subp.output)
	default:
		panic(fmt.Sprintf("%q is not a number or expression\n", p.input[p.pos].value))
	}
	p.pos++
}

type tokenclass string

const (
	tkNum   tokenclass = "num"
	tkId               = "id"
	tkExpr             = "expr"
	tkOp               = "op"
	tkSpace            = "[space]"
)

type token struct {
	class  tokenclass
	value  string
	lexeme string
}

func (tk token) String() string {
	return fmt.Sprintf("{%s %s %q}", tk.class, tk.value, tk.lexeme)
}

func parsetoken(input string, pos int) (*token, int, error) {
	st := pos
	for _, c := range input[pos:] {
		if !unicode.IsSpace(c) {
			break
		}
		st++
	}
	if st >= len(input) {
		return &token{class: tkSpace, value: tkSpace, lexeme: input[pos:]}, len(input[pos:]), nil
	}

	if unicode.IsDigit(rune(input[st])) {
		i := st
		for ; i < len(input); i++ {
			if !unicode.IsDigit(rune(input[i])) {
				break
			}
		}
		if val, err := strconv.ParseUint(input[st:i], 10, 64); err == nil {
			return &token{class: tkNum, lexeme: input[pos:i], value: fmt.Sprintf("%d", val)}, len(input[pos:i]), nil
		}
	}

	if unicode.IsLetter(rune(input[st])) {
		i := st
		for ; i < len(input); i++ {
			if r := rune(input[i]); !unicode.IsLetter(r) && !unicode.IsDigit(r) {
				break
			}
		}
		return &token{class: tkId, lexeme: input[pos:i], value: input[st:i]}, len(input[pos:i]), nil
	}

	isop := func(c byte) bool { return strings.IndexByte("+-*/", c) != -1 }
	if isop(input[st]) {
		cs := fmt.Sprintf("%c", input[st])
		return &token{class: tkOp, lexeme: cs, value: cs}, 1 + st - pos, nil
	}

	if input[st] == '(' {
		for i := st; i < len(input); i++ {
			if input[i] == ')' {
				return &token{class: tkExpr, lexeme: input[pos : i+1], value: input[st+1 : i]}, len(input[pos : i+1]), nil
			}
		}
	}

	return nil, -1, fmt.Errorf("Unknown characters %q", input[pos:st+1])
}

func tokenize(input string) ([]token, error) {
	tokens := []token{}
	for pos := 0; pos < len(input); {
		tk, shift, err := parsetoken(input, pos)
		if err != nil {
			return nil, err
		}
		if tk.class != tkSpace {
			tokens = append(tokens, *tk)
		}
		pos += shift
	}
	return tokens, nil
}

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Infix → RPN converter")
	for {
		fmt.Print("-> ")
		text, _ := reader.ReadString('\n')
		tokens, err := tokenize(text)
		if err != nil {
			fmt.Println(err)
			continue
		}
		p := &parser{input: tokens}
		p.expr()
		fmt.Println(p.output)
	}
}
