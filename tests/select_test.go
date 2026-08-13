package tests

import (
	"testing"

	"github.com/mcastrillom/sqlkata-go/sqlkata"
)

func TestBasicSelect(t *testing.T) {
	s := newTestSupport()
	q := sqlkata.NewQuery().From("users").Select("id", "name")
	c := s.Compile(q)

	assertEqual(t, "SELECT [id], [name] FROM [users]", c[sqlkata.EngineSqlServer])
	assertEqual(t, `SELECT "id", "name" FROM "users"`, c[sqlkata.EnginePostgres])
	assertEqual(t, `SELECT "id", "name" FROM "users"`, c[sqlkata.EngineOracle])
}

func TestBasicSelectWhereBindingIsEmptyOrNull(t *testing.T) {
	s := newTestSupport()
	q := sqlkata.NewQuery().
		From("users").
		Select("id", "name").
		WhereEq("author", "").
		OrWhereNull("author")

	c := s.Compile(q)

	assertEqual(t, "SELECT [id], [name] FROM [users] WHERE [author] = '' OR [author] IS NULL", c[sqlkata.EngineSqlServer])
	assertEqual(t, `SELECT "id", "name" FROM "users" WHERE "author" = '' OR "author" IS NULL`, c[sqlkata.EnginePostgres])
}

func TestBasicSelectWithAlias(t *testing.T) {
	s := newTestSupport()
	q := sqlkata.NewQuery().From("users as u").Select("id", "name")
	c := s.Compile(q)

	assertEqual(t, "SELECT [id], [name] FROM [users] AS [u]", c[sqlkata.EngineSqlServer])
	assertEqual(t, `SELECT "id", "name" FROM "users" AS "u"`, c[sqlkata.EnginePostgres])
	assertEqual(t, `SELECT "id", "name" FROM "users" "u"`, c[sqlkata.EngineOracle])
}

func TestExpandedSelect(t *testing.T) {
	s := newTestSupport()
	q := sqlkata.NewQuery().From("users").Select("users.{id,name, age}")
	c := s.Compile(q)

	assertEqual(t, "SELECT [users].[id], [users].[name], [users].[age] FROM [users]", c[sqlkata.EngineSqlServer])
	assertEqual(t, `SELECT "users"."id", "users"."name", "users"."age" FROM "users"`, c[sqlkata.EnginePostgres])
}

func TestExpandedSelectMultiline(t *testing.T) {
	s := newTestSupport()
	q := sqlkata.NewQuery().From("users").Select(`users.{
                                                                id,
                                                                name as Name,
                                                                age
                                                              }`)
	c := s.Compile(q)

	assertEqual(t, "SELECT [users].[id], [users].[name] AS [Name], [users].[age] FROM [users]", c[sqlkata.EngineSqlServer])
	assertEqual(t, `SELECT "users"."id", "users"."name" AS "Name", "users"."age" FROM "users"`, c[sqlkata.EnginePostgres])
}

func TestExpandedSelectWithSchema(t *testing.T) {
	s := newTestSupport()
	q := sqlkata.NewQuery().From("users").Select("dbo.users.{id,name, age}")
	c := s.Compile(q)

	assertEqual(t, "SELECT [dbo].[users].[id], [dbo].[users].[name], [dbo].[users].[age] FROM [users]", c[sqlkata.EngineSqlServer])
}

func TestSelectRaw(t *testing.T) {
	s := newTestSupport()
	q := sqlkata.NewQuery().From("Users").SelectRaw("[Id], [Name], {Age}")
	c := s.Compile(q)

	assertEqual(t, "SELECT [Id], [Name], [Age] FROM [Users]", c[sqlkata.EngineSqlServer])
	assertEqual(t, `SELECT "Id", "Name", "Age" FROM "Users"`, c[sqlkata.EnginePostgres])
	assertEqual(t, `SELECT "Id", "Name", "Age" FROM "Users"`, c[sqlkata.EngineOracle])
}

func TestNestedSelect(t *testing.T) {
	s := newTestSupport()
	q := sqlkata.NewQuery().
		From("Users").
		Select("Id").
		SelectQuery(sqlkata.NewQuery().From("Orders").AsCount(), "OrderCount")
	c := s.Compile(q)

	assertEqual(t, "SELECT [Id], (SELECT COUNT(*) AS [count] FROM [Orders]) AS [OrderCount] FROM [Users]", c[sqlkata.EngineSqlServer])
	assertEqual(t, `SELECT "Id", (SELECT COUNT(*) AS "count" FROM "Orders") AS "OrderCount" FROM "Users"`, c[sqlkata.EnginePostgres])
}
