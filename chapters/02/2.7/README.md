# 2.7

We desire to implement a translator which can take in a statement such as

    { int x; char y; { bool y; x; y; } x; y; }

and output

    { { x:int; y:bool; } x:int; y:char; }

The form of the output is intended to demonstrate that the translator
managed to identify each of the statements to their respective declaration.

The grammar we shall work with is as follows:

    block → '{' decls stmts '}'
    decls → decl decls | ε
    stmts → stmt stmts | ε
    decl  → type id;
    stmt  → id; | block
    type  → int | bool | char
    id    → /[a-zA-Z][a-zA-Z0-9_]+/
