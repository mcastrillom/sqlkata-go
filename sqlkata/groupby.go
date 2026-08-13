package sqlkata

// GroupBy mirrors Query.GroupBy.
func (q *Query) GroupBy(columns ...string) *Query {
	for _, col := range columns {
		_ = q.AddComponent("group", &Column{Name: col}, nil)
	}
	return q
}

// GroupByRaw mirrors Query.GroupByRaw.
func (q *Query) GroupByRaw(expression string, bindings ...any) *Query {
	return q.AddComponent("group", &RawColumn{Expression: expression, Bindings: bindings}, nil)
}
