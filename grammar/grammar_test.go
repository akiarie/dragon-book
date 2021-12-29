package grammar

/*
func TestBNF(t *testing.T) {
	G := Grammar{
		Nonterminal{"bnfgrammar", []string{"bnfgrammar nonterminal", "nonterminal"}},
		Nonterminal{"nonterminal", []string{"nonterminal `|` production", "name `→` production"}},
		Nonterminal{"production", []string{"production symbol", "symbol"}},
		Nonterminal{"name",
			[]string{
				"name||alpha",
				"name||digit",
				"alpha",
			},
		},
		Nonterminal{"symbol",
			[]string{
				"symbol||symbol",
				"alpha",
				"digit",
				"specialchar",
				"quote||bnfchar||quote",
				"escquote||quote||escquote",
				"ε",
			},
		},
		Nonterminal{"alpha", strings.Split("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz", "")},
		Nonterminal{"digit", strings.Split("0123456789", "")},
		Nonterminal{"specialchar", append(strings.Split("!?@#$%^&*()-_—=+[]{}();:,.<>\"\\/", ""), "escquote")},
		Nonterminal{"escquote", []string{"'"}},
		Nonterminal{"bnfchar", []string{"`|`", "`||`", "`→`"}},
		Nonterminal{"quote", []string{"`"}},
	}
	_ = G
		fmt.Println(G.terminals())

		g216file, err := os.ReadFile("bnf/dragon-216.grm")
		if err != nil {
			t.Errorf("Cannot read file: %s", err)
		}
		tree, err := G.ParseAST(g216file)
		if err != nil {
			t.Errorf("Cannot parse input: %s", err)
		}
		fmt.Println(tree)
}
*/
