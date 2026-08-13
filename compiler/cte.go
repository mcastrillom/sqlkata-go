package compiler

import (
	"strings"

	"github.com/mcastrillom/sqlkata-go/sqlkata"
)

// findCtes mirrors SqlKata.Compilers.CteFinder (nested CTEs first, unique by alias).
func findCtes(q *sqlkata.Query, eng *string, seen map[string]struct{}) []sqlkata.AbstractClause {
	if seen == nil {
		seen = map[string]struct{}{}
	}
	list := sqlkata.GetComponentsAs[sqlkata.AbstractClause](&q.BaseQuery, "cte", eng)
	var result []sqlkata.AbstractClause
	for _, cte := range list {
		alias := cteAlias(cte)
		if alias == "" {
			continue
		}
		if _, ok := seen[alias]; ok {
			continue
		}
		seen[alias] = struct{}{}
		result = append(result, cte)
		if qf, ok := cte.(*sqlkata.QueryFromClause); ok && qf.Query != nil {
			nested := findCtes(qf.Query, eng, seen)
			result = append(nested, result...)
		}
	}
	return result
}

func cteAlias(cl sqlkata.AbstractClause) string {
	switch x := cl.(type) {
	case *sqlkata.QueryFromClause:
		return x.Alias()
	case *sqlkata.RawFromClause:
		return x.Alias()
	case *sqlkata.AdHocTableFromClause:
		return x.Alias()
	default:
		return ""
	}
}

func (c *Compiler) compileCteQuery(ctx *SqlResult, q *sqlkata.Query) {
	ctes := findCtes(q, c.eng(), nil)
	if len(ctes) == 0 {
		return
	}
	var b strings.Builder
	b.WriteString("WITH ")
	var cteBindings []any
	for i, cte := range ctes {
		part, bindings := c.compileCte(cte)
		cteBindings = append(cteBindings, bindings...)
		b.WriteString(strings.TrimSpace(part))
		if i < len(ctes)-1 {
			b.WriteString(",\n")
		}
	}
	b.WriteByte('\n')
	b.WriteString(ctx.RawSQL)
	ctx.Bindings = append(cteBindings, ctx.Bindings...)
	ctx.RawSQL = b.String()
}

func (c *Compiler) compileCte(cte sqlkata.AbstractClause) (sql string, bindings []any) {
	ctx := &SqlResult{}
	switch x := cte.(type) {
	case *sqlkata.RawFromClause:
		ctx.Bindings = append(ctx.Bindings, x.Bindings...)
		sql = c.wrapValue(x.Alias()) + " AS (" + c.wrapIdentifiers(x.Expression) + ")"
	case *sqlkata.QueryFromClause:
		sub := c.compileSelectQuery(ctx, x.Query)
		sql = c.wrapValue(x.Alias()) + " AS (" + sub + ")"
	case *sqlkata.AdHocTableFromClause:
		adhoc := c.compileAdHocQuery(x)
		ctx.Bindings = append(ctx.Bindings, adhoc.Bindings...)
		sql = c.wrapValue(x.Alias()) + " AS (" + adhoc.RawSQL + ")"
	default:
		return "", nil
	}
	return sql, ctx.Bindings
}

func (c *Compiler) compileAdHocQuery(adHoc *sqlkata.AdHocTableFromClause) *SqlResult {
	if c.EngineCode == sqlkata.EngineSqlServer {
		return c.compileAdHocQuerySqlServer(adHoc)
	}
	return c.compileAdHocQueryBase(adHoc)
}

func (c *Compiler) compileAdHocQuerySqlServer(adHoc *sqlkata.AdHocTableFromClause) *SqlResult {
	ctx := &SqlResult{Bindings: append([]any(nil), adHoc.Values...)}
	cols := make([]string, 0, len(adHoc.Columns))
	for _, col := range adHoc.Columns {
		cols = append(cols, c.wrap(col))
	}
	colNames := strings.Join(cols, ", ")
	phRow := make([]string, len(adHoc.Columns))
	for i := range phRow {
		phRow[i] = c.placeholder
	}
	valueRow := "(" + strings.Join(phRow, ", ") + ")"
	nRows := 0
	if len(adHoc.Columns) > 0 {
		nRows = len(adHoc.Values) / len(adHoc.Columns)
	}
	rows := make([]string, nRows)
	for i := range rows {
		rows[i] = valueRow
	}
	ctx.RawSQL = "SELECT " + colNames + " FROM (VALUES " + strings.Join(rows, ", ") + ") AS tbl (" + colNames + ")"
	return ctx
}

func (c *Compiler) compileAdHocQueryBase(adHoc *sqlkata.AdHocTableFromClause) *SqlResult {
	ctx := &SqlResult{Bindings: append([]any(nil), adHoc.Values...)}
	nCols := len(adHoc.Columns)
	if nCols == 0 {
		return ctx
	}
	parts := make([]string, nCols)
	for i, col := range adHoc.Columns {
		parts[i] = c.placeholder + " AS " + c.wrap(col)
	}
	row := "SELECT " + strings.Join(parts, ", ")
	if c.EngineCode == sqlkata.EngineOracle {
		row += " FROM DUAL"
	}
	nRows := len(adHoc.Values) / nCols
	rows := make([]string, nRows)
	for i := range rows {
		rows[i] = row
	}
	ctx.RawSQL = strings.Join(rows, " UNION ALL ")
	return ctx
}
