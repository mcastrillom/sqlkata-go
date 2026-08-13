package compiler

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/mcastrillom/sqlkata-go/sqlkata"
)

// SqlResult holds compiled SQL and bindings (SqlKata.SqlResult subset).
//
// RawSQL and Bindings are produced by Compile. SQL and NamedBindings are derived
// on demand and cached, so compiling stays cheap when only RawSQL is executed.
// A SqlResult is not safe for concurrent use by multiple goroutines.
type SqlResult struct {
	Query    *sqlkata.Query
	RawSQL   string
	Bindings []any

	comp  *Compiler
	sql   string
	named map[string]any
	err   error // first compile-time failure (e.g. missing variable)
}

func (r *SqlResult) fail(err error) {
	if r != nil && r.err == nil && err != nil {
		r.err = err
	}
}

// SQL returns RawSQL with dialect named placeholders (@p0, :p0, …).
func (r *SqlResult) SQL() string {
	if r.sql == "" && r.RawSQL != "" && r.comp != nil {
		r.sql = replacePlaceholders(r.RawSQL, r.comp.placeholder, r.comp.escape, r.comp.paramPrefix)
	}
	return r.sql
}

// NamedBindings maps each binding to its named placeholder (@p0 → value).
func (r *SqlResult) NamedBindings() map[string]any {
	if r.named == nil && r.comp != nil {
		r.named = namedBindings(r.Bindings, r.comp.paramPrefix)
	}
	return r.named
}

// String interpolates bindings into human-readable SQL (similar to SqlResult.ToString()).
func (r *SqlResult) String() string {
	if r.RawSQL == "" {
		return ""
	}
	return replaceBindings(r.RawSQL, r.Bindings)
}

func replaceBindings(raw string, bindings []any) string {
	idx := 0
	var b strings.Builder
	escaped := false
	for i := 0; i < len(raw); i++ {
		ch := raw[i]
		if escaped {
			if ch == '?' {
				b.WriteByte('?')
			} else {
				b.WriteByte('\\')
				b.WriteByte(ch)
			}
			escaped = false
			continue
		}
		if ch == '\\' {
			escaped = true
			continue
		}
		if ch == '?' {
			if idx >= len(bindings) {
				b.WriteString("?/*missing binding*/")
				idx++
				continue
			}
			b.WriteString(formatBinding(bindings[idx]))
			idx++
			continue
		}
		b.WriteByte(ch)
	}
	return b.String()
}

func formatBinding(v any) string {
	if v == nil {
		return "NULL"
	}
	switch x := v.(type) {
	case int:
		return strconv.FormatInt(int64(x), 10)
	case int32:
		return strconv.FormatInt(int64(x), 10)
	case int64:
		return strconv.FormatInt(x, 10)
	case float64:
		return strconv.FormatFloat(x, 'f', -1, 64)
	case bool:
		if x {
			return "1"
		}
		return "0"
	case string:
		return "'" + strings.ReplaceAll(x, "'", "''") + "'"
	default:
		return fmt.Sprint(x)
	}
}
