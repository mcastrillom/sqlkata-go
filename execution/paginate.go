package execution

import (
	"context"
	"fmt"
	"math"

	"github.com/mcastrillom/sqlkata-go/sqlkata"
)

// PaginationResult mirrors SqlKata.Execution.PaginationResult[T].
type PaginationResult[T any] struct {
	Query   *sqlkata.Query
	Count   int64
	List    []T
	Page    int
	PerPage int
}

func (p PaginationResult[T]) TotalPages() int {
	if p.PerPage < 1 {
		return 0
	}
	return int(math.Ceil(float64(p.Count) / float64(p.PerPage)))
}

func (p PaginationResult[T]) IsFirst() bool { return p.Page == 1 }
func (p PaginationResult[T]) IsLast() bool {
	tp := p.TotalPages()
	return tp == 0 || p.Page == tp
}
func (p PaginationResult[T]) HasNext() bool     { return p.Page < p.TotalPages() }
func (p PaginationResult[T]) HasPrevious() bool { return p.Page > 1 }

// NextQuery returns the query for the next page.
func (p PaginationResult[T]) NextQuery() *sqlkata.Query {
	return p.Query.Clone().ForPage(p.Page+1, p.PerPage)
}

// PreviousQuery returns the query for the previous page.
func (p PaginationResult[T]) PreviousQuery() *sqlkata.Query {
	page := p.Page - 1
	if page < 1 {
		page = 1
	}
	return p.Query.Clone().ForPage(page, p.PerPage)
}

// Paginate runs Count + Get for a page (SqlKata.Paginate).
func Paginate[T any](ctx context.Context, f *QueryFactory, q *sqlkata.Query, page, perPage int, opts ...Option) (PaginationResult[T], error) {
	var zero PaginationResult[T]
	if page < 1 {
		return zero, fmt.Errorf("page must be >= 1")
	}
	if perPage < 1 {
		return zero, fmt.Errorf("perPage must be >= 1")
	}
	count, err := f.Count(ctx, q.Clone(), opts...)
	if err != nil {
		return zero, err
	}
	var list []T
	if count > 0 {
		list, err = GetSlice[T](ctx, f, q.Clone().ForPage(page, perPage), opts...)
		if err != nil {
			return zero, err
		}
	}
	return PaginationResult[T]{
		Query:   q,
		Count:   count,
		List:    list,
		Page:    page,
		PerPage: perPage,
	}, nil
}

// Chunk invokes fn for each page until fn returns false (SqlKata.Chunk with Func).
func Chunk[T any](ctx context.Context, f *QueryFactory, q *sqlkata.Query, chunkSize int, fn func(list []T, page int) bool, opts ...Option) error {
	if chunkSize < 1 {
		return fmt.Errorf("chunkSize must be >= 1")
	}
	page := 1
	for {
		res, err := Paginate[T](ctx, f, q, page, chunkSize, opts...)
		if err != nil {
			return err
		}
		if len(res.List) == 0 && page == 1 {
			return nil
		}
		if !fn(res.List, page) {
			return nil
		}
		if !res.HasNext() {
			return nil
		}
		page++
	}
}
