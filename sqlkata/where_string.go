package sqlkata

func (b *BaseQuery) addStringWhere(op, column string, value any, caseSensitive bool, escapeCharacter string) *Query {
	cond := &BasicStringCondition{
		BasicCondition: BasicCondition{
			ConditionBase: ConditionBase{IsOr: b.getOr(), IsNot: b.getNot()},
			Column:        column,
			Operator:      op,
			Value:         value,
		},
		CaseSensitive: caseSensitive,
	}
	if escapeCharacter != "" {
		if err := cond.SetEscapeCharacter(escapeCharacter); err != nil {
			return b.owner.setErr(err)
		}
	}
	return b.AddComponent("where", cond, nil)
}

func (b *BaseQuery) WhereLike(column string, value any, caseSensitive bool, escapeCharacter string) *Query {
	return b.addStringWhere("like", column, value, caseSensitive, escapeCharacter)
}

func (b *BaseQuery) WhereNotLike(column string, value any, caseSensitive bool, escapeCharacter string) *Query {
	return b.Not().WhereLike(column, value, caseSensitive, escapeCharacter)
}

func (b *BaseQuery) OrWhereLike(column string, value any, caseSensitive bool, escapeCharacter string) *Query {
	return b.Or().WhereLike(column, value, caseSensitive, escapeCharacter)
}

func (b *BaseQuery) OrWhereNotLike(column string, value any, caseSensitive bool, escapeCharacter string) *Query {
	return b.Or().Not().WhereLike(column, value, caseSensitive, escapeCharacter)
}

func (b *BaseQuery) WhereStarts(column string, value any, caseSensitive bool, escapeCharacter string) *Query {
	return b.addStringWhere("starts", column, value, caseSensitive, escapeCharacter)
}

func (b *BaseQuery) WhereNotStarts(column string, value any, caseSensitive bool, escapeCharacter string) *Query {
	return b.Not().WhereStarts(column, value, caseSensitive, escapeCharacter)
}

func (b *BaseQuery) OrWhereStarts(column string, value any, caseSensitive bool, escapeCharacter string) *Query {
	return b.Or().WhereStarts(column, value, caseSensitive, escapeCharacter)
}

func (b *BaseQuery) OrWhereNotStarts(column string, value any, caseSensitive bool, escapeCharacter string) *Query {
	return b.Or().Not().WhereStarts(column, value, caseSensitive, escapeCharacter)
}

func (b *BaseQuery) WhereEnds(column string, value any, caseSensitive bool, escapeCharacter string) *Query {
	return b.addStringWhere("ends", column, value, caseSensitive, escapeCharacter)
}

func (b *BaseQuery) WhereNotEnds(column string, value any, caseSensitive bool, escapeCharacter string) *Query {
	return b.Not().WhereEnds(column, value, caseSensitive, escapeCharacter)
}

func (b *BaseQuery) OrWhereEnds(column string, value any, caseSensitive bool, escapeCharacter string) *Query {
	return b.Or().WhereEnds(column, value, caseSensitive, escapeCharacter)
}

func (b *BaseQuery) OrWhereNotEnds(column string, value any, caseSensitive bool, escapeCharacter string) *Query {
	return b.Or().Not().WhereEnds(column, value, caseSensitive, escapeCharacter)
}

func (b *BaseQuery) WhereContains(column string, value any, caseSensitive bool, escapeCharacter string) *Query {
	return b.addStringWhere("contains", column, value, caseSensitive, escapeCharacter)
}

func (b *BaseQuery) WhereNotContains(column string, value any, caseSensitive bool, escapeCharacter string) *Query {
	return b.Not().WhereContains(column, value, caseSensitive, escapeCharacter)
}

func (b *BaseQuery) OrWhereContains(column string, value any, caseSensitive bool, escapeCharacter string) *Query {
	return b.Or().WhereContains(column, value, caseSensitive, escapeCharacter)
}

func (b *BaseQuery) OrWhereNotContains(column string, value any, caseSensitive bool, escapeCharacter string) *Query {
	return b.Or().Not().WhereContains(column, value, caseSensitive, escapeCharacter)
}
