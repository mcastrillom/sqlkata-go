# Expressions

Building blocks for filters and projections. Unlike goqu’s `Ex` / `Op` maps, sqlkata-go follows SqlKata’s **fluent methods** on `*Query`.

## Where basics

```go
q := sqlkata.NewQuery().From("users").
	Where("status", "=", "active").
	OrWhereEq("role", "admin").
	WhereNotNull("email").
	WhereIn("country", "MX", "US", "CA").
	WhereBetween("age", 18, 65)
```

Useful helpers (non-exhaustive):

| Method | Meaning |
|---|---|
| `Where` / `WhereEq` / `WhereNotEq` | column OP value |
| `WhereNull` / `WhereNotNull` | IS NULL / IS NOT NULL |
| `WhereIn` / `WhereNotIn` | IN (…) |
| `WhereBetween` / `WhereNotBetween` | BETWEEN |
| `WhereColumns` | column OP column |
| `WhereTrue` / `WhereFalse` | boolean helpers |
| `WhereRaw` | raw SQL fragment + bindings |
| `WhereNested` / `OrWhereNested` | `( … )` groups |
| `WhereExists` / `WhereNotExists` | EXISTS (subquery) |
| `WhereQuery` / `WhereInQuery` / `WhereSub` | subquery compares |
| `WhereTextEqual` / `WhereTextNotEqual` | long-text / CLOB-safe compare |
| `OrWhereTextEqual` / `OrWhereTextNotEqual` | OR variants |

Short aliases: `WhereEq`, `WhereGt`, `WhereGte`, `WhereLt`, `WhereLte`, …

## Long text / CLOB

Avoid Oracle `ORA-00932` when comparing CLOB columns to binds:

```go
q.WhereTextEqual("ROOT", value).
	OrWhereTextNotEqual("NOTES", other)

q.SelectAsText("ROOT", "RootText")           // Oracle: DBMS_LOB.SUBSTR(..., 4000, 1)
q.SelectAsText("ROOT", "RootText", 2000)     // custom SUBSTR length (Oracle only)
```

* Oracle WHERE → `DBMS_LOB.COMPARE(col, TO_CLOB(?)) = 0` (or `<> 0`)
* SQL Server / Postgres WHERE → normal `=` / `<>`
* Oracle SELECT → `DBMS_LOB.SUBSTR(col, maxLen, 1) alias`; others → `col AS alias`

## Strings

```go
q.WhereStarts("name", "A", false, "").
	OrWhereContains("email", "@example.com", false, "").
	WhereLike("code", "X_Y", true, `\`).
	WhereEnds("name", "son", false, "")
```

* `caseSensitive=false` → `LOWER(col) like …` (Postgres may use `ILIKE`)
* Pass `escapeCharacter` when needed

## Dates

```go
q.WhereDate("created_at", ">=", "2024-01-01").
	WhereDatePart("year", "created_at", "=", 2024)
```

Compilers emit dialect-specific expressions (`CAST` / `DATEPART` / `DATE_PART` / `EXTRACT` / `TO_CHAR`, …).

## Nested groups

```go
q.WhereNested(func(q *sqlkata.Query) *sqlkata.Query {
	return q.WhereEq("dept", "IT").OrWhereEq("dept", "HR")
})
```

## Raw / UnsafeLiteral / Variable

```go
q.WhereRaw("[score] > ?", 10)

q.Define("minAge", 21).
	Where("age", ">=", sqlkata.VariableExpr("minAge")).
	Where("status", "=", sqlkata.UnsafeLiteralExpr("'Active'", false))
```

* `Variable` resolves via `Define` / parent queries at compile time
* `UnsafeLiteral` embeds SQL as-is — **never** pass user input

## Column expand

SqlKata-style expand:

```go
q.Select("users.{id, name as FullName, age}")
// → users.id, users.name AS FullName, users.age
```

Also works with schema prefixes: `dbo.users.{id,name}`.

## Having

Mirrors many Where helpers on aggregates:

```go
q.GroupBy("user_id").
	Having("SUM(total)", ">", 100).
	OrHavingBetween("SUM(total)", 200, 1000).
	HavingNested(func(q *sqlkata.Query) *sqlkata.Query {
		return q.HavingEq("COUNT(*)", 1).OrHavingEq("COUNT(*)", 2)
	})
```

## When / WhenNot

```go
q.When(activeOnly, func(q *sqlkata.Query) *sqlkata.Query {
	return q.WhereEq("active", true)
})
```

## Order / Limit

```go
q.OrderBy("name").OrderByDesc("id").
	Limit(10).Offset(20)
// or
q.ForPage(2, 10) // page is 1-based
q.Take(10).Skip(20)
```
