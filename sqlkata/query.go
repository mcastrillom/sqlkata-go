package sqlkata

// Query is the SqlKata.Query surface for the Go port.
type Query struct {
	BaseQuery
	QueryAlias string
	Method     string // "select", "insert", …
	IsDistinct bool
	comment    *string
	Variables  map[string]any
	Includes   []*Include
	err        error // first builder validation error; checked by compiler.Compile
}

// Err returns the first validation error recorded while building this query.
func (q *Query) Err() error {
	if q == nil {
		return nil
	}
	return q.err
}

// setErr records err if none is set yet (first error wins) and returns q for chaining.
func (q *Query) setErr(err error) *Query {
	if q != nil && q.err == nil && err != nil {
		q.err = err
	}
	return q
}

// NewQuery mirrors new Query().
func NewQuery() *Query {
	q := &Query{Method: "select", Variables: map[string]any{}}
	q.BaseQuery.init(q)
	return q
}

// NewQueryFromTable mirrors new Query(table, comment).
func NewQueryFromTable(table string, comment ...string) *Query {
	q := NewQuery().From(table)
	if len(comment) > 0 && comment[0] != "" {
		q.Comment(comment[0])
	}
	return q
}

// As mirrors As(string alias).
func (q *Query) As(alias string) *Query {
	q.QueryAlias = alias
	return q
}

// Clone mirrors Query.Clone.
func (q *Query) Clone() *Query {
	clone := q.BaseQuery.cloneClauses()
	clone.Parent = q.Parent
	clone.QueryAlias = q.QueryAlias
	clone.Method = q.Method
	clone.IsDistinct = q.IsDistinct
	clone.comment = q.comment
	clone.err = q.err
	if q.Variables != nil {
		clone.Variables = make(map[string]any, len(q.Variables))
		for k, v := range q.Variables {
			clone.Variables[k] = v
		}
	} else {
		clone.Variables = map[string]any{}
	}
	if len(q.Includes) > 0 {
		clone.Includes = make([]*Include, 0, len(q.Includes))
		for _, inc := range q.Includes {
			clone.Includes = append(clone.Includes, inc.Clone())
		}
	}
	return clone
}

// Select adds column clauses (SqlKata.Query.Select) with expression expand.
func (q *Query) Select(columns ...string) *Query {
	q.Method = "select"
	for _, col := range columns {
		for _, expanded := range ExpandExpression(col) {
			_ = q.AddComponent("select", &Column{Name: expanded}, nil)
		}
	}
	return q
}

// SelectRaw adds a raw select expression (SqlKata.Query.SelectRaw).
func (q *Query) SelectRaw(expression string, bindings ...any) *Query {
	q.Method = "select"
	_ = q.AddComponent("select", &RawColumn{Expression: expression, Bindings: Flatten(bindings)}, nil)
	return q
}

// Distinct mirrors Query.Distinct().
func (q *Query) Distinct() *Query {
	q.IsDistinct = true
	return q
}

// Limit mirrors Query.Limit(int).
func (q *Query) Limit(value int) *Query {
	n := &LimitClause{}
	if value > 0 {
		n.Limit = value
	}
	return q.AddOrReplaceComponent("limit", n, nil)
}

// Offset mirrors Query.Offset(long).
func (q *Query) Offset(value int64) *Query {
	n := &OffsetClause{}
	if value > 0 {
		n.Offset = value
	}
	return q.AddOrReplaceComponent("offset", n, nil)
}

// GetLimit returns the limit for the given engine scope (nil uses EngineScope).
func (q *Query) GetLimit(engineCode *string) int {
	cl, ok := GetOneComponentAs[*LimitClause](&q.BaseQuery, "limit", engineCode)
	if !ok || cl == nil {
		return 0
	}
	return cl.Limit
}

// GetOffset returns the offset for the given engine scope (nil uses EngineScope).
func (q *Query) GetOffset(engineCode *string) int64 {
	cl, ok := GetOneComponentAs[*OffsetClause](&q.BaseQuery, "offset", engineCode)
	if !ok || cl == nil {
		return 0
	}
	return cl.Offset
}

// HasOffset mirrors Query.HasOffset.
func (q *Query) HasOffset(engineCode *string) bool {
	return q.GetOffset(engineCode) > 0
}

// HasLimit mirrors Query.HasLimit.
func (q *Query) HasLimit(engineCode *string) bool {
	return q.GetLimit(engineCode) > 0
}

// OrderBy mirrors Query.OrderBy (ascending).
func (q *Query) OrderBy(columns ...string) *Query {
	for _, col := range columns {
		_ = q.AddComponent("order", &OrderBy{Column: col, Ascending: true}, nil)
	}
	return q
}

// OrderByDesc mirrors Query.OrderByDesc (descending).
func (q *Query) OrderByDesc(columns ...string) *Query {
	for _, col := range columns {
		_ = q.AddComponent("order", &OrderBy{Column: col, Ascending: false}, nil)
	}
	return q
}

// OrderByRaw mirrors Query.OrderByRaw.
func (q *Query) OrderByRaw(expression string, bindings ...any) *Query {
	return q.AddComponent("order", &RawOrderBy{Expression: expression, Bindings: Flatten(bindings)}, nil)
}

// OrderByRandom mirrors Query.OrderByRandom (seed ignored, same as C# API).
func (q *Query) OrderByRandom(_ string) *Query {
	return q.AddComponent("order", &OrderByRandom{}, nil)
}
