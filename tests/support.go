package tests

import (
	"fmt"

	"github.com/mcastrillom/sqlkata-go/compiler"
	"github.com/mcastrillom/sqlkata-go/sqlkata"
)

// TestCompilersContainer mirrors SqlKata.Tests.Infrastructure.TestCompilersContainer
// for the dialects implemented in this port.
type TestCompilersContainer struct {
	compilers map[string]*compiler.Compiler
}

// allEngines is the set compiled when a test does not name specific engines.
var allEngines = []string{
	sqlkata.EngineSqlServer,
	sqlkata.EnginePostgres,
	sqlkata.EngineOracle,
	sqlkata.EngineMySql,
	sqlkata.EngineSqlite,
}

// NewTestCompilersContainer builds the default compiler set.
// SqlServer uses modern OFFSET/FETCH (UseLegacyPagination=false), matching
// SqlKata.SqlServerLimitTests rather than the legacy-default TestCompilersContainer.
func NewTestCompilersContainer() *TestCompilersContainer {
	return &TestCompilersContainer{
		compilers: map[string]*compiler.Compiler{
			sqlkata.EngineSqlServer: NewSqlServerCompiler().Compiler,
			sqlkata.EnginePostgres:  NewPostgresCompiler().Compiler,
			sqlkata.EngineOracle:    NewOracleCompiler().Compiler,
			sqlkata.EngineMySql:     NewMySqlCompiler().Compiler,
			sqlkata.EngineSqlite:    NewSqliteCompiler().Compiler,
		},
	}
}

// Get returns the compiler for an engine code.
func (c *TestCompilersContainer) Get(engineCode string) *compiler.Compiler {
	comp, ok := c.compilers[engineCode]
	if !ok {
		panic(fmt.Sprintf("Engine code '%s' is not valid", engineCode))
	}
	return comp
}

// mustCompile compiles or panics (test helper).
func mustCompile(comp *compiler.Compiler, query *sqlkata.Query) *compiler.SqlResult {
	res, err := comp.Compile(query)
	if err != nil {
		panic(err)
	}
	return res
}

// CompileFor compiles against a single engine.
func (c *TestCompilersContainer) CompileFor(engineCode string, query *sqlkata.Query) *compiler.SqlResult {
	return mustCompile(c.Get(engineCode), query)
}

// Compile compiles against the given engines (or all known engines when empty).
func (c *TestCompilersContainer) Compile(query *sqlkata.Query, engineCodes ...string) TestSqlResultContainer {
	codes := engineCodes
	if len(codes) == 0 {
		codes = allEngines
	}
	out := make(TestSqlResultContainer, len(codes))
	for _, code := range codes {
		comp, ok := c.compilers[code]
		if !ok {
			panic(fmt.Sprintf("Invalid engine codes supplied '%s'", code))
		}
		out[code] = mustCompile(comp, query.Clone())
	}
	return out
}

// CompileStrings is the legacy helper: engine → interpolated SQL (SqlResult.ToString()).
func (c *TestCompilersContainer) CompileStrings(query *sqlkata.Query, engineCodes ...string) map[string]string {
	results := c.Compile(query, engineCodes...)
	out := make(map[string]string, len(results))
	for k, v := range results {
		out[k] = v.String()
	}
	return out
}

// TestSqlResultContainer mirrors SqlKata.Tests.Infrastructure.TestSqlResultContainer.
type TestSqlResultContainer map[string]*compiler.SqlResult

// TestSupport embeds the compilers container like SqlKata.Tests.Infrastructure.TestSupport.
type TestSupport struct {
	Compilers *TestCompilersContainer
}

func newTestSupport() TestSupport {
	return TestSupport{Compilers: NewTestCompilersContainer()}
}

// Compile returns interpolated SQL per engine (SqlKata TestSupport.Compile).
func (t TestSupport) Compile(query *sqlkata.Query) map[string]string {
	return t.Compilers.CompileStrings(query)
}

// Thin wrappers so tests can use the same New* names as production code
// while keeping a single place to tweak defaults for the suite.
func NewSqlServerCompiler() *compiler.SqlServerCompiler {
	return compiler.NewSqlServerCompiler()
}

func NewPostgresCompiler() *compiler.PostgresCompiler {
	return compiler.NewPostgresCompiler()
}

func NewOracleCompiler() *compiler.OracleCompiler {
	return compiler.NewOracleCompiler()
}

func NewMySqlCompiler() *compiler.MySqlCompiler {
	return compiler.NewMySqlCompiler()
}

func NewSqliteCompiler() *compiler.SqliteCompiler {
	return compiler.NewSqliteCompiler()
}
