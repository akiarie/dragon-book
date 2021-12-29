package main

import (
	"fmt"
	"log"

	"github.com/akiarie/dragon-tests/grammar"
)

// Exercise 2.4.1
func ex241() {
	gA := grammar.Grammar{grammar.Nonterminal{"S", []string{"+ S S", "- S S", "a"}}}
	/*
		The production
			S → S ( S ) S | ε
		as given is left-recursive. We rewrite using
			S → ε R
			R → ( S ) S R | ε,
		which leads to a further simplification, namely
			S → ( S ) S S | ε.
	*/
	gB := grammar.Grammar{grammar.Nonterminal{"S", []string{"( S ) S S", "ε"}}}
	gC := grammar.Grammar{grammar.Nonterminal{"S", []string{"0 S 1", "0 1"}}}
	for _, G := range []grammar.Grammar{gA, gB, gC} {
		if err := G.Validate(); err != nil {
			log.Fatal(err)
		}
		fmt.Println(G)
	}
	treeA, err := gA.ParseAST([]byte(`+ - + a a + a a a`))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(treeA)
	treeB, err := gB.ParseAST([]byte(`()`))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(treeB)
	treeC, err := gC.ParseAST([]byte(`00 001 111`))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(treeC)
}
