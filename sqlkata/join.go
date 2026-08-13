package sqlkata

import (
	"fmt"
	"strings"
)

// AbstractJoin is the join clause family (SqlKata.AbstractJoin).
type AbstractJoin interface {
	AbstractClause
}

// BaseJoin wraps a Join builder instance.
type BaseJoin struct {
	ClauseBase
	Join *Join
}

func (j *BaseJoin) Clone() AbstractClause {
	var join *Join
	if j.Join != nil {
		join = j.Join.Clone()
	}
	n := &BaseJoin{Join: join}
	j.ClauseBase.copyMeta(&n.ClauseBase)
	return n
}

// DeepJoin mirrors SqlKata.DeepJoin (relation-style join expansion).
type DeepJoin struct {
	ClauseBase
	Type               string
	Expression         string
	SourceKeySuffix    string
	TargetKey          string
	SourceKeyGenerator func(string) string
	TargetKeyGenerator func(string) string
}

func (j *DeepJoin) Clone() AbstractClause {
	n := &DeepJoin{
		Type:               j.Type,
		Expression:         j.Expression,
		SourceKeySuffix:    j.SourceKeySuffix,
		TargetKey:          j.TargetKey,
		SourceKeyGenerator: j.SourceKeyGenerator,
		TargetKeyGenerator: j.TargetKeyGenerator,
	}
	j.ClauseBase.copyMeta(&n.ClauseBase)
	return n
}

// Join is SqlKata.Join — a clause bag used inside BaseJoin (not embedding BaseQuery[*Query]).
type Join struct {
	AbstractQuery
	Clauses     []AbstractClause
	orFlag      bool
	notFlag     bool
	EngineScope *string
	typ         string
	err         error
}

// Err returns the first validation error recorded on this join.
func (j *Join) Err() error {
	if j == nil {
		return nil
	}
	return j.err
}

func (j *Join) setErr(err error) *Join {
	if j != nil && j.err == nil && err != nil {
		j.err = err
	}
	return j
}

// NewJoin constructs a join with default type "inner join".
func NewJoin() *Join {
	return &Join{typ: "inner join"}
}

// Type returns the join type string.
func (j *Join) Type() string { return j.typ }

// SetType sets the join type (uppercased like C#).
func (j *Join) SetType(t string) *Join {
	j.typ = strings.ToUpper(t)
	return j
}

// AsType mirrors Join.AsType.
func (j *Join) AsType(t string) *Join { return j.SetType(t) }

// AsInner / AsOuter / AsLeft / AsRight / AsCross match SqlKata helpers.
func (j *Join) AsInner() *Join { return j.AsType("inner join") }
func (j *Join) AsOuter() *Join { return j.AsType("outer join") }
func (j *Join) AsLeft() *Join  { return j.AsType("left join") }
func (j *Join) AsRight() *Join { return j.AsType("right join") }
func (j *Join) AsCross() *Join { return j.AsType("cross join") }

// Clone deep-copies clauses and join type.
func (j *Join) Clone() *Join {
	n := NewJoin()
	n.typ = j.typ
	n.Parent = j.Parent
	n.EngineScope = j.EngineScope
	n.err = j.err
	for _, c := range j.Clauses {
		n.Clauses = append(n.Clauses, c.Clone())
	}
	return n
}

func (j *Join) Or() *Join {
	j.orFlag = true
	return j
}

func (j *Join) Not(flag ...bool) *Join {
	f := true
	if len(flag) > 0 {
		f = flag[0]
	}
	j.notFlag = f
	return j
}

func (j *Join) getOr() bool {
	r := j.orFlag
	j.orFlag = false
	return r
}

func (j *Join) getNot() bool {
	r := j.notFlag
	j.notFlag = false
	return r
}

func (j *Join) AddComponent(component string, clause AbstractClause, engineCode *string) *Join {
	ec := engineCode
	if ec == nil {
		ec = j.EngineScope
	}
	clause.SetEngine(ec)
	clause.SetComponent(component)
	j.Clauses = append(j.Clauses, clause)
	return j
}

func (j *Join) AddOrReplaceComponent(component string, clause AbstractClause, engineCode *string) *Join {
	ec := engineCode
	if ec == nil {
		ec = j.EngineScope
	}
	for i, c := range j.Clauses {
		if c.GetComponent() == component && strPtrEq(c.GetEngine(), ec) {
			j.Clauses = append(j.Clauses[:i], j.Clauses[i+1:]...)
			break
		}
	}
	return j.AddComponent(component, clause, engineCode)
}

func (j *Join) GetOneComponent(component string, engineCode *string) AbstractClause {
	ec := engineCode
	if ec == nil {
		ec = j.EngineScope
	}
	var fallback AbstractClause
	for _, c := range j.Clauses {
		if c.GetComponent() != component {
			continue
		}
		if strPtrEq(c.GetEngine(), ec) {
			return c
		}
		if c.GetEngine() == nil && fallback == nil {
			fallback = c
		}
	}
	return fallback
}

func (j *Join) GetComponents(component string, engineCode *string) []AbstractClause {
	ec := engineCode
	if ec == nil {
		ec = j.EngineScope
	}
	var out []AbstractClause
	for _, c := range j.Clauses {
		if c.GetComponent() != component {
			continue
		}
		if ec == nil || c.GetEngine() == nil || strPtrEq(ec, c.GetEngine()) {
			out = append(out, c)
		}
	}
	return out
}

// From sets the joined table (same as JoinWith(string)).
func (j *Join) From(table string) *Join {
	return j.AddOrReplaceComponent("from", &FromClause{Table: table}, nil)
}

// FromQuery sets a subquery as the join target.
func (j *Join) FromQuery(query *Query, alias string) *Join {
	if query == nil {
		return j.setErr(fmt.Errorf("query is nil"))
	}
	q := query.Clone()
	q.Parent = j
	if alias != "" {
		q.As(alias)
	}
	return j.AddOrReplaceComponent("from", &QueryFromClause{Query: q}, nil)
}

// JoinWith aliases From(string).
func (j *Join) JoinWith(table string) *Join { return j.From(table) }

// JoinWithQuery aliases FromQuery.
func (j *Join) JoinWithQuery(query *Query, alias string) *Join {
	return j.FromQuery(query, alias)
}

// On adds a two-column ON condition.
func (j *Join) On(first, second string, op ...string) *Join {
	operator := "="
	if len(op) > 0 && op[0] != "" {
		operator = op[0]
	}
	return j.AddComponent("where", &TwoColumnsCondition{
		ConditionBase: ConditionBase{IsOr: j.getOr(), IsNot: j.getNot()},
		First:         first,
		Second:        second,
		Operator:      operator,
	}, nil)
}

// OrOn is Or().On(...).
func (j *Join) OrOn(first, second string, op ...string) *Join {
	return j.Or().On(first, second, op...)
}
