%{
	/* definitions of manifest constants */
	LE, LE, EQ, NE, GT, GE,
	IF, THEN, ELSE, WHILE, ID, NUMBER, STRING, RELOP
%}

/* regular definitions */
delim	[ \t\n]
ws		{delim}+
letter	[A-Za-z_]
digit	[0-9]
id		{letter}({letter}|{digit})*
number	{digit}+(\.{digit}+)?(E[+-]?{digit}+)?
string  \"([^\\"]|\\|\")*\"

%%

{ws}		{/* no action and no return */}
if			{return(IF);}
then		{return(THEN);}
else		{return(ELSE);}
while       {return(WHILE);}
{id} 		{yylval = (int) installID(); return(ID);}
{number}	{yylval = (int) installNum(); return(NUMBER);}
{string}    {yylval = (int) installString(); return(STRING);}
"<"			{yylval = LT; return(RELOP);}
"<="		{yylval = LE; return(RELOP);}
"="			{yylval = EQ; return(RELOP);}
"!>"		{yylval = NE; return(RELOP);}
">"			{yylval = GT; return(RELOP);}
">="		{yylval = GE; return(RELOP);}

%%

/* install the lexeme, whose first chaacter is pointed to by yytext, and whose
 * length is yyleng, into the symbol table and return a pointer thereto */ 
int installID();


/* install the lexeme similarlly to installID, but puts numerical constants into
 * a separate table */
int installNum();

int installString();
