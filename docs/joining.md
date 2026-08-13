# Joining

## Basic joins

```go
q := sqlkata.NewQuery().
	From("posts").
	Select("posts.title", "authors.name").
	Join("authors", "authors.id", "posts.author_id")

// LEFT / RIGHT / CROSS
q.LeftJoin("authors", "authors.id", "posts.author_id")
q.RightJoin("authors", "authors.id", "posts.author_id")
q.CrossJoin("tags")
```

Optional operator / type overrides on `Join(table, first, second, op?, type?)`.

## Callback join (multiple ON conditions)

```go
q.JoinWith("profiles", func(j *sqlkata.Join) *sqlkata.Join {
	return j.On("profiles.user_id", "users.id").
		OrOn("profiles.alt_id", "users.id")
}, "inner join")
```

`On` / `OrOn` live on `*sqlkata.Join`. You can also add where-like constraints on the join builder.

## Subquery join

```go
top := sqlkata.NewQuery().From("comments").OrderByDesc("likes").Limit(10).As("top_comments")
q.LeftJoinQuery(top, func(j *sqlkata.Join) *sqlkata.Join {
	return j.On("top_comments.post_id", "posts.id")
})
```

## DeepJoin

Path-style joins (expanded by the compiler). Default keys: FK `{Table}Id`, PK `Id`.

```go
q := sqlkata.NewQuery().
	From("Books").
	Select("Books.Title", "Author.Name", "Country.Name").
	LeftDeepJoin("Author.Country")
```

Compiles roughly to:

```sql
SELECT … FROM [Books]
LEFT JOIN [Author] ON [Author].[Id] = [Books].[AuthorId]
LEFT JOIN [Country] ON [Country].[Id] = [Author].[CountryId]
```

Customize:

```go
q.DeepJoin("Comment as c",
	sqlkata.DeepJoinGenerators(
		func(string) string { return "Id" },     // Posts.Id
		func(string) string { return "PostId" }, // Comment.PostId
	),
)
```

Also: `DeepJoinType`, `DeepJoinKeys`, `LeftDeepJoin`.

## Include / IncludeMany (metadata)

Registered on the query for execution layers (not emitted as SQL by the compiler):

```go
q.Include("Profile",
	sqlkata.NewQuery().From("profiles").Select("user_id", "bio"),
	"user_id", "id",
).IncludeMany("Orders",
	sqlkata.NewQuery().From("orders").Select("user_id", "total"),
	"user_id", "id",
)
```

Hydration in `execution` is planned (SqlKata.Execution `handleIncludes`); metadata + `Clone` independence already work.
