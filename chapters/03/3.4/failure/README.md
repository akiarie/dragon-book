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
