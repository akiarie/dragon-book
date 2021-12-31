package main

import (
	"fmt"
	"log"
	"strings"
)

/*
	The nonterminal
		expr → expr + term { print('+') }
			 | expr - term { print('-') }
			 | term
	as given is left-recursive. Ignoring the print commands we have
		expr → expr + term
			 | expr - term
			 | term
	If we write
		rest → + term rest
		  	 | - term rest
		  	 | ε
	then we may re-write
		expr → term rest
	Re-introducing the print commands and re-writing in full we obtain
		expr → term rest

		rest → + term { print('+') } rest
		  	 | - term { print('-') } rest
		  	 | ε

		term → 0 { print('0') }
			 | 1 { print('1') }
				...
			 | 9 { print('9') }
*/

type parser struct {
	pos           int
	input, output string
}

type stateFn func(*parser) stateFn

func errormsg(err error, p *parser) stateFn {
	return func(p *parser) stateFn {
		log.Fatalln(err)
		return nil
	}
}

func expr(p *parser) stateFn {
	if err := p.term(); err != nil {
		return errormsg(err, p)
	}
	return rest
}

func rest(p *parser) stateFn {
	var op byte
	switch p.input[p.pos] {
	case '+':
		op = '+'
	case '-':
		op = '-'
	default:
		return errormsg(fmt.Errorf("%q is not '+' or '-'", p.input[p.pos]), p)
	}
	p.pos++
	if err := p.term(); err != nil {
		return errormsg(err, p)
	}
	p.output += fmt.Sprintf("%c", op)
	return rest
}

func (p *parser) term() error {
	if strings.IndexByte("0123456789", p.input[p.pos]) == -1 {
		return fmt.Errorf("%q is not a digit", p.input[p.pos])
	}
	p.output += fmt.Sprintf("%c", p.input[p.pos])
	p.pos++
	return nil
}

func main() {
	p := &parser{input: "9-7+3+5-2-5+2"}
	for state := expr(p); state != nil && p.pos < len(p.input); state = state(p) {
	}
	fmt.Println(p.output)
}
