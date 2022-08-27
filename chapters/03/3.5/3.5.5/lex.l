%{

#include<stdio.h>

SELECT, FROM, WHERE, ID

%}

%option noyywrap

delim	[ \t\n]
ws 		{delim}+
digit	[0-9]
letter	[A-Za-z]
id		{letter}({letter}|{digit})*
select	(select|SELECT)
where	(where|WHERE)
from	(from|FROM)

%%

{ws}		{}
{select}	{return(SELECT);}
{where}		{return(WHERE);}
{from}		{return(FROM);}
{id}		{yylval = (int) installID(); return(ID);}

%%

/* install the lexeme, whose first character is pointed to by yytext, and whose
 * length is yyleng, into the symbol table and return a pointer thereto the
 * insert should be case insensitive
 */ 
int installID();


