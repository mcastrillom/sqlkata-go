package sqlkata

// LimitClause mirrors SqlKata.LimitClause (positive limits only).
type LimitClause struct {
	ClauseBase
	Limit int
}

func (l *LimitClause) Clone() AbstractClause {
	n := &LimitClause{Limit: l.Limit}
	l.ClauseBase.copyMeta(&n.ClauseBase)
	return n
}

// OffsetClause mirrors SqlKata.OffsetClause.
type OffsetClause struct {
	ClauseBase
	Offset int64
}

func (o *OffsetClause) Clone() AbstractClause {
	n := &OffsetClause{Offset: o.Offset}
	o.ClauseBase.copyMeta(&n.ClauseBase)
	return n
}
