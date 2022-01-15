package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
)

type stream []token

type parser struct {
	pos    int
	lines  []int
	raw    string
	input  stream
	output string
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// error is a wretched syntax-error reporter but I don't have the time or
// energy to clean it up right now :(
func (p *parser) error(format string, a ...interface{}) error {
	linelen := len(p.raw)
	linepos := 0
	linenum := 0
	pos := p.input[p.pos].pos
	for i, l := range p.lines {
		if pos < l {
			linelen = l - linepos
			break
		}
		linenum = i
		linepos = l
	}
	fmt.Fprintf(os.Stderr, "error: %s at position %d on line %d:\n", fmt.Sprintf(format, a...), pos-linepos, linenum+1)
	if linenum > 0 {
		fmt.Fprintf(os.Stderr, "%d %s\n", linenum, p.raw[p.lines[linenum-1]+1:linepos])
	}
	bftk := fmt.Sprintf("%d %s", linenum+1, p.raw[linepos+1:pos])
	aftk := fmt.Sprintf("%s\n", p.raw[pos+len(p.input[p.pos].value):linepos+linelen])
	red := color.New(color.FgRed, color.Bold)
	red.Fprint(os.Stderr, bftk)
	color.New(color.Bold).Fprint(os.Stderr, p.input[p.pos].value)
	red.Fprint(os.Stderr, aftk)
	if linenum+1 < len(p.lines) {
		fmt.Fprintf(os.Stderr, "%d %s\n", linenum+2, p.raw[p.lines[linenum+2]+1:p.lines[min(linenum+3, len(p.lines))]])
	}
	os.Exit(1)
	return fmt.Errorf("HACK")
}

func (p *parser) Write(bytes []byte) (n int, err error) {
	p.output += string(bytes)
	return len(bytes), nil
}

type node interface {
	gen() (string, error)
}

func (p *parser) block() {
	p.punct('{')
	for p.pos < len(p.input) {
		// if next token is not end of block, must parse stmt
		if tk := p.input[p.pos]; tk.value == "}" {
			break
		}
		p.parsestmt()
	}
	p.punct('}')
}

func (p *parser) punct(c byte) {
	tk := p.input[p.pos]
	if tk.class != tkPunctuation || tk.value[0] != c {
		p.error(fmt.Sprintf("expected %q got %q near", c, tk.value))
	}
	p.pos++
}

func (p *parser) parsestmt() {
	// expr ;
	if step, err := p.expr(p.input[p.pos:]); err == nil {
		p.pos += step
		p.punct(';')
		return
	}

	// decl ;
	if tk := p.input[p.pos]; tk.class == tkType {
		p.pos++

		// arrays
		if tk := p.input[p.pos]; tk.class == tkPunctuation && tk.value == "[" {
			p.punct('[')
			if tk := p.input[p.pos]; tk.class != tkNum {
				p.error("array declaration must have number as size")
			}
			p.pos++
			p.punct(']')
		}

		if tk := p.input[p.pos]; tk.class != tkId {
			p.error("declaration must have identifier")
		}
		p.pos++
		p.punct(';')
		return
	}

	// if ( expr ) stmt
	if tk := p.input[p.pos]; tk.class == tkKeyword && tk.value == "if" {
		p.pos++
		p.punct('(')
		if step, err := p.expr(p.input[p.pos:]); err == nil {
			p.pos += step
			p.punct(')')
			p.parsestmt()
			return
		} else {
			p.error("if statement must include expression: %v", err)
		}
	}

	// while ( expr ) stmt
	if tk := p.input[p.pos]; tk.class == tkKeyword && tk.value == "while" {
		p.pos++
		p.punct('(')
		if step, err := p.expr(p.input[p.pos:]); err == nil {
			p.pos += step
			p.punct(')')
			p.parsestmt()
			return
		} else {
			p.error("while statement must include expression: %v", err)
		}
	}

	// do stmt while ( expr ) ;
	if tk := p.input[p.pos]; tk.class == tkKeyword && tk.value == "do" {
		p.pos++
		p.parsestmt()
		if tk := p.input[p.pos]; tk.class != tkKeyword || tk.value != "while" {
			p.error("do while statement must include 'while'")
		}
		p.pos++
		p.punct('(')
		if step, err := p.expr(p.input[p.pos:]); err == nil {
			p.pos += step
			p.punct(')')
			p.punct(';')
			return
		} else {
			p.error("do while statement must include expression: %v", err)
		}
	}

	// break
	if tk := p.input[p.pos]; tk.class == tkKeyword && tk.value == "break" {
		p.pos++
		p.punct(';')
		return
	}

	p.block()
}

func (p *parser) expr(input stream) (int, error) {
	// boolean
	if tk := input[0]; tk.class == tkBool {
		return 1, nil
	}
	// rel
	step, err := p.rel(input)
	if err != nil {
		return -1, fmt.Errorf("cannot parse expr: %v")
	}
	// recurse on assign
	if step < len(input) {
		if tk := input[step]; tk.class == tkAssign {
			step++
			tot, err := p.expr(input[step:])
			if err != nil {
				return -1, p.error(err.Error())
			}
			return step + tot, nil
		}
	}
	return step, nil
}

func (p *parser) rel(input stream) (int, error) {
	step, err := p.arithm(input)
	if err != nil {
		return -1, fmt.Errorf("rel must start with arithm: %v", err)
	}
	// recurse on boolop
	if step < len(input) {
		if tk := input[step]; tk.class == tkRel {
			step++
			tot, err := p.rel(input[step:])
			if err != nil {
				return -1, p.error(err.Error())
			}
			return step + tot, nil
		}
	}
	return step, nil
}

func (p *parser) arithm(input stream) (int, error) {
	/*
		In the productions
			arithm  → arithm + term
					| arithm - term
					| term
		and
			term    → term * factor
					| term / factor
					| factor
		we begin by evaluating the last term in each to avoid infinite-recursion.
	*/

	step, err := p.term(input)
	if err != nil {
		return -1, fmt.Errorf("arithm must start with term: %v", err)
	}
	if step < len(input) {
		if strings.IndexByte("+-", input[step].value[0]) != -1 {
			step++
			tot, err := p.arithm(input[step:])
			if err != nil {
				return -1, p.error(err.Error())
			}
			return step + tot, nil
		}
	}
	return step, nil
}

func (p *parser) term(input stream) (int, error) {
	step, err := p.factor(input)
	if err != nil {
		return -1, fmt.Errorf("term must start with factor: %v", err)
	}
	if step < len(input) {
		if strings.IndexByte("*/%", input[step].value[0]) != -1 {
			step++
			tot, err := p.term(input[step:])
			if err != nil {
				return -1, p.error(err.Error())
			}
			return step + tot, nil
		}
	}
	return step, nil
}

func (p *parser) factor(input stream) (int, error) {
	// num | id
	switch input[0].class {
	case tkNum:
		return 1, nil
	case tkId:
		if len(input) > 1 {
			pos := 1
			if tk := input[pos]; tk.class == tkPunctuation && tk.value == "[" {
				pos++
				step, err := p.arithm(input[pos:])
				if err != nil {
					return -1, p.error("array access must be arithmetic: %v", err)
				}
				pos += step
				if tk := input[pos]; tk.class != tkPunctuation || tk.value != "]" {
					return -1, p.error("array access not closed: %v", err)
				}
				return pos + 1, nil
			}
		}
		return 1, nil
	}

	// ( expr )
	if tk := input[0]; tk.class != tkPunctuation || tk.value != "(" {
		return -1, fmt.Errorf("cannot parse factor %q", tk)
	}
	step, err := p.expr(input[1:])
	if err != nil {
		return -1, p.error("bracketed expression parse error: %v", err)
	}
	if tk := input[1+step]; tk.class != tkPunctuation || tk.value != ")" {
		return -1, p.error("bracketed not closed: %v", err)
	}
	return 2 + step, nil
}
