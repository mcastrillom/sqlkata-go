package sqlkata

// Where adds column OP value (SqlKata.Where).
func (b *BaseQuery) Where(column, op string, value any) *Query {
	if value == nil {
		return b.Not(op != "=").WhereNull(column)
	}
	if bv, ok := value.(bool); ok {
		if op != "=" {
			b.Not()
		}
		if bv {
			return b.WhereTrue(column)
		}
		return b.WhereFalse(column)
	}
	return b.AddComponent("where", &BasicCondition{
		ConditionBase: ConditionBase{IsOr: b.getOr(), IsNot: b.getNot()},
		Column:        column,
		Operator:      op,
		Value:         value,
	}, nil)
}

// WhereEq is Where(column, "=", value).
func (b *BaseQuery) WhereEq(column string, value any) *Query {
	return b.Where(column, "=", value)
}

func (b *BaseQuery) WhereNot(column, op string, value any) *Query {
	return b.Not().Where(column, op, value)
}

func (b *BaseQuery) WhereNotEq(column string, value any) *Query {
	return b.WhereNot(column, "=", value)
}

func (b *BaseQuery) OrWhere(column, op string, value any) *Query {
	return b.Or().Where(column, op, value)
}

func (b *BaseQuery) OrWhereEq(column string, value any) *Query {
	return b.OrWhere(column, "=", value)
}

func (b *BaseQuery) OrWhereNot(column, op string, value any) *Query {
	return b.Or().Not().Where(column, op, value)
}

func (b *BaseQuery) OrWhereNotEq(column string, value any) *Query {
	return b.OrWhereNot(column, "=", value)
}

func (b *BaseQuery) WhereRaw(sql string, bindings ...any) *Query {
	return b.AddComponent("where", &RawCondition{
		ConditionBase: ConditionBase{IsOr: b.getOr(), IsNot: b.getNot()},
		Expression:    sql,
		Bindings:      bindings,
	}, nil)
}

func (b *BaseQuery) OrWhereRaw(sql string, bindings ...any) *Query {
	return b.Or().WhereRaw(sql, bindings...)
}

func (b *BaseQuery) WhereNull(column string) *Query {
	return b.AddComponent("where", &NullCondition{
		ConditionBase: ConditionBase{IsOr: b.getOr(), IsNot: b.getNot()},
		Column:        column,
	}, nil)
}

func (b *BaseQuery) WhereNotNull(column string) *Query {
	return b.Not().WhereNull(column)
}

func (b *BaseQuery) OrWhereNull(column string) *Query {
	return b.Or().WhereNull(column)
}

func (b *BaseQuery) OrWhereNotNull(column string) *Query {
	return b.Or().Not().WhereNull(column)
}

func (b *BaseQuery) WhereTrue(column string) *Query {
	return b.AddComponent("where", &BooleanCondition{
		ConditionBase: ConditionBase{IsOr: b.getOr(), IsNot: b.getNot()},
		Column:        column,
		Value:         true,
	}, nil)
}

func (b *BaseQuery) OrWhereTrue(column string) *Query {
	return b.Or().WhereTrue(column)
}

func (b *BaseQuery) WhereFalse(column string) *Query {
	return b.AddComponent("where", &BooleanCondition{
		ConditionBase: ConditionBase{IsOr: b.getOr(), IsNot: b.getNot()},
		Column:        column,
		Value:         false,
	}, nil)
}

func (b *BaseQuery) OrWhereFalse(column string) *Query {
	return b.Or().WhereFalse(column)
}

func (b *BaseQuery) WhereColumns(first, op, second string) *Query {
	return b.AddComponent("where", &TwoColumnsCondition{
		ConditionBase: ConditionBase{IsOr: b.getOr(), IsNot: b.getNot()},
		First:         first,
		Operator:      op,
		Second:        second,
	}, nil)
}

func (b *BaseQuery) OrWhereColumns(first, op, second string) *Query {
	return b.Or().WhereColumns(first, op, second)
}

func (b *BaseQuery) WhereBetween(column string, lower, higher any) *Query {
	return b.AddComponent("where", &BetweenCondition{
		ConditionBase: ConditionBase{IsOr: b.getOr(), IsNot: b.getNot()},
		Column:        column,
		Lower:         lower,
		Higher:        higher,
	}, nil)
}

func (b *BaseQuery) OrWhereBetween(column string, lower, higher any) *Query {
	return b.Or().WhereBetween(column, lower, higher)
}

func (b *BaseQuery) WhereNotBetween(column string, lower, higher any) *Query {
	return b.Not().WhereBetween(column, lower, higher)
}

func (b *BaseQuery) OrWhereNotBetween(column string, lower, higher any) *Query {
	return b.Or().Not().WhereBetween(column, lower, higher)
}

func (b *BaseQuery) WhereIn(column string, values ...any) *Query {
	return b.AddComponent("where", &InCondition{
		ConditionBase: ConditionBase{IsOr: b.getOr(), IsNot: b.getNot()},
		Column:        column,
		Values:        values,
	}, nil)
}

func (b *BaseQuery) OrWhereIn(column string, values ...any) *Query {
	return b.Or().WhereIn(column, values...)
}

func (b *BaseQuery) WhereNotIn(column string, values ...any) *Query {
	return b.Not().WhereIn(column, values...)
}

func (b *BaseQuery) OrWhereNotIn(column string, values ...any) *Query {
	return b.Or().Not().WhereIn(column, values...)
}

func (b *BaseQuery) WhereInQuery(column string, query *Query) *Query {
	return b.AddComponent("where", &InQueryCondition{
		ConditionBase: ConditionBase{IsOr: b.getOr(), IsNot: b.getNot()},
		Column:        column,
		Query:         query,
	}, nil)
}

func (b *BaseQuery) WhereExists(query *Query) *Query {
	if err := mustHaveFrom(query); err != nil {
		return b.owner.setErr(err)
	}
	return b.AddComponent("where", &ExistsCondition{
		ConditionBase: ConditionBase{IsOr: b.getOr(), IsNot: b.getNot()},
		Query:         query,
	}, nil)
}

func (b *BaseQuery) WhereNotExists(query *Query) *Query {
	return b.Not().WhereExists(query)
}

// WhereNested applies a nested WHERE group (SqlKata.Where(Func)).
func (b *BaseQuery) WhereNested(callback func(*Query) *Query) *Query {
	child := b.NewChild()
	q := callback(child)
	hasWhere := false
	for _, c := range q.Clauses {
		if c.GetComponent() == "where" {
			hasWhere = true
			break
		}
	}
	if !hasWhere {
		return b.owner
	}
	return b.AddComponent("where", &NestedCondition{
		ConditionBase: ConditionBase{IsOr: b.getOr(), IsNot: b.getNot()},
		Nested:        q,
	}, nil)
}

func (b *BaseQuery) WhereNotNested(callback func(*Query) *Query) *Query {
	return b.Not().WhereNested(callback)
}

func (b *BaseQuery) OrWhereNested(callback func(*Query) *Query) *Query {
	return b.Or().WhereNested(callback)
}

func (b *BaseQuery) OrWhereNotNested(callback func(*Query) *Query) *Query {
	return b.Not().Or().WhereNested(callback)
}

// WhereMap applies equality constraints from a map (SqlKata.Where(IEnumerable KVP)).
func (b *BaseQuery) WhereMap(values map[string]any) *Query {
	orFlag := b.getOr()
	notFlag := b.getNot()
	q := b.owner
	for k, v := range values {
		if orFlag {
			q = q.Or()
		}
		q = q.Not(notFlag).WhereEq(k, v)
	}
	return q
}
