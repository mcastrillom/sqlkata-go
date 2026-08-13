package compiler

import (
	"fmt"
	"strings"

	"github.com/mcastrillom/sqlkata-go/sqlkata"
)

func (c *Compiler) parameter(ctx *SqlResult, value any) string {
	switch v := value.(type) {
	case *sqlkata.UnsafeLiteral:
		return v.Value
	case sqlkata.UnsafeLiteral:
		return v.Value
	case *sqlkata.Variable:
		resolved, err := ctx.Query.FindVariable(v.Name)
		if err != nil {
			ctx.fail(err)
			return c.placeholder
		}
		ctx.Bindings = append(ctx.Bindings, resolved)
		return c.placeholder
	case sqlkata.Variable:
		resolved, err := ctx.Query.FindVariable(v.Name)
		if err != nil {
			ctx.fail(err)
			return c.placeholder
		}
		ctx.Bindings = append(ctx.Bindings, resolved)
		return c.placeholder
	}
	ctx.Bindings = append(ctx.Bindings, value)
	return c.placeholder
}

// resolve mirrors SqlKata.Compiler.Resolve (Variable / UnsafeLiteral → concrete value).
func (c *Compiler) resolve(ctx *SqlResult, value any) any {
	switch v := value.(type) {
	case *sqlkata.UnsafeLiteral:
		return v.Value
	case sqlkata.UnsafeLiteral:
		return v.Value
	case *sqlkata.Variable:
		resolved, err := ctx.Query.FindVariable(v.Name)
		if err != nil {
			ctx.fail(err)
			return nil
		}
		return resolved
	case sqlkata.Variable:
		resolved, err := ctx.Query.FindVariable(v.Name)
		if err != nil {
			ctx.fail(err)
			return nil
		}
		return resolved
	default:
		return value
	}
}

func (c *Compiler) parameterize(ctx *SqlResult, values []any) string {
	parts := make([]string, 0, len(values))
	for _, v := range values {
		parts = append(parts, c.parameter(ctx, v))
	}
	return strings.Join(parts, ", ")
}

func (c *Compiler) checkOperator(op string) string {
	return strings.ToLower(op)
}

func (c *Compiler) compileTrue() string {
	switch c.EngineCode {
	case sqlkata.EngineSqlServer:
		return "cast(1 as bit)"
	case sqlkata.EngineSqlite:
		return "1" // SQLite has no boolean literal before 3.23
	default:
		return "true"
	}
}

func (c *Compiler) compileFalse() string {
	switch c.EngineCode {
	case sqlkata.EngineSqlServer:
		return "cast(0 as bit)"
	case sqlkata.EngineSqlite:
		return "0"
	default:
		return "false"
	}
}

func (c *Compiler) compileWheres(ctx *SqlResult, q *sqlkata.Query) string {
	eng := c.eng()
	if !q.HasComponent("where", eng) {
		return ""
	}
	conds := sqlkata.GetComponentsAs[sqlkata.AbstractCondition](&q.BaseQuery, "where", eng)
	sql := strings.TrimSpace(c.compileConditions(ctx, conds))
	if sql == "" {
		return ""
	}
	return "WHERE " + sql
}

func (c *Compiler) compileConditions(ctx *SqlResult, conditions []sqlkata.AbstractCondition) string {
	var b strings.Builder
	b.Grow(24 * len(conditions))
	for i, cond := range conditions {
		compiled := c.compileCondition(ctx, cond)
		if compiled == "" {
			continue
		}
		if b.Len() > 0 {
			b.WriteByte(' ')
		}
		if i > 0 {
			if cond.GetIsOr() {
				b.WriteString("OR ")
			} else {
				b.WriteString("AND ")
			}
		}
		b.WriteString(compiled)
	}
	return b.String()
}

func (c *Compiler) compileCondition(ctx *SqlResult, clause sqlkata.AbstractCondition) string {
	switch x := clause.(type) {
	case *sqlkata.RawCondition:
		ctx.Bindings = append(ctx.Bindings, x.Bindings...)
		return c.wrapIdentifiers(x.Expression)
	case *sqlkata.BasicStringCondition:
		return c.compileBasicStringCondition(ctx, x)
	case *sqlkata.BasicDateCondition:
		return c.compileBasicDateCondition(ctx, x)
	case *sqlkata.BasicCondition:
		sql := c.wrap(x.Column) + " " + c.checkOperator(x.Operator) + " " + c.parameter(ctx, x.Value)
		if x.IsNot {
			return "NOT (" + sql + ")"
		}
		return sql
	case *sqlkata.TwoColumnsCondition:
		op := ""
		if x.IsNot {
			op = "NOT "
		}
		return op + c.wrap(x.First) + " " + c.checkOperator(x.Operator) + " " + c.wrap(x.Second)
	case *sqlkata.BetweenCondition:
		between := "BETWEEN"
		if x.IsNot {
			between = "NOT BETWEEN"
		}
		return c.wrap(x.Column) + " " + between + " " + c.parameter(ctx, x.Lower) + " AND " + c.parameter(ctx, x.Higher)
	case *sqlkata.InCondition:
		if len(x.Values) == 0 {
			if x.IsNot {
				return "1 = 1 /* NOT IN [empty list] */"
			}
			return "1 = 0 /* IN [empty list] */"
		}
		inOp := "IN"
		if x.IsNot {
			inOp = "NOT IN"
		}
		return c.wrap(x.Column) + " " + inOp + " (" + c.parameterize(ctx, x.Values) + ")"
	case *sqlkata.InQueryCondition:
		inOp := "IN"
		if x.IsNot {
			inOp = "NOT IN"
		}
		sub := c.compileSelectQuery(ctx, x.Query)
		return c.wrap(x.Column) + " " + inOp + " (" + sub + ")"
	case *sqlkata.NullCondition:
		op := "IS NULL"
		if x.IsNot {
			op = "IS NOT NULL"
		}
		return c.wrap(x.Column) + " " + op
	case *sqlkata.BooleanCondition:
		val := c.compileFalse()
		if x.Value {
			val = c.compileTrue()
		}
		op := "="
		if x.IsNot {
			op = "!="
		}
		return c.wrap(x.Column) + " " + op + " " + val
	case *sqlkata.ExistsCondition:
		op := "EXISTS"
		if x.IsNot {
			op = "NOT EXISTS"
		}
		query := x.Query.Clone()
		_ = query.ClearComponent("select", nil)
		_ = query.SelectRaw("1")
		sub := c.compileSelectQuery(ctx, query)
		return op + " (" + sub + ")"
	case *sqlkata.QueryCondition:
		sub := c.compileSelectQuery(ctx, x.Query)
		return c.wrap(x.Column) + " " + c.checkOperator(x.Operator) + " (" + sub + ")"
	case *sqlkata.SubQueryCondition:
		sub := c.compileSelectQuery(ctx, x.Query)
		return "(" + sub + ") " + c.checkOperator(x.Operator) + " " + c.parameter(ctx, x.Value)
	case *sqlkata.NestedCondition:
		return c.compileNestedCondition(ctx, x)
	case *sqlkata.TextCondition:
		return c.compileTextCondition(ctx, x)
	default:
		return ""
	}
}

// compileTextCondition emits portable long-text equality.
// Oracle: DBMS_LOB.COMPARE(col, TO_CLOB(?)) = 0  (avoids ORA-00932).
// Others: col = ? / col <> ?.
func (c *Compiler) compileTextCondition(ctx *SqlResult, x *sqlkata.TextCondition) string {
	col := c.wrap(x.Column)
	ph := c.parameter(ctx, x.Value)
	var sql string
	switch c.EngineCode {
	case sqlkata.EngineOracle:
		op := "= 0"
		if !x.Equal {
			op = "<> 0"
		}
		sql = "DBMS_LOB.COMPARE(" + col + ", TO_CLOB(" + ph + ")) " + op
	default:
		op := "="
		if !x.Equal {
			op = "<>"
		}
		sql = col + " " + op + " " + ph
	}
	if x.IsNot {
		return "NOT (" + sql + ")"
	}
	return sql
}

func (c *Compiler) compileBasicStringCondition(ctx *SqlResult, x *sqlkata.BasicStringCondition) string {
	column := c.wrap(x.Column)
	resolved := c.resolve(ctx, x.Value)
	value, _ := resolved.(string)
	if value == "" && resolved != nil {
		value = fmt.Sprint(resolved)
	}
	method := x.Operator
	switch x.Operator {
	case "starts", "ends", "contains", "like", "ilike":
		method = "LIKE"
		switch x.Operator {
		case "starts":
			value = value + "%"
		case "ends":
			value = "%" + value
		case "contains":
			value = "%" + value + "%"
		}
	}
	// Postgres: non-case-sensitive string ops use ILIKE (SqlKata.PostgresCompiler).
	if c.EngineCode == sqlkata.EnginePostgres && !x.CaseSensitive {
		switch x.Operator {
		case "starts", "ends", "contains", "like", "ilike":
			method = "ILIKE"
		}
	} else if !x.CaseSensitive {
		column = "LOWER(" + column + ")"
		value = strings.ToLower(value)
	}
	var sql string
	if _, ok := x.Value.(*sqlkata.UnsafeLiteral); ok {
		sql = column + " " + c.checkOperator(method) + " " + value
	} else if _, ok := x.Value.(sqlkata.UnsafeLiteral); ok {
		sql = column + " " + c.checkOperator(method) + " " + value
	} else {
		sql = column + " " + c.checkOperator(method) + " " + c.parameter(ctx, value)
	}
	if x.EscapeCharacter != nil && *x.EscapeCharacter != "" {
		sql += " ESCAPE '" + *x.EscapeCharacter + "'"
	}
	if x.IsNot {
		return "NOT (" + sql + ")"
	}
	return sql
}

// sqliteDatePart maps SqlKata date parts to strftime format specifiers.
var sqliteDatePart = map[string]string{
	"year":   "%Y",
	"month":  "%m",
	"day":    "%d",
	"hour":   "%H",
	"minute": "%M",
	"second": "%S",
}

func (c *Compiler) compileBasicDateCondition(ctx *SqlResult, x *sqlkata.BasicDateCondition) string {
	column := c.wrap(x.Column)
	part := strings.ToLower(x.Part)
	var left string
	switch c.EngineCode {
	case sqlkata.EngineSqlServer:
		if part == "time" || part == "date" {
			left = "CAST(" + column + " AS " + strings.ToUpper(part) + ")"
		} else {
			left = "DATEPART(" + strings.ToUpper(part) + ", " + column + ")"
		}
	case sqlkata.EnginePostgres:
		if part == "time" {
			left = column + "::time"
		} else if part == "date" {
			left = column + "::date"
		} else {
			left = "DATE_PART('" + strings.ToUpper(part) + "', " + column + ")"
		}
	case sqlkata.EngineOracle:
		switch part {
		case "date":
			left = "TO_CHAR(" + column + ", 'YY-MM-DD')"
		case "time":
			left = "TO_CHAR(" + column + ", 'HH24:MI:SS')"
		case "year", "month", "day", "hour", "minute", "second":
			left = "EXTRACT(" + strings.ToUpper(part) + " FROM " + column + ")"
		default:
			left = column
		}
	case sqlkata.EngineSqlite:
		// SQLite has no YEAR()/MONTH() functions, and strftime returns text,
		// so numeric parts are cast to compare against numeric bindings.
		switch {
		case part == "date" || part == "time":
			left = part + "(" + column + ")"
		case sqliteDatePart[part] != "":
			left = "CAST(strftime('" + sqliteDatePart[part] + "', " + column + ") AS INTEGER)"
		default:
			left = column
		}
	default:
		left = strings.ToUpper(part) + "(" + column + ")"
	}
	sql := left + " " + c.checkOperator(x.Operator) + " " + c.parameter(ctx, x.Value)
	if c.EngineCode == sqlkata.EngineOracle && part == "date" {
		// SqlKata compares TO_CHAR on both sides for date strings; keep value as param for simplicity.
	}
	if x.IsNot {
		return "NOT (" + sql + ")"
	}
	return sql
}

func (c *Compiler) compileNestedCondition(ctx *SqlResult, x *sqlkata.NestedCondition) string {
	var conds []sqlkata.AbstractCondition
	switch n := x.Nested.(type) {
	case *sqlkata.Query:
		eng := c.eng()
		clause := "where"
		if !n.HasComponent("where", eng) && n.HasComponent("having", eng) {
			clause = "having"
		}
		conds = sqlkata.GetComponentsAs[sqlkata.AbstractCondition](&n.BaseQuery, clause, eng)
	default:
		return ""
	}
	if len(conds) == 0 {
		return ""
	}
	sql := c.compileConditions(ctx, conds)
	if x.IsNot {
		return "NOT (" + sql + ")"
	}
	return "(" + sql + ")"
}

func (c *Compiler) compileGroups(ctx *SqlResult, q *sqlkata.Query) string {
	eng := c.eng()
	if !q.HasComponent("group", eng) {
		return ""
	}
	cols := sqlkata.GetComponentsAs[sqlkata.AbstractColumn](&q.BaseQuery, "group", eng)
	parts := make([]string, 0, len(cols))
	for _, col := range cols {
		parts = append(parts, c.compileColumn(ctx, col))
	}
	if len(parts) == 0 {
		return ""
	}
	return "GROUP BY " + strings.Join(parts, ", ")
}

func (c *Compiler) compileHaving(ctx *SqlResult, q *sqlkata.Query) string {
	eng := c.eng()
	if !q.HasComponent("having", eng) {
		return ""
	}
	conds := sqlkata.GetComponentsAs[sqlkata.AbstractCondition](&q.BaseQuery, "having", eng)
	var parts []string
	for _, cond := range conds {
		compiled := c.compileCondition(ctx, cond)
		if compiled == "" {
			continue
		}
		prefix := ""
		if len(parts) > 0 {
			if cond.GetIsOr() {
				prefix = "OR "
			} else {
				prefix = "AND "
			}
		}
		parts = append(parts, prefix+compiled)
	}
	if len(parts) == 0 {
		return ""
	}
	return "HAVING " + strings.Join(parts, " ")
}

func (c *Compiler) compileJoins(ctx *SqlResult, q *sqlkata.Query) string {
	eng := c.eng()
	if !q.HasComponent("join", eng) {
		return ""
	}
	raw := q.GetComponents("join", eng)
	parts := make([]string, 0, len(raw))
	prev := c.fromAlias(q)
	for _, cl := range raw {
		switch x := cl.(type) {
		case *sqlkata.BaseJoin:
			if x == nil || x.Join == nil {
				continue
			}
			parts = append(parts, c.compileJoin(ctx, x.Join))
			if alias := joinFromAlias(x.Join, eng); alias != "" {
				prev = alias
			}
		case *sqlkata.DeepJoin:
			if x == nil {
				continue
			}
			for _, j := range expandDeepJoin(x, &prev) {
				parts = append(parts, c.compileJoin(ctx, j))
			}
		}
	}
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, " ")
}

// fromAlias returns the FROM table alias/name used as DeepJoin start.
func (c *Compiler) fromAlias(q *sqlkata.Query) string {
	eng := c.eng()
	cl := q.GetOneComponent("from", eng)
	switch f := cl.(type) {
	case *sqlkata.FromClause:
		return f.Alias()
	case *sqlkata.QueryFromClause:
		return f.Alias()
	case *sqlkata.RawFromClause:
		return f.Alias()
	case *sqlkata.AdHocTableFromClause:
		return f.Alias()
	default:
		return ""
	}
}

func joinFromAlias(join *sqlkata.Join, eng *string) string {
	if join == nil {
		return ""
	}
	switch f := join.GetOneComponent("from", eng).(type) {
	case *sqlkata.FromClause:
		return f.Alias()
	case *sqlkata.QueryFromClause:
		return f.Alias()
	case *sqlkata.RawFromClause:
		return f.Alias()
	default:
		return ""
	}
}

// expandDeepJoin turns Expression "Author.Country" into Join clauses.
// Keys: ON next.TargetKey = prev.SourceKey where SourceKey = SourceKeyGenerator(nextTable).
func expandDeepJoin(dj *sqlkata.DeepJoin, prev *string) []*sqlkata.Join {
	expr := strings.TrimSpace(dj.Expression)
	if expr == "" || prev == nil {
		return nil
	}
	srcGen := dj.SourceKeyGenerator
	if srcGen == nil {
		suffix := dj.SourceKeySuffix
		if suffix == "" {
			suffix = "Id"
		}
		srcGen = func(table string) string { return table + suffix }
	}
	tgtGen := dj.TargetKeyGenerator
	if tgtGen == nil {
		key := dj.TargetKey
		if key == "" {
			key = "Id"
		}
		tgtGen = func(string) string { return key }
	}
	typ := dj.Type
	if typ == "" {
		typ = "inner join"
	}
	segments := strings.Split(expr, ".")
	out := make([]*sqlkata.Join, 0, len(segments))
	for _, seg := range segments {
		seg = strings.TrimSpace(seg)
		if seg == "" {
			continue
		}
		table, alias := parseTableAs(seg)
		srcKey := srcGen(table)
		tgtKey := tgtGen(table)
		j := sqlkata.NewJoin().JoinWith(seg).On(alias+"."+tgtKey, *prev+"."+srcKey).AsType(typ)
		out = append(out, j)
		*prev = alias
	}
	return out
}

func parseTableAs(expr string) (table, alias string) {
	lower := strings.ToLower(expr)
	if idx := strings.Index(lower, " as "); idx >= 0 {
		table = strings.TrimSpace(expr[:idx])
		alias = strings.TrimSpace(expr[idx+4:])
		if alias == "" {
			alias = table
		}
		return table, alias
	}
	return expr, expr
}

func (c *Compiler) compileJoin(ctx *SqlResult, join *sqlkata.Join) string {
	from := join.GetOneComponent("from", c.eng())
	joinTable := ""
	if from != nil {
		joinTable = c.compileTableExpression(ctx, from)
	}
	rawConds := join.GetComponents("where", c.eng())
	conds := make([]sqlkata.AbstractCondition, 0, len(rawConds))
	for _, cl := range rawConds {
		if ac, ok := cl.(sqlkata.AbstractCondition); ok {
			conds = append(conds, ac)
		}
	}
	on := ""
	if len(conds) > 0 {
		on = " ON " + c.compileConditions(ctx, conds)
	}
	return join.Type() + " " + joinTable + on
}

func (c *Compiler) compileUnion(ctx *SqlResult, q *sqlkata.Query) string {
	eng := c.eng()
	combines := sqlkata.GetComponentsAs[sqlkata.AbstractCombine](&q.BaseQuery, "combine", eng)
	if len(combines) == 0 {
		return ""
	}
	parts := make([]string, 0, len(combines))
	for _, cl := range combines {
		switch x := cl.(type) {
		case *sqlkata.Combine:
			op := strings.ToUpper(x.Operation) + " "
			if x.All {
				op += "ALL "
			}
			sub := c.compileSelectQuery(ctx, x.Query)
			parts = append(parts, op+sub)
		case *sqlkata.RawCombine:
			ctx.Bindings = append(ctx.Bindings, x.Bindings...)
			parts = append(parts, c.wrapIdentifiers(x.Expression))
		}
	}
	return strings.Join(parts, " ")
}

func (c *Compiler) compileInsertQuery(ctx *SqlResult, q *sqlkata.Query) string {
	eng := c.eng()
	from := q.GetOneComponent("from", eng)
	if from == nil {
		return "-- insert: no table"
	}
	table := c.compileTableExpression(ctx, from)
	inserts := sqlkata.GetComponentsAs[sqlkata.AbstractInsertClause](&q.BaseQuery, "insert", eng)
	if len(inserts) == 0 {
		return "-- insert: missing values"
	}
	if iq, ok := inserts[0].(*sqlkata.InsertQueryClause); ok {
		cols := make([]string, 0, len(iq.Columns))
		for _, col := range iq.Columns {
			cols = append(cols, c.wrap(col))
		}
		colList := ""
		if len(cols) > 0 {
			colList = " (" + strings.Join(cols, ", ") + ")"
		}
		sub := c.compileSelectQuery(ctx, iq.Query)
		return "INSERT INTO " + table + colList + " " + sub
	}
	first, ok := inserts[0].(*sqlkata.InsertClause)
	if !ok || first == nil {
		return "-- insert: missing values"
	}
	cols := make([]string, 0, len(first.Columns))
	for _, col := range first.Columns {
		cols = append(cols, c.wrap(col))
	}
	colList := ""
	if len(cols) > 0 {
		colList = " (" + strings.Join(cols, ", ") + ")"
	}
	sql := "INSERT INTO " + table + colList + " VALUES (" + c.parameterize(ctx, first.Values) + ")"
	for _, cl := range inserts[1:] {
		row, ok := cl.(*sqlkata.InsertClause)
		if !ok || row == nil {
			continue
		}
		sql += ", (" + c.parameterize(ctx, row.Values) + ")"
	}
	if first.ReturnId && c.LastID != "" && len(inserts) == 1 {
		sql += ";" + c.LastID
	}
	return sql
}
