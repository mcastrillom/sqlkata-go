package sqlkata

import (
	"fmt"
	"strings"
)

// With adds a CTE from a query that already has QueryAlias (SqlKata.With(Query)).
func (q *Query) With(cte *Query) *Query {
	if cte == nil || strings.TrimSpace(cte.QueryAlias) == "" {
		return q.setErr(fmt.Errorf("no alias found for the CTE query"))
	}
	cloned := cte.Clone()
	alias := strings.TrimSpace(cloned.QueryAlias)
	cloned.QueryAlias = ""
	clause := &QueryFromClause{Query: cloned}
	clause.SetAlias(alias)
	return q.AddComponent("cte", clause, nil)
}

// WithAlias adds a CTE with an explicit alias (SqlKata.With(string, Query)).
func (q *Query) WithAlias(alias string, cte *Query) *Query {
	return q.With(cte.Clone().As(alias))
}

// WithFunc builds a CTE via callback (SqlKata.With(Func)).
func (q *Query) WithFunc(fn func(*Query) *Query) *Query {
	return q.With(fn(NewQuery()))
}

// WithAliasFunc builds a named CTE via callback.
func (q *Query) WithAliasFunc(alias string, fn func(*Query) *Query) *Query {
	return q.WithAlias(alias, fn(NewQuery()))
}

// WithRaw adds a raw CTE expression (SqlKata.WithRaw).
func (q *Query) WithRaw(alias, sql string, bindings ...any) *Query {
	clause := &RawFromClause{Expression: sql, Bindings: bindings}
	clause.SetAlias(alias)
	return q.AddComponent("cte", clause, nil)
}

// WithValues adds an ad-hoc VALUES CTE (SqlKata.With(alias, columns, valuesCollection)).
func (q *Query) WithValues(alias string, columns []string, rows [][]any) *Query {
	if len(columns) == 0 || len(rows) == 0 {
		return q.setErr(fmt.Errorf("columns and valuesCollection cannot be null or empty"))
	}
	clause := &AdHocTableFromClause{
		Columns: append([]string(nil), columns...),
		Values:  nil,
	}
	clause.SetAlias(alias)
	for _, row := range rows {
		if len(row) != len(columns) {
			return q.setErr(fmt.Errorf("columns count should be equal to each Values count"))
		}
		clause.Values = append(clause.Values, row...)
	}
	return q.AddComponent("cte", clause, nil)
}
