package compiler

import (
	"errors"
	"strconv"
	"strings"

	"github.com/mcastrillom/sqlkata-go/sqlkata"
)

// limitStyle selects how LIMIT/OFFSET are emitted.
type limitStyle int

const (
	limitOffsetFetch limitStyle = iota // SQL Server / Oracle 12c+: OFFSET … FETCH NEXT …
	limitLimitOffset                   // PostgreSQL / MySQL / SQLite: LIMIT … OFFSET …
)

// Compiler is the shared SELECT compiler (SqlKata.Compiler subset) used by dialect wrappers.
type Compiler struct {
	OpeningIdentifier   string
	ClosingIdentifier   string
	ColumnAsKeyword     string // "AS " or "" (Oracle)
	TableAsKeyword      string // "AS " or "" (Oracle)
	LastID              string
	EngineCode          string
	SupportsFilter      bool
	UseLegacyPagination bool
	placeholder         string
	escape              string
	paramPrefix         string
	limitStyle          limitStyle
	randomSQL           string
	safeOrderSQL        string // used when OFFSET/FETCH and no ORDER BY
	// unboundedLimit is the row count meaning "no upper bound" for dialects that
	// reject a bare OFFSET (MySQL, SQLite). Empty means OFFSET may stand alone.
	unboundedLimit   string
	applyLegacyLimit func(c *Compiler, ctx *SqlResult, q *sqlkata.Query, raw string) string
}

func namedBindings(bindings []any, prefix string) map[string]any {
	out := make(map[string]any, len(bindings))
	for i, v := range bindings {
		out[prefix+strconv.Itoa(i)] = v
	}
	return out
}

func replacePlaceholders(raw, match, escape, prefix string) string {
	if raw == "" || !strings.Contains(raw, match) {
		return raw
	}
	idx := 0
	var b strings.Builder
	b.Grow(len(raw) + 8)
	esc := []byte(escape)
	ml := len(match)
	for i := 0; i < len(raw); {
		if ml > 0 && i+ml <= len(raw) && raw[i:i+ml] == match && !isEscaped(raw, i, esc) {
			b.WriteString(prefix)
			b.WriteString(strconv.Itoa(idx))
			idx++
			i += ml
			continue
		}
		b.WriteByte(raw[i])
		i++
	}
	return b.String()
}

func isEscaped(raw string, pos int, esc []byte) bool {
	if len(esc) == 0 {
		return false
	}
	n := 0
	for p := pos - len(esc); p >= 0 && strings.HasPrefix(raw[p:], string(esc)); p -= len(esc) {
		n++
	}
	return n%2 == 1
}

// eng returns the engine scope pointer for component lookups.
// Taking the address of the field avoids an allocation per lookup.
func (c *Compiler) eng() *string {
	return &c.EngineCode
}

// Compile builds RawSQL (? placeholders) and Bindings.
// Returns an error when the query has a builder validation error or compile fails
// (for example a missing Define variable).
//
// Compile never mutates q, so SqlResult.Query references the caller's query.
func (c *Compiler) Compile(q *sqlkata.Query) (*SqlResult, error) {
	if q == nil {
		return nil, errors.New("query is nil")
	}
	if err := q.Err(); err != nil {
		return nil, err
	}
	ctx := &SqlResult{Query: q, comp: c}
	var raw string
	switch q.Method {
	case "select", "aggregate":
		raw = c.compileSelectQuery(ctx, q)
	case "insert":
		raw = c.compileInsertQuery(ctx, q)
	case "update":
		raw = c.compileUpdateQuery(ctx, q)
	case "delete":
		raw = c.compileDeleteQuery(ctx, q)
	default:
		raw = "-- unsupported method: " + q.Method
	}
	if cmt := q.GetComment(); cmt != "" {
		raw = "/* " + strings.ReplaceAll(cmt, "*/", "* /") + " */\n" + raw
	}
	ctx.RawSQL = raw
	if q.HasComponent("cte", c.eng()) {
		c.compileCteQuery(ctx, q)
	}
	if ctx.err != nil {
		return nil, ctx.err
	}
	return ctx, nil
}

func (c *Compiler) compileSelectQuery(ctx *SqlResult, q *sqlkata.Query) string {
	// SqlKata creates a fresh SqlResult{Query: subquery} per CompileSelectQuery so
	// Variable resolution (FindVariable) sees the subquery's Define map.
	prev := ctx.Query
	ctx.Query = q
	defer func() { ctx.Query = prev }()

	raw := c.compileSelectQueryCore(ctx, q)
	if c.UseLegacyPagination && c.applyLegacyLimit != nil {
		return c.applyLegacyLimit(c, ctx, q, raw)
	}
	return raw
}

func (c *Compiler) compileSelectQueryCore(ctx *SqlResult, q *sqlkata.Query) string {
	parts := [9]string{
		c.compileColumns(ctx, q),
		c.compileFrom(ctx, q),
		c.compileJoins(ctx, q),
		c.compileWheres(ctx, q),
		c.compileGroups(ctx, q),
		c.compileHaving(ctx, q),
		c.compileOrders(ctx, q),
		c.compileLimit(ctx, q),
		c.compileUnion(ctx, q),
	}
	size := 0
	for _, p := range parts {
		if p != "" {
			size += len(p) + 1
		}
	}
	var b strings.Builder
	b.Grow(size)
	for _, p := range parts {
		if p == "" || strings.TrimSpace(p) == "" {
			continue
		}
		if b.Len() > 0 {
			b.WriteByte(' ')
		}
		b.WriteString(p)
	}
	return b.String()
}

func (c *Compiler) compileColumns(ctx *SqlResult, q *sqlkata.Query) string {
	eng := c.eng()
	if q.HasComponent("aggregate", eng) {
		agg, ok := sqlkata.GetOneComponentAs[*sqlkata.AggregateClause](&q.BaseQuery, "aggregate", eng)
		if ok && agg != nil {
			cols := agg.Columns
			if len(cols) == 0 {
				cols = []string{"*"}
			}
			parts := make([]string, 0, len(cols))
			for _, col := range cols {
				parts = append(parts, c.compileColumn(ctx, &sqlkata.Column{Name: col}))
			}
			inner := strings.Join(parts, ", ")
			if q.IsDistinct && len(parts) == 1 {
				inner = "DISTINCT " + inner
			}
			return "SELECT " + strings.ToUpper(agg.Type) + "(" + inner + ") " + c.ColumnAsKeyword + c.wrapValue(agg.Type)
		}
	}
	cols := sqlkata.GetComponentsAs[sqlkata.AbstractColumn](&q.BaseQuery, "select", eng)
	dist := ""
	if q.IsDistinct {
		dist = "DISTINCT "
	}
	if len(cols) == 0 {
		return "SELECT " + dist + "*"
	}
	var b strings.Builder
	b.Grow(16 * (len(cols) + 1))
	b.WriteString("SELECT ")
	b.WriteString(dist)
	for i, col := range cols {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(c.compileColumn(ctx, col))
	}
	return b.String()
}

func (c *Compiler) compileColumn(ctx *SqlResult, col sqlkata.AbstractColumn) string {
	switch x := col.(type) {
	case *sqlkata.RawColumn:
		ctx.Bindings = append(ctx.Bindings, x.Bindings...)
		return c.wrapIdentifiers(x.Expression)
	case *sqlkata.TextColumn:
		return c.compileTextColumn(x)
	case *sqlkata.QueryColumn:
		alias := ""
		if x.Query != nil && x.Query.QueryAlias != "" {
			alias = " " + c.ColumnAsKeyword + c.wrapValue(x.Query.QueryAlias)
		}
		sub := c.compileSelectQuery(ctx, x.Query)
		return "(" + sub + ")" + alias
	case *sqlkata.AggregatedColumn:
		agg := strings.ToUpper(x.Aggregate)
		inner := ""
		if x.Column != nil {
			inner = c.compileColumn(ctx, x.Column)
		}
		filterSQL := ""
		if x.Filter != nil {
			conds := sqlkata.GetComponentsAs[sqlkata.AbstractCondition](&x.Filter.BaseQuery, "where", c.eng())
			filterSQL = c.compileConditions(ctx, conds)
		}
		if filterSQL == "" {
			return agg + "(" + inner + ")"
		}
		if c.SupportsFilter {
			return agg + "(" + inner + ") FILTER (WHERE " + filterSQL + ")"
		}
		return agg + "(CASE WHEN " + filterSQL + " THEN " + inner + " END)"
	case *sqlkata.Column:
		return c.wrap(x.Name)
	default:
		return ""
	}
}

// compileTextColumn projects long text as a string-safe expression.
// Oracle: DBMS_LOB.SUBSTR(col, maxLen, 1); others: wrap(col).
func (c *Compiler) compileTextColumn(x *sqlkata.TextColumn) string {
	col := c.wrap(x.Name)
	expr := col
	if c.EngineCode == sqlkata.EngineOracle {
		maxLen := x.MaxLen
		if maxLen <= 0 {
			maxLen = 4000
		}
		expr = "DBMS_LOB.SUBSTR(" + col + ", " + strconv.Itoa(maxLen) + ", 1)"
	}
	if x.Alias == "" {
		return expr
	}
	return expr + " " + c.ColumnAsKeyword + c.wrapValue(x.Alias)
}

func (c *Compiler) compileFrom(ctx *SqlResult, q *sqlkata.Query) string {
	eng := c.eng()
	if !q.HasComponent("from", eng) {
		return ""
	}
	cl := q.GetOneComponent("from", eng)
	if cl == nil {
		return ""
	}
	return "FROM " + c.compileTableExpression(ctx, cl)
}

func (c *Compiler) compileTableExpression(ctx *SqlResult, cl sqlkata.AbstractClause) string {
	switch x := cl.(type) {
	case *sqlkata.RawFromClause:
		ctx.Bindings = append(ctx.Bindings, x.Bindings...)
		return c.wrapIdentifiers(x.Expression)
	case *sqlkata.QueryFromClause:
		if x.Query == nil {
			return "/*nil subquery*/"
		}
		alias := ""
		if x.Query.QueryAlias != "" {
			alias = " " + c.TableAsKeyword + c.wrapValue(x.Query.QueryAlias)
		}
		sub := c.compileSelectQuery(ctx, x.Query)
		return "(" + sub + ")" + alias
	case *sqlkata.FromClause:
		return c.wrap(x.Table)
	default:
		return ""
	}
}

func (c *Compiler) compileOrders(ctx *SqlResult, q *sqlkata.Query) string {
	eng := c.eng()
	if !q.HasComponent("order", eng) {
		return ""
	}
	orders := sqlkata.GetComponentsAs[sqlkata.AbstractOrderBy](&q.BaseQuery, "order", eng)
	var b strings.Builder
	b.Grow(16 * (len(orders) + 1))
	b.WriteString("ORDER BY ")
	written := 0
	for _, o := range orders {
		switch x := o.(type) {
		case *sqlkata.RawOrderBy:
			if written > 0 {
				b.WriteString(", ")
			}
			ctx.Bindings = append(ctx.Bindings, x.Bindings...)
			b.WriteString(c.wrapIdentifiers(x.Expression))
			written++
		case *sqlkata.OrderByRandom:
			if written > 0 {
				b.WriteString(", ")
			}
			b.WriteString(c.randomSQL)
			written++
		case *sqlkata.OrderBy:
			if written > 0 {
				b.WriteString(", ")
			}
			b.WriteString(c.wrap(x.Column))
			if !x.Ascending {
				b.WriteString(" DESC")
			}
			written++
		}
	}
	if written == 0 {
		return ""
	}
	return b.String()
}

func (c *Compiler) compileLimit(ctx *SqlResult, q *sqlkata.Query) string {
	if c.UseLegacyPagination {
		return ""
	}
	eng := c.eng()
	limit := q.GetLimit(eng)
	offset := q.GetOffset(eng)
	if limit == 0 && offset == 0 {
		return ""
	}

	switch c.limitStyle {
	case limitLimitOffset:
		return c.compileLimitAndOffset(ctx, limit, offset)
	default:
		return c.compileLimitOffsetFetch(ctx, q, eng, limit, offset)
	}
}

// CompileLimit mirrors SqlKata.Compiler.CompileLimit (used by dialect limit unit tests).
func (c *Compiler) CompileLimit(ctx *SqlResult) string {
	if ctx == nil || ctx.Query == nil {
		return ""
	}
	return c.compileLimit(ctx, ctx.Query)
}

func (c *Compiler) compileLimitAndOffset(ctx *SqlResult, limit int, offset int64) string {
	if offset == 0 {
		ctx.Bindings = append(ctx.Bindings, limit)
		return "LIMIT " + c.placeholder
	}
	if limit == 0 {
		ctx.Bindings = append(ctx.Bindings, offset)
		if c.unboundedLimit != "" {
			return "LIMIT " + c.unboundedLimit + " OFFSET " + c.placeholder
		}
		return "OFFSET " + c.placeholder
	}
	ctx.Bindings = append(ctx.Bindings, limit, offset)
	return "LIMIT " + c.placeholder + " OFFSET " + c.placeholder
}

func (c *Compiler) compileLimitOffsetFetch(ctx *SqlResult, q *sqlkata.Query, eng *string, limit int, offset int64) string {
	safeOrder := ""
	if !q.HasComponent("order", eng) && c.safeOrderSQL != "" {
		safeOrder = c.safeOrderSQL
	}
	if limit == 0 {
		ctx.Bindings = append(ctx.Bindings, offset)
		return safeOrder + "OFFSET " + c.placeholder + " ROWS"
	}
	ctx.Bindings = append(ctx.Bindings, offset, int64(limit))
	return safeOrder + "OFFSET " + c.placeholder + " ROWS FETCH NEXT " + c.placeholder + " ROWS ONLY"
}

func (c *Compiler) wrap(value string) string {
	if idx := lastIndexAsKeyword(value); idx > 0 {
		return c.wrap(strings.TrimSpace(value[:idx])) + " " + c.ColumnAsKeyword + c.wrapValue(strings.TrimSpace(value[idx+4:]))
	}
	if strings.IndexByte(value, '.') < 0 {
		return c.wrapValue(value)
	}
	var b strings.Builder
	b.Grow(len(value) + 8)
	start := 0
	for i := 0; i <= len(value); i++ {
		if i < len(value) && value[i] != '.' {
			continue
		}
		if start > 0 {
			b.WriteByte('.')
		}
		b.WriteString(c.wrapValue(strings.TrimSpace(value[start:i])))
		start = i + 1
	}
	return b.String()
}

// lastIndexAsKeyword finds the last case-insensitive " as " without allocating.
func lastIndexAsKeyword(value string) int {
	for i := len(value) - 4; i >= 0; i-- {
		if value[i] != ' ' || value[i+3] != ' ' {
			continue
		}
		if (value[i+1] == 'a' || value[i+1] == 'A') && (value[i+2] == 's' || value[i+2] == 'S') {
			return i
		}
	}
	return -1
}

func (c *Compiler) wrapValue(value string) string {
	v := strings.TrimSpace(value)
	if v == "*" {
		return v
	}
	if c.OpeningIdentifier == "" && c.ClosingIdentifier == "" {
		return v
	}
	esc := c.ClosingIdentifier
	if esc != "" && strings.Contains(v, esc) {
		v = strings.ReplaceAll(v, esc, esc+esc)
	}
	return c.OpeningIdentifier + v + c.ClosingIdentifier
}

func (c *Compiler) wrapIdentifiers(input string) string {
	out := sqlkata.ReplaceIdentifierUnlessEscaped(input, c.escape, "{", c.OpeningIdentifier)
	out = sqlkata.ReplaceIdentifierUnlessEscaped(out, c.escape, "}", c.ClosingIdentifier)
	out = sqlkata.ReplaceIdentifierUnlessEscaped(out, c.escape, "[", c.OpeningIdentifier)
	out = sqlkata.ReplaceIdentifierUnlessEscaped(out, c.escape, "]", c.ClosingIdentifier)
	return out
}
