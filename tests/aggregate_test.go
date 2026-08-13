package tests

import (
	"testing"

	"github.com/mcastrillom/sqlkata-go/sqlkata"
)

func TestCount(t *testing.T) {
	s := newTestSupport()
	q := sqlkata.NewQuery().From("A").AsCount()
	c := s.Compile(q)

	assertEqual(t, "SELECT COUNT(*) AS [count] FROM [A]", c[sqlkata.EngineSqlServer])
	assertEqual(t, `SELECT COUNT(*) AS "count" FROM "A"`, c[sqlkata.EnginePostgres])
	assertEqual(t, `SELECT COUNT(*) "count" FROM "A"`, c[sqlkata.EngineOracle])
}

func TestAverage(t *testing.T) {
	s := newTestSupport()
	q := sqlkata.NewQuery().From("A").AsAvg("TTL")
	c := s.Compile(q)

	assertEqual(t, "SELECT AVG([TTL]) AS [avg] FROM [A]", c[sqlkata.EngineSqlServer])
}

func TestSum(t *testing.T) {
	s := newTestSupport()
	q := sqlkata.NewQuery().From("A").AsSum("PacketsDropped")
	c := s.Compile(q)

	assertEqual(t, "SELECT SUM([PacketsDropped]) AS [sum] FROM [A]", c[sqlkata.EngineSqlServer])
}

func TestMax(t *testing.T) {
	s := newTestSupport()
	q := sqlkata.NewQuery().From("A").AsMax("LatencyMs")
	c := s.Compile(q)

	assertEqual(t, "SELECT MAX([LatencyMs]) AS [max] FROM [A]", c[sqlkata.EngineSqlServer])
}

func TestMin(t *testing.T) {
	s := newTestSupport()
	q := sqlkata.NewQuery().From("A").AsMin("LatencyMs")
	c := s.Compile(q)

	assertEqual(t, "SELECT MIN([LatencyMs]) AS [min] FROM [A]", c[sqlkata.EngineSqlServer])
}

func TestSelectSumWithFilter(t *testing.T) {
	s := newTestSupport()
	q := sqlkata.NewQuery().From("Orders").
		Select("UserId").
		SelectSum("Total", func(q *sqlkata.Query) *sqlkata.Query {
			return q.WhereEq("Paid", true)
		}).
		GroupBy("UserId")
	c := s.Compile(q)

	assertEqual(t,
		"SELECT [UserId], SUM(CASE WHEN [Paid] = cast(1 as bit) THEN [Total] END) FROM [Orders] GROUP BY [UserId]",
		c[sqlkata.EngineSqlServer])
	assertEqual(t,
		`SELECT "UserId", SUM("Total") FILTER (WHERE "Paid" = true) FROM "Orders" GROUP BY "UserId"`,
		c[sqlkata.EnginePostgres])
}
