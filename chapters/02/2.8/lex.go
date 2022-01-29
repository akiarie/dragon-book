package main

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

type tokenclass string

const (
	tkPunctuation tokenclass = "punct"
	tkSpace                  = "[space]"
	tkKeyword                = "keyword"
	tkType                   = "type"
	tkId                     = "id" // distinct from type & keyword b/c cannot be parsed as type
	tkNum                    = "num"
	tkBool                   = "bool"
	tkOp                     = "op"
	tkRel                    = "rel"
	tkAssign                 = "assign"
)

type token struct {
	class tokenclass
	value string
	pos   int // position of lexeme in input stream
}

func (tk token) String() string {
	return tk.value
}

type lexer struct {
	input string
	pos   int
	lines []int
}

func parsetoken(l *lexer) (*token, error) {
	// space
	st := l.pos
	for i, c := range l.input[l.pos:] {
		if !unicode.IsSpace(c) {
			break
		}
		if c == '\n' {
			l.lines = append(l.lines, i+l.pos)
		}
		st++
	}
	if st > l.pos {
		if st >= len(l.input) {
			tk := &token{class: tkSpace, value: tkSpace, pos: l.pos}
			l.pos += len(l.input[l.pos:])
			return tk, nil
		}
		// recurse & increment
		l.pos += st - l.pos
		return parsetoken(l)
	}

	// punct
	if strings.IndexByte("{}()[];", l.input[l.pos]) != -1 {
		tk := &token{class: tkPunctuation, value: fmt.Sprintf("%c", l.input[l.pos]), pos: l.pos}
		l.pos++
		return tk, nil
	}

	// rel
	if strings.IndexByte("=<>!", l.input[l.pos]) != -1 {
		if l.pos+1 < len(l.input) && l.input[l.pos+1] == '=' {
			tk := &token{class: tkRel, value: l.input[l.pos : l.pos+2], pos: l.pos}
			l.pos += 2
			return tk, nil
		}
		switch c := l.input[l.pos]; c {
		case '=':
			tk := &token{class: tkAssign, value: "=", pos: l.pos}
			l.pos++
			return tk, nil
		case '<', '>':
			tk := &token{class: tkRel, value: fmt.Sprintf("%c", c), pos: l.pos}
			l.pos++
			return tk, nil
		}
	}

	// op
	if strings.IndexByte("+-*/", l.input[l.pos]) != -1 {
		tk := &token{class: tkOp, value: fmt.Sprintf("%c", l.input[l.pos]), pos: l.pos}
		l.pos++
		return tk, nil
	}

	// type
	for _, t := range []string{"int", "float"} {
		if strings.Index(l.input[l.pos:], t) == 0 {
			tk := &token{class: tkType, value: t, pos: l.pos}
			l.pos += len(t)
			return tk, nil
		}
	}

	// keyword
	for _, t := range []string{"do", "while", "if", "break"} {
		if strings.Index(l.input[l.pos:], t) == 0 {
			tk := &token{class: tkKeyword, value: t, pos: l.pos}
			l.pos += len(t)
			return tk, nil
		}
	}

	// bool
	for _, t := range []string{"false", "true"} {
		if strings.Index(l.input[l.pos:], t) == 0 {
			tk := &token{class: tkNum, value: t, pos: l.pos}
			l.pos += len(t)
			return tk, nil
		}
	}

	// id
	re := regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*`)
	if match := re.FindString(l.input[l.pos:]); match != "" {
		tk := &token{class: tkId, value: match, pos: l.pos}
		l.pos += len(match)
		return tk, nil
	}

	// num
	re = regexp.MustCompile(`[0-9]+`)
	if match := re.FindString(l.input[l.pos:]); match != "" {
		tk := &token{class: tkNum, value: match, pos: l.pos}
		l.pos += len(match)
		return tk, nil
	}

	return nil, fmt.Errorf("Unknown characters %q", l.input[l.pos:])
}

func tokenize(input string) ([]token, []int, error) {
	l := &lexer{input: input}
	tokens := []token{}
	for l.pos < len(input) {
		tk, err := parsetoken(l)
		if err != nil {
			return nil, nil, err
		}
		if tk.class != tkSpace {
			tokens = append(tokens, *tk)
		}
	}
	return tokens, l.lines, nil
}
