package main

import (
	"fmt"
	"log"

	"github.com/akiarie/dragon-tests/grammar"
)

func main() {
	ex241()
}

// Figure 2.16
func fig216() {
	G := grammar.Grammar{
		grammar.Nonterminal{
			"stmt",
			[]string{
				"expr ;",
				"if ( expr ) stmt",
				"for ( optexpr ; optexpr ; optexpr ) stmt",
				"other",
			},
		},
		grammar.Nonterminal{
			"optexpr",
			[]string{
				"ε",
				"expr",
			},
		},
	}
	if err := G.Validate(); err != nil {
		log.Fatal(err)
	}
	tree, err := G.ParseAST([]byte(`for ( ; expr ; expr ) other`))
	if err != nil {
		log.Fatalf("Cannot parse input: %s\n", err)
	}
	fmt.Println(tree)
}
