package sqlkata

// AbstractInsertClause is the insert clause family.
type AbstractInsertClause interface {
	AbstractClause
}

// InsertClause holds column/value rows for INSERT.
type InsertClause struct {
	ClauseBase
	Columns  []string
	Values   []any
	ReturnId bool
}

func (c *InsertClause) Clone() AbstractClause {
	n := &InsertClause{
		Columns:  append([]string(nil), c.Columns...),
		Values:   append([]any(nil), c.Values...),
		ReturnId: c.ReturnId,
	}
	c.ClauseBase.copyMeta(&n.ClauseBase)
	return n
}

// InsertQueryClause is INSERT … SELECT.
type InsertQueryClause struct {
	ClauseBase
	Columns []string
	Query   *Query
}

func (c *InsertQueryClause) Clone() AbstractClause {
	var q *Query
	if c.Query != nil {
		q = c.Query.Clone()
	}
	n := &InsertQueryClause{
		Columns: append([]string(nil), c.Columns...),
		Query:   q,
	}
	c.ClauseBase.copyMeta(&n.ClauseBase)
	return n
}

// IncrementClause is UPDATE col = col +/- value (SqlKata.IncrementClause).
type IncrementClause struct {
	ClauseBase
	Column string
	Value  int
}

func (c *IncrementClause) Clone() AbstractClause {
	n := &IncrementClause{Column: c.Column, Value: c.Value}
	c.ClauseBase.copyMeta(&n.ClauseBase)
	return n
}

// NewIncrementClause defaults Value to 1 like C#.
func NewIncrementClause(column string, value ...int) *IncrementClause {
	v := 1
	if len(value) > 0 {
		v = value[0]
	}
	return &IncrementClause{Column: column, Value: v}
}
