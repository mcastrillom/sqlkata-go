package execution_test

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/mcastrillom/sqlkata-go/compiler"
	"github.com/mcastrillom/sqlkata-go/execution"
	"github.com/mcastrillom/sqlkata-go/sqlkata"
)

type userRow struct {
	ID   int    `db:"Id"`
	Name string `db:"Name"`
}

func newFactory(t *testing.T) (*execution.QueryFactory, sqlmock.Sqlmock, func()) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatal(err)
	}
	xdb := sqlx.NewDb(db, "sqlmock")
	f := execution.New(xdb, compiler.NewSqlServerCompiler())
	f.Logger = func(r *compiler.SqlResult) {
		t.Logf("SQL=%s Bindings=%v", r.SQL(), r.Bindings)
	}
	return f, mock, func() { _ = xdb.Close() }
}

func TestQueryFactoryGet(t *testing.T) {
	f, mock, done := newFactory(t)
	defer done()

	mock.ExpectQuery(`(?i)SELECT.*FROM.*Users`).
		WithArgs("Active").
		WillReturnRows(sqlmock.NewRows([]string{"Id", "Name"}).
			AddRow(1, "Ada").
			AddRow(2, "Grace"))

	q := f.Query("Users").Select("Id", "Name").WhereEq("Status", "Active")
	rows, err := execution.GetSlice[userRow](context.Background(), f, q)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 || rows[0].Name != "Ada" {
		t.Fatalf("rows=%v", rows)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestQueryFactoryFirst(t *testing.T) {
	f, mock, done := newFactory(t)
	defer done()

	mock.ExpectQuery(`(?i)SELECT.*FROM.*Users`).
		WithArgs(1, 0, 1).
		WillReturnRows(sqlmock.NewRows([]string{"Id", "Name"}).AddRow(1, "Ada"))

	q := f.Query("Users").Select("Id", "Name").WhereEq("Id", 1)
	u, err := execution.FirstValue[userRow](context.Background(), f, q)
	if err != nil {
		t.Fatal(err)
	}
	if u.Name != "Ada" {
		t.Fatalf("got %+v", u)
	}
}

func TestQueryFactoryFirstNoRows(t *testing.T) {
	f, mock, done := newFactory(t)
	defer done()

	mock.ExpectQuery(`(?i)SELECT.*FROM.*Users`).
		WillReturnRows(sqlmock.NewRows([]string{"Id", "Name"}))

	q := f.Query("Users").Select("Id", "Name")
	_, err := execution.FirstValue[userRow](context.Background(), f, q)
	if err != execution.ErrNoRows {
		t.Fatalf("err=%v", err)
	}
}

func TestQueryFactoryExecute(t *testing.T) {
	f, mock, done := newFactory(t)
	defer done()

	mock.ExpectExec(`(?i)UPDATE.*Users`).
		WithArgs("Ada", 1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	q := f.Query("Users").WhereEq("Id", 1).AsUpdate(map[string]any{"Name": "Ada"})
	n, err := f.ExecuteAffected(context.Background(), q)
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("affected=%d", n)
	}
}

func TestQueryFactoryExists(t *testing.T) {
	f, mock, done := newFactory(t)
	defer done()

	mock.ExpectQuery(`(?i)SELECT.*1.*Exists`).
		WithArgs(1, 0, 1).
		WillReturnRows(sqlmock.NewRows([]string{"Exists"}).AddRow(1))

	q := f.Query("Users").WhereEq("Id", 1)
	ok, err := f.Exists(context.Background(), q)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("expected exists")
	}
}

func TestQueryFactoryCount(t *testing.T) {
	f, mock, done := newFactory(t)
	defer done()

	mock.ExpectQuery(`(?i)SELECT COUNT`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(42)))

	q := f.Query("Users")
	n, err := f.Count(context.Background(), q)
	if err != nil {
		t.Fatal(err)
	}
	if n != 42 {
		t.Fatalf("count=%d", n)
	}
}

func TestQueryFactoryPaginate(t *testing.T) {
	f, mock, done := newFactory(t)
	defer done()

	mock.ExpectQuery(`(?i)SELECT COUNT`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(3)))
	mock.ExpectQuery(`(?i)SELECT.*FROM.*Users`).
		WillReturnRows(sqlmock.NewRows([]string{"Id", "Name"}).
			AddRow(1, "A").
			AddRow(2, "B"))

	q := f.Query("Users").Select("Id", "Name").OrderBy("Id")
	page, err := execution.Paginate[userRow](context.Background(), f, q, 1, 2)
	if err != nil {
		t.Fatal(err)
	}
	if page.Count != 3 || len(page.List) != 2 || !page.HasNext() {
		t.Fatalf("%+v", page)
	}
}

func TestXQueryGet(t *testing.T) {
	f, mock, done := newFactory(t)
	defer done()

	mock.ExpectQuery(`(?i)SELECT.*FROM.*Users`).
		WillReturnRows(sqlmock.NewRows([]string{"Id", "Name"}).AddRow(1, "Ada"))

	xq := f.XQuery("Users")
	xq.Select("Id", "Name")
	var rows []userRow
	if err := xq.Get(context.Background(), &rows); err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("%v", rows)
	}
}

func TestWithTx(t *testing.T) {
	f, mock, done := newFactory(t)
	defer done()

	mock.ExpectBegin()
	mock.ExpectExec(`(?i)DELETE FROM.*Users`).
		WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectCommit()

	err := f.WithTx(context.Background(), func(tx *execution.QueryFactory) error {
		_, err := tx.ExecuteAffected(context.Background(), tx.Query("Users").AsDelete())
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestCompileLogger(t *testing.T) {
	f, _, done := newFactory(t)
	defer done()
	var logged bool
	f.Logger = func(r *compiler.SqlResult) {
		logged = true
		if r.RawSQL == "" {
			t.Fatal("empty RawSQL")
		}
	}
	if _, err := f.Compile(sqlkata.NewQuery().From("Users").Select("Id")); err != nil {
		t.Fatal(err)
	}
	if !logged {
		t.Fatal("logger not called")
	}
}
