package sqlkata

import "strings"

// WhereDatePart mirrors SqlKata.WhereDatePart.
func (b *BaseQuery) WhereDatePart(part, column, op string, value any) *Query {
	return b.AddComponent("where", &BasicDateCondition{
		BasicCondition: BasicCondition{
			ConditionBase: ConditionBase{IsOr: b.getOr(), IsNot: b.getNot()},
			Column:        column,
			Operator:      op,
			Value:         value,
		},
		Part: strings.ToLower(part),
	}, nil)
}

func (b *BaseQuery) WhereNotDatePart(part, column, op string, value any) *Query {
	return b.Not().WhereDatePart(part, column, op, value)
}

func (b *BaseQuery) OrWhereDatePart(part, column, op string, value any) *Query {
	return b.Or().WhereDatePart(part, column, op, value)
}

func (b *BaseQuery) OrWhereNotDatePart(part, column, op string, value any) *Query {
	return b.Or().Not().WhereDatePart(part, column, op, value)
}

func (b *BaseQuery) WhereDate(column, op string, value any) *Query {
	return b.WhereDatePart("date", column, op, value)
}

func (b *BaseQuery) WhereNotDate(column, op string, value any) *Query {
	return b.Not().WhereDate(column, op, value)
}

func (b *BaseQuery) OrWhereDate(column, op string, value any) *Query {
	return b.Or().WhereDate(column, op, value)
}

func (b *BaseQuery) OrWhereNotDate(column, op string, value any) *Query {
	return b.Or().Not().WhereDate(column, op, value)
}

func (b *BaseQuery) WhereTime(column, op string, value any) *Query {
	return b.WhereDatePart("time", column, op, value)
}

func (b *BaseQuery) WhereNotTime(column, op string, value any) *Query {
	return b.Not().WhereTime(column, op, value)
}

func (b *BaseQuery) OrWhereTime(column, op string, value any) *Query {
	return b.Or().WhereTime(column, op, value)
}

func (b *BaseQuery) OrWhereNotTime(column, op string, value any) *Query {
	return b.Or().Not().WhereTime(column, op, value)
}

// Equality helpers (SqlKata overloads with implicit "=").
func (b *BaseQuery) WhereDateEq(column string, value any) *Query {
	return b.WhereDate(column, "=", value)
}

func (b *BaseQuery) WhereTimeEq(column string, value any) *Query {
	return b.WhereTime(column, "=", value)
}

func (b *BaseQuery) WhereDatePartEq(part, column string, value any) *Query {
	return b.WhereDatePart(part, column, "=", value)
}
