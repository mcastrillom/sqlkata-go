package sqlkata

// AggregateClause is COUNT/MAX/MIN/… over columns (SqlKata.AggregateClause).
type AggregateClause struct {
	ClauseBase
	Columns []string
	Type    string
}

func (c *AggregateClause) Clone() AbstractClause {
	n := &AggregateClause{
		Columns: append([]string(nil), c.Columns...),
		Type:    c.Type,
	}
	c.ClauseBase.copyMeta(&n.ClauseBase)
	return n
}
