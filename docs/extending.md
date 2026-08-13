# Extending

## AddComponent (SqlKata style)

Public APIs that only attach clauses stay extendable without forking the compiler:

```go
func WhereActive(q *sqlkata.Query) *sqlkata.Query {
	return q.WhereEq("active", true)
}

// Or add a method in your package:
func WhereTenant(q *sqlkata.Query, tenantID string) *sqlkata.Query {
	return q.AddComponent("where", &sqlkata.BasicCondition{
		Column:   "tenant_id",
		Operator: "=",
		Value:    tenantID,
	}, nil)
}
```

You also have `AddOrReplaceComponent`, `GetComponents`, `ClearComponent`, `Clone`.

## Dialect helpers (`For*`)

```go
q.ForOracle(func(q *sqlkata.Query) *sqlkata.Query {
	return q.WhereRaw(`/* oracle-only tweak */ 1=1`)
})
```

For portable CLOB / long-text filters prefer built-ins (`WhereTextEqual`, `SelectAsText`) instead of raw `ForOracle` + `DBMS_LOB` — the compiler already emits the Oracle-safe form.

## Raw escape hatches

* `SelectRaw` / `WhereRaw` / `HavingRaw` / `FromRaw` / `OrderByRaw` / `GroupByRaw`
* `UnsafeLiteral` for trusted SQL literals
* `Comment` for tracing

## Custom condition types

In C#, SqlKata discovers `CompileXxx` via reflection. In Go, `compileCondition` uses a **type switch**.

Options today:

1. Prefer `WhereRaw` / existing clauses for business helpers
2. Open a PR to register a condition compiler hook
3. Wrap `*compiler.Compiler` and override compile entry points for a private dialect

Query-side extension (`AddComponent`) already works; compiler plugins are the remaining gap vs C#.

## QueryFactory customization

```go
db := execution.New(sqlxDB, compiler.NewOracleCompiler())
db.BindStyle = execution.BindOraclePositional
db.QueryTimeout = time.Minute
db.Logger = myLogger
```

## Package layout

```
sqlkata-go/
  sqlkata/     # Query builder + clauses
  compiler/    # Dialect compilers + SqlResult
  execution/   # QueryFactory + sqlx
  docs/        # This documentation
  example/     # Runnable samples
  tests/       # Compiler / builder tests
```
