package sqlkata

import (
	"fmt"
	"sort"
)

// Insert builds an INSERT from a map (alias of AsInsert).
func (q *Query) Insert(values map[string]any) *Query {
	return q.AsInsert(values, false)
}

// AsInsert mirrors SqlKata.AsInsert(IEnumerable KeyValuePair, returnId).
func (q *Query) AsInsert(values map[string]any, returnId bool) *Query {
	if len(values) == 0 {
		return q.setErr(fmt.Errorf("values argument cannot be null or empty"))
	}
	keys := make([]string, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	vals := make([]any, 0, len(keys))
	for _, k := range keys {
		vals = append(vals, values[k])
	}
	q.Method = "insert"
	_ = q.ClearComponent("insert", nil)
	return q.AddComponent("insert", &InsertClause{Columns: keys, Values: vals, ReturnId: returnId}, nil)
}

// AsInsertColumns mirrors AsInsert(columns, values).
func (q *Query) AsInsertColumns(columns []string, values []any) *Query {
	if len(columns) == 0 || len(values) == 0 {
		return q.setErr(fmt.Errorf("columns and values cannot be null or empty"))
	}
	if len(columns) != len(values) {
		return q.setErr(fmt.Errorf("columns and values count mismatch"))
	}
	q.Method = "insert"
	_ = q.ClearComponent("insert", nil)
	return q.AddComponent("insert", &InsertClause{
		Columns: append([]string(nil), columns...),
		Values:  append([]any(nil), values...),
	}, nil)
}

// InsertColumns is an alias for AsInsertColumns.
func (q *Query) InsertColumns(columns []string, values []any) *Query {
	return q.AsInsertColumns(columns, values)
}

// AsInsertRows inserts multiple value rows (SqlKata.AsInsert(columns, rowsValues)).
func (q *Query) AsInsertRows(columns []string, rows [][]any) *Query {
	if len(columns) == 0 || len(rows) == 0 {
		return q.setErr(fmt.Errorf("columns and rowsValues cannot be null or empty"))
	}
	q.Method = "insert"
	_ = q.ClearComponent("insert", nil)
	for _, row := range rows {
		if len(row) != len(columns) {
			return q.setErr(fmt.Errorf("columns count should be equal to each rowsValues entry count"))
		}
		_ = q.AddComponent("insert", &InsertClause{
			Columns: append([]string(nil), columns...),
			Values:  append([]any(nil), row...),
		}, nil)
	}
	return q
}

// AsInsertQuery inserts from a select subquery (SqlKata.AsInsert(columns, Query)).
func (q *Query) AsInsertQuery(columns []string, query *Query) *Query {
	if query == nil {
		return q.setErr(fmt.Errorf("query is nil"))
	}
	q.Method = "insert"
	_ = q.ClearComponent("insert", nil)
	return q.AddComponent("insert", &InsertQueryClause{
		Columns: append([]string(nil), columns...),
		Query:   query.Clone(),
	}, nil)
}
