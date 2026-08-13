package tests

import (
	"testing"

	"github.com/mcastrillom/sqlkata-go/sqlkata"
)

func TestGroupedWhereFilters(t *testing.T) {
	s := newTestSupport()
	q := sqlkata.NewQuery().From("Table1").
		WhereNested(func(q *sqlkata.Query) *sqlkata.Query {
			return q.WhereEq("Column1", 10).OrWhereEq("Column2", 20)
		}).
		WhereEq("Column3", 30)

	c := s.Compile(q)

	assertEqual(t,
		`SELECT * FROM "Table1" WHERE ("Column1" = 10 OR "Column2" = 20) AND "Column3" = 30`,
		c[sqlkata.EnginePostgres])
}

func TestGroupedHavingFilters(t *testing.T) {
	s := newTestSupport()
	q := sqlkata.NewQuery().From("Table1").
		HavingNested(func(q *sqlkata.Query) *sqlkata.Query {
			return q.HavingRaw("SUM([Column1]) = ?", 10).OrHavingRaw("SUM([Column2]) = ?", 20)
		}).
		HavingRaw("SUM([Column3]) = ?", 30)

	c := s.Compile(q)

	assertEqual(t,
		`SELECT * FROM "Table1" HAVING (SUM("Column1") = 10 OR SUM("Column2") = 20) AND SUM("Column3") = 30`,
		c[sqlkata.EnginePostgres])
}

func TestBasicWhere(t *testing.T) {
	s := newTestSupport()
	q := sqlkata.NewQuery().From("users").Select("id").Where("id", "=", 1)
	c := s.Compile(q)

	assertEqual(t, "SELECT [id] FROM [users] WHERE [id] = 1", c[sqlkata.EngineSqlServer])
	assertEqual(t, `SELECT "id" FROM "users" WHERE "id" = 1`, c[sqlkata.EnginePostgres])
}

func TestWhereIn(t *testing.T) {
	s := newTestSupport()
	q := sqlkata.NewQuery().From("users").Select("id").WhereIn("id", 1, 2, 3)
	c := s.Compile(q)

	assertEqual(t, "SELECT [id] FROM [users] WHERE [id] IN (1, 2, 3)", c[sqlkata.EngineSqlServer])
}

func TestWhereBetween(t *testing.T) {
	s := newTestSupport()
	q := sqlkata.NewQuery().From("users").Select("id").WhereBetween("age", 18, 65)
	c := s.Compile(q)

	assertEqual(t, "SELECT [id] FROM [users] WHERE [age] BETWEEN 18 AND 65", c[sqlkata.EngineSqlServer])
}

func TestWhereNull(t *testing.T) {
	s := newTestSupport()
	q := sqlkata.NewQuery().From("users").Select("id").WhereNull("email").OrWhereNotNull("phone")
	c := s.Compile(q)

	assertEqual(t, "SELECT [id] FROM [users] WHERE [email] IS NULL OR [phone] IS NOT NULL", c[sqlkata.EngineSqlServer])
}

func TestWhereExists(t *testing.T) {
	s := newTestSupport()
	sub := sqlkata.NewQuery().From("Orders").WhereColumns("Orders.UserId", "=", "Users.Id")
	q := sqlkata.NewQuery().From("Users").Select("Id").WhereExists(sub)
	c := s.Compile(q)

	assertEqual(t,
		"SELECT [Id] FROM [Users] WHERE EXISTS (SELECT 1 FROM [Orders] WHERE [Orders].[UserId] = [Users].[Id])",
		c[sqlkata.EngineSqlServer])
}

func TestWhereStartsContains(t *testing.T) {
	s := newTestSupport()
	q := sqlkata.NewQuery().From("Users").Select("Id").
		WhereStarts("Name", "A", false, "").
		OrWhereContains("Email", "@example.com", false, "")
	c := s.Compile(q)

	assertEqual(t,
		"SELECT [Id] FROM [Users] WHERE LOWER([Name]) like 'a%' OR LOWER([Email]) like '%@example.com%'",
		c[sqlkata.EngineSqlServer])
}
