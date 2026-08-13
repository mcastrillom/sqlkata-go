package sqlkata

import (
	"strings"
)

// AbstractFrom is the Go equivalent of SqlKata.AbstractFrom.
type AbstractFrom struct {
	ClauseBase
	alias *string
}

func (f *AbstractFrom) GetAliasPtr() *string  { return f.alias }
func (f *AbstractFrom) SetAliasPtr(a *string) { f.alias = a }

// SetAlias stores a non-empty alias on the FROM/CTE clause.
func (f *AbstractFrom) SetAlias(name string) {
	if name == "" {
		f.alias = nil
		return
	}
	f.alias = strPtr(name)
}

func (f *AbstractFrom) storedAlias() string {
	if f.alias != nil {
		return *f.alias
	}
	return ""
}

// FromClause is a simple table FROM ("table" or "table as alias").
type FromClause struct {
	AbstractFrom
	Table string
}

func (f *FromClause) Alias() string {
	idx := strings.Index(strings.ToLower(f.Table), " as ")
	if idx >= 0 {
		parts := strings.Fields(f.Table)
		if len(parts) >= 3 {
			return parts[2]
		}
	}
	if f.alias != nil {
		return *f.alias
	}
	return f.Table
}

func (f *FromClause) Clone() AbstractClause {
	n := &FromClause{Table: f.Table}
	f.AbstractFrom.ClauseBase.copyMeta(&n.ClauseBase)
	n.alias = f.alias
	return n
}

// QueryFromClause is FROM (subquery).
type QueryFromClause struct {
	AbstractFrom
	Query *Query
}

func (q *QueryFromClause) Alias() string {
	if q.alias != nil && *q.alias != "" {
		return *q.alias
	}
	if q.Query != nil && q.Query.QueryAlias != "" {
		return q.Query.QueryAlias
	}
	return ""
}

func (q *QueryFromClause) Clone() AbstractClause {
	var cloned *Query
	if q.Query != nil {
		cloned = q.Query.Clone()
	}
	n := &QueryFromClause{Query: cloned}
	q.AbstractFrom.ClauseBase.copyMeta(&n.ClauseBase)
	n.alias = q.alias
	return n
}

// RawFromClause is FROM raw SQL with bindings.
type RawFromClause struct {
	AbstractFrom
	Expression string
	Bindings   []any
}

func (r *RawFromClause) Clone() AbstractClause {
	b := append([]any(nil), r.Bindings...)
	n := &RawFromClause{Expression: r.Expression, Bindings: b}
	r.AbstractFrom.ClauseBase.copyMeta(&n.ClauseBase)
	n.alias = r.alias
	return n
}

func (r *RawFromClause) Alias() string { return r.storedAlias() }

// AdHocTableFromClause is an ad-hoc VALUES table / CTE (SqlKata.AdHocTableFromClause).
type AdHocTableFromClause struct {
	AbstractFrom
	Columns []string
	Values  []any
}

func (a *AdHocTableFromClause) Alias() string { return a.storedAlias() }

func (a *AdHocTableFromClause) Clone() AbstractClause {
	n := &AdHocTableFromClause{
		Columns: append([]string(nil), a.Columns...),
		Values:  append([]any(nil), a.Values...),
	}
	a.AbstractFrom.ClauseBase.copyMeta(&n.ClauseBase)
	n.alias = a.alias
	return n
}
