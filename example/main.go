package main

import (
	"fmt"

	"github.com/mcastrillom/sqlkata-go/compiler"
	"github.com/mcastrillom/sqlkata-go/sqlkata"
)

func main() {
	c := compiler.NewSqlServerCompiler()

	// --- dialects (limit/offset/order) ---
	qPage := sqlkata.NewQuery().
		From("Users").
		Select("ID", "NAME as FullName").
		OrderBy("NAME").
		OrderByDesc("ID").
		Limit(10).
		Offset(20)
	printDialect("SQL Server page", mustCompile(compiler.NewSqlServerCompiler(), qPage))
	printDialect("Oracle page", mustCompile(compiler.NewOracleCompiler(), qPage))
	printDialect("PostgreSQL page", mustCompile(compiler.NewPostgresCompiler(), qPage))

	// --- WHERE (conditions) ---
	qWhere := sqlkata.NewQuery().
		From("Users").
		Select("Id", "Name").
		Where("Status", "=", "Active").
		OrWhereEq("Role", "Admin").
		WhereNotNull("Email").
		WhereIn("Country", "MX", "US", "CA").
		WhereBetween("Age", 18, 65).
		WhereRaw("[Score] > ?", 10).
		WhereNested(func(q *sqlkata.Query) *sqlkata.Query {
			return q.WhereEq("Dept", "IT").OrWhereEq("Dept", "HR")
		})
	printExample("WHERE (basic / in / between / raw / nested)", mustCompile(c, qWhere))

	qWhereStr := sqlkata.NewQuery().
		From("Users").
		Select("Id", "Name").
		WhereStarts("Name", "A", false, "").
		OrWhereContains("Email", "@example.com", false, "").
		WhereLike("Code", "X_Y", true, "\\").
		WhereDate("CreatedAt", ">=", "2024-01-01").
		WhereDatePart("year", "CreatedAt", "=", 2024)
	printExample("WHERE like/starts/contains/date", mustCompile(c, qWhereStr))

	subExists := sqlkata.NewQuery().From("Orders").WhereColumns("Orders.UserId", "=", "Users.Id")
	qExists := sqlkata.NewQuery().
		From("Users").
		Select("Id").
		WhereExists(subExists).
		WhereInQuery("Id", sqlkata.NewQuery().From("VIP").Select("UserId")).
		WhereQuery("Id", ">", sqlkata.NewQuery().From("Thresholds").Select("MinId")).
		WhereSub(sqlkata.NewQuery().From("Stats").AsCount(), ">", 0)
	printExample("WHERE EXISTS / IN / Query / Sub", mustCompile(c, qExists))

	// --- JOIN ---
	qJoin := sqlkata.NewQuery().
		From("Users as u").
		Select("u.Id", "o.Total").
		LeftJoin("Orders as o", "o.UserId", "u.Id").
		Where("u.Active", "=", true)
	printExample("LEFT JOIN + WHERE bool", mustCompile(c, qJoin))

	qJoinCB := sqlkata.NewQuery().
		From("Users").
		Select("Users.Id", "Profiles.Bio").
		JoinWith("Profiles", func(j *sqlkata.Join) *sqlkata.Join {
			return j.On("Profiles.UserId", "Users.Id").OrOn("Profiles.AltId", "Users.Id")
		}, "inner join")
	printExample("JOIN with callback (On / OrOn)", mustCompile(c, qJoinCB))

	qDeep := sqlkata.NewQuery().
		From("Books").
		Select("Books.Title", "Author.Name", "Country.Name").
		LeftDeepJoin("Author.Country")
	printExample("LeftDeepJoin Author.Country", mustCompile(c, qDeep))

	qDeepKeys := sqlkata.NewQuery().
		From("Posts as p").
		Select("p.Id", "c.Body").
		DeepJoin("Comment as c", sqlkata.DeepJoinGenerators(
			func(string) string { return "Id" },     // Posts.Id
			func(string) string { return "PostId" }, // Comment.PostId
		))
	printExample("DeepJoin custom generators", mustCompile(c, qDeepKeys))

	// Include/IncludeMany are metadata for execution layers (not emitted as SQL).
	qInclude := sqlkata.NewQuery().
		From("Users").
		Select("Id", "Name").
		Include("Profile", sqlkata.NewQuery().From("Profiles").Select("UserId", "Bio"), "UserId", "Id").
		IncludeMany("Orders", sqlkata.NewQuery().From("Orders").Select("UserId", "Total"), "UserId", "Id")
	printExample("Include/IncludeMany (SQL unchanged; Includes on query)", mustCompile(c, qInclude))
	fmt.Printf("Includes count: %d (first=%s isMany=%v, second=%s isMany=%v)\n\n",
		len(qInclude.Includes),
		qInclude.Includes[0].Name, qInclude.Includes[0].IsMany,
		qInclude.Includes[1].Name, qInclude.Includes[1].IsMany,
	)

	// --- UNION / COMBINE ---
	qUnion := sqlkata.NewQuery().From("A").Select("Id").
		UnionAll(sqlkata.NewQuery().From("B").Select("Id")).
		Union(sqlkata.NewQuery().From("C").Select("Id"))
	printExample("UNION ALL / UNION", mustCompile(c, qUnion))

	// --- AGGREGATE ---
	qAgg := sqlkata.NewQuery().From("Orders").AsCount()
	printExample("AsCount (AggregateClause)", mustCompile(c, qAgg))

	qSelAgg := sqlkata.NewQuery().
		From("Orders").
		Select("UserId").
		SelectAggregate("sum", "Total").
		OrderBy("UserId")
	printExample("SelectAggregate (AggregatedColumn)", mustCompile(c, qSelAgg))

	qSubCol := sqlkata.NewQuery().
		From("Users").
		Select("Id").
		SelectQuery(sqlkata.NewQuery().From("Orders").AsCount(), "OrderCount")
	printExample("SelectQuery (QueryColumn)", mustCompile(c, qSubCol))

	// --- GROUP BY + HAVING ---
	qGroup := sqlkata.NewQuery().
		From("Orders").
		Select("UserId").
		SelectAggregate("sum", "Total").
		GroupBy("UserId").
		Having("SUM(Total)", ">", 100).
		OrHavingBetween("SUM(Total)", 200, 1000).
		HavingNested(func(q *sqlkata.Query) *sqlkata.Query {
			return q.HavingEq("COUNT(*)", 1).OrHavingEq("COUNT(*)", 2)
		}).
		OrderBy("UserId")
	printExample("GROUP BY + HAVING", mustCompile(c, qGroup))

	qGroupRaw := sqlkata.NewQuery().
		From("Orders").
		SelectRaw("YEAR(CreatedAt) as Y").
		GroupByRaw("YEAR(CreatedAt)").
		HavingRaw("COUNT(*) >= ?", 5)
	printExample("GROUP BY RAW + HAVING RAW", mustCompile(c, qGroupRaw))

	// --- UPDATE / DELETE ---
	qUpd := sqlkata.NewQuery().
		From("Users").
		WhereEq("Id", 42).
		WhereIn("Id", 1, 2, 3).
		AsUpdate(map[string]any{
			"Name":   "Ada",
			"Active": true,
		})
	printExample("UPDATE (AsUpdate)", mustCompile(c, qUpd))

	qInc := sqlkata.NewQuery().
		From("Accounts").
		WhereEq("Id", 1).
		AsIncrement("Balance", 50)
	printExample("UPDATE increment", mustCompile(c, qInc))

	qDec := sqlkata.NewQuery().
		From("Accounts").
		WhereEq("Id", 1).
		AsDecrement("Balance", 10)
	printExample("UPDATE decrement", mustCompile(c, qDec))

	qDel := sqlkata.NewQuery().
		From("Users").
		WhereEq("Id", 42).
		AsDelete()
	printExample("DELETE", mustCompile(c, qDel))

	qDelJoin := sqlkata.NewQuery().
		From("Users as u").
		LeftJoin("Orders as o", "o.UserId", "u.Id").
		WhereNull("o.Id").
		AsDelete()
	printExample("DELETE with JOIN", mustCompile(c, qDelJoin))

	// --- INSERT ---
	qIns := sqlkata.NewQuery().
		From("Users").
		Insert(map[string]any{
			"Name":   "Ada",
			"Email":  "ada@example.com",
			"Active": true,
		})
	printExample("INSERT (InsertClause)", mustCompile(c, qIns))

	// --- CTE (WITH) ---
	qCte := sqlkata.NewQuery().
		WithAlias("ActiveUsers", sqlkata.NewQuery().From("Users").Select("Id", "Name").WhereEq("Active", true)).
		From("ActiveUsers").
		Select("Id", "Name").
		OrderBy("Name")
	printExample("CTE WithAlias", mustCompile(c, qCte))

	qCteRaw := sqlkata.NewQuery().
		WithRaw("Filtered", "SELECT Id FROM Users WHERE Status = ?", "Active").
		From("Filtered").
		Select("Id")
	printExample("CTE WithRaw", mustCompile(c, qCteRaw))

	qCteVals := sqlkata.NewQuery().
		WithValues("Codes", []string{"Code", "Label"}, [][]any{
			{"A", "Alpha"},
			{"B", "Beta"},
		}).
		From("Codes").
		Select("Code", "Label")
	printExample("CTE WithValues (AdHoc)", mustCompile(c, qCteVals))
	printExample("CTE WithValues on Postgres", mustCompile(compiler.NewPostgresCompiler(), qCteVals))

	// --- When / For / ForPage / Comment ---
	activeOnly := true
	qWhen := sqlkata.NewQuery().
		From("Users").
		Select("Id", "Name").
		Comment("active users page").
		When(activeOnly, func(q *sqlkata.Query) *sqlkata.Query {
			return q.WhereEq("Active", true)
		}).
		ForPage(2, 10)
	printExample("When + ForPage + Comment", mustCompile(c, qWhen))

	qFor := sqlkata.NewQuery().
		From("Users").
		Select("Id").
		ForSqlServer(func(q *sqlkata.Query) *sqlkata.Query {
			return q.WhereRaw("[LegacyFlag] = 1")
		}).
		ForOracle(func(q *sqlkata.Query) *sqlkata.Query {
			return q.WhereRaw(`"LegacyFlag" = 1`)
		}).
		ForPostgreSql(func(q *sqlkata.Query) *sqlkata.Query {
			return q.WhereRaw(`"LegacyFlag" = 1`)
		})
	printExample("ForSqlServer scoped clause", mustCompile(c, qFor))
	printExample("ForOracle scoped clause", mustCompile(compiler.NewOracleCompiler(), qFor))
	printExample("ForPostgreSql scoped clause", mustCompile(compiler.NewPostgresCompiler(), qFor))

	// --- Select aggregates with filter + expand ---
	qSel := sqlkata.NewQuery().
		From("Orders").
		Select("UserId").
		Select("Orders.{Total, Qty}").
		SelectSum("Total", func(q *sqlkata.Query) *sqlkata.Query {
			return q.WhereEq("Paid", true)
		}).
		GroupBy("UserId")
	printExample("Select expand + SelectSum(filter)", mustCompile(c, qSel))
	printExample("SelectSum(filter) on Postgres (FILTER)", mustCompile(compiler.NewPostgresCompiler(), qSel))

	// --- Insert multi / query / returnId ---
	qInsMulti := sqlkata.NewQuery().
		From("Users").
		AsInsertRows([]string{"Name", "Email"}, [][]any{
			{"Ada", "ada@example.com"},
			{"Grace", "grace@example.com"},
		})
	printExample("INSERT multi-row", mustCompile(c, qInsMulti))

	qInsSel := sqlkata.NewQuery().
		From("UsersArchive").
		AsInsertQuery([]string{"Id", "Name"}, sqlkata.NewQuery().From("Users").Select("Id", "Name").WhereEq("Active", false))
	printExample("INSERT…SELECT", mustCompile(c, qInsSel))

	qInsRet := sqlkata.NewQuery().
		From("Users").
		AsInsert(map[string]any{"Name": "Ada"}, true)
	printExample("INSERT ReturnId", mustCompile(c, qInsRet))

	// --- Variable + UnsafeLiteral ---
	qVar := sqlkata.NewQuery().
		From("Users").
		Define("minAge", 21).
		Select("Id").
		Where("Age", ">=", sqlkata.VariableExpr("minAge")).
		Where("Status", "=", sqlkata.UnsafeLiteralExpr("'Active'", false))
	printExample("Variable + UnsafeLiteral", mustCompile(c, qVar))

	// date dialect sample
	qDate := sqlkata.NewQuery().From("Users").Select("Id").WhereDate("CreatedAt", ">=", "2024-01-01")
	printExample("WhereDate SQL Server", mustCompile(compiler.NewSqlServerCompiler(), qDate))
	printExample("WhereDate Postgres", mustCompile(compiler.NewPostgresCompiler(), qDate))
	printExample("WhereDate Oracle", mustCompile(compiler.NewOracleCompiler(), qDate))
}

func mustCompile(c QueryCompiler, q *sqlkata.Query) *compiler.SqlResult {
	r, err := c.Compile(q)
	if err != nil {
		panic(err)
	}
	return r
}

func printDialect(name string, r *compiler.SqlResult) {
	fmt.Println("---", name, "---")
	fmt.Println("SQL:", r.SQL())
	fmt.Println()
}

func printExample(name string, r *compiler.SqlResult) {
	fmt.Println("---", name, "---")
	fmt.Println("SQL:", r.SQL())
	fmt.Println("Bindings:", r.Bindings)
	fmt.Println("Interpolated:", r.String())
	fmt.Println()
}

type QueryCompiler interface {
	Compile(q *sqlkata.Query) (*compiler.SqlResult, error)
}
