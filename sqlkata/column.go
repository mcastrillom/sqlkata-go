package sqlkata

// AbstractColumn is the clause type family for SELECT columns (SqlKata.AbstractColumn).
type AbstractColumn interface {
	AbstractClause
}

// Column is a named SELECT / GROUP column (SqlKata.Column).
type Column struct {
	ClauseBase
	Name string
}

func (c *Column) Clone() AbstractClause {
	n := &Column{Name: c.Name}
	c.ClauseBase.copyMeta(&n.ClauseBase)
	return n
}

// RawColumn is a raw SELECT expression with bindings (SqlKata.RawColumn).
type RawColumn struct {
	ClauseBase
	Expression string
	Bindings   []any
}

func (r *RawColumn) Clone() AbstractClause {
	b := append([]any(nil), r.Bindings...)
	n := &RawColumn{Expression: r.Expression, Bindings: b}
	r.ClauseBase.copyMeta(&n.ClauseBase)
	return n
}

// QueryColumn is a SELECT subquery column (SqlKata.QueryColumn).
type QueryColumn struct {
	ClauseBase
	Query *Query
}

func (c *QueryColumn) Clone() AbstractClause {
	var q *Query
	if c.Query != nil {
		q = c.Query.Clone()
	}
	n := &QueryColumn{Query: q}
	c.ClauseBase.copyMeta(&n.ClauseBase)
	return n
}

// AggregatedColumn is AGG(column) [FILTER ...] (SqlKata.AggregatedColumn).
type AggregatedColumn struct {
	ClauseBase
	Filter    *Query
	Aggregate string
	Column    AbstractColumn
}

func (c *AggregatedColumn) Clone() AbstractClause {
	var filter *Query
	if c.Filter != nil {
		filter = c.Filter.Clone()
	}
	var col AbstractColumn
	if c.Column != nil {
		col = c.Column.Clone().(AbstractColumn)
	}
	n := &AggregatedColumn{Filter: filter, Aggregate: c.Aggregate, Column: col}
	c.ClauseBase.copyMeta(&n.ClauseBase)
	return n
}

// TextColumn projects a long-text column as a portable string expression.
// Oracle: DBMS_LOB.SUBSTR(col, MaxLen, 1); others: wrap(col).
type TextColumn struct {
	ClauseBase
	Name   string
	Alias  string
	MaxLen int // Oracle SUBSTR length; default 4000 when unset
}

func (c *TextColumn) Clone() AbstractClause {
	n := &TextColumn{Name: c.Name, Alias: c.Alias, MaxLen: c.MaxLen}
	c.ClauseBase.copyMeta(&n.ClauseBase)
	return n
}
