package compiler

import "github.com/mcastrillom/sqlkata-go/sqlkata"

// PostgresCompiler compiles queries for PostgreSQL (SqlKata.PostgresCompiler).
type PostgresCompiler struct {
	*Compiler
}

// NewPostgresCompiler returns defaults matching SqlKata.Compilers.PostgresCompiler.
func NewPostgresCompiler() *PostgresCompiler {
	return &PostgresCompiler{
		Compiler: &Compiler{
			OpeningIdentifier: "\"",
			ClosingIdentifier: "\"",
			ColumnAsKeyword:   "AS ",
			TableAsKeyword:    "AS ",
			LastID:            "SELECT lastval() AS id",
			EngineCode:        sqlkata.EnginePostgres,
			SupportsFilter:    true,
			placeholder:       "?",
			escape:            "\\",
			paramPrefix:       "@p",
			limitStyle:        limitLimitOffset,
			randomSQL:         "RANDOM()",
		},
	}
}
