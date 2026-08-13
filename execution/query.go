package execution

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/mcastrillom/sqlkata-go/sqlkata"
)

// ErrNoRows is returned by First when the query matches no rows (SqlKata First throws).
var ErrNoRows = sql.ErrNoRows

// Get scans all rows into dest (must be *[]T or *[]map[string]any).
// Mirrors QueryFactory.Get<T>.
func (f *QueryFactory) Get(ctx context.Context, q *sqlkata.Query, dest any, opts ...Option) error {
	ctx, cancel := f.withTimeout(ctx)
	defer cancel()
	res, err := f.Compile(q)
	if err != nil {
		return err
	}
	query, args := f.bind(res)
	ext := f.resolveExt(opts)
	return sqlx.SelectContext(ctx, ext, dest, query, args...)
}

// GetSlice is a generic helper: rows, err := execution.GetSlice[User](ctx, db, q).
func GetSlice[T any](ctx context.Context, f *QueryFactory, q *sqlkata.Query, opts ...Option) ([]T, error) {
	var dest []T
	if err := f.Get(ctx, q, &dest, opts...); err != nil {
		return nil, err
	}
	return dest, nil
}

// FirstOrDefault scans at most one row into dest (*T). ok=false when no row.
func (f *QueryFactory) FirstOrDefault(ctx context.Context, q *sqlkata.Query, dest any, opts ...Option) (bool, error) {
	ctx, cancel := f.withTimeout(ctx)
	defer cancel()
	res, err := f.Compile(q.Clone().Limit(1))
	if err != nil {
		return false, err
	}
	query, args := f.bind(res)
	ext := f.resolveExt(opts)
	err = sqlx.GetContext(ctx, ext, dest, query, args...)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// First is FirstOrDefault that returns ErrNoRows when empty (SqlKata.First).
func (f *QueryFactory) First(ctx context.Context, q *sqlkata.Query, dest any, opts ...Option) error {
	ok, err := f.FirstOrDefault(ctx, q, dest, opts...)
	if err != nil {
		return err
	}
	if !ok {
		return ErrNoRows
	}
	return nil
}

// FirstValue is a generic First helper.
func FirstValue[T any](ctx context.Context, f *QueryFactory, q *sqlkata.Query, opts ...Option) (T, error) {
	var dest T
	err := f.First(ctx, q, &dest, opts...)
	return dest, err
}

// FirstOrDefaultValue is a generic FirstOrDefault helper.
func FirstOrDefaultValue[T any](ctx context.Context, f *QueryFactory, q *sqlkata.Query, opts ...Option) (T, bool, error) {
	var dest T
	ok, err := f.FirstOrDefault(ctx, q, &dest, opts...)
	return dest, ok, err
}

// GetMaps returns rows as []map[string]any (SqlKata.GetDictionary-ish).
func (f *QueryFactory) GetMaps(ctx context.Context, q *sqlkata.Query, opts ...Option) ([]map[string]any, error) {
	ctx, cancel := f.withTimeout(ctx)
	defer cancel()
	res, err := f.Compile(q)
	if err != nil {
		return nil, err
	}
	query, args := f.bind(res)
	ext := f.resolveExt(opts)

	rows, err := ext.QueryxContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []map[string]any
	for rows.Next() {
		m := map[string]any{}
		if err := rows.MapScan(m); err != nil {
			return nil, err
		}
		// MapScan leaves []byte for text on some drivers; normalize later if needed.
		out = append(out, m)
	}
	return out, rows.Err()
}

// Execute runs INSERT/UPDATE/DELETE (SqlKata.Execute) and returns rows affected.
func (f *QueryFactory) Execute(ctx context.Context, q *sqlkata.Query, opts ...Option) (sql.Result, error) {
	ctx, cancel := f.withTimeout(ctx)
	defer cancel()
	res, err := f.Compile(q)
	if err != nil {
		return nil, err
	}
	query, args := f.bind(res)
	ext := f.resolveExt(opts)
	return ext.ExecContext(ctx, query, args...)
}

// ExecuteAffected is Execute returning only rows affected.
func (f *QueryFactory) ExecuteAffected(ctx context.Context, q *sqlkata.Query, opts ...Option) (int64, error) {
	r, err := f.Execute(ctx, q, opts...)
	if err != nil {
		return 0, err
	}
	return r.RowsAffected()
}

// ExecuteScalar compiles Limit(1) and scans a single scalar column into dest.
func (f *QueryFactory) ExecuteScalar(ctx context.Context, q *sqlkata.Query, dest any, opts ...Option) error {
	ctx, cancel := f.withTimeout(ctx)
	defer cancel()
	res, err := f.Compile(q.Clone().Limit(1))
	if err != nil {
		return err
	}
	query, args := f.bind(res)
	ext := f.resolveExt(opts)
	return sqlx.GetContext(ctx, ext, dest, query, args...)
}

// ExecuteScalarValue is a generic ExecuteScalar.
func ExecuteScalarValue[T any](ctx context.Context, f *QueryFactory, q *sqlkata.Query, opts ...Option) (T, error) {
	var dest T
	err := f.ExecuteScalar(ctx, q, &dest, opts...)
	return dest, err
}

// Exists returns true if the query would return at least one row (SqlKata.Exists).
func (f *QueryFactory) Exists(ctx context.Context, q *sqlkata.Query, opts ...Option) (bool, error) {
	clone := q.Clone().ClearComponent("select", nil).SelectRaw("1 as Exists").Limit(1)
	maps, err := f.GetMaps(ctx, clone, opts...)
	if err != nil {
		return false, err
	}
	return len(maps) > 0, nil
}

// Count runs AsCount and returns the scalar (SqlKata.Count).
func (f *QueryFactory) Count(ctx context.Context, q *sqlkata.Query, opts ...Option) (int64, error) {
	return ExecuteScalarValue[int64](ctx, f, q.Clone().AsCount(), opts...)
}

// Aggregate runs AsAggregate(op, columns...) and returns a scalar.
func (f *QueryFactory) Aggregate(ctx context.Context, q *sqlkata.Query, op string, columns []string, opts ...Option) (float64, error) {
	qq := q.Clone().AsAggregate(op, columns...)
	return ExecuteScalarValue[float64](ctx, f, qq, opts...)
}
