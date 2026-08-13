package sqlkata

import (
	"errors"
	"fmt"
)

// BaseQuery is the Go equivalent of SqlKata.BaseQuery<Query>.
// Go does not support the C# CRTP constraint BaseQuery<Q> where Q : BaseQuery<Q> without
// an illegal recursive interface; this implementation is concrete for *Query. When you
// migrate Join, mirror these methods on a joinBase struct or duplicate the pattern.
type BaseQuery struct {
	owner *Query
	AbstractQuery
	Clauses     []AbstractClause
	orFlag      bool
	notFlag     bool
	EngineScope *string

	// index groups Clauses by component name so lookups skip the linear scan.
	// It is only populated once a query holds clauseIndexThreshold clauses, and
	// is maintained by the mutating methods so that reads stay side-effect free
	// (compiling a query concurrently must not write to it).
	index    map[string][]AbstractClause
	indexLen int
}

// clauseIndexThreshold is the clause count at which the component index starts
// paying for itself. Below it a scan over the slice beats hashing a string, so
// small queries never allocate the map.
const clauseIndexThreshold = 16

func (b *BaseQuery) init(owner *Query) {
	b.owner = owner
}

// buildIndex groups every clause by component, replacing any previous index.
func (b *BaseQuery) buildIndex() {
	idx := make(map[string][]AbstractClause, 8)
	for _, c := range b.Clauses {
		comp := c.GetComponent()
		idx[comp] = append(idx[comp], c)
	}
	b.index = idx
	b.indexLen = len(b.Clauses)
}

// indexIfLarge builds the index when the query has grown past the threshold.
func (b *BaseQuery) indexIfLarge() {
	if b.index == nil && len(b.Clauses) >= clauseIndexThreshold {
		b.buildIndex()
	}
}

// dropIndex discards the index; the next mutation past the threshold rebuilds it.
func (b *BaseQuery) dropIndex() {
	b.index = nil
	b.indexLen = 0
}

// lookup returns the clauses to iterate for component. The second result
// reports whether the list is already filtered by component (index hit), in
// which case it must be treated as read-only. A stale index is ignored rather
// than rebuilt so that readers never mutate the query.
func (b *BaseQuery) lookup(component string) ([]AbstractClause, bool) {
	if b.index != nil && b.indexLen == len(b.Clauses) {
		return b.index[component], true
	}
	return b.Clauses, false
}

// SetEngineScope mirrors SetEngineScope(string engine).
// In C#, engine may be null to clear EngineScope; in Go use SetEngineScopePtr(nil) for that.
func (b *BaseQuery) SetEngineScope(engine string) *Query {
	if engine == "" {
		b.EngineScope = nil
		return b.owner
	}
	b.EngineScope = strPtr(engine)
	return b.owner
}

// SetEngineScopePtr sets EngineScope from a nullable string pointer (nil clears).
func (b *BaseQuery) SetEngineScopePtr(engine *string) *Query {
	b.EngineScope = engine
	return b.owner
}

// cloneClauses mirrors the C# BaseQuery.Clone body (only clauses).
func (b *BaseQuery) cloneClauses() *Query {
	q := NewQuery()
	if len(b.Clauses) > 0 {
		q.Clauses = make([]AbstractClause, 0, len(b.Clauses))
		for _, c := range b.Clauses {
			q.Clauses = append(q.Clauses, c.Clone())
		}
	}
	q.indexIfLarge()
	return q
}

// SetParent mirrors SetParent(AbstractQuery parent).
func (b *BaseQuery) SetParent(parent any) *Query {
	if parent != nil && parent == any(b.owner) {
		panic(fmt.Sprintf("cannot set the same %T as a parent of itself", parent))
	}
	b.Parent = parent
	return b.owner
}

// NewChild mirrors NewChild().
func (b *BaseQuery) NewChild() *Query {
	nq := NewQuery()
	_ = nq.SetParent(b.owner)
	nq.EngineScope = b.EngineScope
	return nq
}

// AddComponent mirrors AddComponent.
func (b *BaseQuery) AddComponent(component string, clause AbstractClause, engineCode *string) *Query {
	ec := engineCode
	if ec == nil {
		ec = b.EngineScope
	}
	clause.SetEngine(ec)
	clause.SetComponent(component)
	b.Clauses = append(b.Clauses, clause)
	if b.index != nil {
		if b.indexLen == len(b.Clauses)-1 {
			b.index[component] = append(b.index[component], clause)
			b.indexLen = len(b.Clauses)
		} else {
			b.dropIndex()
		}
	}
	b.indexIfLarge()
	return b.owner
}

// AddOrReplaceComponent mirrors AddOrReplaceComponent.
func (b *BaseQuery) AddOrReplaceComponent(component string, clause AbstractClause, engineCode *string) *Query {
	ec := engineCode
	if ec == nil {
		ec = b.EngineScope
	}
	list := b.getComponents(component, nil)
	var current AbstractClause
	for _, c := range list {
		if strPtrEq(c.GetEngine(), ec) {
			current = c
			break
		}
	}
	if current != nil {
		b.removeClause(current)
	}
	return b.AddComponent(component, clause, engineCode)
}

func (b *BaseQuery) removeClause(target AbstractClause) {
	inSync := b.index != nil && b.indexLen == len(b.Clauses)
	var kept []AbstractClause
	for _, c := range b.Clauses {
		if c != target {
			kept = append(kept, c)
		}
	}
	b.Clauses = kept
	if b.index == nil {
		return
	}
	if !inSync {
		b.dropIndex()
		return
	}
	comp := target.GetComponent()
	bucket := b.index[comp]
	for i, c := range bucket {
		if c == target {
			b.index[comp] = append(bucket[:i], bucket[i+1:]...)
			break
		}
	}
	b.indexLen = len(b.Clauses)
}

func (b *BaseQuery) getComponents(component string, engineCode *string) []AbstractClause {
	ec := engineCode
	if ec == nil {
		ec = b.EngineScope
	}
	list, indexed := b.lookup(component)
	var out []AbstractClause
	for _, x := range list {
		if !indexed && x.GetComponent() != component {
			continue
		}
		if ec == nil || x.GetEngine() == nil || strPtrEq(ec, x.GetEngine()) {
			// An index hit knows the exact upper bound; a scan would over-allocate.
			if out == nil && indexed {
				out = make([]AbstractClause, 0, len(list))
			}
			out = append(out, x)
		}
	}
	return out
}

// GetComponents mirrors GetComponents(string component, string engineCode = null).
func (b *BaseQuery) GetComponents(component string, engineCode *string) []AbstractClause {
	return b.getComponents(component, engineCode)
}

// GetComponentsAs mirrors GetComponents<C>(string component, string engineCode = null).
func GetComponentsAs[C AbstractClause](b *BaseQuery, component string, engineCode *string) []C {
	ec := engineCode
	if ec == nil {
		ec = b.EngineScope
	}
	list, indexed := b.lookup(component)
	var out []C
	for _, x := range list {
		if !indexed && x.GetComponent() != component {
			continue
		}
		if ec == nil || x.GetEngine() == nil || strPtrEq(ec, x.GetEngine()) {
			if c, ok := x.(C); ok {
				if out == nil && indexed {
					out = make([]C, 0, len(list))
				}
				out = append(out, c)
			}
		}
	}
	return out
}

// GetOneComponent mirrors GetOneComponent(string component, string engineCode = null).
func (b *BaseQuery) GetOneComponent(component string, engineCode *string) AbstractClause {
	c, _ := GetOneComponentAs[AbstractClause](b, component, engineCode)
	return c
}

// GetOneComponentAs mirrors GetOneComponent<C>(...).
func GetOneComponentAs[C AbstractClause](b *BaseQuery, component string, engineCode *string) (C, bool) {
	ec := engineCode
	if ec == nil {
		ec = b.EngineScope
	}
	var zero C
	var fallback C
	hasFallback := false
	list, indexed := b.lookup(component)
	for _, x := range list {
		if !indexed && x.GetComponent() != component {
			continue
		}
		eng := x.GetEngine()
		if !(ec == nil || eng == nil || strPtrEq(ec, eng)) {
			continue
		}
		typed, ok := x.(C)
		if !ok {
			continue
		}
		if strPtrEq(eng, ec) {
			return typed, true
		}
		if eng == nil && !hasFallback {
			fallback, hasFallback = typed, true
		}
	}
	if hasFallback {
		return fallback, true
	}
	return zero, false
}

// HasComponent mirrors HasComponent.
func (b *BaseQuery) HasComponent(component string, engineCode *string) bool {
	ec := engineCode
	if ec == nil {
		ec = b.EngineScope
	}
	list, indexed := b.lookup(component)
	for _, x := range list {
		if !indexed && x.GetComponent() != component {
			continue
		}
		if ec == nil || x.GetEngine() == nil || strPtrEq(ec, x.GetEngine()) {
			return true
		}
	}
	return false
}

// ClearComponent mirrors ClearComponent.
func (b *BaseQuery) ClearComponent(component string, engineCode *string) *Query {
	ec := engineCode
	if ec == nil {
		ec = b.EngineScope
	}
	var kept []AbstractClause
	for _, x := range b.Clauses {
		if x.GetComponent() == component && (ec == nil || x.GetEngine() == nil || strPtrEq(ec, x.GetEngine())) {
			continue
		}
		kept = append(kept, x)
	}
	b.Clauses = kept
	if b.index != nil {
		b.buildIndex()
	}
	return b.owner
}

// Or mirrors Or().
func (b *BaseQuery) Or() *Query {
	b.orFlag = true
	return b.owner
}

// Not mirrors Not(bool flag = true).
func (b *BaseQuery) Not(flag ...bool) *Query {
	f := true
	if len(flag) > 0 {
		f = flag[0]
	}
	b.notFlag = f
	return b.owner
}

// getOr consumes the Or() flag (SqlKata protected GetOr).
func (b *BaseQuery) getOr() bool {
	ret := b.orFlag
	b.orFlag = false
	return ret
}

// getNot consumes the Not() flag (SqlKata protected GetNot).
func (b *BaseQuery) getNot() bool {
	ret := b.notFlag
	b.notFlag = false
	return ret
}

// From mirrors From(string table).
func (b *BaseQuery) From(table string) *Query {
	return b.AddOrReplaceComponent("from", &FromClause{
		AbstractFrom: AbstractFrom{},
		Table:        table,
	}, nil)
}

// FromQuery mirrors From(Query query, string alias = null).
func (b *BaseQuery) FromQuery(query *Query, alias string) *Query {
	if query == nil {
		return b.owner.setErr(errors.New("query is nil"))
	}
	q := query.Clone()
	_ = q.SetParent(b.owner)
	if alias != "" {
		q.As(alias)
	}
	return b.AddOrReplaceComponent("from", &QueryFromClause{Query: q}, nil)
}

// FromRaw mirrors FromRaw(string sql, params object[] bindings).
func (b *BaseQuery) FromRaw(sql string, bindings ...any) *Query {
	return b.AddOrReplaceComponent("from", &RawFromClause{
		Expression: sql,
		Bindings:   bindings,
	}, nil)
}

// FromFunc mirrors From(Func<Query, Query> callback, string alias = null).
func (b *BaseQuery) FromFunc(callback func(*Query) *Query, alias string) *Query {
	q := NewQuery()
	_ = q.SetParent(b.owner)
	out := callback(q)
	return b.FromQuery(out, alias)
}
