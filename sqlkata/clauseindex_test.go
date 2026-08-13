package sqlkata

import (
	"fmt"
	"sync"
	"testing"
)

func componentsByScan(q *Query, component string) []AbstractClause {
	var out []AbstractClause
	for _, c := range q.Clauses {
		if c.GetComponent() == component {
			out = append(out, c)
		}
	}
	return out
}

// assertIndexConsistent checks that lookups return exactly what a plain scan
// over Clauses would return, and that the index covers every clause.
func assertIndexConsistent(t *testing.T, q *Query, step string) {
	t.Helper()
	if q.index != nil && q.indexLen != len(q.Clauses) {
		t.Fatalf("%s: stale index, indexLen=%d clauses=%d", step, q.indexLen, len(q.Clauses))
	}
	for _, component := range []string{"select", "from", "where", "order", "group", "limit", "offset"} {
		want := componentsByScan(q, component)
		got := q.GetComponents(component, nil)
		if len(got) != len(want) {
			t.Fatalf("%s: %q returned %d clauses, scan found %d", step, component, len(got), len(want))
		}
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("%s: %q clause %d differs from scan order", step, component, i)
			}
		}
		if q.HasComponent(component, nil) != (len(want) > 0) {
			t.Fatalf("%s: HasComponent(%q) disagrees with scan", step, component)
		}
	}
	if q.index != nil {
		total := 0
		for _, bucket := range q.index {
			total += len(bucket)
		}
		if total != len(q.Clauses) {
			t.Fatalf("%s: index holds %d clauses, slice holds %d", step, total, len(q.Clauses))
		}
	}
}

func TestClauseIndexStaysConsistent(t *testing.T) {
	q := NewQuery().From("users")
	for i := 0; i < clauseIndexThreshold+10; i++ {
		q = q.WhereEq(fmt.Sprintf("col%d", i), i).Select(fmt.Sprintf("col%d", i))
	}
	assertIndexConsistent(t, q, "after build")
	if q.index == nil {
		t.Fatal("expected index on a query past the threshold")
	}

	// AddOrReplaceComponent removes the previous clause and appends a new one.
	q = q.Limit(10).Limit(25).Offset(50)
	assertIndexConsistent(t, q, "after limit/offset replace")
	if got := q.GetLimit(nil); got != 25 {
		t.Fatalf("limit = %d, want 25", got)
	}

	q = q.From("customers")
	assertIndexConsistent(t, q, "after from replace")

	q = q.ClearComponent("select", nil)
	assertIndexConsistent(t, q, "after clear select")
	if len(q.GetComponents("select", nil)) != 0 {
		t.Fatal("select clauses survived ClearComponent")
	}

	q = q.WhereEq("extra", 1).OrderBy("extra")
	assertIndexConsistent(t, q, "after post-clear additions")

	assertIndexConsistent(t, q.Clone(), "clone")
}

func TestSmallQueryHasNoIndex(t *testing.T) {
	q := NewQuery().From("users").Select("id", "name").WhereEq("status", "active")
	if q.index != nil {
		t.Fatalf("small query allocated an index with %d clauses", len(q.Clauses))
	}
	assertIndexConsistent(t, q, "small query")
}

// TestConcurrentLookupsAreReadOnly guards the invariant that lookups never
// touch the index, so a prebuilt query can be compiled from several goroutines.
// Meaningful under -race.
func TestConcurrentLookupsAreReadOnly(t *testing.T) {
	q := NewQuery().From("users")
	for i := 0; i < clauseIndexThreshold*3; i++ {
		q = q.WhereEq(fmt.Sprintf("col%d", i), i).Select(fmt.Sprintf("col%d", i))
	}

	var wg sync.WaitGroup
	for g := 0; g < 8; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 200; i++ {
				q.GetComponents("where", nil)
				q.HasComponent("select", nil)
				q.GetOneComponent("from", nil)
			}
		}()
	}
	wg.Wait()
}

// TestStaleIndexFallsBackToScan covers direct writes to the exported Clauses
// field, which cannot update the index.
func TestStaleIndexFallsBackToScan(t *testing.T) {
	q := NewQuery().From("users")
	for i := 0; i < clauseIndexThreshold+1; i++ {
		q = q.WhereEq(fmt.Sprintf("col%d", i), i)
	}
	extra := &Column{Name: "sneaky"}
	extra.SetComponent("select")
	q.Clauses = append(q.Clauses, extra)

	got := q.GetComponents("select", nil)
	if len(got) != 1 || got[0] != extra {
		t.Fatalf("stale index hid a directly appended clause: got %d clauses", len(got))
	}
}
