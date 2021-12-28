package main

import (
	"fmt"
	"strings"
)

func main() {
	G := grammar{
		nonterminal{
			"stmt",
			[]production{
				"expr ;",
				"if ( expr ) stmt",
				"for ( optexpr ; optexpr ; optexpr ) stmt",
				"other",
			},
		},
		nonterminal{
			"optexpr",
			[]production{
				"ε",
				"expr",
			},
		},
	}
	fmt.Println(G)
	input := `for ( ; expr ; expr ) other`
	lex := &lexer{
		input:  strings.TrimSpace(input),
		tokens: []token{},
	}
	for state := stateFn(tokenize); state != nil; state = state(lex) {
	}
	/*
		tree, err := G.parsetree(lex.tokens)
		if err != nil {
			log.Fatalln(err)
		}
		fmt.Println(tree)
	*/
}
