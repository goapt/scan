package require

import (
	"reflect"
	"testing"
)

func NoError(t testing.TB, err error, _ ...any) {
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func Error(t testing.TB, err error, _ ...any) {
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func Len(t testing.TB, v any, n int, _ ...any) {
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Array, reflect.Slice, reflect.Map, reflect.String:
		if rv.Len() != n {
			t.Fatalf("expected len: %d, actual len: %d", n, rv.Len())
		}
	default:
		t.Fatalf("len not supported for kind: %s", rv.Kind().String())
	}
}
