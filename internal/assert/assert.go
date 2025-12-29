package assert

import (
	"reflect"
	"testing"
)

func Equal(t testing.TB, expected, actual any, _ ...any) {
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("expected: %v, actual: %v", expected, actual)
	}
}

func NotEqual(t testing.TB, expected, actual any, _ ...any) {
	if reflect.DeepEqual(expected, actual) {
		t.Errorf("unexpected equal: %v", actual)
	}
}

func EqualValues(t testing.TB, expected, actual any, _ ...any) {
	if !equalValues(expected, actual) {
		t.Errorf("expected values: %v, actual values: %v", expected, actual)
	}
}

func Len(t testing.TB, v any, n int, _ ...any) {
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Array, reflect.Slice, reflect.Map, reflect.String:
		if rv.Len() != n {
			t.Errorf("expected len: %d, actual len: %d", n, rv.Len())
		}
	default:
		t.Errorf("len not supported for kind: %s", rv.Kind().String())
	}
}

func NoError(t testing.TB, err error, _ ...any) {
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func Error(t testing.TB, err error, _ ...any) {
	if err == nil {
		t.Errorf("expected error, got nil")
	}
}

func Contains(t testing.TB, s, substr string, _ ...any) {
	if !contains(s, substr) {
		t.Errorf("expected %q to contain %q", s, substr)
	}
}

func Nil(t testing.TB, v any, _ ...any) {
	if !isNil(v) {
		t.Errorf("expected nil, got: %v", v)
	}
}

func NotNil(t testing.TB, v any, _ ...any) {
	if isNil(v) {
		t.Errorf("unexpected nil")
	}
}

func contains(s, substr string) bool {
	return len(substr) == 0 || (len(s) >= len(substr) && (indexOf(s, substr) >= 0))
}

func indexOf(s, substr string) int {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func isNil(v any) bool {
	if v == nil {
		return true
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Pointer, reflect.Map, reflect.Slice, reflect.Func, reflect.Interface, reflect.Chan:
		return rv.IsNil()
	default:
		return false
	}
}

func equalValues(a, b any) bool {
	if a == nil || b == nil {
		return a == b
	}
	ka := reflect.ValueOf(a).Kind()
	kb := reflect.ValueOf(b).Kind()
	if isNumber(ka) && isNumber(kb) {
		return toFloat64(a) == toFloat64(b)
	}
	return reflect.DeepEqual(a, b)
}

func isNumber(k reflect.Kind) bool {
	switch k {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return true
	default:
		return false
	}
}

func toFloat64(v any) float64 {
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(rv.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return float64(rv.Uint())
	case reflect.Float32, reflect.Float64:
		return rv.Float()
	default:
		return 0
	}
}
