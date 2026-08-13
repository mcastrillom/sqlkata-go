package tests

import (
	"reflect"
	"strings"
	"testing"
)

func assertEqual(t *testing.T, want, got string) {
	t.Helper()
	if want != got {
		t.Fatalf("\nwant: %s\ngot:  %s", want, got)
	}
}

func assertBinding(t *testing.T, want, got any) {
	t.Helper()
	if !bindingEqual(want, got) {
		t.Fatalf("binding want %v (%T), got %v (%T)", want, want, got, got)
	}
}

func bindingEqual(want, got any) bool {
	if reflect.DeepEqual(want, got) {
		return true
	}
	// int / int64 interchangeability (SqlKata often uses long for offsets).
	wn, wok := toInt64(want)
	gn, gok := toInt64(got)
	return wok && gok && wn == gn
}

func toInt64(v any) (int64, bool) {
	switch x := v.(type) {
	case int:
		return int64(x), true
	case int32:
		return int64(x), true
	case int64:
		return x, true
	default:
		return 0, false
	}
}

func normalizeSpace(s string) string {
	return strings.Join(strings.Fields(s), " ")
}
