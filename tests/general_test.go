package tests

import (
	"testing"

	"github.com/mcastrillom/sqlkata-go/sqlkata"
)

func TestColumnsEscaping(t *testing.T) {
	s := newTestSupport()
	q := sqlkata.NewQuery().From("users").Select("mycol[isthis]")
	c := s.Compile(q)

	assertEqual(t, "SELECT [mycol[isthis]]] FROM [users]", c[sqlkata.EngineSqlServer])
}

func TestInnerScopeEngineWithinCTE(t *testing.T) {
	s := newTestSupport()
	series := sqlkata.NewQuery().From("table").
		ForPostgreSql(func(q *sqlkata.Query) *sqlkata.Query {
			return q.WhereRaw("postgres = true")
		}).
		ForSqlServer(func(q *sqlkata.Query) *sqlkata.Query {
			return q.WhereRaw("sqlsrv = 1")
		}).
		ForOracle(func(q *sqlkata.Query) *sqlkata.Query {
			return q.WhereRaw("oracle = 1")
		})
	query := sqlkata.NewQuery().From("series").WithAlias("series", series)
	c := s.Compile(query)

	assertEqual(t, "WITH [series] AS (SELECT * FROM [table] WHERE sqlsrv = 1)\nSELECT * FROM [series]", c[sqlkata.EngineSqlServer])
	assertEqual(t, "WITH \"series\" AS (SELECT * FROM \"table\" WHERE postgres = true)\nSELECT * FROM \"series\"", c[sqlkata.EnginePostgres])
	assertEqual(t, "WITH \"series\" AS (SELECT * FROM \"table\" WHERE oracle = 1)\nSELECT * FROM \"series\"", c[sqlkata.EngineOracle])
}

func TestInnerScopeEngineWithinSubQuery(t *testing.T) {
	s := newTestSupport()
	series := sqlkata.NewQuery().From("table").
		ForPostgreSql(func(q *sqlkata.Query) *sqlkata.Query {
			return q.WhereRaw("postgres = true")
		}).
		ForSqlServer(func(q *sqlkata.Query) *sqlkata.Query {
			return q.WhereRaw("sqlsrv = 1")
		})
	query := sqlkata.NewQuery().FromQuery(series, "series")
	c := s.Compile(query)

	assertEqual(t, "SELECT * FROM (SELECT * FROM [table] WHERE sqlsrv = 1) AS [series]", c[sqlkata.EngineSqlServer])
	assertEqual(t, `SELECT * FROM (SELECT * FROM "table" WHERE postgres = true) AS "series"`, c[sqlkata.EnginePostgres])
}

func TestShouldEqualAfterMultipleCompile(t *testing.T) {
	s := newTestSupport()
	query := sqlkata.NewQuery().
		Select("Id", "Name").
		From("Table").
		OrderBy("Name").
		Limit(20).
		Offset(1)

	first := s.Compile(query)
	second := s.Compile(query)

	assertEqual(t, first[sqlkata.EngineSqlServer], second[sqlkata.EngineSqlServer])
	assertEqual(t, first[sqlkata.EnginePostgres], second[sqlkata.EnginePostgres])
	assertEqual(t, first[sqlkata.EngineOracle], second[sqlkata.EngineOracle])

	assertEqual(t,
		"SELECT [Id], [Name] FROM [Table] ORDER BY [Name] OFFSET 1 ROWS FETCH NEXT 20 ROWS ONLY",
		first[sqlkata.EngineSqlServer])
	assertEqual(t,
		`SELECT "Id", "Name" FROM "Table" ORDER BY "Name" LIMIT 20 OFFSET 1`,
		first[sqlkata.EnginePostgres])
}

func TestRawWrapIdentifiers(t *testing.T) {
	s := newTestSupport()
	query := sqlkata.NewQuery().From("Users").SelectRaw("[Id], [Name], {Age}")
	c := s.Compile(query)

	assertEqual(t, "SELECT [Id], [Name], [Age] FROM [Users]", c[sqlkata.EngineSqlServer])
	assertEqual(t, `SELECT "Id", "Name", "Age" FROM "Users"`, c[sqlkata.EnginePostgres])
}

func TestCloneShouldProduceIndependentIncludesList(t *testing.T) {
	query := sqlkata.NewQuery().From("users").
		Include("posts", sqlkata.NewQuery().From("posts"), "user_id", "id")
	clone := query.Clone()
	clone.Include("comments", sqlkata.NewQuery().From("comments"), "user_id", "id")

	if len(query.Includes) != 1 {
		t.Fatalf("original Includes = %d, want 1", len(query.Includes))
	}
	if len(clone.Includes) != 2 {
		t.Fatalf("clone Includes = %d, want 2", len(clone.Includes))
	}
}

func TestCloneShouldProduceIndependentIncludeObjects(t *testing.T) {
	query := sqlkata.NewQuery().From("users").
		Include("posts", sqlkata.NewQuery().From("posts"), "user_id", "id")
	clone := query.Clone()
	clone.Includes[0].Name = "modified_name"

	assertEqual(t, "posts", query.Includes[0].Name)
	assertEqual(t, "modified_name", clone.Includes[0].Name)
}

func TestCloneShouldProduceIndependentVariablesDictionary(t *testing.T) {
	query := sqlkata.NewQuery().From("users").Define("limit", 10)
	clone := query.Clone()
	clone.Define("offset", 5)

	if _, ok := query.Variables["offset"]; ok {
		t.Fatal("original should not have offset")
	}
	if _, ok := clone.Variables["offset"]; !ok {
		t.Fatal("clone should have offset")
	}
}

func TestCloneShouldPreserveAllProperties(t *testing.T) {
	query := sqlkata.NewQuery().From("users").
		Select("id", "name").
		WhereEq("active", true).
		Distinct().
		As("u").
		Include("posts", sqlkata.NewQuery().From("posts"), "user_id", "id").
		Define("myvar", 42)
	clone := query.Clone()

	assertEqual(t, query.QueryAlias, clone.QueryAlias)
	if query.IsDistinct != clone.IsDistinct {
		t.Fatal("IsDistinct mismatch")
	}
	assertEqual(t, query.Method, clone.Method)
	if len(query.Includes) != len(clone.Includes) {
		t.Fatal("Includes count mismatch")
	}
	assertEqual(t, query.Includes[0].Name, clone.Includes[0].Name)
	if query.Variables["myvar"] != clone.Variables["myvar"] {
		t.Fatal("Variables mismatch")
	}
}

func TestUnion(t *testing.T) {
	s := newTestSupport()
	q := sqlkata.NewQuery().From("A").Select("Id").
		UnionAll(sqlkata.NewQuery().From("B").Select("Id")).
		Union(sqlkata.NewQuery().From("C").Select("Id"))
	c := s.Compile(q)

	assertEqual(t, "SELECT [Id] FROM [A] UNION ALL SELECT [Id] FROM [B] UNION SELECT [Id] FROM [C]", c[sqlkata.EngineSqlServer])
}
