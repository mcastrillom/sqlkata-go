package tests

import (
	"testing"

	"github.com/mcastrillom/sqlkata-go/sqlkata"
)

func TestDefineWhere(t *testing.T) {
	s := newTestSupport()
	q := sqlkata.NewQuery().From("Products").
		Define("@name", "Anto").
		Where("ProductName", "=", sqlkata.VariableExpr("@name"))
	c := s.Compile(q)

	assertEqual(t, "SELECT * FROM [Products] WHERE [ProductName] = 'Anto'", c[sqlkata.EngineSqlServer])
}

func TestDefineSubQuery(t *testing.T) {
	s := newTestSupport()
	sub := sqlkata.NewQuery().From("Products").
		AsAvg("unitprice").
		Define("@UnitsInSt", 10).
		Where("UnitsInStock", ">", sqlkata.VariableExpr("@UnitsInSt"))
	q := sqlkata.NewQuery().From("Products").
		WhereQuery("unitprice", ">", sub).
		Where("UnitsOnOrder", ">", 5)
	c := s.Compile(q)

	assertEqual(t,
		"SELECT * FROM [Products] WHERE [unitprice] > (SELECT AVG([unitprice]) AS [avg] FROM [Products] WHERE [UnitsInStock] > 10) AND [UnitsOnOrder] > 5",
		c[sqlkata.EngineSqlServer])
}

func TestDefineWhereEnds(t *testing.T) {
	s := newTestSupport()
	q1 := sqlkata.NewQuery().From("Products").
		Select("ProductId").
		Define("@product", "Coffee").
		WhereEnds("ProductName", sqlkata.VariableExpr("@product"), false, "")
	q2 := sqlkata.NewQuery().From("Products").
		Select("ProductId", "ProductName").
		Define("@product", "Coffee").
		WhereEnds("ProductName", sqlkata.VariableExpr("@product"), true, "")

	c1 := s.Compile(q1)
	c2 := s.Compile(q2)

	assertEqual(t, "SELECT [ProductId] FROM [Products] WHERE LOWER([ProductName]) like '%coffee'", c1[sqlkata.EngineSqlServer])
	assertEqual(t, "SELECT [ProductId], [ProductName] FROM [Products] WHERE [ProductName] like '%Coffee'", c2[sqlkata.EngineSqlServer])
}

func TestDefineWhereStarts(t *testing.T) {
	s := newTestSupport()
	q := sqlkata.NewQuery().From("Products").
		Select("ProductId").
		Define("@perUnit", "12").
		WhereStarts("QuantityPerUnit", sqlkata.VariableExpr("@perUnit"), false, "")
	c := s.Compile(q)

	assertEqual(t, "SELECT [ProductId] FROM [Products] WHERE LOWER([QuantityPerUnit]) like '12%'", c[sqlkata.EngineSqlServer])
}

func TestUnsafeLiteral(t *testing.T) {
	s := newTestSupport()
	q := sqlkata.NewQuery().From("Users").
		Select("Id").
		Where("Status", "=", sqlkata.UnsafeLiteralExpr("'Active'", false))
	c := s.Compile(q)

	assertEqual(t, "SELECT [Id] FROM [Users] WHERE [Status] = 'Active'", c[sqlkata.EngineSqlServer])
}
