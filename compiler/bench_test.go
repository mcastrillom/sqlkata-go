package compiler_test

import (
	"fmt"
	"testing"

	"github.com/mcastrillom/sqlkata-go/compiler"
	"github.com/mcastrillom/sqlkata-go/sqlkata"
)

func benchCompile(b *testing.B, build func() *sqlkata.Query) {
	b.Helper()
	c := compiler.NewSqlServerCompiler()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q := build()
		res, err := c.Compile(q)
		if err != nil {
			b.Fatal(err)
		}
		if res.RawSQL == "" {
			b.Fatal("empty SQL")
		}
	}
}

// benchCompileReuse measures compilation only, reusing one prebuilt query.
func benchCompileReuse(b *testing.B, q *sqlkata.Query) {
	b.Helper()
	c := compiler.NewSqlServerCompiler()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		res, err := c.Compile(q)
		if err != nil {
			b.Fatal(err)
		}
		if res.RawSQL == "" {
			b.Fatal("empty SQL")
		}
	}
}

func simpleSelect() *sqlkata.Query {
	return sqlkata.NewQuery().From("users").Select("id", "name").WhereEq("status", "active")
}

func selectWhereJoin() *sqlkata.Query {
	return sqlkata.NewQuery().
		From("users").
		Select("users.id", "users.name", "profiles.bio").
		LeftJoin("profiles", "profiles.user_id", "users.id").
		WhereEq("users.status", "active").
		WhereIn("users.role", "admin", "editor", "viewer").
		WhereNotNull("users.email").
		OrderBy("users.name").
		Limit(50)
}

func paginatedSelect() *sqlkata.Query {
	return sqlkata.NewQuery().
		From("orders").
		Select("id", "total").
		WhereEq("customer_id", 42).
		OrderByDesc("created_at").
		ForPage(3, 25)
}

func insertQuery() *sqlkata.Query {
	return sqlkata.NewQuery().From("users").Insert(map[string]any{
		"name":   "Ada",
		"email":  "ada@example.com",
		"active": true,
	})
}

// bigSelect returns a builder for a report-style query with n columns, n where
// conditions, n order clauses and 3 joins, so clause count grows with n.
func bigSelect(n int) func() *sqlkata.Query {
	cols := make([]string, n)
	names := make([]string, n)
	for i := 0; i < n; i++ {
		cols[i] = fmt.Sprintf("users.col%d", i)
		names[i] = fmt.Sprintf("col%d", i)
	}
	return func() *sqlkata.Query {
		q := sqlkata.NewQuery().
			From("users").
			Select(cols...).
			LeftJoin("profiles", "profiles.user_id", "users.id").
			LeftJoin("orders", "orders.user_id", "users.id").
			LeftJoin("countries", "countries.id", "users.country_id")
		for i, name := range names {
			q = q.WhereEq(name, i)
		}
		q = q.OrderBy(names...).GroupBy("users.id").Limit(100)
		return q
	}
}

// benchBuild measures query construction only (no compile), to catch
// regressions in the clause-append path.
func benchBuild(b *testing.B, build func() *sqlkata.Query) {
	b.Helper()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if q := build(); q == nil {
			b.Fatal("nil query")
		}
	}
}

func BenchmarkCompileBigSelect10(b *testing.B)  { benchCompile(b, bigSelect(10)) }
func BenchmarkCompileBigSelect40(b *testing.B)  { benchCompile(b, bigSelect(40)) }
func BenchmarkCompileBigSelect100(b *testing.B) { benchCompile(b, bigSelect(100)) }

func BenchmarkCompileReuseBigSelect40(b *testing.B)  { benchCompileReuse(b, bigSelect(40)()) }
func BenchmarkCompileReuseBigSelect100(b *testing.B) { benchCompileReuse(b, bigSelect(100)()) }

func BenchmarkBuildSimpleSelect(b *testing.B) { benchBuild(b, simpleSelect) }
func BenchmarkBuildBigSelect40(b *testing.B)  { benchBuild(b, bigSelect(40)) }

func BenchmarkCompileSimpleSelect(b *testing.B)    { benchCompile(b, simpleSelect) }
func BenchmarkCompileSelectWhereJoin(b *testing.B) { benchCompile(b, selectWhereJoin) }
func BenchmarkCompilePaginatedSelect(b *testing.B) { benchCompile(b, paginatedSelect) }
func BenchmarkCompileInsert(b *testing.B)          { benchCompile(b, insertQuery) }
func BenchmarkCompileReuseSimple(b *testing.B)     { benchCompileReuse(b, simpleSelect()) }
func BenchmarkCompileReuseWhereJoin(b *testing.B)  { benchCompileReuse(b, selectWhereJoin()) }
