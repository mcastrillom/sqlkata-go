# Execution / QueryFactory

The `execution` module mirrors [SqlKata.Execution](https://sqlkata.com/docs/execution/setup) + goqu’s “Database” idea: one factory holds connection + compiler and runs queries.

It is a **separate Go module** from the builder/compilers (`github.com/mcastrillom/sqlkata-go`). Install it only when you need to run queries:

```sh
go get github.com/mcastrillom/sqlkata-go/execution
```

Built on [sqlx](https://github.com/jmoiron/sqlx) (pulled in by this module only; the root module stays dependency-light).

## Setup

```go
import (
	"github.com/jmoiron/sqlx"
	"github.com/mcastrillom/sqlkata-go/compiler"
	"github.com/mcastrillom/sqlkata-go/execution"
)

sqlxDB, err := sqlx.Open("godror", dsn) // or sqlserver / pgx …
db := execution.New(sqlxDB, compiler.NewOracleCompiler())
db.QueryTimeout = 30 * time.Second
db.Logger = func(r *compiler.SqlResult) { /* optional */ }
```

Always match compiler ↔ database dialect.

## Building queries

```go
q := db.Query("users").Select("id", "name").WhereEq("active", true)
```

`Query` / `FromQuery` return `*sqlkata.Query` (builder). Execution methods take that query explicitly (idiomatic Go; avoids losing wrapper type after fluent calls).

## Get / First

```go
type User struct {
	ID   int    `db:"id"`
	Name string `db:"name"`
}

var users []User
err = db.Get(ctx, q, &users)

users, err = execution.GetSlice[User](ctx, db, q)

var one User
err = db.First(ctx, q.Clone().WhereEq("id", 1), &one) // ErrNoRows if empty
ok, err := db.FirstOrDefault(ctx, q, &one)

one, err = execution.FirstValue[User](ctx, db, q)
```

### Maps

```go
rows, err := db.GetMaps(ctx, q) // []map[string]any
```

## Execute (DML)

```go
res, err := db.Execute(ctx, q.AsDelete())
n, err := db.ExecuteAffected(ctx, q.AsUpdate(map[string]any{"name": "Ada"}))
```

## Count / Exists / Aggregate

```go
n, err := db.Count(ctx, db.Query("users").WhereEq("active", true))
ok, err := db.Exists(ctx, db.Query("users").WhereEq("id", 1))
avg, err := db.Aggregate(ctx, db.Query("orders"), "avg", []string{"total"})
```

## Paginate / Chunk

```go
page, err := execution.Paginate[User](ctx, db, q.OrderBy("id"), 1, 25)
fmt.Println(page.Count, page.TotalPages(), page.HasNext())

err = execution.Chunk[User](ctx, db, q, 100, func(list []User, page int) bool {
	// process list; return false to stop
	return true
})
```

## Transactions

```go
err = db.WithTx(ctx, func(tx *execution.QueryFactory) error {
	_, err := tx.ExecuteAffected(ctx, tx.Query("users").WhereEq("id", 1).AsDelete())
	return err
})
```

Or pass a transaction into a single call:

```go
err = db.Get(ctx, q, &users, execution.WithTx(sqlxTx))
```

## XQuery terminals

```go
xq := db.XQuery("users")
xq.Select("id", "name")
err = xq.Get(ctx, &users)
err = xq.First(ctx, &one)
n, err = xq.Count(ctx)
```

**Note:** further fluent methods on the embedded `*Query` return `*sqlkata.Query`. Prefer mutating `xq.Query` then calling terminals, or use `db.Get(ctx, q, …)`.

## Include hydration

`Include` / `IncludeMany` metadata is stored on the query. Automatic hydration (SqlKata `handleIncludes`) is **not** fully wired yet — see `execution/include.go`.
