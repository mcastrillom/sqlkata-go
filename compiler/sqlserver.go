package compiler

import "github.com/mcastrillom/sqlkata-go/sqlkata"

// SqlServerCompiler compiles queries for SQL Server (SqlKata.SqlServerCompiler).
type SqlServerCompiler struct {
	*Compiler
}

// NewSqlServerCompiler returns defaults matching SqlKata.Compilers.SqlServerCompiler.
func NewSqlServerCompiler() *SqlServerCompiler {
	return &SqlServerCompiler{
		Compiler: &Compiler{
			OpeningIdentifier: "[",
			ClosingIdentifier: "]",
			ColumnAsKeyword:   "AS ",
			TableAsKeyword:    "AS ",
			LastID:            "SELECT scope_identity() as Id",
			EngineCode:        sqlkata.EngineSqlServer,
			placeholder:       "?",
			escape:            "\\",
			paramPrefix:       "@p",
			limitStyle:        limitOffsetFetch,
			randomSQL:         "NEWID()",
			safeOrderSQL:      "ORDER BY (SELECT 0) ",
		},
	}
}
