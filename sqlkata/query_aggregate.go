package sqlkata

// AsAggregate mirrors Query.AsAggregate.
func (q *Query) AsAggregate(typ string, columns ...string) *Query {
	q.Method = "aggregate"
	_ = q.ClearComponent("aggregate", nil)
	cols := append([]string(nil), columns...)
	return q.AddComponent("aggregate", &AggregateClause{Type: typ, Columns: cols}, nil)
}

func (q *Query) AsCount(columns ...string) *Query {
	if len(columns) == 0 {
		columns = []string{"*"}
	}
	return q.AsAggregate("count", columns...)
}

func (q *Query) AsSum(column string) *Query { return q.AsAggregate("sum", column) }
func (q *Query) AsAvg(column string) *Query { return q.AsAggregate("avg", column) }
func (q *Query) AsMax(column string) *Query { return q.AsAggregate("max", column) }
func (q *Query) AsMin(column string) *Query { return q.AsAggregate("min", column) }

// SelectAggregate adds AGG(column) [with optional filter query] (SqlKata.SelectAggregate).
func (q *Query) SelectAggregate(aggregate, column string, filter ...*Query) *Query {
	q.Method = "select"
	var f *Query
	if len(filter) > 0 {
		f = filter[0]
	}
	return q.AddComponent("select", &AggregatedColumn{
		Aggregate: aggregate,
		Column:    &Column{Name: column},
		Filter:    f,
	}, nil)
}

// SelectAggregateFunc adds AGG with a filter callback.
func (q *Query) SelectAggregateFunc(aggregate, column string, filter func(*Query) *Query) *Query {
	if filter == nil {
		return q.SelectAggregate(aggregate, column)
	}
	return q.SelectAggregate(aggregate, column, filter(q.NewChild()))
}

func (q *Query) SelectSum(column string, filter ...func(*Query) *Query) *Query {
	return q.selectAgg("sum", column, filter...)
}
func (q *Query) SelectCount(column string, filter ...func(*Query) *Query) *Query {
	return q.selectAgg("count", column, filter...)
}
func (q *Query) SelectAvg(column string, filter ...func(*Query) *Query) *Query {
	return q.selectAgg("avg", column, filter...)
}
func (q *Query) SelectMin(column string, filter ...func(*Query) *Query) *Query {
	return q.selectAgg("min", column, filter...)
}
func (q *Query) SelectMax(column string, filter ...func(*Query) *Query) *Query {
	return q.selectAgg("max", column, filter...)
}

func (q *Query) selectAgg(name, column string, filter ...func(*Query) *Query) *Query {
	if len(filter) > 0 && filter[0] != nil {
		return q.SelectAggregateFunc(name, column, filter[0])
	}
	return q.SelectAggregate(name, column)
}

// SelectQuery adds a subquery select column (QueryColumn / Select(Query, alias)).
func (q *Query) SelectQuery(sub *Query, alias string) *Query {
	q.Method = "select"
	sub = sub.Clone()
	if alias != "" {
		sub.As(alias)
	}
	return q.AddComponent("select", &QueryColumn{Query: sub}, nil)
}

// SelectQueryFunc is Select(Func, alias).
func (q *Query) SelectQueryFunc(callback func(*Query) *Query, alias string) *Query {
	return q.SelectQuery(callback(q.NewChild()), alias)
}
