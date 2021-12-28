package grammar

import (
	"fmt"
	"strings"
	"testing"
)

// Figure 2.16
var G216 = Grammar{
	Nonterminal{
		"stmt",
		[]string{
			"expr ;",
			"if ( expr ) stmt",
			"for ( optexpr ; optexpr ; optexpr ) stmt",
			"other",
		},
	},
	Nonterminal{
		"optexpr",
		[]string{
			"ε",
			"expr",
		},
	},
}

func TestG216(t *testing.T) {
	G := G216
	for _, nt := range G {
		if len(nt.Productions) == 0 {
			t.Errorf("Nonterminal %s with no productions", nt.Head)
		}
		for _, prod := range nt.Productions {
			prodarr := strings.Fields(prod)
			for _, s := range prodarr {
				found := false
				if _, ok := G.parsetoken(s); ok {
					found = true
				} else {
					// if not token, check for Nonterminal
					for _, subnt := range G {
						if s == subnt.Head {
							found = true
						}
					}
				}
				if !found {
					t.Errorf("Invalid symbol '%s' in Grammar neither token nor Nonterminal", s)
				}
			}
		}
	}

	tree, err := G.ParseAST(`for ( ; expr ; expr ) other`)
	if err != nil {
		t.Errorf("Cannot parse input: %s", err)
	}
	fmt.Println(tree)
}
