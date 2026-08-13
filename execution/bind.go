package execution

import (
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/mcastrillom/sqlkata-go/compiler"
	"github.com/mcastrillom/sqlkata-go/sqlkata"
)

// Placeholder styles for QueryFactory (sqlx bind types + Oracle positional).
const (
	// BindAuto picks style from compiler engine, then driver name.
	BindAuto = -1
	// BindOraclePositional rewrites ? → :1, :2, … (godror / Oracle database/sql).
	BindOraclePositional = 100
)

func init() {
	// Ensure common Oracle driver names use named binds if Rebind is used.
	for _, name := range []string{"godror", "oracle", "oci8", "ora", "goracle"} {
		sqlx.BindDriver(name, sqlx.NAMED)
	}
}

func engineCode(c QueryCompiler) string {
	switch x := c.(type) {
	case *compiler.OracleCompiler:
		if x.Compiler != nil {
			return x.EngineCode
		}
		return sqlkata.EngineOracle
	case *compiler.SqlServerCompiler:
		if x.Compiler != nil {
			return x.EngineCode
		}
		return sqlkata.EngineSqlServer
	case *compiler.PostgresCompiler:
		if x.Compiler != nil {
			return x.EngineCode
		}
		return sqlkata.EnginePostgres
	case *compiler.MySqlCompiler:
		if x.Compiler != nil {
			return x.EngineCode
		}
		return sqlkata.EngineMySql
	case *compiler.SqliteCompiler:
		if x.Compiler != nil {
			return x.EngineCode
		}
		return sqlkata.EngineSqlite
	case *compiler.Compiler:
		return x.EngineCode
	default:
		type engineCoder interface{ GetEngineCode() string }
		if e, ok := c.(engineCoder); ok {
			return e.GetEngineCode()
		}
		return ""
	}
}

func (f *QueryFactory) resolveBindStyle() int {
	// Explicit override (including sqlx.QUESTION=1). BindAuto=-1 means detect.
	if f.BindStyle != BindAuto {
		if f.BindStyle == 0 {
			// legacy zero value → auto
		} else {
			return f.BindStyle
		}
	}
	switch engineCode(f.Compiler) {
	case sqlkata.EngineOracle:
		return BindOraclePositional
	case sqlkata.EnginePostgres:
		return sqlx.DOLLAR
	case sqlkata.EngineMySql, sqlkata.EngineSqlite:
		return sqlx.QUESTION
	case sqlkata.EngineSqlServer:
		if name := f.driverName(); name != "" {
			if t := sqlx.BindType(name); t != sqlx.UNKNOWN {
				return t
			}
		}
		return sqlx.QUESTION
	}
	if name := f.driverName(); name != "" {
		if t := sqlx.BindType(name); t != sqlx.UNKNOWN {
			return t
		}
	}
	return sqlx.QUESTION
}

func (f *QueryFactory) driverName() string {
	if f.DB != nil {
		return f.DB.DriverName()
	}
	if f.Tx != nil {
		return f.Tx.DriverName()
	}
	return ""
}

// rebindQuery converts RawSQL "?" placeholders to the target dialect.
func rebindQuery(style int, query string) string {
	switch style {
	case sqlx.QUESTION, sqlx.UNKNOWN:
		return query
	case BindOraclePositional:
		return rebindOraclePositional(query)
	default:
		return sqlx.Rebind(style, query)
	}
}

// rebindOraclePositional turns ? into :1, :2, … (Oracle / godror positional binds).
// sqlx.NAMED would produce :arg1 which some stacks handle poorly with []any args.
func rebindOraclePositional(query string) string {
	if !strings.Contains(query, "?") {
		return query
	}
	var b strings.Builder
	b.Grow(len(query) + 8)
	n := 1
	for i := 0; i < len(query); i++ {
		if query[i] == '?' {
			b.WriteByte(':')
			b.WriteString(strconv.Itoa(n))
			n++
			continue
		}
		b.WriteByte(query[i])
	}
	return b.String()
}
