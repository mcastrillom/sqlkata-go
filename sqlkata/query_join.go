package sqlkata

// Join adds an INNER JOIN with a simple ON (SqlKata.Join).
func (q *Query) Join(table, first, second string, extras ...string) *Query {
	op, typ := "=", "inner join"
	if len(extras) > 0 && extras[0] != "" {
		op = extras[0]
	}
	if len(extras) > 1 && extras[1] != "" {
		typ = extras[1]
	}
	j := NewJoin().JoinWith(table).On(first, second, op).AsType(typ)
	return q.addJoin(j)
}

// LeftJoin mirrors Query.LeftJoin.
func (q *Query) LeftJoin(table, first, second string, op ...string) *Query {
	operator := "="
	if len(op) > 0 && op[0] != "" {
		operator = op[0]
	}
	return q.Join(table, first, second, operator, "left join")
}

// RightJoin mirrors Query.RightJoin.
func (q *Query) RightJoin(table, first, second string, op ...string) *Query {
	operator := "="
	if len(op) > 0 && op[0] != "" {
		operator = op[0]
	}
	return q.Join(table, first, second, operator, "right join")
}

// CrossJoin mirrors Query.CrossJoin.
func (q *Query) CrossJoin(table string) *Query {
	j := NewJoin().JoinWith(table).AsCross()
	return q.addJoin(j)
}

// JoinWith builds a join via callback (SqlKata.Join(table, Func<Join,Join>)).
func (q *Query) JoinWith(table string, callback func(*Join) *Join, typ ...string) *Query {
	t := "inner join"
	if len(typ) > 0 && typ[0] != "" {
		t = typ[0]
	}
	j := NewJoin().JoinWith(table).AsType(t)
	j = callback(j)
	return q.addJoin(j)
}

// JoinQuery joins a subquery (SqlKata.Join(Query, Func, type)).
func (q *Query) JoinQuery(sub *Query, onCallback func(*Join) *Join, typ ...string) *Query {
	t := "inner join"
	if len(typ) > 0 && typ[0] != "" {
		t = typ[0]
	}
	j := NewJoin().JoinWithQuery(sub, sub.QueryAlias).AsType(t)
	j = onCallback(j)
	return q.addJoin(j)
}

func (q *Query) addJoin(j *Join) *Query {
	if err := j.Err(); err != nil {
		return q.setErr(err)
	}
	return q.AddComponent("join", &BaseJoin{Join: j}, nil)
}

// LeftJoinQuery is JoinQuery with left join.
func (q *Query) LeftJoinQuery(sub *Query, onCallback func(*Join) *Join) *Query {
	return q.JoinQuery(sub, onCallback, "left join")
}

// DeepJoin adds a path-style join clause (SqlKata.DeepJoin).
// Expression is a dotted path, e.g. "Author.Country".
// Default keys: source = relatedTable + sourceKeySuffix ("AuthorId"), target = targetKey ("Id").
func (q *Query) DeepJoin(expression string, opts ...DeepJoinOption) *Query {
	dj := &DeepJoin{
		Type:            "inner join",
		Expression:      expression,
		SourceKeySuffix: "Id",
		TargetKey:       "Id",
	}
	dj.SourceKeyGenerator = func(table string) string { return table + dj.SourceKeySuffix }
	dj.TargetKeyGenerator = func(string) string { return dj.TargetKey }
	for _, opt := range opts {
		opt(dj)
	}
	return q.AddComponent("join", dj, nil)
}

// LeftDeepJoin is DeepJoin with type left join.
func (q *Query) LeftDeepJoin(expression string, opts ...DeepJoinOption) *Query {
	opts = append([]DeepJoinOption{DeepJoinType("left join")}, opts...)
	return q.DeepJoin(expression, opts...)
}

// DeepJoinOption customizes a DeepJoin clause.
type DeepJoinOption func(*DeepJoin)

// DeepJoinType sets the join type ("inner join", "left join", …).
func DeepJoinType(typ string) DeepJoinOption {
	return func(d *DeepJoin) { d.Type = typ }
}

// DeepJoinKeys sets SourceKeySuffix and TargetKey (and default generators).
func DeepJoinKeys(sourceKeySuffix, targetKey string) DeepJoinOption {
	return func(d *DeepJoin) {
		if sourceKeySuffix != "" {
			d.SourceKeySuffix = sourceKeySuffix
		}
		if targetKey != "" {
			d.TargetKey = targetKey
		}
		d.SourceKeyGenerator = func(table string) string { return table + d.SourceKeySuffix }
		d.TargetKeyGenerator = func(string) string { return d.TargetKey }
	}
}

// DeepJoinGenerators overrides key generator funcs.
func DeepJoinGenerators(source, target func(string) string) DeepJoinOption {
	return func(d *DeepJoin) {
		if source != nil {
			d.SourceKeyGenerator = source
		}
		if target != nil {
			d.TargetKeyGenerator = target
		}
	}
}
