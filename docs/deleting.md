# Deleting

## Basic delete

```go
q := sqlkata.NewQuery().From("posts").AsDelete()
fmt.Println(compiler.NewSqlServerCompiler().Compile(q).String())
// DELETE FROM [posts]
```

## Delete with where

```go
q := sqlkata.NewQuery().From("posts").WhereEq("id", 5).AsDelete()
// DELETE FROM [posts] WHERE [id] = 5
```

## Delete with join

```go
q := sqlkata.NewQuery().From("posts").
	Join("authors", "authors.id", "posts.author_id").
	WhereEq("authors.id", 5).
	AsDelete()
// DELETE [posts] FROM [posts] INNER JOIN [authors] ON … WHERE …
```

### Alias form

```go
q := sqlkata.NewQuery().From("posts as p").
	Join("authors", "authors.id", "p.author_id").
	WhereEq("authors.id", 5).
	AsDelete()
// DELETE [p] FROM [posts] AS [p] INNER JOIN … WHERE …
```

## Executing deletes

```go
n, err := db.ExecuteAffected(ctx, q)
```
