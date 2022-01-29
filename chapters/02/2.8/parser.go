package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
)

type table struct {
	labels []string
	vars   []string
}

func (t *table) newvar() string {
	t.vars = append(t.vars, fmt.Sprintf("t%d", len(t.vars)))
	return t.vars[len(t.vars)-1]
}

func (t *table) newlabel() string {
	t.labels = append(t.labels, fmt.Sprintf("L%d", len(t.labels)))
	return t.labels[len(t.labels)-1]
}

type node interface {
	gen(*parser, *table) error
}

type block []node

func (b block) gen(p *parser, t *table) error {
	for _, stmt := range b {
		if err := stmt.gen(p, t); err != nil {
			return err
		}
	}
	return nil
}

type command int

const comBreak command = iota

func (com command) gen(p *parser, t *table) error {
	return fmt.Errorf("command NOT IMPLEMENTED")
}

type decl struct {
	id,
	_type token
	num *token // ptr to indicate w/not this is array
}

func (d decl) String() string {
	if d.num == nil {
		return fmt.Sprintf("decl{%s %s}", d._type, d.id)
	}
	return fmt.Sprintf("decl{%s[%s] %s}", d._type, d.num, d.id)
}

func (d decl) gen(p *parser, t *table) error {
	return nil // TODO: does this make sense?
}

type ifstmt struct {
	expr expr
	stmt node
}

func (_if ifstmt) gen(p *parser, t *table) error {
	after := t.newlabel()
	_bool, err := p.rvalue(_if.expr, t)
	if err != nil {
		return err
	}
	fmt.Fprintf(p, "ifFalse %s goto %s\n", _bool, after)
	err = _if.stmt.gen(p, t)
	if err != nil {
		return err
	}
	fmt.Fprintf(p, "%s:\n", after)
	return nil
}

type whilestmt struct {
	expr expr
	stmt node
}

func (stmt whilestmt) String() string {
	return fmt.Sprintf("while ( %v ) { %v }", stmt.expr, stmt.stmt)
}

func (while whilestmt) gen(p *parser, t *table) error {
	before, after := t.newlabel(), t.newlabel()
	fmt.Fprintf(p, "%s:\n", before)
	_bool, err := p.rvalue(while.expr, t)
	if err != nil {
		return err
	}
	fmt.Fprintf(p, "ifFalse %s goto %s\n", _bool, after)
	err = while.stmt.gen(p, t)
	if err != nil {
		return err
	}
	fmt.Fprintf(p, "goto %s:\n", before)
	fmt.Fprintf(p, "%s:\n", after)
	return nil
}

type dostmt struct {
	expr expr
	stmt node
}

func (do dostmt) gen(p *parser, t *table) error {
	before, after := t.newlabel(), t.newlabel()
	fmt.Fprintf(p, "%s:\n", before)
	err := do.stmt.gen(p, t)
	if err != nil {
		return err
	}
	_bool, err := p.rvalue(do.expr, t)
	if err != nil {
		return err
	}
	fmt.Fprintf(p, "ifFalse %s goto %s\n", _bool, after)
	fmt.Fprintf(p, "goto %s:\n", before)
	fmt.Fprintf(p, "%s:\n", after)
	return nil
}

type composite interface {
	h() interface{}
	t() composite
	compose(string, string) string
}

type expr []rel

func (expr expr) h() interface{} { return expr[0] }

func (expr expr) t() composite {
	if len(expr) > 1 {
		return expr[1:]
	}
	return nil
}

func (expr expr) compose(a, b string) string { return fmt.Sprintf("%s = %s", a, b) }

func (expr expr) gen(p *parser, t *table) error {
	return fmt.Errorf("expr NOT IMPLEMENTED")
}

func (ex expr) String() string {
	rels := make([]string, len(ex))
	for i := 0; i < len(rels); i++ {
		rels[i] = fmt.Sprintf("%s", ex[i])
	}
	return strings.Join(rels, " = ")
}

type rel struct {
	head arithm
	tail *struct {
		boolop token
		rel
	}
}

func (rel rel) h() interface{} { return rel.head }
func (rel rel) t() composite   { return rel.tail }

func (rel rel) compose(a, b string) string {
	return fmt.Sprintf("%s %s %s", a, rel.tail.boolop, b)
}

func (t rel) String() string {
	if t.tail == nil {
		return fmt.Sprintf("%v", t.head)
	}
	return fmt.Sprintf("%s %v %s", t.head, t.tail.boolop, t.tail.rel)
}

type arithm struct {
	head term
	tail *struct {
		sign token
		arithm
	}
}

func (arithm arithm) h() interface{} { return arithm.head }
func (arithm arithm) t() composite   { return arithm.tail }
func (arithm arithm) compose(a, b string) string {
	return fmt.Sprintf("%s %v %s", a, arithm.tail.sign, b)
}

func (t arithm) String() string {
	if t.tail == nil {
		return fmt.Sprintf("%v", t.head)
	}
	return fmt.Sprintf("%s %v %s", t.head, t.tail.sign, t.tail.arithm)
}

type term struct {
	head factor
	tail *struct {
		op token
		term
	}
}

func (term term) h() interface{} { return term.head }
func (term term) t() composite   { return term.tail }
func (term term) compose(a, b string) string {
	return fmt.Sprintf("%s %v %s", a, term.tail.op, b)
}

func (t term) String() string {
	if t.tail == nil {
		return fmt.Sprintf("%v", t.head)
	}
	return fmt.Sprintf("%s %v %s", t.head, t.tail.op, t.tail.term)
}

type factype int

const (
	factypeId factype = iota
	factypeBool
	factypeConst
	factypeAccess
	factypeExpr
)

type factor struct {
	ftype factype
	node  interface{}
}

func (f factor) rvalue(p *parser, t *table) (string, error) {
	switch f.ftype {
	case factypeId, factypeBool, factypeConst:
		return fmt.Sprintf("%v", f.node), nil
	case factypeAccess:
		arr, ok := f.node.(access)
		if !ok {
			panic(fmt.Sprintf("invalid access factype for %v", f))
		}
		tmp := t.newvar()
		lval, err := p.lvalue(arr, t)
		if err != nil {
			return "", err
		}
		fmt.Fprintf(p, "%s = %s\n", tmp, lval)
		return tmp, nil
	case factypeExpr:
		expr, ok := f.node.(expr)
		if !ok {
			panic(fmt.Sprintf("invalid expr factype for %v", f))
		}

	default:
		return "", fmt.Errorf("unknown factor %T as %v", f, f)
	}

}

type access struct {
	id string
	arithm
}

type stream []token

type parser struct {
	pos    int
	lines  []int
	raw    string
	input  stream
	output string
}

func (p *parser) Write(b []byte) (int, error) {
	fmt.Printf(string(b))
	p.output += string(b)
	return len(b), nil
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

func (p *parser) lvalue(stmt interface{}, t *table) (string, error) {
	return "", fmt.Errorf("lvalue NOT IMPLEMENTED")
}

func (p *parser) rvalue(stmt interface{}, t *table) (string, error) {
	switch stmt.(type) {
	case factor:
		f, _ := stmt.(factor)
		return f.rvalue(p, t)
	case composite:
	default:
		return "", fmt.Errorf("no rvalue for %T as %v", stmt, stmt)
	}
}

func (p *parser) block() block {
	p.punct('{')
	stmts := []node{}
	for p.pos < len(p.input) {
		// if next token is not end of block, must parse stmt
		if tk := p.input[p.pos]; tk.value == "}" {
			break
		}
		stmts = append(stmts, p.stmt())
	}
	p.punct('}')
	return stmts
}

func (p *parser) punct(c byte) {
	tk := p.input[p.pos]
	if tk.class != tkPunctuation || tk.value[0] != c {
		p.error(fmt.Sprintf("expected %q got %q near", c, tk.value))
	}
	p.pos++
}

func (p *parser) stmt() node {
	// expr ;
	if expr, step, err := p.expr(p.input[p.pos:]); err == nil {
		p.pos += step
		p.punct(';')
		return expr
	}

	// decl ;
	if _type := p.input[p.pos]; _type.class == tkType {
		p.pos++

		// arrays
		var num *token
		if tk := p.input[p.pos]; tk.class == tkPunctuation && tk.value == "[" {
			p.punct('[')
			if tk := p.input[p.pos]; tk.class != tkNum {
				p.error("array declaration must have number as size")
			} else {
				num = &tk
			}
			p.pos++
			p.punct(']')
		}

		id := p.input[p.pos]
		if id.class != tkId {
			p.error("declaration must have identifier")
		}
		p.pos++
		p.punct(';')
		return &decl{id, _type, num}
	}

	// if ( expr ) stmt
	if tk := p.input[p.pos]; tk.class == tkKeyword && tk.value == "if" {
		p.pos++
		p.punct('(')
		if expr, step, err := p.expr(p.input[p.pos:]); err == nil {
			p.pos += step
			p.punct(')')
			return ifstmt{expr, p.stmt()}
		} else {
			p.error("if statement must include expression: %v", err)
		}
	}

	// while ( expr ) stmt
	if tk := p.input[p.pos]; tk.class == tkKeyword && tk.value == "while" {
		p.pos++
		p.punct('(')
		if expr, step, err := p.expr(p.input[p.pos:]); err == nil {
			p.pos += step
			p.punct(')')
			return whilestmt{expr, p.stmt()}
		} else {
			p.error("while statement must include expression: %v", err)
		}
	}

	// do stmt while ( expr ) ;
	if tk := p.input[p.pos]; tk.class == tkKeyword && tk.value == "do" {
		p.pos++
		stmt := p.stmt()
		if tk := p.input[p.pos]; tk.class != tkKeyword || tk.value != "while" {
			p.error("do while statement must include 'while'")
		}
		p.pos++
		p.punct('(')
		if expr, step, err := p.expr(p.input[p.pos:]); err == nil {
			p.pos += step
			p.punct(')')
			p.punct(';')
			return dostmt{expr, stmt}
		} else {
			p.error("do while statement must include expression: %v", err)
		}
	}

	// break
	if tk := p.input[p.pos]; tk.class == tkKeyword && tk.value == "break" {
		p.pos++
		p.punct(';')
		return comBreak
	}

	return p.block()
}

func (p *parser) expr(input stream) (expr, int, error) {
	// rel
	rel, step, err := p.rel(input)
	if err != nil {
		return nil, -1, fmt.Errorf("cannot parse expr: %v", err)
	}
	// recurse on assign
	if step < len(input) {
		if tk := input[step]; tk.class == tkAssign {
			step++
			prev, tot, err := p.expr(input[step:])
			if err != nil {
				return nil, -1, p.error(err.Error())
			}
			return append(expr{*rel}, prev...), step + tot, nil
		}
	}
	return expr{*rel}, step, nil
}

func (p *parser) rel(input stream) (*rel, int, error) {
	arithm, step, err := p.arithm(input)
	if err != nil {
		return nil, -1, fmt.Errorf("rel must start with arithm: %v", err)
	}
	// recurse on boolop
	if step < len(input) {
		if tk := input[step]; tk.class == tkRel {
			step++
			prev, tot, err := p.rel(input[step:])
			if err != nil {
				return nil, -1, p.error(err.Error())
			}
			tail := struct {
				boolop token
				rel
			}{tk, *prev}
			return &rel{*arithm, &tail}, step + tot, nil
		}
	}
	return &rel{head: *arithm}, step, nil
}

func (p *parser) arithm(input stream) (*arithm, int, error) {
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

	term, step, err := p.term(input)
	if err != nil {
		return nil, -1, fmt.Errorf("arithm must start with term: %v", err)
	}
	if step < len(input) {
		if tk := input[step]; strings.IndexByte("+-", tk.value[0]) != -1 {
			step++
			prev, tot, err := p.arithm(input[step:])
			if err != nil {
				return nil, -1, p.error(err.Error())
			}
			tail := struct {
				sign token
				arithm
			}{tk, *prev}
			return &arithm{*term, &tail}, step + tot, nil
		}
	}
	return &arithm{head: *term}, step, nil
}

func (p *parser) term(input stream) (*term, int, error) {
	factor, step, err := p.factor(input)
	if err != nil {
		return nil, -1, fmt.Errorf("term must start with factor: %v", err)
	}
	if step < len(input) {
		if tk := input[step]; strings.IndexByte("*/%", tk.value[0]) != -1 {
			step++
			prev, tot, err := p.term(input[step:])
			if err != nil {
				return nil, -1, p.error(err.Error())
			}
			tail := struct {
				op token
				term
			}{tk, *prev}
			return &term{*factor, &tail}, step + tot, nil
		}
	}
	return &term{head: *factor}, step, nil
}

func (p *parser) factor(input stream) (*factor, int, error) {
	// num | id
	switch input[0].class {
	case tkBool:
		return &factor{factypeBool, input[0].value == "true"}, 1, nil
	case tkNum:
		return &factor{factypeConst, input[0].value}, 1, nil
	case tkId:
		if len(input) > 1 {
			pos := 1
			if tk := input[pos]; tk.class == tkPunctuation && tk.value == "[" {
				pos++
				arithm, step, err := p.arithm(input[pos:])
				if err != nil {
					return nil, -1, p.error("array access must be arithmetic: %v", err)
				}
				pos += step
				if tk := input[pos]; tk.class != tkPunctuation || tk.value != "]" {
					return nil, -1, p.error("array access not closed: %v", err)
				}
				return &factor{factypeAccess, access{input[0].value, *arithm}}, pos + 1, nil
			}
		}
		return &factor{factypeId, input[0].value}, 1, nil
	}

	// ( expr )
	if tk := input[0]; tk.class != tkPunctuation || tk.value != "(" {
		return nil, -1, fmt.Errorf("cannot parse factor %q", tk)
	}
	expr, step, err := p.expr(input[1:])
	if err != nil {
		return nil, -1, p.error("bracketed expression parse error: %v", err)
	}
	if tk := input[1+step]; tk.class != tkPunctuation || tk.value != ")" {
		return nil, -1, p.error("bracketed not closed: %v", err)
	}
	return &factor{factypeExpr, expr}, 2 + step, nil
}
