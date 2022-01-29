# 2.8 — Intermediate Code Generation.
Given a block of code such as
```java
{
    int i; int j; float[100] a; float v; float x;
    while ( true ) {
        do i = i+1; while ( a[i] < v );
        do j = j-1; while ( a[j] > v ); 
        if ( i >= j ) break;
        x = a[i]; a[i] = a[j]; a[j] = x;
    } 
}
```
we intend to translate to something of the form
```C
01. i = i + 1
02. t1 = a [ i ]
03. if t1 < v goto 1
04. j = j - 1
05. t2 = a [ j ]
06. if t2 > v goto 4
07. ifFalse i >= j goto 9
08. goto 14
09. x = a [ i ]
10. t3 = a [ j ]
11. a [ i ] = t3
12. a [ j ] = x
13. goto 1
14.
```
The second representation is known as _three-address code_.

## Grammar
We begin with a modified version of the grammar in _Figure 2.39_. In essence, the idea is to extend
the operators to match those above (in the block of Java code).
```
program → block
block   → '{' stmts '}'

stmts   → stmts stmt 
        | ε

stmt    → expr ;
        | decl ;
        | command ;
        | if ( expr ) stmt
        | while ( expr ) stmt
        | do stmt while ( expr );
        | block

command → 'break'

decl    → type id
        | type '[' num ']' id

type    → int | float

expr    → rel = expr
        | rel

rel     → arithm
        | rel boolop arithm

boolop  → < | > | <= | >= | == | !=

arithm  → arithm + term
        | arithm - term
        | term

term    → term * factor
        | term / factor
        | term % factor
        | factor

factor  → ( expr )
        | num
        | boolean
        | id
        | id '[' arithm ']'

boolean → true | false
id      → [a-zA-Z_][a-zA-Z0-9_]+
num     → [0-9]+
```
I give the productions for _id_ and _num_ as regular expressions to avoid prolixity. Also, the
`boolean` type is defined to have the numerical values `true = 1` and `false = 0`, as one would
expect.

The syntax tree (following the layout in the book) will consist of nonterminals each represented as
a `node`,
```go
type node interface {
    gen() (string, error)
}
```
with appropriate concretisations `stmt` and `expr`
```go
type stmt node

type expr node

```
