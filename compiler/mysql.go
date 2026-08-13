package compiler

import "github.com/mcastrillom/sqlkata-go/sqlkata"

// mysqlUnboundedLimit is the max BIGINT UNSIGNED, MySQL's documented way of
// saying "all rows from this offset on" (SqlKata.MySqlCompiler.CompileLimit).
const mysqlUnboundedLimit = "18446744073709551615"

// MySqlCompiler compiles queries for MySQL and MariaDB (SqlKata.MySqlCompiler).
type MySqlCompiler struct {
	*Compiler
}

// NewMySqlCompiler returns defaults matching SqlKata.Compilers.MySqlCompiler.
func NewMySqlCompiler() *MySqlCompiler {
	return &MySqlCompiler{
		Compiler: &Compiler{
			OpeningIdentifier: "`",
			ClosingIdentifier: "`",
			ColumnAsKeyword:   "AS ",
			TableAsKeyword:    "AS ",
			LastID:            "SELECT last_insert_id() as Id",
			EngineCode:        sqlkata.EngineMySql,
			placeholder:       "?",
			escape:            "\\",
			paramPrefix:       "@p",
			limitStyle:        limitLimitOffset,
			// SqlKata emits RANDOM(); MySQL only has RAND().
			randomSQL:      "RAND()",
			unboundedLimit: mysqlUnboundedLimit,
		},
	}
}
