# Inserting

## Insert from map

```go
q := sqlkata.NewQuery().From("users").AsInsert(map[string]any{
	"name":   "Ada",
	"email":  "ada@example.com",
	"active": true,
}, false)

fmt.Println(compiler.NewSqlServerCompiler().Compile(q).String())
// INSERT INTO [users] (…) VALUES (…)
```

`AsInsert(values, returnId)` — when `returnId` is true, compilers may append identity SQL (e.g. SQL Server `scope_identity()`).

Also: `Insert(map)` helper on Query where available.

## Insert columns + values

```go
q.AsInsertColumns(
	[]string{"id", "author", "isbn", "date"},
	[]any{1, "Author 1", "123456", nil},
)
```

## Multi-row insert

```go
q.AsInsertRows(
	[]string{"name", "brand", "year"},
	[][]any{
		{"Chiron", "Bugatti", nil},
		{"Huayra", "Pagani", 2012},
	},
)
```

## Insert…Select

```go
q := sqlkata.NewQuery().From("users_archive").AsInsertQuery(
	[]string{"id", "name"},
	sqlkata.NewQuery().From("users").Select("id", "name").WhereEq("active", false),
)
```

## With CTE + insert

```go
q := sqlkata.NewQuery().
	WithAlias("old_cars", sqlkata.NewQuery().From("all_cars").Where("year", "<", 2000)).
	From("expensive_cars").
	AsInsertQuery([]string{"name", "model", "year"},
		sqlkata.NewQuery().From("old_cars").Where("price", ">", 100).ForPage(2, 10),
	)
```

## Executing inserts

```go
db := execution.New(sqlxDB, compiler.NewSqlServerCompiler())
_, err := db.ExecuteAffected(ctx, q)
```

See [execution.md](./execution.md).
