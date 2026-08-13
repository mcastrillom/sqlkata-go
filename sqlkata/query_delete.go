package sqlkata

// AsDelete mirrors Query.AsDelete.
func (q *Query) AsDelete() *Query {
	q.Method = "delete"
	return q
}
