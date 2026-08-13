package execution

import (
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/mcastrillom/sqlkata-go/compiler"
	"github.com/mcastrillom/sqlkata-go/sqlkata"
)

func mustFactoryCompile(t *testing.T, f *QueryFactory, q *sqlkata.Query) *compiler.SqlResult {
	t.Helper()
	res, err := f.Compile(q)
	if err != nil {
		t.Fatal(err)
	}
	return res
}

func TestBindOracleUsesPositionalColon(t *testing.T) {
	f := New(nil, compiler.NewOracleCompiler())
	res := mustFactoryCompile(t, f, sqlkata.NewQuery().From("T").Select("Id").WhereEq("A", 1).WhereEq("B", 2))
	sql, args := f.bind(res)
	if want := `SELECT "Id" FROM "T" WHERE "A" = :1 AND "B" = :2`; sql != want {
		t.Fatalf("sql=\n%s\nwant\n%s", sql, want)
	}
	if len(args) != 2 {
		t.Fatalf("args=%v", args)
	}
}

func TestBindPostgresUsesDollar(t *testing.T) {
	f := New(nil, compiler.NewPostgresCompiler())
	res := mustFactoryCompile(t, f, sqlkata.NewQuery().From("t").Select("id").WhereEq("a", 1))
	sql, _ := f.bind(res)
	if want := `SELECT "id" FROM "t" WHERE "a" = $1`; sql != want {
		t.Fatalf("sql=%s", sql)
	}
}

func TestBindSqlServerKeepsQuestionWithoutDriver(t *testing.T) {
	f := New(nil, compiler.NewSqlServerCompiler())
	res := mustFactoryCompile(t, f, sqlkata.NewQuery().From("t").Select("id").WhereEq("a", 1))
	sql, _ := f.bind(res)
	if want := `SELECT [id] FROM [t] WHERE [a] = ?`; sql != want {
		t.Fatalf("sql=%s", sql)
	}
}

func TestBindMySqlAndSqliteKeepQuestion(t *testing.T) {
	q := sqlkata.NewQuery().From("t").Select("id").WhereEq("a", 1)

	f := New(nil, compiler.NewMySqlCompiler())
	sql, _ := f.bind(mustFactoryCompile(t, f, q))
	if want := "SELECT `id` FROM `t` WHERE `a` = ?"; sql != want {
		t.Fatalf("mysql sql=%s", sql)
	}

	f = New(nil, compiler.NewSqliteCompiler())
	sql, _ = f.bind(mustFactoryCompile(t, f, q))
	if want := `SELECT "id" FROM "t" WHERE "a" = ?`; sql != want {
		t.Fatalf("sqlite sql=%s", sql)
	}
}

func TestBindStyleOverride(t *testing.T) {
	f := New(nil, compiler.NewOracleCompiler())
	f.BindStyle = sqlx.QUESTION
	res := mustFactoryCompile(t, f, sqlkata.NewQuery().From("T").WhereEq("A", 1))
	sql, _ := f.bind(res)
	if !containsByte(sql, '?') {
		t.Fatalf("expected ? override, got %s", sql)
	}
}

func containsByte(s string, c byte) bool {
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return true
		}
	}
	return false
}
