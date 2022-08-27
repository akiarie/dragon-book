%{

#include<stdio.h>

void printcons();
void printvowel();
void printyy();

%}

%option noyywrap

delim	[ \t\n]
ws 		{delim}+
letter	[A-Za-z]
vowel	[AEIOUaeiou]
cons	[BCDFGHJKLMNPQRSTVWXYZbcdfghjklmnpqrstvwxyz]
vword	{vowel}{letter}*
cword	{cons}{letter}*

%%

{ws}	{printyy();}
{vword}	{printvowel();}
{cword}	{printcons();}

%%

void printyy()
{
	printf("%.*s", (int)yyleng, yytext);
}

void printvowel()
{
	printf("%.*say", (int)yyleng, yytext);
}

void printcons()
{
	printf("%.*s%cay", (int)yyleng, yytext+1, *yytext);
}

int main(int argc, char* argv[])
{
	if (argc != 2) {
		perror("must supply input string");
		exit(1);
	}
	yy_scan_string(argv[1]);
	yylex(); 
	printf("\n");
}
