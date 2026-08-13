package sqlkata

import "fmt"

// Combine mirrors Query.Combine.
func (q *Query) Combine(operation string, all bool, other *Query) *Query {
	if other == nil {
		return q.setErr(fmt.Errorf("combine query is nil"))
	}
	if q.Method != "select" || other.Method != "select" {
		return q.setErr(fmt.Errorf("only select queries can be combined"))
	}
	return q.AddComponent("combine", &Combine{
		Query:     other,
		Operation: operation,
		All:       all,
	}, nil)
}

func (q *Query) Union(other *Query, all ...bool) *Query {
	a := false
	if len(all) > 0 {
		a = all[0]
	}
	return q.Combine("union", a, other)
}

func (q *Query) UnionAll(other *Query) *Query {
	return q.Union(other, true)
}

func (q *Query) Except(other *Query, all ...bool) *Query {
	a := false
	if len(all) > 0 {
		a = all[0]
	}
	return q.Combine("except", a, other)
}

func (q *Query) Intersect(other *Query, all ...bool) *Query {
	a := false
	if len(all) > 0 {
		a = all[0]
	}
	return q.Combine("intersect", a, other)
}

func (q *Query) CombineRaw(sql string, bindings ...any) *Query {
	if q.Method != "select" {
		return q.setErr(fmt.Errorf("only select queries can be combined"))
	}
	return q.AddComponent("combine", &RawCombine{Expression: sql, Bindings: bindings}, nil)
}
