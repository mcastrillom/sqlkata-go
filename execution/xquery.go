package execution

import (
	"context"

	"github.com/mcastrillom/sqlkata-go/sqlkata"
)

// XQuery is an executable query handle bound to a QueryFactory (SqlKata.XQuery).
// Build with factory.XQuery / XFromQuery, then call terminal methods.
//
// Note: fluent methods on the embedded *sqlkata.Query return *sqlkata.Query.
// Prefer factory.Get/First with that query, or call terminals on XQuery before
// further fluent calls that discard the wrapper — or use:
//
//	xq := db.XQuery("Users")
//	xq.Query.WhereEq("Id", 1)
//	err := xq.Get(ctx, &rows)
type XQuery struct {
	*sqlkata.Query
	Factory *QueryFactory
}

// XQuery starts an XQuery (SqlKata.QueryFactory.Query → XQuery).
func (f *QueryFactory) XQuery(table ...string) *XQuery {
	return &XQuery{Query: f.Query(table...), Factory: f}
}

// XFromQuery wraps/clones a query as XQuery (SqlKata.FromQuery).
func (f *QueryFactory) XFromQuery(q *sqlkata.Query) *XQuery {
	return &XQuery{Query: f.FromQuery(q), Factory: f}
}

func (x *XQuery) Get(ctx context.Context, dest any, opts ...Option) error {
	return x.Factory.Get(ctx, x.Query, dest, opts...)
}

func (x *XQuery) First(ctx context.Context, dest any, opts ...Option) error {
	return x.Factory.First(ctx, x.Query, dest, opts...)
}

func (x *XQuery) FirstOrDefault(ctx context.Context, dest any, opts ...Option) (bool, error) {
	return x.Factory.FirstOrDefault(ctx, x.Query, dest, opts...)
}

func (x *XQuery) Execute(ctx context.Context, opts ...Option) (int64, error) {
	return x.Factory.ExecuteAffected(ctx, x.Query, opts...)
}

func (x *XQuery) Exists(ctx context.Context, opts ...Option) (bool, error) {
	return x.Factory.Exists(ctx, x.Query, opts...)
}

func (x *XQuery) Count(ctx context.Context, opts ...Option) (int64, error) {
	return x.Factory.Count(ctx, x.Query, opts...)
}

func (x *XQuery) GetMaps(ctx context.Context, opts ...Option) ([]map[string]any, error) {
	return x.Factory.GetMaps(ctx, x.Query, opts...)
}
