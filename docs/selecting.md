# Selecting

## Basic select

```go
q := sqlkata.NewQuery().From("users").Select("id", "name")
r := compiler.NewSqlServerCompiler().Compile(q)
fmt.Println(r.String())
// SELECT [id], [name] FROM [users]
```

### Table alias

```go
q := sqlkata.NewQuery().From("users as u").Select("u.id", "u.name")
// FROM [users] AS [u]
```

### Distinct

```go
q.Distinct().Select("country")
```

### SelectRaw / expand

```go
q.SelectRaw("[id], [name], {age}")
q.Select("orders.{total, qty}")
```

### Subquery column

```go
q := sqlkata.NewQuery().From("users").Select("id").
	SelectQuery(sqlkata.NewQuery().From("orders").AsCount(), "order_count")
```

## Aggregates

```go
sqlkata.NewQuery().From("orders").AsCount()
// SELECT COUNT(*) AS [count] FROM [orders]

sqlkata.NewQuery().From("orders").AsSum("total")
sqlkata.NewQuery().From("orders").AsAvg("total")
sqlkata.NewQuery().From("orders").AsMin("total")
sqlkata.NewQuery().From("orders").AsMax("total")
```

### Select aggregates (+ filter)

```go
q := sqlkata.NewQuery().From("orders").
	Select("user_id").
	SelectSum("total", func(q *sqlkata.Query) *sqlkata.Query {
		return q.WhereEq("paid", true)
	}).
	GroupBy("user_id")
```

* SQL Server / Oracle: `SUM(CASE WHEN … THEN total END)`
* PostgreSQL (`SupportsFilter`): `SUM(total) FILTER (WHERE …)`

## From subquery

```go
inner := sqlkata.NewQuery().From("users").Select("id").WhereEq("active", true)
q := sqlkata.NewQuery().FromQuery(inner, "u").Select("u.id")
```

## CTE (`WITH`)

```go
q := sqlkata.NewQuery().
	WithAlias("active_users",
		sqlkata.NewQuery().From("users").Select("id", "name").WhereEq("active", true),
	).
	From("active_users").
	Select("id", "name")
```

Also: `WithRaw`, `WithValues` (ad-hoc / VALUES CTE).

## Union / combine

```go
q := sqlkata.NewQuery().From("a").Select("id").
	UnionAll(sqlkata.NewQuery().From("b").Select("id")).
	Union(sqlkata.NewQuery().From("c").Select("id"))
```

Also: `Except`, `Intersect` (+ raw combine helpers).

## Pagination

```go
q.OrderBy("id").Limit(10).Offset(20)
q.ForPage(2, 25)
```

See [dialect.md](./dialect.md) for `OFFSET/FETCH` vs `LIMIT/OFFSET`.

## Compiling only

```go
c := compiler.NewPostgresCompiler()
r, err := c.Compile(q)
if err != nil {
	panic(err)
}
fmt.Println(r.RawSQL)   // ? placeholders
fmt.Println(r.SQL())    // named placeholders (lazy)
fmt.Println(r.Bindings) // ordered args
fmt.Println(r.String()) // interpolated preview
```

## Executing a select

```go
db := execution.New(sqlxDB, compiler.NewSqlServerCompiler())
q := db.Query("users").Select("id", "name").WhereEq("active", true)

var rows []User
err := db.Get(ctx, q, &rows)

// or
rows, err := execution.GetSlice[User](ctx, db, q)
user, err := execution.FirstValue[User](ctx, db, q.Clone().WhereEq("id", 1))
```

See [execution.md](./execution.md).
