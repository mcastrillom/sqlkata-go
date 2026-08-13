package sqlkata

import (
	"fmt"
	"sort"
)

// AsUpdate sets Method=update with an InsertClause of columns/values (SqlKata.AsUpdate).
func (q *Query) AsUpdate(values map[string]any) *Query {
	if len(values) == 0 {
		return q.setErr(fmt.Errorf("values cannot be null or empty"))
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
	return q.AsUpdateColumns(keys, vals)
}

// AsUpdateColumns mirrors AsUpdate(IEnumerable columns, IEnumerable values).
func (q *Query) AsUpdateColumns(columns []string, values []any) *Query {
	if len(columns) == 0 || len(values) == 0 {
		return q.setErr(fmt.Errorf("columns and values cannot be null or empty"))
	}
	if len(columns) != len(values) {
		return q.setErr(fmt.Errorf("columns count should be equal to values count"))
	}
	q.Method = "update"
	_ = q.ClearComponent("update", nil)
	return q.AddComponent("update", &InsertClause{
		Columns: append([]string(nil), columns...),
		Values:  append([]any(nil), values...),
	}, nil)
}

// AsIncrement mirrors Query.AsIncrement.
func (q *Query) AsIncrement(column string, value ...int) *Query {
	q.Method = "update"
	inc := NewIncrementClause(column, value...)
	return q.AddOrReplaceComponent("update", inc, nil)
}

// AsDecrement mirrors Query.AsDecrement.
func (q *Query) AsDecrement(column string, value ...int) *Query {
	v := 1
	if len(value) > 0 {
		v = value[0]
	}
	return q.AsIncrement(column, -v)
}

// Increment is an alias for AsIncrement.
func (q *Query) Increment(column string, value ...int) *Query {
	return q.AsIncrement(column, value...)
}
