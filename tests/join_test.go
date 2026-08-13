package tests

import (
	"testing"

	"github.com/mcastrillom/sqlkata-go/sqlkata"
)

func TestLeftJoin(t *testing.T) {
	s := newTestSupport()
	q := sqlkata.NewQuery().
		From("Users as u").
		Select("u.Id", "o.Total").
		LeftJoin("Orders as o", "o.UserId", "u.Id")
	c := s.Compile(q)

	assertEqual(t,
		"SELECT [u].[Id], [o].[Total] FROM [Users] AS [u] LEFT JOIN [Orders] AS [o] ON [o].[UserId] = [u].[Id]",
		c[sqlkata.EngineSqlServer])
}

func TestJoinWithCallback(t *testing.T) {
	s := newTestSupport()
	q := sqlkata.NewQuery().
		From("Users").
		Select("Users.Id", "Profiles.Bio").
		JoinWith("Profiles", func(j *sqlkata.Join) *sqlkata.Join {
			return j.On("Profiles.UserId", "Users.Id").OrOn("Profiles.AltId", "Users.Id")
		}, "inner join")
	c := s.Compile(q)

	assertEqual(t,
		"SELECT [Users].[Id], [Profiles].[Bio] FROM [Users] INNER JOIN [Profiles] ON [Profiles].[UserId] = [Users].[Id] OR [Profiles].[AltId] = [Users].[Id]",
		c[sqlkata.EngineSqlServer])
}

func TestLeftDeepJoin(t *testing.T) {
	s := newTestSupport()
	q := sqlkata.NewQuery().
		From("Books").
		Select("Books.Title", "Author.Name", "Country.Name").
		LeftDeepJoin("Author.Country")
	c := s.Compile(q)

	assertEqual(t,
		"SELECT [Books].[Title], [Author].[Name], [Country].[Name] FROM [Books] LEFT JOIN [Author] ON [Author].[Id] = [Books].[AuthorId] LEFT JOIN [Country] ON [Country].[Id] = [Author].[CountryId]",
		c[sqlkata.EngineSqlServer])
}

func TestDeepJoinCustomGenerators(t *testing.T) {
	s := newTestSupport()
	q := sqlkata.NewQuery().
		From("Posts as p").
		Select("p.Id", "c.Body").
		DeepJoin("Comment as c", sqlkata.DeepJoinGenerators(
			func(string) string { return "Id" },
			func(string) string { return "PostId" },
		))
	c := s.Compile(q)

	assertEqual(t,
		"SELECT [p].[Id], [c].[Body] FROM [Posts] AS [p] INNER JOIN [Comment] AS [c] ON [c].[PostId] = [p].[Id]",
		c[sqlkata.EngineSqlServer])
}

func TestIncludeDoesNotAffectSQL(t *testing.T) {
	s := newTestSupport()
	q := sqlkata.NewQuery().
		From("Users").
		Select("Id", "Name").
		Include("Profile", sqlkata.NewQuery().From("Profiles").Select("UserId", "Bio"), "UserId", "Id").
		IncludeMany("Orders", sqlkata.NewQuery().From("Orders").Select("UserId", "Total"), "UserId", "Id")
	c := s.Compile(q)

	assertEqual(t, "SELECT [Id], [Name] FROM [Users]", c[sqlkata.EngineSqlServer])
	if len(q.Includes) != 2 {
		t.Fatalf("Includes = %d, want 2", len(q.Includes))
	}
	if q.Includes[0].IsMany {
		t.Fatal("Profile should not be many")
	}
	if !q.Includes[1].IsMany {
		t.Fatal("Orders should be many")
	}
}
