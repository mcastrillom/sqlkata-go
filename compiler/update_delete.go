package compiler

import (
	"fmt"
	"math"
	"strings"

	"github.com/mcastrillom/sqlkata-go/sqlkata"
)

func (c *Compiler) resolveTable(ctx *SqlResult, q *sqlkata.Query) (table string, from sqlkata.AbstractClause, err error) {
	eng := c.eng()
	if !q.HasComponent("from", eng) {
		return "", nil, fmt.Errorf("no table set")
	}
	cl := q.GetOneComponent("from", eng)
	if cl == nil {
		return "", nil, fmt.Errorf("invalid table expression")
	}
	switch x := cl.(type) {
	case *sqlkata.FromClause:
		return c.wrap(x.Table), cl, nil
	case *sqlkata.RawFromClause:
		ctx.Bindings = append(ctx.Bindings, x.Bindings...)
		return c.wrapIdentifiers(x.Expression), cl, nil
	default:
		return "", nil, fmt.Errorf("invalid table expression")
	}
}

func (c *Compiler) compileUpdateQuery(ctx *SqlResult, q *sqlkata.Query) string {
	table, _, err := c.resolveTable(ctx, q)
	if err != nil {
		return "-- update: " + err.Error()
	}
	eng := c.eng()
	clause := q.GetOneComponent("update", eng)

	if inc, ok := clause.(*sqlkata.IncrementClause); ok && inc != nil {
		col := c.wrap(inc.Column)
		sign := "+"
		if inc.Value < 0 {
			sign = "-"
		}
		val := c.parameter(ctx, int(math.Abs(float64(inc.Value))))
		where := c.compileWheres(ctx, q)
		if where != "" {
			where = " " + where
		}
		return "UPDATE " + table + " SET " + col + " = " + col + " " + sign + " " + val + where
	}

	upd, ok := clause.(*sqlkata.InsertClause)
	if !ok || upd == nil {
		return "-- update: missing values"
	}
	parts := make([]string, 0, len(upd.Columns))
	for i, col := range upd.Columns {
		var v any
		if i < len(upd.Values) {
			v = upd.Values[i]
		}
		parts = append(parts, c.wrap(col)+" = "+c.parameter(ctx, v))
	}
	where := c.compileWheres(ctx, q)
	if where != "" {
		where = " " + where
	}
	return "UPDATE " + table + " SET " + strings.Join(parts, ", ") + where
}

func (c *Compiler) compileDeleteQuery(ctx *SqlResult, q *sqlkata.Query) string {
	table, from, err := c.resolveTable(ctx, q)
	if err != nil {
		return "-- delete: " + err.Error()
	}

	joins := c.compileJoins(ctx, q)
	where := c.compileWheres(ctx, q)
	if where != "" {
		where = " " + where
	}

	if strings.TrimSpace(joins) == "" {
		return "DELETE FROM " + table + where
	}

	// SqlKata: if FromClause has alias ("table as alias"), DELETE alias FROM ...
	if fc, ok := from.(*sqlkata.FromClause); ok {
		if strings.Contains(strings.ToLower(fc.Table), " as ") {
			return "DELETE " + c.wrap(fc.Alias()) + " FROM " + table + " " + joins + where
		}
	}
	return "DELETE " + table + " FROM " + table + " " + joins + where
}
