package sqlkata

// Include holds eager-load metadata (SqlKata.Include). Compilers ignore it;
// execution layers use Includes to fetch related rows.
type Include struct {
	Name       string
	Query      *Query
	ForeignKey string
	LocalKey   string
	IsMany     bool
}

// Clone deep-copies the include (including the related query).
func (i *Include) Clone() *Include {
	if i == nil {
		return nil
	}
	var q *Query
	if i.Query != nil {
		q = i.Query.Clone()
	}
	return &Include{
		Name:       i.Name,
		Query:      q,
		ForeignKey: i.ForeignKey,
		LocalKey:   i.LocalKey,
		IsMany:     i.IsMany,
	}
}

// Include registers a related query (one-to-one by default).
func (q *Query) Include(relationName string, related *Query, foreignKey string, localKey string, isMany ...bool) *Query {
	many := false
	if len(isMany) > 0 {
		many = isMany[0]
	}
	if localKey == "" {
		localKey = "Id"
	}
	q.Includes = append(q.Includes, &Include{
		Name:       relationName,
		Query:      related,
		ForeignKey: foreignKey,
		LocalKey:   localKey,
		IsMany:     many,
	})
	return q
}

// IncludeMany registers a one-to-many related query (SqlKata.IncludeMany).
func (q *Query) IncludeMany(relationName string, related *Query, foreignKey string, localKey string) *Query {
	return q.Include(relationName, related, foreignKey, localKey, true)
}
