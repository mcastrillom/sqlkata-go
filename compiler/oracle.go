package compiler

import "github.com/mcastrillom/sqlkata-go/sqlkata"

// OracleCompiler compiles queries for Oracle (SqlKata.OracleCompiler).
type OracleCompiler struct {
	*Compiler
}

// NewOracleCompiler returns defaults matching SqlKata.Compilers.OracleCompiler.
func NewOracleCompiler() *OracleCompiler {
	c := &Compiler{
		OpeningIdentifier: "\"",
		ClosingIdentifier: "\"",
		ColumnAsKeyword:   "", // Oracle omits AS for column aliases
		TableAsKeyword:    "", // Oracle omits AS for table aliases
		EngineCode:        sqlkata.EngineOracle,
		placeholder:       "?",
		escape:            "\\",
		paramPrefix:       ":p",
		limitStyle:        limitOffsetFetch,
		randomSQL:         "RANDOM()",
		safeOrderSQL:      "ORDER BY (SELECT 0 FROM DUAL) ",
	}
	c.applyLegacyLimit = applyOracleLegacyLimit
	return &OracleCompiler{Compiler: c}
}

// applyOracleLegacyLimit mirrors OracleCompiler.ApplyLegacyLimit (pre-12c ROWNUM).
func applyOracleLegacyLimit(c *Compiler, ctx *SqlResult, q *sqlkata.Query, raw string) string {
	eng := c.eng()
	limit := q.GetLimit(eng)
	offset := q.GetOffset(eng)
	if limit == 0 && offset == 0 {
		return raw
	}

	ph := c.placeholder
	if limit == 0 {
		ctx.Bindings = append(ctx.Bindings, offset)
		return `SELECT * FROM (SELECT "results_wrapper".*, ROWNUM "row_num" FROM (` + raw + `) "results_wrapper") WHERE "row_num" > ` + ph
	}
	if offset == 0 {
		ctx.Bindings = append(ctx.Bindings, limit)
		return `SELECT * FROM (` + raw + `) WHERE ROWNUM <= ` + ph
	}
	ctx.Bindings = append(ctx.Bindings, int64(limit)+offset, offset)
	return `SELECT * FROM (SELECT "results_wrapper".*, ROWNUM "row_num" FROM (` + raw + `) "results_wrapper" WHERE ROWNUM <= ` + ph + `) WHERE "row_num" > ` + ph
}
