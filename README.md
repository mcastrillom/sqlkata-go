# sqlkata-go

```
           _ _         _
 ___  __ _| | | ____ _| |_ __ _
/ __|/ _` | | |/ / _` | __/ _` |
\__ \ (_| | |   < (_| | || (_| |
|___/\__, |_|_|\_\__,_|\__\__,_|
        |_|           -go
```

`sqlkata-go` is an **unofficial** Go port of [SqlKata](https://sqlkata.com/) (C#), created by [Ahmad Moussawi](https://github.com/ahmad-moussawi) — an expressive SQL query builder with dialect compilers and an optional execution layer (similar in spirit to [goqu](https://doug-martin.github.io/goqu/)).

**Not affiliated with or supported by the official SqlKata team.** Maintained solely by [mcastrillom](https://github.com/mcastrillom).

It focuses on a fluent API close to SqlKata/.NET: native identifier quoting, `WhereIn` / `WhereNull` / joins / CTEs, and `QueryFactory` execution via [sqlx](https://github.com/jmoiron/sqlx).

## Installation

Two Go modules (same layout as SqlKata / SqlKata.Execution):

```sh
# Builder + compilers (no sqlx)
go get github.com/mcastrillom/sqlkata-go@v0.1.1

# Optional QueryFactory (pulls sqlx)
go get github.com/mcastrillom/sqlkata-go/execution@v0.1.1
```

```go
import (
	"github.com/mcastrillom/sqlkata-go/sqlkata"
	"github.com/mcastrillom/sqlkata-go/compiler"
	"github.com/mcastrillom/sqlkata-go/execution" // optional
)
```

## Features

* Fluent query builder (`Select`, `Where*`, `Join`, `CTE`, `Union`, …)
* Dialect compilers: **SQL Server**, **PostgreSQL**, **Oracle**, **MySQL/MariaDB**, **SQLite**
* Automatic identifier wrapping (`&#91;col&#93;`, `"col"`, `` `col` ``) — no manual `Flavor().Quote()`
* Parameterized SQL + interpolated preview (`SqlResult.String()`)
* Engine-scoped clauses: `ForSqlServer` / `ForPostgreSql` / `ForOracle` / `ForMySql` / `ForSqlite`
* Separate `execution` module: `QueryFactory` (`Get`, `First`, `Execute`, `Paginate`, `WithTx`, …)
* Oracle-safe bind rewriting (`?` → `:1`, `:2`, …) in execution

This library is a **query builder + optional executor**, not a full ORM. For associations / hooks, pair it with your own repositories (or hydrate `Include` metadata in a later layer).

## Why?

Thin SQL string builders often force you to quote every identifier and stitch `Where` fragments by hand. SqlKata’s model keeps a single fluent chain and lets the **compiler** own dialect differences. `sqlkata-go` brings that model to Go, especially for teams migrating from SqlKata/.NET or from wrappers around huandu/go-sqlbuilder.

## Docs

* [Dialects / Compilers](./docs/dialect.md) — SQL Server, PostgreSQL, Oracle, MySQL, SQLite
* [Expressions](./docs/expressions.md) — Where, Having, raw, variables, expand columns
* [Selecting](./docs/selecting.md) — SELECT, aggregates, pagination, CTE, UNION
* [Joining](./docs/joining.md) — Join, Left/Right/Cross, DeepJoin, Include metadata
* [Inserting](./docs/inserting.md) — INSERT map / rows / SELECT
* [Updating](./docs/updating.md) — UPDATE, increment / decrement
* [Deleting](./docs/deleting.md) — DELETE (+ join)
* [Prepared / Interpolation](./docs/interpolation.md) — `SqlResult`, bindings, placeholders
* [Execution / QueryFactory](./docs/execution.md) — run queries with sqlx
* [Extending](./docs/extending.md) — `AddComponent`, `For*`, custom helpers

## Quick examples

### Select

```go
c := compiler.NewSqlServerCompiler()
q := sqlkata.NewQuery().
	From("users").
	Select("id", "name").
	WhereEq("status", "active")

r, err := c.Compile(q)
if err != nil {
	panic(err)
}
fmt.Println(r.String())
// SELECT [id], [name] FROM [users] WHERE [status] = 'active'

fmt.Println(r.SQL(), r.Bindings)
// SELECT [id], [name] FROM [users] WHERE [status] = @p0 [active]
```

Builder validation errors (empty insert map, missing CTE alias, …) are stored on the query (`q.Err()`) and returned by `Compile`.

### Insert

```go
q := sqlkata.NewQuery().From("users").AsInsert(map[string]any{
	"name":  "Ada",
	"email": "ada@example.com",
}, false)
r, _ := compiler.NewSqlServerCompiler().Compile(q)
fmt.Println(r.String())
// INSERT INTO [users] ([email], [name]) VALUES ('ada@example.com', 'Ada')
```

### Update

```go
q := sqlkata.NewQuery().From("users").
	WhereEq("id", 1).
	AsUpdate(map[string]any{"name": "Ada"})
r, _ := compiler.NewSqlServerCompiler().Compile(q)
fmt.Println(r.String())
// UPDATE [users] SET [name] = 'Ada' WHERE [id] = 1
```

### Delete

```go
q := sqlkata.NewQuery().From("users").WhereEq("id", 1).AsDelete()
r, _ := compiler.NewSqlServerCompiler().Compile(q)
fmt.Println(r.String())
// DELETE FROM [users] WHERE [id] = 1
```

### Execute (QueryFactory)

```go
db := execution.New(sqlxDB, compiler.NewOracleCompiler())
q := db.Query("users").Select("id", "name").WhereEq("status", "active")

users, err := execution.GetSlice[User](ctx, db, q)
one, err := execution.FirstValue[User](ctx, db, q.Clone().WhereEq("id", 1))
n, err := db.ExecuteAffected(ctx, db.Query("users").WhereEq("id", 1).AsDelete())
```

## Running tests

```sh
cd sqlkata-go
go test ./...
```

## License

[MIT](./LICENSE).

This repository is an unofficial Go port of [sqlkata/querybuilder](https://github.com/sqlkata/querybuilder) by **Ahmad Moussawi** (SqlKata). It is **not** endorsed, maintained, or supported by the official SqlKata team — only by **mcastrillom**.

Copyright (c) 2026 mcastrillom  
Copyright (c) 2017 SqlKata (original C# Query Builder by Ahmad Moussawi)
