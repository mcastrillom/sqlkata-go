package sqlkata_test

import (
	"errors"
	"testing"

	"github.com/mcastrillom/sqlkata-go/compiler"
	"github.com/mcastrillom/sqlkata-go/sqlkata"
)

func TestAsInsertEmptySetsErr(t *testing.T) {
	q := sqlkata.NewQuery().From("t").AsInsert(map[string]any{}, false)
	if q.Err() == nil {
		t.Fatal("expected Err")
	}
	if _, err := compiler.NewSqlServerCompiler().Compile(q); err == nil {
		t.Fatal("expected Compile error")
	}
}

func TestMissingVariableCompileError(t *testing.T) {
	q := sqlkata.NewQuery().From("t").WhereEq("id", sqlkata.Variable{Name: "missing"})
	_, err := compiler.NewSqlServerCompiler().Compile(q)
	if err == nil {
		t.Fatal("expected missing variable error")
	}
	if !errors.Is(err, sqlkata.ErrVariableNotFound) {
		t.Fatalf("want ErrVariableNotFound, got %v", err)
	}
}

func TestFindVariableErrorIsSentinel(t *testing.T) {
	_, err := sqlkata.NewQuery().FindVariable("nope")
	if !errors.Is(err, sqlkata.ErrVariableNotFound) {
		t.Fatalf("want ErrVariableNotFound, got %v", err)
	}
}
