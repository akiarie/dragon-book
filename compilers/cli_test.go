package main

import (
	"strings"
	"testing"
)

func TestCli(t *testing.T) {
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
	for _, nt := range G {
		if len(nt.prods) == 0 {
			t.Errorf("Nonterminal %s with no productions", nt.name)
		}
		for _, prod := range nt.prods {
			prodarr := strings.Fields(string(prod))
			for _, s := range prodarr {
				found := false
				if _, ok := parsetoken(s); ok {
					found = true
				} else {
					// if not token, check for nonterminal
					for _, subnt := range G {
						if s == subnt.name {
							found = true
						}
					}
				}
				if !found {
					t.Errorf("Invalid symbol '%s' in grammar neither token nor nonterminal", s)
				}
			}
		}
	}
}
