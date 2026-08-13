package sqlkata

import "fmt"

// WhereQuery is column OP (subquery) — SqlKata.Where(column, op, Query).
func (b *BaseQuery) WhereQuery(column, op string, query *Query) *Query {
	return b.AddComponent("where", &QueryCondition{
		ConditionBase: ConditionBase{IsOr: b.getOr(), IsNot: b.getNot()},
		Column:        column,
		Operator:      op,
		Query:         query,
	}, nil)
}

// WhereQueryFunc is column OP (callback subquery).
func (b *BaseQuery) WhereQueryFunc(column, op string, callback func(*Query) *Query) *Query {
	return b.WhereQuery(column, op, callback(b.NewChild()))
}

func (b *BaseQuery) OrWhereQuery(column, op string, query *Query) *Query {
	return b.Or().WhereQuery(column, op, query)
}

func (b *BaseQuery) OrWhereQueryFunc(column, op string, callback func(*Query) *Query) *Query {
	return b.Or().WhereQueryFunc(column, op, callback)
}

// WhereSub is (subquery) OP value — SqlKata.WhereSub.
func (b *BaseQuery) WhereSub(query *Query, op string, value any) *Query {
	return b.AddComponent("where", &SubQueryCondition{
		ConditionBase: ConditionBase{IsOr: b.getOr(), IsNot: b.getNot()},
		Query:         query,
		Operator:      op,
		Value:         value,
	}, nil)
}

func (b *BaseQuery) WhereSubEq(query *Query, value any) *Query {
	return b.WhereSub(query, "=", value)
}

func (b *BaseQuery) OrWhereSub(query *Query, op string, value any) *Query {
	return b.Or().WhereSub(query, op, value)
}

func (b *BaseQuery) OrWhereSubEq(query *Query, value any) *Query {
	return b.Or().WhereSubEq(query, value)
}

// WhereInFunc is column IN (callback) — SqlKata.WhereIn(column, Func).
func (b *BaseQuery) WhereInFunc(column string, callback func(*Query) *Query) *Query {
	child := NewQuery()
	child.Parent = b.owner
	return b.WhereInQuery(column, callback(child))
}

func (b *BaseQuery) OrWhereInQuery(column string, query *Query) *Query {
	return b.Or().WhereInQuery(column, query)
}

func (b *BaseQuery) OrWhereInFunc(column string, callback func(*Query) *Query) *Query {
	return b.Or().WhereInFunc(column, callback)
}

func (b *BaseQuery) WhereNotInQuery(column string, query *Query) *Query {
	return b.Not().WhereInQuery(column, query)
}

func (b *BaseQuery) WhereNotInFunc(column string, callback func(*Query) *Query) *Query {
	return b.Not().WhereInFunc(column, callback)
}

func (b *BaseQuery) OrWhereNotInQuery(column string, query *Query) *Query {
	return b.Or().Not().WhereInQuery(column, query)
}

func (b *BaseQuery) OrWhereNotInFunc(column string, callback func(*Query) *Query) *Query {
	return b.Or().Not().WhereInFunc(column, callback)
}

// WhereExistsFunc mirrors WhereExists(Func).
func (b *BaseQuery) WhereExistsFunc(callback func(*Query) *Query) *Query {
	child := NewQuery()
	child.Parent = b.owner
	return b.WhereExists(callback(child))
}

func (b *BaseQuery) WhereNotExistsFunc(callback func(*Query) *Query) *Query {
	return b.Not().WhereExistsFunc(callback)
}

func (b *BaseQuery) OrWhereExists(query *Query) *Query {
	return b.Or().WhereExists(query)
}

func (b *BaseQuery) OrWhereExistsFunc(callback func(*Query) *Query) *Query {
	return b.Or().WhereExistsFunc(callback)
}

func (b *BaseQuery) OrWhereNotExists(query *Query) *Query {
	return b.Or().Not().WhereExists(query)
}

func (b *BaseQuery) OrWhereNotExistsFunc(callback func(*Query) *Query) *Query {
	return b.Or().Not().WhereExistsFunc(callback)
}

// ensure From is present for Exists (SqlKata validation).
func mustHaveFrom(query *Query) error {
	if query == nil || !query.HasComponent("from", nil) {
		return fmt.Errorf("'FromClause' cannot be empty if used inside a 'WhereExists' condition")
	}
	return nil
}
