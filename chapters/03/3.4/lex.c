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
#include <stdbool.h>
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

bool isreserved(char* tk)
{
	return strcmp(tk, "if") == 0
		|| strcmp(tk, "then") == 0
		|| strcmp(tk, "else") == 0;
}

void* identifier(char *input)
{
	/* state 10 */
	char *f = input + 1;
	while (*f != '\0') {
		if (( 'a' <= *f && *f <= 'z') ||
			( 'A' <= *f && *f <= 'Z') ||
			( '0' <= *f && *f <= '9')) {
			f++;
			continue; /* state 10 */
		}
		break; /* state 11 */
	}
	int len = f - input;
	char* id = (char*) malloc(sizeof(char) * (len + 1));
	snprintf(id, len + 1, "%s", input);
	if (isreserved(id)) {
		printf("reserved keyword '%s'\n", id);
	} 
	else {
		printf("identifier '%s'\n", id);
	}
	free(id);
	return initial(f);
}

/* natscan: scan any number of digits and return the number */
int natscan(char *f)
{	
	char *st = f;
	while (*f != '\0') {
		if ('0' <= *f && *f <= '9') {
			f++;
			continue;
		}
		break;
	}
	return f - st;
}

#define STATE13 13 /* . | E | digit | other */
#define STATE14 14 /* digit */
#define STATE15 15 /* E | digit | other */
#define STATE16 16 /* [+-] | digit */
#define STATE17 17 /* digit */
#define STATE18 18 /* digit | other */

int numstate(char c, int state) {
	switch (c) {
		case '.':
			if (state == STATE13) {
				return STATE14;
			}
			break;
		case 'E':
			if (state == STATE13 || state == STATE15) {
				return STATE16;
			}
			break;
		case '+': case '-':
			if (state == STATE16) {
				return STATE17;
			}
			break;
	}
	fprintf(stderr, "'%c' not allowed in state %d\n", c, state);
	exit(1);
}

char* scannumber(char *f) {
	int state = 13;
	while (true) {
		/* after the scan, *f is a non-digit, so it is either one of the
		 * permitted number symbols (based on the state) or an "other" char
		 * indicating termination of the number token */
		int ndigits = natscan(f);
		f += ndigits;
		switch (state) {
			/* for states 14 and 17 we must have ndigits == 1 */
			case STATE14: case STATE17:
				if (ndigits != 1) {
					fprintf(stderr, "1 digit needed in state %d\n", state);
					exit(1);
				}
				state++;
				break;
			case STATE16: /* for state 16 we must have ndigits ≤ 1 */
				if (ndigits > 1) {
					fprintf(stderr, "1≤ digit needed in state %d\n", state);
					exit(1);
				}
				if (ndigits == 1) {
					state = 18;
				}
				break;
		}
		switch (*f) {
			case '.': case 'E': case '+': case '-':
				state = numstate(*f, state);
				f++;
				break;
			case '\0':
			default: /* terminate if state permits it */
				if (state == STATE13 || state == STATE15 || state == STATE18) {
					return f;
				}
				fprintf(stderr, "'%c' not allowed in state %d\n", *f, state);
				exit(1);
		}
	}
}

void* number(char *input)
{
	char *f = scannumber(input + 1);
	int len = f - input;
	char* num = (char*) malloc(sizeof(char) * (len + 1));
	snprintf(num, len + 1, "%s", input);
	printf("number %s\n", num);
	free(num);
	return initial(f);
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
