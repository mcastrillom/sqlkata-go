# Prepared statements / Interpolation

Compiling a query produces a `*compiler.SqlResult`:

| Member | Meaning |
|---|---|
| `RawSQL` | SQL with `?` placeholders (builder-neutral) |
| `Bindings` | Ordered argument values |
| `SQL()` | SQL with dialect named placeholders (`@p0`, `:p0`, …) |
| `NamedBindings()` | Map of named args |
| `String()` | Interpolated preview (debugging — not for execution) |

`SQL()` and `NamedBindings()` are computed on first call and cached, so compiling stays cheap when you only execute `RawSQL` + `Bindings`. A `SqlResult` is not safe for concurrent use.

```go
c := compiler.NewSqlServerCompiler()
r, err := c.Compile(sqlkata.NewQuery().From("users").WhereEq("id", 1))
if err != nil {
	panic(err) // also covers q.Err() from builder validation
}

fmt.Println(r.RawSQL)
// SELECT * FROM [users] WHERE [id] = ?

fmt.Println(r.SQL())
// SELECT * FROM [users] WHERE [id] = @p0

fmt.Println(r.Bindings)
// [1]

fmt.Println(r.String())
// SELECT * FROM [users] WHERE [id] = 1
```

## Execution binding

`execution.QueryFactory` always starts from `RawSQL` (`?`) and rewrites placeholders from the **compiler engine**:

| Engine | Sent to driver |
|---|---|
| Oracle | `:1`, `:2`, … |
| PostgreSQL | `$1`, `$2`, … |
| SQL Server | driver default (`?` or `@pN`) |
| MySQL / MariaDB | `?` |
| SQLite | `?` |

This avoids Oracle **ORA-00911** when sqlx would otherwise leave `?`.

```go
db := execution.New(sqlxDB, compiler.NewOracleCompiler())
// BindStyle defaults to BindAuto

// Force override if needed:
db.BindStyle = execution.BindOraclePositional
```

## Logging

```go
db.Logger = func(r *compiler.SqlResult) {
	log.Printf("%s | %v", r.SQL(), r.Bindings)
}
```

Called on every `Compile` inside the factory (Get / Execute / …).

## Security notes

* Prefer parameterized `Where` / bindings over string concat
* Use `UnsafeLiteral` only for trusted fragments
* Operator whitelist (SqlKata-style) is a recommended hardening step before untrusted operators
