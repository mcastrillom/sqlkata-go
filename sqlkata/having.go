package sqlkata

// Having adds column OP value to the having component (SqlKata.Having).
func (q *Query) Having(column, op string, value any) *Query {
	if value == nil {
		return q.Not(op != "=").HavingNull(column)
	}
	return q.AddComponent("having", &BasicCondition{
		ConditionBase: ConditionBase{IsOr: q.getOr(), IsNot: q.getNot()},
		Column:        column,
		Operator:      op,
		Value:         value,
	}, nil)
}

func (q *Query) HavingEq(column string, value any) *Query {
	return q.Having(column, "=", value)
}

func (q *Query) HavingNot(column, op string, value any) *Query {
	return q.Not().Having(column, op, value)
}

func (q *Query) OrHaving(column, op string, value any) *Query {
	return q.Or().Having(column, op, value)
}

func (q *Query) OrHavingEq(column string, value any) *Query {
	return q.OrHaving(column, "=", value)
}

func (q *Query) OrHavingNot(column, op string, value any) *Query {
	return q.Or().Not().Having(column, op, value)
}

func (q *Query) HavingRaw(sql string, bindings ...any) *Query {
	return q.AddComponent("having", &RawCondition{
		ConditionBase: ConditionBase{IsOr: q.getOr(), IsNot: q.getNot()},
		Expression:    sql,
		Bindings:      bindings,
	}, nil)
}

func (q *Query) OrHavingRaw(sql string, bindings ...any) *Query {
	return q.Or().HavingRaw(sql, bindings...)
}

func (q *Query) HavingNull(column string) *Query {
	return q.AddComponent("having", &NullCondition{
		ConditionBase: ConditionBase{IsOr: q.getOr(), IsNot: q.getNot()},
		Column:        column,
	}, nil)
}

func (q *Query) HavingNotNull(column string) *Query {
	return q.Not().HavingNull(column)
}

func (q *Query) OrHavingNull(column string) *Query {
	return q.Or().HavingNull(column)
}

func (q *Query) HavingTrue(column string) *Query {
	return q.AddComponent("having", &BooleanCondition{
		ConditionBase: ConditionBase{IsOr: q.getOr(), IsNot: q.getNot()},
		Column:        column,
		Value:         true,
	}, nil)
}

func (q *Query) HavingFalse(column string) *Query {
	return q.AddComponent("having", &BooleanCondition{
		ConditionBase: ConditionBase{IsOr: q.getOr(), IsNot: q.getNot()},
		Column:        column,
		Value:         false,
	}, nil)
}

func (q *Query) HavingColumns(first, op, second string) *Query {
	return q.AddComponent("having", &TwoColumnsCondition{
		ConditionBase: ConditionBase{IsOr: q.getOr(), IsNot: q.getNot()},
		First:         first,
		Operator:      op,
		Second:        second,
	}, nil)
}

func (q *Query) OrHavingColumns(first, op, second string) *Query {
	return q.Or().HavingColumns(first, op, second)
}

func (q *Query) HavingBetween(column string, lower, higher any) *Query {
	return q.AddComponent("having", &BetweenCondition{
		ConditionBase: ConditionBase{IsOr: q.getOr(), IsNot: q.getNot()},
		Column:        column,
		Lower:         lower,
		Higher:        higher,
	}, nil)
}

func (q *Query) OrHavingBetween(column string, lower, higher any) *Query {
	return q.Or().HavingBetween(column, lower, higher)
}

func (q *Query) HavingNotBetween(column string, lower, higher any) *Query {
	return q.Not().HavingBetween(column, lower, higher)
}

func (q *Query) HavingIn(column string, values ...any) *Query {
	return q.AddComponent("having", &InCondition{
		ConditionBase: ConditionBase{IsOr: q.getOr(), IsNot: q.getNot()},
		Column:        column,
		Values:        values,
	}, nil)
}

func (q *Query) OrHavingIn(column string, values ...any) *Query {
	return q.Or().HavingIn(column, values...)
}

func (q *Query) HavingNotIn(column string, values ...any) *Query {
	return q.Not().HavingIn(column, values...)
}

func (q *Query) HavingInQuery(column string, query *Query) *Query {
	return q.AddComponent("having", &InQueryCondition{
		ConditionBase: ConditionBase{IsOr: q.getOr(), IsNot: q.getNot()},
		Column:        column,
		Query:         query,
	}, nil)
}

func (q *Query) HavingExists(query *Query) *Query {
	return q.AddComponent("having", &ExistsCondition{
		ConditionBase: ConditionBase{IsOr: q.getOr(), IsNot: q.getNot()},
		Query:         query,
	}, nil)
}

func (q *Query) HavingNotExists(query *Query) *Query {
	return q.Not().HavingExists(query)
}

func (q *Query) HavingLike(column string, value any, caseSensitive bool, escapeCharacter string) *Query {
	cond := &BasicStringCondition{
		BasicCondition: BasicCondition{
			ConditionBase: ConditionBase{IsOr: q.getOr(), IsNot: q.getNot()},
			Column:        column,
			Operator:      "like",
			Value:         value,
		},
		CaseSensitive: caseSensitive,
	}
	if escapeCharacter != "" {
		if err := cond.SetEscapeCharacter(escapeCharacter); err != nil {
			return q.setErr(err)
		}
	}
	return q.AddComponent("having", cond, nil)
}

func (q *Query) HavingStarts(column string, value any, caseSensitive bool, escapeCharacter string) *Query {
	cond := &BasicStringCondition{
		BasicCondition: BasicCondition{
			ConditionBase: ConditionBase{IsOr: q.getOr(), IsNot: q.getNot()},
			Column:        column,
			Operator:      "starts",
			Value:         value,
		},
		CaseSensitive: caseSensitive,
	}
	if escapeCharacter != "" {
		if err := cond.SetEscapeCharacter(escapeCharacter); err != nil {
			return q.setErr(err)
		}
	}
	return q.AddComponent("having", cond, nil)
}

func (q *Query) HavingContains(column string, value any, caseSensitive bool, escapeCharacter string) *Query {
	cond := &BasicStringCondition{
		BasicCondition: BasicCondition{
			ConditionBase: ConditionBase{IsOr: q.getOr(), IsNot: q.getNot()},
			Column:        column,
			Operator:      "contains",
			Value:         value,
		},
		CaseSensitive: caseSensitive,
	}
	if escapeCharacter != "" {
		if err := cond.SetEscapeCharacter(escapeCharacter); err != nil {
			return q.setErr(err)
		}
	}
	return q.AddComponent("having", cond, nil)
}

// HavingNested applies a nested HAVING group (SqlKata.Having(Func)).
func (q *Query) HavingNested(callback func(*Query) *Query) *Query {
	child := q.NewChild()
	nested := callback(child)
	return q.AddComponent("having", &NestedCondition{
		ConditionBase: ConditionBase{IsOr: q.getOr(), IsNot: q.getNot()},
		Nested:        nested,
	}, nil)
}

func (q *Query) OrHavingNested(callback func(*Query) *Query) *Query {
	return q.Or().HavingNested(callback)
}

func (q *Query) HavingNotNested(callback func(*Query) *Query) *Query {
	return q.Not().HavingNested(callback)
}
