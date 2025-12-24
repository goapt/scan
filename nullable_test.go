package scan

import (
	"testing"

	"github.com/goapt/scan/internal/assert"
)

func TestNullable_Scan(t *testing.T) {
	a := 10
	v := Nullable(&a).(nullable)

	_ = v.Scan(5)
	assert.Equal(t, *v.dest.(*int), 5)
}

func TestNullable(t *testing.T) {
	a := 10
	v := Nullable(&a)
	assert.Equal(t, nullable{dest: &a}, v)
}

type customScanner int

func (*customScanner) Scan(any) error {
	return nil
}

func TestNullable_nullable(t *testing.T) {
	a := customScanner(10)
	v := Nullable(&a)
	assert.Equal(t, &a, v)
}

func TestNullable_ptr(t *testing.T) {
	var a *int
	v := Nullable(&a)
	assert.Equal(t, &a, v)
}
