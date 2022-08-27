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

{ws}		{printyy();}
{vword} 	{printvowel();}
{cword} 	{printcons();}

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

/* read_file: reads contents of file and returns them
 * caller must free returned string 
 * see https://stackoverflow.com/a/14002993 */
char* read_file(char *path)
{
    FILE *f = fopen(path, "rb");
    fseek(f, 0, SEEK_END);
    long fsize = ftell(f);
    fseek(f, 0, SEEK_SET);  /* same as rewind(f); */
    char *str = malloc(fsize + 1);
    fread(str, fsize, 1, f);
    fclose(f);
    str[fsize] = '\0';
    return str;
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
