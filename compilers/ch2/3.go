package main

import (
	"fmt"

	"github.com/akiarie/dragon-tests/grammar"
)

// Exercise 2.3.1
func ex231() {
	digits := []string{}
	for i := 0; i < 10; i++ {
		digits = append(digits, fmt.Sprintf("%d {print('%d')}", i, i))
	}
	G := grammar.Grammar{
		grammar.Nonterminal{"expr",
			[]string{
				"{print('+')} expr + term",
				"{print('-')} expr - term",
				"term",
			},
		},
		grammar.Nonterminal{"term",
			[]string{
				"{print('*')} term * factor",
				"{print('/')} term / factor",
				"factor",
			},
		},
		grammar.Nonterminal{"factor",
			[]string{
				"digit",
				"{print('(')} ( expr ) {print(')')}",
			},
		},
		grammar.Nonterminal{"digit", digits},
	}
	fmt.Println(G)
}
