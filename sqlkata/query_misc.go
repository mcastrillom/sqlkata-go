package sqlkata

import "fmt"

// Comment sets a SQL comment (SqlKata.Comment).
func (q *Query) Comment(comment string) *Query {
	q.comment = &comment
	return q
}

// GetComment returns the query comment.
func (q *Query) GetComment() string {
	if q.comment == nil {
		return ""
	}
	return *q.comment
}

// Take aliases Limit.
func (q *Query) Take(limit int) *Query { return q.Limit(limit) }

// Skip aliases Offset.
func (q *Query) Skip(offset int) *Query { return q.Offset(int64(offset)) }

// ForPage sets Skip/Take for a 1-based page (SqlKata.ForPage).
func (q *Query) ForPage(page int, perPage ...int) *Query {
	pp := 15
	if len(perPage) > 0 && perPage[0] > 0 {
		pp = perPage[0]
	}
	if page < 1 {
		page = 1
	}
	return q.Skip((page - 1) * pp).Take(pp)
}

// When applies whenTrue if condition is true, else optional whenFalse (SqlKata.When).
func (q *Query) When(condition bool, whenTrue func(*Query) *Query, whenFalse ...func(*Query) *Query) *Query {
	if condition && whenTrue != nil {
		return whenTrue(q)
	}
	if !condition && len(whenFalse) > 0 && whenFalse[0] != nil {
		return whenFalse[0](q)
	}
	return q
}

// WhenNot applies callback when condition is false (SqlKata.WhenNot).
func (q *Query) WhenNot(condition bool, callback func(*Query) *Query) *Query {
	if !condition && callback != nil {
		return callback(q)
	}
	return q
}

// For scopes EngineScope for dialect-specific clauses (SqlKata.For).
func (q *Query) For(engine string, fn func(*Query) *Query) *Query {
	q.EngineScope = strPtr(engine)
	out := fn(q)
	out.EngineScope = nil
	return out
}

// Define stores a named variable (SqlKata.Define).
func (q *Query) Define(name string, value any) *Query {
	if q.Variables == nil {
		q.Variables = map[string]any{}
	}
	q.Variables[name] = value
	return q
}

// FindVariable resolves a variable from this query or parents (SqlKata.FindVariable).
// Returns ErrVariableNotFound (wrapped with the name) when undefined.
func (q *Query) FindVariable(name string) (any, error) {
	if q.Variables != nil {
		if v, ok := q.Variables[name]; ok {
			return v, nil
		}
	}
	if p, ok := q.Parent.(*Query); ok && p != nil {
		return p.FindVariable(name)
	}
	return nil, fmt.Errorf("%w: %q", ErrVariableNotFound, name)
}
