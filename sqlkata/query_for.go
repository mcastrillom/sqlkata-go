package sqlkata

// Dialect-scoped helpers mirroring SqlKata.Extensions.QueryForExtensions.
// Clauses added inside the callback are tagged with that engine code and only
// compile when the matching Compiler is used.

// ForFirebird scopes clauses for EngineFirebird.
func (q *Query) ForFirebird(fn func(*Query) *Query) *Query {
	return q.For(EngineFirebird, fn)
}

// ForMySql scopes clauses for EngineMySql.
func (q *Query) ForMySql(fn func(*Query) *Query) *Query {
	return q.For(EngineMySql, fn)
}

// ForOracle scopes clauses for EngineOracle.
func (q *Query) ForOracle(fn func(*Query) *Query) *Query {
	return q.For(EngineOracle, fn)
}

// ForPostgreSql scopes clauses for EnginePostgres (SqlKata.ForPostgreSql).
func (q *Query) ForPostgreSql(fn func(*Query) *Query) *Query {
	return q.For(EnginePostgres, fn)
}

// ForPostgres is an alias of ForPostgreSql.
func (q *Query) ForPostgres(fn func(*Query) *Query) *Query {
	return q.ForPostgreSql(fn)
}

// ForSqlite scopes clauses for EngineSqlite.
func (q *Query) ForSqlite(fn func(*Query) *Query) *Query {
	return q.For(EngineSqlite, fn)
}

// ForSqlServer scopes clauses for EngineSqlServer.
func (q *Query) ForSqlServer(fn func(*Query) *Query) *Query {
	return q.For(EngineSqlServer, fn)
}
