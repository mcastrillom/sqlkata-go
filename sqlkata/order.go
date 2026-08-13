package sqlkata

// AbstractOrderBy is the clause family for ORDER BY (SqlKata.AbstractOrderBy).
type AbstractOrderBy interface {
	AbstractClause
}

// OrderBy is a column ORDER BY clause.
type OrderBy struct {
	ClauseBase
	Column    string
	Ascending bool
}

func (o *OrderBy) Clone() AbstractClause {
	n := &OrderBy{Column: o.Column, Ascending: o.Ascending}
	o.ClauseBase.copyMeta(&n.ClauseBase)
	return n
}

// RawOrderBy is a raw ORDER BY expression with bindings.
type RawOrderBy struct {
	ClauseBase
	Expression string
	Bindings   []any
}

func (r *RawOrderBy) Clone() AbstractClause {
	b := append([]any(nil), r.Bindings...)
	n := &RawOrderBy{Expression: r.Expression, Bindings: b}
	r.ClauseBase.copyMeta(&n.ClauseBase)
	return n
}

// OrderByRandom mirrors SqlKata.OrderByRandom (compiled as NEWID() on SQL Server).
type OrderByRandom struct {
	ClauseBase
}

func (o *OrderByRandom) Clone() AbstractClause {
	n := &OrderByRandom{}
	o.ClauseBase.copyMeta(&n.ClauseBase)
	return n
}
