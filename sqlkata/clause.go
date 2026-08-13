package sqlkata

// AbstractClause is the Go equivalent of SqlKata.AbstractClause.
type AbstractClause interface {
	Clone() AbstractClause
	GetEngine() *string
	GetComponent() string
	SetEngine(*string)
	SetComponent(string)
}

// ClauseBase holds Engine and Component shared by all clauses.
type ClauseBase struct {
	Engine    *string
	Component string
}

func (c *ClauseBase) GetEngine() *string    { return c.Engine }
func (c *ClauseBase) GetComponent() string  { return c.Component }
func (c *ClauseBase) SetEngine(e *string)   { c.Engine = e }
func (c *ClauseBase) SetComponent(s string) { c.Component = s }
func (c *ClauseBase) copyMeta(dst *ClauseBase) {
	dst.Engine = c.Engine
	dst.Component = c.Component
}
