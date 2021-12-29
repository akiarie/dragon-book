package main

import (
	"fmt"

	"github.com/akiarie/dragon-tests/grammar"
)

// Exercise 2.1.5
func fig215() {
	/*
		The nonterminal
			expr → expr + term {print('+')}
				 | expr - term {print('-')}
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

			rest → + term {print('+')} rest
			  	 | - term {print('-')} rest
			  	 | ε

			term → 0 {print('0')}
				 | 1 {print('1')}
					...
				 | 9 {print('9')}
	*/
	digits := []string{}
	for i := 0; i < 10; i++ {
		digits = append(digits, fmt.Sprintf("%d {print('%d')}", i, i))
	}
	G := grammar.Grammar{
		grammar.Nonterminal{"expr", []string{"term rest"}},
		grammar.Nonterminal{"rest",
			[]string{
				"+ term {print('+')} rest",
				"- term {print('-')} rest",
				"ε",
			},
		},
		grammar.Nonterminal{"term", digits},
	}
	fmt.Println(G)
}
