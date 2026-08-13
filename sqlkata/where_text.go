package sqlkata

// WhereTextEqual filters a long-text column for equality.
// Oracle avoids ORA-00932 (CLOB = :bind) via DBMS_LOB.COMPARE + TO_CLOB.
func (b *BaseQuery) WhereTextEqual(column string, value any) *Query {
	return b.AddComponent("where", &TextCondition{
		ConditionBase: ConditionBase{IsOr: b.getOr(), IsNot: b.getNot()},
		Column:        column,
		Value:         value,
		Equal:         true,
	}, nil)
}

// WhereTextNotEqual filters a long-text column for inequality.
func (b *BaseQuery) WhereTextNotEqual(column string, value any) *Query {
	return b.AddComponent("where", &TextCondition{
		ConditionBase: ConditionBase{IsOr: b.getOr(), IsNot: b.getNot()},
		Column:        column,
		Value:         value,
		Equal:         false,
	}, nil)
}

// OrWhereTextEqual is Or().WhereTextEqual(...).
func (b *BaseQuery) OrWhereTextEqual(column string, value any) *Query {
	return b.Or().WhereTextEqual(column, value)
}

// OrWhereTextNotEqual is Or().WhereTextNotEqual(...).
func (b *BaseQuery) OrWhereTextNotEqual(column string, value any) *Query {
	return b.Or().WhereTextNotEqual(column, value)
}

// SelectAsText selects a long-text column as a portable string expression.
// alias optional; maxLen optional (Oracle SUBSTR length, default 4000).
func (q *Query) SelectAsText(column string, alias string, maxLen ...int) *Query {
	q.Method = "select"
	ml := 0
	if len(maxLen) > 0 {
		ml = maxLen[0]
	}
	return q.AddComponent("select", &TextColumn{
		Name:   column,
		Alias:  alias,
		MaxLen: ml,
	}, nil)
}
