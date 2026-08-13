# Dialects / Compilers

Compilers turn a `*sqlkata.Query` into dialect-specific SQL (identifiers, pagination, booleans, placeholders).

Supported out of the box:

| Compiler | Package constructor | Engine code | Identifiers | Placeholders (compiled `SQL`) |
|---|---|---|---|---|
| SQL Server | `compiler.NewSqlServerCompiler()` | `sqlsrv` | `&#91;col&#93;` | `@p0`, `@p1`, … |
| PostgreSQL | `compiler.NewPostgresCompiler()` | `postgres` | `"col"` | `@p0`, … (execution rebinds to `$1`) |
| Oracle | `compiler.NewOracleCompiler()` | `oracle` | `"col"` | `:p0`, … (execution rebinds to `:1`) |
| MySQL / MariaDB | `compiler.NewMySqlCompiler()` | `mysql` | `` `col` `` | `@p0`, … (execution keeps `?`) |
| SQLite | `compiler.NewSqliteCompiler()` | `sqlite` | `"col"` | `@p0`, … (execution keeps `?`) |

```go
import (
	"fmt"
	"github.com/mcastrillom/sqlkata-go/compiler"
	"github.com/mcastrillom/sqlkata-go/sqlkata"
)

q := sqlkata.NewQuery().From("users").Select("id").WhereEq("id", 10)

fmt.Println(compiler.NewSqlServerCompiler().Compile(q).String())
// SELECT [id] FROM [users] WHERE [id] = 10

fmt.Println(compiler.NewPostgresCompiler().Compile(q).String())
// SELECT "id" FROM "users" WHERE "id" = 10

fmt.Println(compiler.NewOracleCompiler().Compile(q).String())
// SELECT "id" FROM "users" WHERE "id" = 10

fmt.Println(compiler.NewMySqlCompiler().Compile(q).String())
// SELECT `id` FROM `users` WHERE `id` = 10

fmt.Println(compiler.NewSqliteCompiler().Compile(q).String())
// SELECT "id" FROM "users" WHERE "id" = 10
```

## Engine codes

```go
sqlkata.EngineSqlServer // "sqlsrv"
sqlkata.EnginePostgres  // "postgres"
sqlkata.EngineOracle    // "oracle"
sqlkata.EngineMySql     // "mysql"
sqlkata.EngineSqlite    // "sqlite"
```

Also defined for scoped clauses without a compiler yet: `EngineFirebird`, `EngineGeneric`.

## Dialect-scoped clauses (`For*`)

Like SqlKata’s `ForSqlServer` / `ForPostgreSql` / `ForOracle`: clauses added inside the callback are tagged with that engine and only appear when compiling with the matching compiler.

```go
q := sqlkata.NewQuery().From("users").Select("id").
	ForSqlServer(func(q *sqlkata.Query) *sqlkata.Query {
		return q.WhereRaw("[LegacyFlag] = 1")
	}).
	ForOracle(func(q *sqlkata.Query) *sqlkata.Query {
		return q.WhereRaw(`"LegacyFlag" = 1`)
	}).
	ForPostgreSql(func(q *sqlkata.Query) *sqlkata.Query {
		return q.WhereEq("legacy_flag", 1)
	})

compiler.NewSqlServerCompiler().Compile(q)
// … WHERE [LegacyFlag] = 1

compiler.NewOracleCompiler().Compile(q)
// … WHERE "LegacyFlag" = 1
```

Helpers: `ForSqlServer`, `ForPostgreSql` / `ForPostgres`, `ForOracle`, `ForMySql`, `ForSqlite`, `ForFirebird`, and generic `For(engine, fn)`.

## Pagination differences

Same query builder; different SQL:

```go
q := sqlkata.NewQuery().From("users").OrderBy("id").Limit(10).Offset(20)
```

* **SQL Server / Oracle (modern):** `OFFSET … ROWS FETCH NEXT … ROWS ONLY` (injects a safe `ORDER BY` if missing on SQL Server/Oracle)
* **PostgreSQL / MySQL / SQLite:** `LIMIT … OFFSET …`

Oracle also supports legacy `ROWNUM` when `UseLegacyPagination` is enabled on the compiler.

`Offset` without `Limit` is the one case where the three `LIMIT … OFFSET …` dialects differ, because MySQL and SQLite reject a bare `OFFSET`:

| Engine | `Offset(20)` alone |
|---|---|
| PostgreSQL | `OFFSET ?` |
| MySQL | `LIMIT 18446744073709551615 OFFSET ?` |
| SQLite | `LIMIT -1 OFFSET ?` |

## Other dialect differences

| Feature | SQL Server | PostgreSQL | Oracle | MySQL | SQLite |
|---|---|---|---|---|---|
| `WhereTrue` | `cast(1 as bit)` | `true` | `true` | `true` | `1` |
| `OrderByRandom` | `NEWID()` | `RANDOM()` | `RANDOM()` | `RAND()` | `RANDOM()` |
| `WhereDatePart("year", …)` | `DATEPART(YEAR, …)` | `DATE_PART('YEAR', …)` | `EXTRACT(YEAR FROM …)` | `YEAR(…)` | `CAST(strftime('%Y', …) AS INTEGER)` |
| Insert `returnId` | `scope_identity()` | `lastval()` | — | `last_insert_id()` | `last_insert_rowid()` |

Two of these deviate from SqlKata C# on purpose, because the C# output is not valid in that engine: MySQL uses `RAND()` (C# emits `RANDOM()`), and SQLite uses `strftime` (C# emits `YEAR()`, which SQLite does not define).

## Execution + dialect

Always pair `execution.New` with the **same** dialect compiler you intend to run:

```go
db := execution.New(sqlxDB, compiler.NewOracleCompiler())
```

The execution layer rewrites `?` placeholders according to the compiler engine (Oracle → `:1`, Postgres → `$1`, MySQL/SQLite/SQL Server → `?`). See [execution.md](./execution.md) and [interpolation.md](./interpolation.md).

## Custom compilers

Start from `*compiler.Compiler` (shared SELECT/DML core) and set:

* `OpeningIdentifier` / `ClosingIdentifier`
* `ColumnAsKeyword` / `TableAsKeyword`
* `EngineCode`
* optional `SupportsFilter` (Postgres `FILTER (WHERE …)` on aggregates)

Then wrap it like `SqlServerCompiler` / `MySqlCompiler` / `SqliteCompiler`.

Placeholder style, pagination style and the random/unbounded-limit literals are package-internal fields, so a brand new dialect belongs in the `compiler` package next to `mysql.go` and `sqlite.go` — those two are the smallest examples to copy.
