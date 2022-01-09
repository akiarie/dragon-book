# 2.6
In this case we begin with the grammar
```
expr → expr + term { print('+') } | expr - term { print('-') } | term

term → term * factor { print('*') } | term / factor { print('/') } | factor

factor → ( expr ) | num { print(num.value) } | id { print(id.lexeme) }
```

Reasoning as previously, we re-write expr and term:
```
    expr → term rest

    rest → + term { print('+') } rest | - term { print('-') } rest | ε

    term → factor rise

    rise → * factor { print('*') } rise | / factor { print('/') } rise | ε
```
The productions for factor remain unchanged.
