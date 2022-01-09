package main

import (
	"fmt"
	"log"
	"strings"
)

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
