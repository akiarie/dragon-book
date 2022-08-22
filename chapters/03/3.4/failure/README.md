# Clarifying failure function computation.

```C
int nextf(int s, int t) 
{
	if (b[s+1] == b[t+1]) {
		return t+1;
	}
	if (t == 0) {
		return 0;
	}
	return nextf(s, f(t));
}

f(1) = 0;
for (int s = 1; s != n; s++) {
	f(s+1) = nextf(s, f(s));
}
```

The string matching algorithm, accordingly, becomes the following:

```C
int nexts(int s, int i) {
	if (a[i] == b[s+1]) {
		return s+1;
	}
	if (s == 0) {
		return 0;
	}
	return nexts(f(s), i);
}

int s = 0;
for (int i = 1; i <= m; i++) {
	s = nexts(s, i);
	if (s == n) {
		return "yes";
	}
}
return "no";
```
