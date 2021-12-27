package main

import (
	"strings"
	"testing"
)

func TestCli(t *testing.T) {
	for _, nt := range grammar {
		for _, prod := range nt {
			prodarr := strings.Fields(string(prod))
			for _, s := range prodarr {
				found := false
				if _, ok := parsetoken(s); ok {
					found = true
				} else {
					// if not token, check for nonterminal
					for subntname := range grammar {
						if s == subntname {
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
