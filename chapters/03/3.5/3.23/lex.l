%{
	/* definitions of manifest constants */
	LE, LE, EQ, NE, GT, GE,
	IF, THEN, ELSE, ID, NUMBER, RELOP
%}

/* regular definitions */
delim	[ \t\n]
ws		{delim}+
letter	[A-Za-z]
digit	[0-9]
id		{letter}({letter}|{digit})*
number	{digit}+(\.{digit}+)?(E[+-]?{digit}+)?

%%

{ws}		{/* no action and no return */}
if			{return(IF);}
then		{return(THEN);}
else		{return(ELSE);}
{id} 		{yylval = (int) installID(); return(ID);}
{number}	{yylval = (int) installNum(); return(NUMBER);}
"<"			{yylval = LT; return(RELOP);}
"<="		{yylval = LE; return(RELOP);}
"="			{yylval = EQ; return(RELOP);}
"<>"		{yylval = NE; return(RELOP);}
">"			{yylval = GT; return(RELOP);}
">="		{yylval = GE; return(RELOP);}

%%

/* install the lexeme, whose first chaacter is pointed to by yytext, and whose
 * length is yyleng, into the symbol table and return a pointer thereto */ 
int installID();


/* install the lexeme similarlly to installID, but puts numerical constants into
 * a separate table */
int installNum();
