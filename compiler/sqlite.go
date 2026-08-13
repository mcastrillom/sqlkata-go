package compiler

import "github.com/mcastrillom/sqlkata-go/sqlkata"

// SqliteCompiler compiles queries for SQLite (SqlKata.SqliteCompiler).
type SqliteCompiler struct {
	*Compiler
}

// NewSqliteCompiler returns defaults matching SqlKata.Compilers.SqliteCompiler.
func NewSqliteCompiler() *SqliteCompiler {
	return &SqliteCompiler{
		Compiler: &Compiler{
			OpeningIdentifier: "\"",
			ClosingIdentifier: "\"",
			ColumnAsKeyword:   "AS ",
			TableAsKeyword:    "AS ",
			LastID:            "select last_insert_rowid() as id",
			EngineCode:        sqlkata.EngineSqlite,
			placeholder:       "?",
			escape:            "\\",
			paramPrefix:       "@p",
			limitStyle:        limitLimitOffset,
			randomSQL:         "RANDOM()",
			// SQLite reads -1 as "no limit", the documented way to use OFFSET alone.
			unboundedLimit: "-1",
		},
	}
}
