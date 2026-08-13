# Updating

## Update from map

```go
q := sqlkata.NewQuery().From("users").
	WhereEq("id", 1).
	AsUpdate(map[string]any{
		"name":   "Ada",
		"active": true,
	})

fmt.Println(compiler.NewSqlServerCompiler().Compile(q).String())
// UPDATE [users] SET [name] = 'Ada', [active] = cast(1 as bit) WHERE [id] = 1
```

**Note:** Go map iteration order is random; for stable column order use `AsUpdateColumns`.

## Update columns + values

```go
q.AsUpdateColumns(
	[]string{"author", "date", "version"},
	[]any{"Author 1", nil, nil},
)
```

## Increment / decrement

```go
q.AsIncrement("balance", 50)
q.AsDecrement("balance", 10)
```

```sql
UPDATE [accounts] SET [balance] = [balance] + 50 WHERE …
```

## Executing updates

```go
n, err := db.ExecuteAffected(ctx, q)
```
