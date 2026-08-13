package sqlkata

// AbstractCombine is the UNION/EXCEPT/INTERSECT clause family.
type AbstractCombine interface {
	AbstractClause
}

// Combine is UNION / EXCEPT / INTERSECT with another query.
type Combine struct {
	ClauseBase
	Query     *Query
	Operation string
	All       bool
}

func (c *Combine) Clone() AbstractClause {
	var q *Query
	if c.Query != nil {
		q = c.Query.Clone()
	}
	n := &Combine{Query: q, Operation: c.Operation, All: c.All}
	c.ClauseBase.copyMeta(&n.ClauseBase)
	return n
}

// RawCombine is a raw combine expression.
type RawCombine struct {
	ClauseBase
	Expression string
	Bindings   []any
}

func (c *RawCombine) Clone() AbstractClause {
	n := &RawCombine{
		Expression: c.Expression,
		Bindings:   append([]any(nil), c.Bindings...),
	}
	c.ClauseBase.copyMeta(&n.ClauseBase)
	return n
}
