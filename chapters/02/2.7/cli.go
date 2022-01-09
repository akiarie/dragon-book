package main

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"unicode"
)

type stream []token

type parser struct {
	pos    int
	raw    string
	input  stream
	output string
}

func (p *parser) error(msg string) {
	panic(fmt.Sprintf("Error: %s at %d %q in %q", msg, p.input[p.pos].pos, p.input[p.pos].value, p.raw))
}

func (p *parser) Write(bytes []byte) (n int, err error) {
	p.output += string(bytes)
	return len(bytes), nil
}

type table struct {
	m    map[string]string
	prev *table
}

func (t *table) getval(key string, p *parser) string {
	if val, ok := t.m[key]; ok {
		return val
	} else if t.prev != nil {
		return t.prev.getval(key, p)
	}
	p.error(fmt.Sprintf("no type stored for %s", key))
	return ""
}

func (p *parser) block(prev *table) {
	symtbl := &table{m: make(map[string]string), prev: prev}
	p.punct('{')
	fmt.Fprintf(p, "{ ")
	p.decls(symtbl)
	p.stmts(symtbl)
	p.punct('}')
	fmt.Fprintf(p, "}")
}

func (p *parser) punct(c byte) {
	tk := p.input[p.pos]
	if tk.class != tkPunctuation || tk.value[0] != c {
		p.error(fmt.Sprintf("expected %q got %q", c, tk.value))
	}
	p.pos++
}

func (p *parser) decls(symtbl *table) {
	for p.pos+2 < len(p.input) {
		var typ, id string
		if tk := p.input[p.pos]; tk.class != tkType {
			break
		} else {
			typ = tk.value
		}
		p.pos++
		if tk := p.input[p.pos]; tk.class != tkId {
			p.error(fmt.Sprintf("expected identifier got %q", tk.value))
		} else {
			id = tk.value
		}
		symtbl.m[id] = typ
		p.pos++
		p.punct(';') // includes its own increment of pos
	}
}

func (p *parser) stmts(symtbl *table) {
	for p.pos < len(p.input) {
		switch tk := p.input[p.pos]; tk.class {
		case tkPunctuation:
			switch tk.value[0] {
			case '}':
				return
			case ';':
				p.error(fmt.Sprintf("unexpected ;"))
			}
			p.block(symtbl)
			fmt.Fprintf(p, " ")
			continue
		case tkId:
			if p.pos+1 >= len(p.input) {
				p.error(fmt.Sprintf("invalid length identifier"))
			}
			p.pos++
			fmt.Fprintf(p, "%s:%s", tk.value, symtbl.getval(tk.value, p))
			p.punct(';')
			fmt.Fprintf(p, "; ")
			continue
		default:
			return
		}
	}
}

type tokenclass string

const (
	tkPunctuation tokenclass = "punct"
	tkSpace                  = "[space]"
	tkType                   = "type"
	tkId                     = "id" // distinct from type b/c cannot be parsed as type
)

type token struct {
	class tokenclass
	value string
	pos   int // position of lexeme in input stream
}

func (tk token) String() string {
	return tk.value
}

func parsetoken(input string, pos int) (*token, int, error) {
	st := pos
	for _, c := range input[pos:] {
		if !unicode.IsSpace(c) {
			break
		}
		st++
	}
	if st > pos {
		if st >= len(input) {
			return &token{class: tkSpace, value: tkSpace, pos: pos}, len(input[pos:]), nil
		}
		// recurse & increment
		tk, shift, err := parsetoken(input, st)
		return tk, (st - pos) + shift, err
	}

	switch c := input[pos]; c {
	case '{', '}', ';':
		return &token{class: tkPunctuation, value: fmt.Sprintf("%c", c), pos: pos}, 1, nil
	}

	for _, t := range []string{"int", "bool", "char"} {
		if strings.Index(input[pos:], t) == 0 {
			return &token{class: tkType, value: t, pos: pos}, len(t), nil
		}
	}

	re := regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*`)
	if match := re.FindString(input[pos:]); match != "" {
		return &token{class: tkId, value: match, pos: pos}, len(match), nil
	}
	return nil, -1, fmt.Errorf("Unknown characters %q", input[pos:])
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
	raw := "{ int x; char y; { bool y; x; y; } x; y; }"
	tokens, err := tokenize(raw)
	if err != nil {
		log.Fatalln(err)
	}
	p := &parser{input: tokens, raw: raw}
	p.block(&table{m: make(map[string]string)})
	fmt.Println(p.output)
}
