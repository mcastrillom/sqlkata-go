package sqlkata

import (
	"reflect"
	"regexp"
	"strings"
)

var expandExprRe = regexp.MustCompile(`^(?:\w+\.){1,2}\{([^}]*)\}`)

// ExpandExpression expands "table.{a,b}" into ["table.a","table.b"] (SqlKata.Helper.ExpandExpression).
func ExpandExpression(expression string) []string {
	m := expandExprRe.FindStringSubmatch(expression)
	if m == nil {
		return []string{expression}
	}
	idx := strings.Index(expression, ".{")
	if idx < 0 {
		return []string{expression}
	}
	table := expression[:idx]
	parts := regexp.MustCompile(`\s*,\s*`).Split(m[1], -1)
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		out = append(out, table+"."+p)
	}
	if len(out) == 0 {
		return []string{expression}
	}
	return out
}

// Flatten flattens one level of nested slices (SqlKata.Helper.Flatten).
func Flatten(values []any) []any {
	var out []any
	for _, v := range values {
		if isFlatArray(v) {
			rv := reflect.ValueOf(v)
			for i := 0; i < rv.Len(); i++ {
				out = append(out, rv.Index(i).Interface())
			}
		} else {
			out = append(out, v)
		}
	}
	return out
}

func isFlatArray(v any) bool {
	if v == nil {
		return false
	}
	switch v.(type) {
	case string, []byte:
		return false
	}
	rv := reflect.ValueOf(v)
	kind := rv.Kind()
	return kind == reflect.Slice || kind == reflect.Array
}

// ReplaceIdentifierUnlessEscaped mirrors Helper.ReplaceIdentifierUnlessEscaped.
func ReplaceIdentifierUnlessEscaped(input, escapeCharacter, identifier, newIdentifier string) string {
	if identifier == "" {
		return input
	}
	// Non-escaped: replace identifier not preceded by escape
	var b strings.Builder
	esc := escapeCharacter
	id := identifier
	for i := 0; i < len(input); {
		if esc != "" && i+len(esc)+len(id) <= len(input) && input[i:i+len(esc)] == esc && input[i+len(esc):i+len(esc)+len(id)] == id {
			b.WriteString(id) // drop escape, keep identifier
			i += len(esc) + len(id)
			continue
		}
		if i+len(id) <= len(input) && input[i:i+len(id)] == id {
			// check not escaped
			if esc != "" && i >= len(esc) && input[i-len(esc):i] == esc {
				b.WriteByte(input[i])
				i++
				continue
			}
			b.WriteString(newIdentifier)
			i += len(id)
			continue
		}
		b.WriteByte(input[i])
		i++
	}
	return b.String()
}
