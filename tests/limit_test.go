package tests

import (
	"strings"
	"testing"

	"github.com/mcastrillom/sqlkata-go/compiler"
	"github.com/mcastrillom/sqlkata-go/sqlkata"
)

func TestSqlServerLimitOnly(t *testing.T) {
	comp := NewSqlServerCompiler()
	query := sqlkata.NewQuery().From("Table").Limit(10)
	ctx := &compiler.SqlResult{Query: query}

	got := comp.CompileLimit(ctx)
	if !strings.HasSuffix(got, "OFFSET ? ROWS FETCH NEXT ? ROWS ONLY") {
		t.Fatalf("CompileLimit = %q", got)
	}
	if len(ctx.Bindings) != 2 {
		t.Fatalf("bindings = %v", ctx.Bindings)
	}
	assertBinding(t, int64(0), ctx.Bindings[0])
	assertBinding(t, int64(10), ctx.Bindings[1])
}

func TestSqlServerOffsetOnly(t *testing.T) {
	comp := NewSqlServerCompiler()
	query := sqlkata.NewQuery().From("Table").Offset(20)
	ctx := &compiler.SqlResult{Query: query}

	got := comp.CompileLimit(ctx)
	if !strings.HasSuffix(got, "OFFSET ? ROWS") {
		t.Fatalf("CompileLimit = %q", got)
	}
	if len(ctx.Bindings) != 1 {
		t.Fatalf("bindings = %v", ctx.Bindings)
	}
	assertBinding(t, int64(20), ctx.Bindings[0])
}

func TestSqlServerLimitAndOffset(t *testing.T) {
	comp := NewSqlServerCompiler()
	query := sqlkata.NewQuery().From("Table").Limit(5).Offset(20)
	ctx := &compiler.SqlResult{Query: query}

	got := comp.CompileLimit(ctx)
	if !strings.HasSuffix(got, "OFFSET ? ROWS FETCH NEXT ? ROWS ONLY") {
		t.Fatalf("CompileLimit = %q", got)
	}
	assertBinding(t, int64(20), ctx.Bindings[0])
	assertBinding(t, int64(5), ctx.Bindings[1])
}

func TestSqlServerEmulateOrderByIfNoOrderByProvided(t *testing.T) {
	comp := NewSqlServerCompiler()
	query := sqlkata.NewQuery().From("Table").Limit(5).Offset(20)
	sql := mustCompile(comp.Compiler, query).String()
	if !strings.Contains(sql, "ORDER BY (SELECT 0)") {
		t.Fatalf("expected safe ORDER BY, got %s", sql)
	}
}

func TestSqlServerKeepOrdersIfPaginationProvided(t *testing.T) {
	comp := NewSqlServerCompiler()
	query := sqlkata.NewQuery().From("Table").OrderBy("Id").Limit(5).Offset(20)
	sql := mustCompile(comp.Compiler, query).String()
	if !strings.Contains(sql, "ORDER BY [Id]") {
		t.Fatalf("expected ORDER BY [Id], got %s", sql)
	}
	if strings.Contains(sql, "ORDER BY (SELECT 0)") {
		t.Fatalf("should not inject safe order: %s", sql)
	}
}

func TestPostgresLimitOnly(t *testing.T) {
	comp := NewPostgresCompiler()
	query := sqlkata.NewQuery().From("Table").Limit(10)
	ctx := &compiler.SqlResult{Query: query}

	assertEqual(t, "LIMIT ?", comp.CompileLimit(ctx))
	assertBinding(t, 10, ctx.Bindings[0])
}

func TestPostgresOffsetOnly(t *testing.T) {
	comp := NewPostgresCompiler()
	query := sqlkata.NewQuery().From("Table").Offset(20)
	ctx := &compiler.SqlResult{Query: query}

	assertEqual(t, "OFFSET ?", comp.CompileLimit(ctx))
	assertBinding(t, int64(20), ctx.Bindings[0])
}

func TestPostgresLimitAndOffset(t *testing.T) {
	comp := NewPostgresCompiler()
	query := sqlkata.NewQuery().From("Table").Limit(5).Offset(20)
	ctx := &compiler.SqlResult{Query: query}

	assertEqual(t, "LIMIT ? OFFSET ?", comp.CompileLimit(ctx))
	assertBinding(t, 5, ctx.Bindings[0])
	assertBinding(t, int64(20), ctx.Bindings[1])
}

func TestOracleLimitOnly(t *testing.T) {
	comp := NewOracleCompiler()
	query := sqlkata.NewQuery().From("Table").Limit(10)
	ctx := &compiler.SqlResult{Query: query}

	got := comp.CompileLimit(ctx)
	if !strings.HasSuffix(got, "OFFSET ? ROWS FETCH NEXT ? ROWS ONLY") {
		t.Fatalf("CompileLimit = %q", got)
	}
	assertBinding(t, int64(0), ctx.Bindings[0])
	assertBinding(t, int64(10), ctx.Bindings[1])
}

func TestOracleOffsetOnly(t *testing.T) {
	comp := NewOracleCompiler()
	query := sqlkata.NewQuery().From("Table").Offset(20)
	ctx := &compiler.SqlResult{Query: query}

	got := comp.CompileLimit(ctx)
	if !strings.HasSuffix(got, "OFFSET ? ROWS") {
		t.Fatalf("CompileLimit = %q", got)
	}
	assertBinding(t, int64(20), ctx.Bindings[0])
}

func TestOracleLimitAndOffset(t *testing.T) {
	comp := NewOracleCompiler()
	query := sqlkata.NewQuery().From("Table").Limit(5).Offset(20)
	ctx := &compiler.SqlResult{Query: query}

	got := comp.CompileLimit(ctx)
	if !strings.HasSuffix(got, "OFFSET ? ROWS FETCH NEXT ? ROWS ONLY") {
		t.Fatalf("CompileLimit = %q", got)
	}
	assertBinding(t, int64(20), ctx.Bindings[0])
	assertBinding(t, int64(5), ctx.Bindings[1])
}
