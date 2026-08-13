package tests

import (
	"strings"
	"testing"

	"github.com/mcastrillom/sqlkata-go/sqlkata"
)

func TestBasicDelete(t *testing.T) {
	s := newTestSupport()
	q := sqlkata.NewQuery().From("Posts").AsDelete()
	c := s.Compile(q)

	assertEqual(t, "DELETE FROM [Posts]", c[sqlkata.EngineSqlServer])
}

func TestDeleteWithJoin(t *testing.T) {
	s := newTestSupport()
	q := sqlkata.NewQuery().From("Posts").
		Join("Authors", "Authors.Id", "Posts.AuthorId").
		WhereEq("Authors.Id", 5).
		AsDelete()
	c := s.Compile(q)

	got := normalizeSpace(c[sqlkata.EngineSqlServer])
	want := normalizeSpace("DELETE [Posts] FROM [Posts] INNER JOIN [Authors] ON [Authors].[Id] = [Posts].[AuthorId] WHERE [Authors].[Id] = 5")
	assertEqual(t, want, got)
}

func TestDeleteWithJoinAndAlias(t *testing.T) {
	s := newTestSupport()
	q := sqlkata.NewQuery().From("Posts as P").
		Join("Authors", "Authors.Id", "P.AuthorId").
		WhereEq("Authors.Id", 5).
		AsDelete()
	c := s.Compile(q)

	got := normalizeSpace(c[sqlkata.EngineSqlServer])
	want := normalizeSpace("DELETE [P] FROM [Posts] AS [P] INNER JOIN [Authors] ON [Authors].[Id] = [P].[AuthorId] WHERE [Authors].[Id] = 5")
	assertEqual(t, want, got)
}

func TestUpdateObject(t *testing.T) {
	s := newTestSupport()
	q := sqlkata.NewQuery().From("Table").AsUpdate(map[string]any{
		"Name": "The User",
		"Age":  "2018-01-01",
	})
	c := s.Compile(q)

	// Map iteration order is nondeterministic; accept either column order.
	got := c[sqlkata.EngineSqlServer]
	if !strings.Contains(got, "UPDATE [Table] SET ") ||
		!strings.Contains(got, "[Name] = 'The User'") ||
		!strings.Contains(got, "[Age] = '2018-01-01'") {
		t.Fatalf("unexpected UPDATE SQL:\n%s", got)
	}
}

func TestUpdateWithNullValues(t *testing.T) {
	s := newTestSupport()
	q := sqlkata.NewQuery().From("Books").WhereEq("Id", 1).AsUpdateColumns(
		[]string{"Author", "Date", "Version"},
		[]any{"Author 1", nil, nil},
	)
	c := s.Compile(q)

	assertEqual(t,
		"UPDATE [Books] SET [Author] = 'Author 1', [Date] = NULL, [Version] = NULL WHERE [Id] = 1",
		c[sqlkata.EngineSqlServer])
}

func TestUpdateWithEmptyString(t *testing.T) {
	s := newTestSupport()
	q := sqlkata.NewQuery().From("Books").WhereEq("Id", 1).AsUpdateColumns(
		[]string{"Author", "Description"},
		[]any{"Author 1", ""},
	)
	c := s.Compile(q)

	assertEqual(t,
		"UPDATE [Books] SET [Author] = 'Author 1', [Description] = '' WHERE [Id] = 1",
		c[sqlkata.EngineSqlServer])
}

func TestInsertMultiRecords(t *testing.T) {
	s := newTestSupport()
	q := sqlkata.NewQuery().From("expensive_cars").AsInsertRows(
		[]string{"name", "brand", "year"},
		[][]any{
			{"Chiron", "Bugatti", nil},
			{"Huayra", "Pagani", 2012},
			{"Reventon roadster", "Lamborghini", 2009},
		},
	)
	c := s.Compile(q)

	assertEqual(t,
		"INSERT INTO [expensive_cars] ([name], [brand], [year]) VALUES ('Chiron', 'Bugatti', NULL), ('Huayra', 'Pagani', 2012), ('Reventon roadster', 'Lamborghini', 2009)",
		c[sqlkata.EngineSqlServer])
}

func TestInsertWithNullValues(t *testing.T) {
	s := newTestSupport()
	q := sqlkata.NewQuery().From("Books").AsInsertColumns(
		[]string{"Id", "Author", "ISBN", "Date"},
		[]any{1, "Author 1", "123456", nil},
	)
	c := s.Compile(q)

	assertEqual(t,
		"INSERT INTO [Books] ([Id], [Author], [ISBN], [Date]) VALUES (1, 'Author 1', '123456', NULL)",
		c[sqlkata.EngineSqlServer])
}

func TestInsertWithEmptyString(t *testing.T) {
	s := newTestSupport()
	q := sqlkata.NewQuery().From("Books").AsInsertColumns(
		[]string{"Id", "Author", "ISBN", "Description"},
		[]any{1, "Author 1", "123456", ""},
	)
	c := s.Compile(q)

	assertEqual(t,
		"INSERT INTO [Books] ([Id], [Author], [ISBN], [Description]) VALUES (1, 'Author 1', '123456', '')",
		c[sqlkata.EngineSqlServer])
}

func TestInsertFromSubQuery(t *testing.T) {
	s := newTestSupport()
	q := sqlkata.NewQuery().From("expensive_cars").
		WithAlias("old_cars", sqlkata.NewQuery().From("all_cars").Where("year", "<", 2000)).
		AsInsertQuery(
			[]string{"name", "model", "year"},
			sqlkata.NewQuery().From("old_cars").Where("price", ">", 100).ForPage(2, 10),
		)
	c := s.Compile(q)

	assertEqual(t,
		"WITH [old_cars] AS (SELECT * FROM [all_cars] WHERE [year] < 2000)\nINSERT INTO [expensive_cars] ([name], [model], [year]) SELECT * FROM [old_cars] WHERE [price] > 100 ORDER BY (SELECT 0) OFFSET 10 ROWS FETCH NEXT 10 ROWS ONLY",
		c[sqlkata.EngineSqlServer])
	assertEqual(t,
		`WITH "old_cars" AS (SELECT * FROM "all_cars" WHERE "year" < 2000)
INSERT INTO "expensive_cars" ("name", "model", "year") SELECT * FROM "old_cars" WHERE "price" > 100 LIMIT 10 OFFSET 10`,
		c[sqlkata.EnginePostgres])
}
