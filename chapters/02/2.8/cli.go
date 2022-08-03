package main

import (
	"fmt"
	"log"
)

func main() {
	raw := `
{
    int i; int j; float[100] a; float v; float x;
    while ( true ) {
		j = j + 1;
        do i = i+1; while ( a[i] < v );
        do j = j-1; while ( a[j] > v );
        if ( i >= j ) break;
        x = a[i]; a[i] = a[j]; a[j] = x;
    }
}
`
	tokens, lines, err := tokenize(raw)
	if err != nil {
		log.Fatalln(err)
	}
	p := &parser{input: tokens, lines: lines, raw: raw}
	bl := p.block()
	if err := bl.gen(p, newtable()); err != nil {
		log.Fatalln(err)
	}
	fmt.Println(p.output)
}
