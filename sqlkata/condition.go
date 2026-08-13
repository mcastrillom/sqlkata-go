package sqlkata

import (
	"fmt"
	"strings"
)

// AbstractCondition is the condition clause family (SqlKata.AbstractCondition).
type AbstractCondition interface {
	AbstractClause
	GetIsOr() bool
	GetIsNot() bool
	SetIsOr(bool)
	SetIsNot(bool)
}

// ConditionBase holds IsOr / IsNot shared by all conditions.
type ConditionBase struct {
	ClauseBase
	IsOr  bool
	IsNot bool
}

func (c *ConditionBase) GetIsOr() bool   { return c.IsOr }
func (c *ConditionBase) GetIsNot() bool  { return c.IsNot }
func (c *ConditionBase) SetIsOr(v bool)  { c.IsOr = v }
func (c *ConditionBase) SetIsNot(v bool) { c.IsNot = v }

func (c *ConditionBase) copyCond(dst *ConditionBase) {
	c.ClauseBase.copyMeta(&dst.ClauseBase)
	dst.IsOr = c.IsOr
	dst.IsNot = c.IsNot
}

// BasicCondition is column OP value.
type BasicCondition struct {
	ConditionBase
	Column   string
	Operator string
	Value    any
}

func (c *BasicCondition) Clone() AbstractClause {
	n := &BasicCondition{Column: c.Column, Operator: c.Operator, Value: c.Value}
	c.copyCond(&n.ConditionBase)
	return n
}

// BasicStringCondition extends BasicCondition with case/escape options.
type BasicStringCondition struct {
	BasicCondition
	CaseSensitive   bool
	EscapeCharacter *string
}

func (c *BasicStringCondition) SetEscapeCharacter(v string) error {
	if strings.TrimSpace(v) == "" {
		c.EscapeCharacter = nil
		return nil
	}
	if len([]rune(v)) > 1 {
		return fmt.Errorf("EscapeCharacter can only contain a single character, got %q", v)
	}
	c.EscapeCharacter = strPtr(v)
	return nil
}

func (c *BasicStringCondition) Clone() AbstractClause {
	n := &BasicStringCondition{
		BasicCondition: BasicCondition{
			Column:   c.Column,
			Operator: c.Operator,
			Value:    c.Value,
		},
		CaseSensitive: c.CaseSensitive,
	}
	c.copyCond(&n.ConditionBase)
	if c.EscapeCharacter != nil {
		n.EscapeCharacter = strPtr(*c.EscapeCharacter)
	}
	return n
}

// BasicDateCondition compares a date part of a column.
type BasicDateCondition struct {
	BasicCondition
	Part string
}

func (c *BasicDateCondition) Clone() AbstractClause {
	n := &BasicDateCondition{
		BasicCondition: BasicCondition{
			Column:   c.Column,
			Operator: c.Operator,
			Value:    c.Value,
		},
		Part: c.Part,
	}
	c.copyCond(&n.ConditionBase)
	return n
}

// TwoColumnsCondition is first OP second (column vs column).
type TwoColumnsCondition struct {
	ConditionBase
	First    string
	Operator string
	Second   string
}

func (c *TwoColumnsCondition) Clone() AbstractClause {
	n := &TwoColumnsCondition{First: c.First, Operator: c.Operator, Second: c.Second}
	c.copyCond(&n.ConditionBase)
	return n
}

// QueryCondition is column OP (subquery).
type QueryCondition struct {
	ConditionBase
	Column   string
	Operator string
	Query    *Query
}

func (c *QueryCondition) Clone() AbstractClause {
	var q *Query
	if c.Query != nil {
		q = c.Query.Clone()
	}
	n := &QueryCondition{Column: c.Column, Operator: c.Operator, Query: q}
	c.copyCond(&n.ConditionBase)
	return n
}

// SubQueryCondition is (subquery) OP value.
type SubQueryCondition struct {
	ConditionBase
	Value    any
	Operator string
	Query    *Query
}

func (c *SubQueryCondition) Clone() AbstractClause {
	var q *Query
	if c.Query != nil {
		q = c.Query.Clone()
	}
	n := &SubQueryCondition{Value: c.Value, Operator: c.Operator, Query: q}
	c.copyCond(&n.ConditionBase)
	return n
}

// InCondition is column IN (...values).
type InCondition struct {
	ConditionBase
	Column string
	Values []any
}

func (c *InCondition) Clone() AbstractClause {
	n := &InCondition{Column: c.Column, Values: append([]any(nil), c.Values...)}
	c.copyCond(&n.ConditionBase)
	return n
}

// InQueryCondition is column IN (subquery).
type InQueryCondition struct {
	ConditionBase
	Column string
	Query  *Query
}

func (c *InQueryCondition) Clone() AbstractClause {
	var q *Query
	if c.Query != nil {
		q = c.Query.Clone()
	}
	n := &InQueryCondition{Column: c.Column, Query: q}
	c.copyCond(&n.ConditionBase)
	return n
}

// BetweenCondition is column BETWEEN lower AND higher.
type BetweenCondition struct {
	ConditionBase
	Column string
	Higher any
	Lower  any
}

func (c *BetweenCondition) Clone() AbstractClause {
	n := &BetweenCondition{Column: c.Column, Higher: c.Higher, Lower: c.Lower}
	c.copyCond(&n.ConditionBase)
	return n
}

// NullCondition is column IS NULL / IS NOT NULL.
type NullCondition struct {
	ConditionBase
	Column string
}

func (c *NullCondition) Clone() AbstractClause {
	n := &NullCondition{Column: c.Column}
	c.copyCond(&n.ConditionBase)
	return n
}

// BooleanCondition is column IS true/false.
type BooleanCondition struct {
	ConditionBase
	Column string
	Value  bool
}

func (c *BooleanCondition) Clone() AbstractClause {
	n := &BooleanCondition{Column: c.Column, Value: c.Value}
	c.copyCond(&n.ConditionBase)
	return n
}

// NestedCondition wraps nested WHERE/ON groups. Nested is *Query or *Join.
type NestedCondition struct {
	ConditionBase
	Nested any
}

func (c *NestedCondition) Clone() AbstractClause {
	var nested any
	switch x := c.Nested.(type) {
	case *Query:
		if x != nil {
			nested = x.Clone()
		}
	case *Join:
		if x != nil {
			nested = x.Clone()
		}
	default:
		nested = c.Nested
	}
	n := &NestedCondition{Nested: nested}
	c.copyCond(&n.ConditionBase)
	return n
}

// ExistsCondition is EXISTS (subquery).
type ExistsCondition struct {
	ConditionBase
	Query *Query
}

func (c *ExistsCondition) Clone() AbstractClause {
	var q *Query
	if c.Query != nil {
		q = c.Query.Clone()
	}
	n := &ExistsCondition{Query: q}
	c.copyCond(&n.ConditionBase)
	return n
}

// RawCondition is a raw SQL condition fragment.
type RawCondition struct {
	ConditionBase
	Expression string
	Bindings   []any
}

func (c *RawCondition) Clone() AbstractClause {
	n := &RawCondition{Expression: c.Expression, Bindings: append([]any(nil), c.Bindings...)}
	c.copyCond(&n.ConditionBase)
	return n
}

// TextCondition compares long text (CLOB / VARCHAR(MAX) / TEXT).
// Oracle compilers emit DBMS_LOB.COMPARE(..., TO_CLOB(?)) to avoid ORA-00932.
type TextCondition struct {
	ConditionBase
	Column string
	Value  any
	// Equal is true for equality (COMPARE = 0 / =), false for inequality.
	Equal bool
}

func (c *TextCondition) Clone() AbstractClause {
	n := &TextCondition{Column: c.Column, Value: c.Value, Equal: c.Equal}
	c.copyCond(&n.ConditionBase)
	return n
}
