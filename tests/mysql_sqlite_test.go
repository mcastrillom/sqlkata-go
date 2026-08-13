package tests

import (
	"testing"

	"github.com/mcastrillom/sqlkata-go/compiler"
	"github.com/mcastrillom/sqlkata-go/sqlkata"
)

func TestMySqlBasicSelect(t *testing.T) {
	s := newTestSupport()
	q := sqlkata.NewQuery().From("users").Select("id", "name").WhereEq("status", "active")

	assertEqual(t, "SELECT `id`, `name` FROM `users` WHERE `status` = 'active'", s.Compile(q)[sqlkata.EngineMySql])
}

func TestSqliteBasicSelect(t *testing.T) {
	s := newTestSupport()
	q := sqlkata.NewQuery().From("users").Select("id", "name").WhereEq("status", "active")

	assertEqual(t, `SELECT "id", "name" FROM "users" WHERE "status" = 'active'`, s.Compile(q)[sqlkata.EngineSqlite])
}

func TestMySqlSelectWithAlias(t *testing.T) {
	s := newTestSupport()
	q := sqlkata.NewQuery().From("users as u").Select("u.id")

	assertEqual(t, "SELECT `u`.`id` FROM `users` AS `u`", s.Compile(q)[sqlkata.EngineMySql])
}

// TestMySqlEscapesIdentifierQuote guards against a backtick in a column name
// breaking out of the quoted identifier.
func TestMySqlEscapesIdentifierQuote(t *testing.T) {
	s := newTestSupport()
	q := sqlkata.NewQuery().From("users").Select("we`ird")

	assertEqual(t, "SELECT `we``ird` FROM `users`", s.Compile(q)[sqlkata.EngineMySql])
}

func TestMySqlLimitOnly(t *testing.T) {
	comp := NewMySqlCompiler()
	ctx := &compiler.SqlResult{Query: sqlkata.NewQuery().From("Table").Limit(10)}

	assertEqual(t, "LIMIT ?", comp.CompileLimit(ctx))
	assertBinding(t, 10, ctx.Bindings[0])
}

// TestMySqlOffsetOnly covers the MySQL rule that OFFSET needs a LIMIT, so the
// compiler emits the max BIGINT UNSIGNED as upper bound.
func TestMySqlOffsetOnly(t *testing.T) {
	comp := NewMySqlCompiler()
	ctx := &compiler.SqlResult{Query: sqlkata.NewQuery().From("Table").Offset(20)}

	assertEqual(t, "LIMIT 18446744073709551615 OFFSET ?", comp.CompileLimit(ctx))
	assertBinding(t, int64(20), ctx.Bindings[0])
}

func TestMySqlLimitAndOffset(t *testing.T) {
	comp := NewMySqlCompiler()
	ctx := &compiler.SqlResult{Query: sqlkata.NewQuery().From("Table").Limit(5).Offset(20)}

	assertEqual(t, "LIMIT ? OFFSET ?", comp.CompileLimit(ctx))
	assertBinding(t, 5, ctx.Bindings[0])
	assertBinding(t, int64(20), ctx.Bindings[1])
}

func TestSqliteLimitOnly(t *testing.T) {
	comp := NewSqliteCompiler()
	ctx := &compiler.SqlResult{Query: sqlkata.NewQuery().From("Table").Limit(10)}

	assertEqual(t, "LIMIT ?", comp.CompileLimit(ctx))
	assertBinding(t, 10, ctx.Bindings[0])
}

// TestSqliteOffsetOnly covers SQLite's "-1 means no limit" idiom.
func TestSqliteOffsetOnly(t *testing.T) {
	comp := NewSqliteCompiler()
	ctx := &compiler.SqlResult{Query: sqlkata.NewQuery().From("Table").Offset(20)}

	assertEqual(t, "LIMIT -1 OFFSET ?", comp.CompileLimit(ctx))
	assertBinding(t, int64(20), ctx.Bindings[0])
}

func TestSqliteLimitAndOffset(t *testing.T) {
	comp := NewSqliteCompiler()
	ctx := &compiler.SqlResult{Query: sqlkata.NewQuery().From("Table").Limit(5).Offset(20)}

	assertEqual(t, "LIMIT ? OFFSET ?", comp.CompileLimit(ctx))
	assertBinding(t, 5, ctx.Bindings[0])
	assertBinding(t, int64(20), ctx.Bindings[1])
}

func TestPaginationIsDialectSpecific(t *testing.T) {
	s := newTestSupport()
	q := sqlkata.NewQuery().From("users").OrderBy("id").Limit(10).Offset(20)
	c := s.Compile(q)

	assertEqual(t, "SELECT * FROM `users` ORDER BY `id` LIMIT 10 OFFSET 20", c[sqlkata.EngineMySql])
	assertEqual(t, `SELECT * FROM "users" ORDER BY "id" LIMIT 10 OFFSET 20`, c[sqlkata.EngineSqlite])
}

func TestBooleanConditionPerDialect(t *testing.T) {
	s := newTestSupport()
	q := sqlkata.NewQuery().From("users").WhereTrue("active")
	c := s.Compile(q)

	assertEqual(t, "SELECT * FROM `users` WHERE `active` = true", c[sqlkata.EngineMySql])
	assertEqual(t, `SELECT * FROM "users" WHERE "active" = 1`, c[sqlkata.EngineSqlite])
}

func TestOrderByRandomPerDialect(t *testing.T) {
	s := newTestSupport()
	q := sqlkata.NewQuery().From("users").OrderByRandom("")
	c := s.Compile(q)

	assertEqual(t, "SELECT * FROM `users` ORDER BY RAND()", c[sqlkata.EngineMySql])
	assertEqual(t, `SELECT * FROM "users" ORDER BY RANDOM()`, c[sqlkata.EngineSqlite])
}

func TestDatePartPerDialect(t *testing.T) {
	s := newTestSupport()
	q := sqlkata.NewQuery().From("posts").WhereDatePartEq("year", "created_at", 2024)
	c := s.Compile(q)

	assertEqual(t, "SELECT * FROM `posts` WHERE YEAR(`created_at`) = 2024", c[sqlkata.EngineMySql])
	assertEqual(t, `SELECT * FROM "posts" WHERE CAST(strftime('%Y', "created_at") AS INTEGER) = 2024`, c[sqlkata.EngineSqlite])
}

func TestWholeDatePerDialect(t *testing.T) {
	s := newTestSupport()
	q := sqlkata.NewQuery().From("posts").WhereDateEq("created_at", "2024-01-31")
	c := s.Compile(q)

	assertEqual(t, "SELECT * FROM `posts` WHERE DATE(`created_at`) = '2024-01-31'", c[sqlkata.EngineMySql])
	assertEqual(t, `SELECT * FROM "posts" WHERE date("created_at") = '2024-01-31'`, c[sqlkata.EngineSqlite])
}

func TestInsertReturnsLastIdPerDialect(t *testing.T) {
	s := newTestSupport()
	q := sqlkata.NewQuery().From("users").AsInsert(map[string]any{"name": "Ada"}, true)
	c := s.Compile(q)

	assertEqual(t, "INSERT INTO `users` (`name`) VALUES ('Ada');SELECT last_insert_id() as Id", c[sqlkata.EngineMySql])
	assertEqual(t, `INSERT INTO "users" ("name") VALUES ('Ada');select last_insert_rowid() as id`, c[sqlkata.EngineSqlite])
}

func TestForMySqlAndForSqliteScopedClauses(t *testing.T) {
	s := newTestSupport()
	q := sqlkata.NewQuery().From("users").
		ForMySql(func(q *sqlkata.Query) *sqlkata.Query { return q.Limit(5) }).
		ForSqlite(func(q *sqlkata.Query) *sqlkata.Query { return q.Limit(9) })
	c := s.Compile(q)

	assertEqual(t, "SELECT * FROM `users` LIMIT 5", c[sqlkata.EngineMySql])
	assertEqual(t, `SELECT * FROM "users" LIMIT 9`, c[sqlkata.EngineSqlite])
	assertEqual(t, "SELECT * FROM [users]", c[sqlkata.EngineSqlServer])
}
