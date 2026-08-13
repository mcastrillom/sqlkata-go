package execution

import (
	"context"
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/mcastrillom/sqlkata-go/compiler"
	"github.com/mcastrillom/sqlkata-go/sqlkata"
)

// QueryCompiler compiles a Query to SqlResult (SqlKata.Compiler surface used by Execution).
type QueryCompiler interface {
	Compile(q *sqlkata.Query) (*compiler.SqlResult, error)
}

// QueryFactory mirrors SqlKata.Execution.QueryFactory: connection + compiler + execute.
type QueryFactory struct {
	DB           *sqlx.DB
	Tx           *sqlx.Tx // optional; when set, commands run inside this transaction
	Compiler     QueryCompiler
	QueryTimeout time.Duration
	Logger       func(*compiler.SqlResult)
	// BindStyle selects placeholder rewriting for RawSQL ("?").
	// BindAuto (default): Oracle → :1,:2; Postgres → $1; SqlServer → driver / ?.
	// Override with sqlx.QUESTION, sqlx.DOLLAR, sqlx.AT, sqlx.NAMED, or BindOraclePositional.
	BindStyle int
}

// New creates a QueryFactory (SqlKata: new QueryFactory(connection, compiler, timeout)).
func New(db *sqlx.DB, comp QueryCompiler, timeout ...time.Duration) *QueryFactory {
	t := 30 * time.Second
	if len(timeout) > 0 && timeout[0] > 0 {
		t = timeout[0]
	}
	return &QueryFactory{
		DB:           db,
		Compiler:     comp,
		QueryTimeout: t,
		Logger:       func(*compiler.SqlResult) {},
		BindStyle:    BindAuto,
	}
}

// NewFromStd wraps database/sql.DB with a driver name (sqlx.NewDb).
func NewFromStd(db *sql.DB, driverName string, comp QueryCompiler, timeout ...time.Duration) *QueryFactory {
	return New(sqlx.NewDb(db, driverName), comp, timeout...)
}

// Query starts a new select query, optionally From(table) (SqlKata.QueryFactory.Query).
func (f *QueryFactory) Query(table ...string) *sqlkata.Query {
	q := sqlkata.NewQuery()
	if len(table) > 0 && table[0] != "" {
		q.From(table[0])
	}
	return q
}

// FromQuery returns a clone of q for execution against this factory (SqlKata.FromQuery).
func (f *QueryFactory) FromQuery(q *sqlkata.Query) *sqlkata.Query {
	if q == nil {
		return sqlkata.NewQuery()
	}
	return q.Clone()
}

// Compile compiles and logs (SqlKata CompileAndLog).
func (f *QueryFactory) Compile(q *sqlkata.Query) (*compiler.SqlResult, error) {
	res, err := f.Compiler.Compile(q)
	if err != nil {
		return nil, err
	}
	if f.Logger != nil {
		f.Logger(res)
	}
	return res, nil
}

// ext returns the active sqlx.ExtContext (Tx preferred).
func (f *QueryFactory) ext() sqlx.ExtContext {
	if f.Tx != nil {
		return f.Tx
	}
	return f.DB
}

func (f *QueryFactory) withTimeout(parent context.Context) (context.Context, context.CancelFunc) {
	if parent == nil {
		parent = context.Background()
	}
	if f.QueryTimeout <= 0 {
		return parent, func() {}
	}
	// Honor parent deadline if tighter.
	if deadline, ok := parent.Deadline(); ok {
		if time.Until(deadline) <= f.QueryTimeout {
			return parent, func() {}
		}
	}
	return context.WithTimeout(parent, f.QueryTimeout)
}

// bind prepares driver-specific SQL + args from a compiled result.
// Always starts from RawSQL ("?") and rewrites placeholders from the compiler engine
// (not only sqlx.DB.Rebind), so Oracle never keeps "?" (ORA-00911).
func (f *QueryFactory) bind(res *compiler.SqlResult) (string, []any) {
	query := res.RawSQL
	if query == "" {
		query = res.SQL()
	}
	style := f.resolveBindStyle()
	query = rebindQuery(style, query)
	return query, res.Bindings
}

// Option configures a single execution call.
type Option func(*execConfig)

type execConfig struct {
	tx *sqlx.Tx
}

// WithTx runs this call on the given transaction (overrides factory.Tx for the call).
func WithTx(tx *sqlx.Tx) Option {
	return func(c *execConfig) { c.tx = tx }
}

func (f *QueryFactory) resolveExt(opts []Option) sqlx.ExtContext {
	cfg := &execConfig{}
	for _, o := range opts {
		o(cfg)
	}
	if cfg.tx != nil {
		return cfg.tx
	}
	return f.ext()
}

// WithTx runs fn inside a transaction bound to a child QueryFactory (commit/rollback).
func (f *QueryFactory) WithTx(ctx context.Context, fn func(txFactory *QueryFactory) error) (err error) {
	if f.DB == nil {
		return sql.ErrConnDone
	}
	if ctx == nil {
		ctx = context.Background()
	}
	tx, err := f.DB.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	child := &QueryFactory{
		DB:           f.DB,
		Tx:           tx,
		Compiler:     f.Compiler,
		QueryTimeout: f.QueryTimeout,
		Logger:       f.Logger,
		BindStyle:    f.BindStyle,
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		} else if err != nil {
			_ = tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()
	err = fn(child)
	return err
}
