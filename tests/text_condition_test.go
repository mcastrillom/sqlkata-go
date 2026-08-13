package tests

import (
	"testing"

	"github.com/mcastrillom/sqlkata-go/sqlkata"
)

func TestWhereTextEqual(t *testing.T) {
	s := newTestSupport()
	q := sqlkata.NewQuery().From("docs").Select("id").WhereTextEqual("body", "hello")
	c := s.Compile(q)

	assertEqual(t,
		"SELECT [id] FROM [docs] WHERE [body] = 'hello'",
		c[sqlkata.EngineSqlServer])
	assertEqual(t,
		`SELECT "id" FROM "docs" WHERE "body" = 'hello'`,
		c[sqlkata.EnginePostgres])
	assertEqual(t,
		`SELECT "id" FROM "docs" WHERE DBMS_LOB.COMPARE("body", TO_CLOB('hello')) = 0`,
		c[sqlkata.EngineOracle])
}

func TestWhereTextNotEqual(t *testing.T) {
	s := newTestSupport()
	q := sqlkata.NewQuery().From("docs").Select("id").WhereTextNotEqual("body", "x")
	c := s.Compile(q)

	assertEqual(t,
		"SELECT [id] FROM [docs] WHERE [body] <> 'x'",
		c[sqlkata.EngineSqlServer])
	assertEqual(t,
		`SELECT "id" FROM "docs" WHERE DBMS_LOB.COMPARE("body", TO_CLOB('x')) <> 0`,
		c[sqlkata.EngineOracle])
}

func TestOrWhereTextEqual(t *testing.T) {
	s := newTestSupport()
	q := sqlkata.NewQuery().From("docs").Select("id").
		WhereEq("id", 1).
		OrWhereTextEqual("body", "a")
	c := s.Compile(q)

	assertEqual(t,
		"SELECT [id] FROM [docs] WHERE [id] = 1 OR [body] = 'a'",
		c[sqlkata.EngineSqlServer])
	assertEqual(t,
		`SELECT "id" FROM "docs" WHERE "id" = 1 OR DBMS_LOB.COMPARE("body", TO_CLOB('a')) = 0`,
		c[sqlkata.EngineOracle])
}

func TestSelectAsText(t *testing.T) {
	s := newTestSupport()
	q := sqlkata.NewQuery().From("docs").SelectAsText("body", "BodyText")
	c := s.Compile(q)

	assertEqual(t,
		"SELECT [body] AS [BodyText] FROM [docs]",
		c[sqlkata.EngineSqlServer])
	assertEqual(t,
		`SELECT "body" AS "BodyText" FROM "docs"`,
		c[sqlkata.EnginePostgres])
	assertEqual(t,
		`SELECT DBMS_LOB.SUBSTR("body", 4000, 1) "BodyText" FROM "docs"`,
		c[sqlkata.EngineOracle])
}

func TestSelectAsTextCustomMaxLen(t *testing.T) {
	s := newTestSupport()
	q := sqlkata.NewQuery().From("docs").SelectAsText("body", "BodyText", 2000)
	c := s.Compile(q)

	assertEqual(t,
		`SELECT DBMS_LOB.SUBSTR("body", 2000, 1) "BodyText" FROM "docs"`,
		c[sqlkata.EngineOracle])
	// Non-Oracle ignores MaxLen and projects the column.
	assertEqual(t,
		`SELECT "body" AS "BodyText" FROM "docs"`,
		c[sqlkata.EnginePostgres])
}
