/*
 * lex input string and output tokens for simplicity:
 *
 * Grammar:
 * 		  stmt  → if expr then stmt	
 * 		 	    | if expr then stmt else stmt
 * 		 	    | ε
 * 		  expr  → term relop term
 * 		 	    | term
 * 		  term  → id
 * 			    | number
 *
 * 		  digit → [0-9]
 * 		 digits → digit+
 * 		 number → digits ( . digits )? ( E [+-]? digits )?
 * 		 letter → [A-Za-z]
 * 		     id → letter ( letter | digit )*
 * 		     if → if
 * 		   then → then
 * 		   else → else
 * 		  relop → < | > | <= | >= | = | <>
 */

#include <stdio.h>
#include <stdlib.h>
#include <string.h>

void* initial(char* input);

/* state 0 */
void* relop(char* input)
{
	switch (*input) {
		case '=': /* state 5 */
			printf("relop EQ\n");
			return initial(++input);
		case '<': /* state 1 */
			switch (*++input) {
				case '=': /* state 2 */
					printf("relop LE: %s\n", input);
					input++;
					break;
				case '>': /* state 3 */
					printf("relop NE\n");
					input++;
					break;
				default: /* state 4 */
					printf("relop LT\n");
			}
			return initial(input); /* incremented above */
		case '>': /* state 6 */
			switch (*++input) {
				case '=': /* state 7 */
					printf("relop GE\n");
					input++;
					break;
				default: /* state 8 */
					printf("relop GT\n");
			}
			return initial(input); /* incremented above */
	}
	fprintf(stderr, "unknown relop char '%c'\n", *input);
	exit(1);
}

void* identifier(char *input)
{
	fprintf(stderr, "identifier NOT IMPLEMENTED\n");
	exit(1);
}

void* number(char *input)
{
	fprintf(stderr, "number NOT IMPLEMENTED\n");
	exit(1);
}

void* initial(char* input)
{
	switch (input[0]) {
		case '\0':
			return NULL;
		case ' ': case '\t': case '\n':
			return initial(++input);
		case '<': case '=': case '>':
			return relop(input);
	}
	if (('a' <= input[0] && input[0] <= 'z') || 
		('A' <= input[0] && input[0] <= 'Z')) {
		printf("'%c' is an identifier\n", input[0]);
		return identifier(input);
	}
	if ('0' <= input[0] && input[0] <= '9') {
		return number(input);
	}
	fprintf(stderr, "unknown char '%c'\n", input[0]);
	exit(1);
}


int main(int argc, char *argv[])
{
	if (argc != 2) {
		fprintf(stderr, "must provide input as string\n");
		return 1;
	}
	char* input = argv[1];
	typedef void* (*state)(char*);
	for (state st = &initial; st != NULL; st = (state)st(input));
}
