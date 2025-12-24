package scan_test

import (
	"database/sql"
	"errors"
	"reflect"
	"testing"

	"github.com/goapt/scan"
	"github.com/goapt/scan/internal/assert"
	"github.com/goapt/scan/internal/require"
)

func TestRowsConvertsColumnNamesToTitleText(t *testing.T) {
	type Item struct {
		First string
	}

	expected := "Brett Jones"
	rows := fakeRowsWithRecords(t, []string{"First"},
		[]any{expected},
	)
	defer rows.Close()

	item, err := scan.Row[Item](rows)
	require.NoError(t, err)
	assert.Equal(t, 1, rows.ScanCallCount())
	assert.Equal(t, expected, item.First)
}

func TestRowsUsesTagName(t *testing.T) {
	expected := "Brett Jones"
	rows := fakeRowsWithRecords(t, []string{"first_and_last_name"},
		[]any{expected},
	)
	defer rows.Close()

	type Item struct {
		FirstAndLastName string `db:"first_and_last_name"`
	}

	item, err := scan.Row[Item](rows)
	require.NoError(t, err)
	assert.Equal(t, 1, rows.ScanCallCount())
	assert.Equal(t, expected, item.FirstAndLastName)
}

func TestRowsIgnoresUnsetableColumns(t *testing.T) {
	expected := "Brett Jones"
	rows := fakeRowsWithRecords(t, []string{"first_and_last_name"},
		[]any{expected},
	)
	defer rows.Close()

	type Item struct {
		// private, unsetable
		firstAndLastName string `db:"first_and_last_name"`
	}

	item, err := scan.Row[Item](rows)
	require.NoError(t, err)
	assert.NotEqual(t, expected, item.firstAndLastName)
}

func TestErrorsWhenScanErrors(t *testing.T) {
	expected := errors.New("asdf")
	rows := fakeRowsWithColumns(t, 1, "first_and_last_name")
	rows.ScanStub = func(...any) error {
		return expected
	}
	defer rows.Close()

	type Item struct {
		FirstAndLastName string `db:"first_and_last_name"`
	}
	_, err := scan.Row[Item](rows)
	assert.Equal(t, expected, err)
}

func TestRowsErrorsWhenNotGivenAPointer(t *testing.T) {
	rows := fakeRowsWithColumns(t, 1, "name")
	defer rows.Close()

	_, err := scan.Rows[string](rows)
	require.NoError(t, err)
}

func TestRowsErrorsWhenNotGivenAPointerToSlice(t *testing.T) {
	rows := fakeRowsWithColumns(t, 1, "name")
	defer rows.Close()

	_, err := scan.Rows[struct{}](rows)
	require.NoError(t, err)
}

func TestErrorsWhenColumnsReturnsError(t *testing.T) {
	expected := errors.New("asdf")
	rows := &FakeRowsScanner{
		ColumnsStub: func() ([]string, error) {
			return nil, expected
		},
	}
	defer rows.Close()

	type Item struct {
		Name string
		Age  int
	}
	_, err := scan.Rows[Item](rows)
	assert.Equal(t, expected, err)
}

func TestDoesNothingWhenNoColumns(t *testing.T) {
	rows := fakeRowsWithColumns(t, 1)
	defer rows.Close()

	type Item struct {
		Name string
		Age  int
	}
	items, err := scan.Rows[Item](rows)
	assert.NoError(t, err)
	assert.Nil(t, items)
}

func TestDoesNothingWhenNextIsFalse(t *testing.T) {
	rows := fakeRowsWithColumns(t, 0, "Name")
	defer rows.Close()

	type Item struct {
		Name string
		Age  int
	}
	items, err := scan.Rows[Item](rows)
	assert.NoError(t, err)
	assert.Nil(t, items)
}

func TestIgnoresColumnsThatDoNotHaveFields(t *testing.T) {
	rows := fakeRowsWithRecords(t, []string{"First", "Last", "Age"},
		[]any{"Brett", "Jones"},
		[]any{"Fred", "Jones"},
	)
	defer rows.Close()

	type Item struct {
		First string
		Last  string
	}
	items, err := scan.Rows[Item](rows)
	require.NoError(t, err)
	require.Len(t, items, 2)
	assert.Equal(t, "Brett", items[0].First)
	assert.Equal(t, "Jones", items[0].Last)
	assert.Equal(t, "Fred", items[1].First)
	assert.Equal(t, "Jones", items[1].Last)
}

func TestIgnoresFieldsThatDoNotHaveColumns(t *testing.T) {
	rows := fakeRowsWithRecords(t, []string{"first", "age"},
		[]any{"Brett", int8(40)},
		[]any{"Fred", int8(50)},
	)
	defer rows.Close()

	type Item struct {
		First string
		Last  string
		Age   int8
	}
	items, err := scan.Rows[Item](rows)
	require.NoError(t, err)
	require.Len(t, items, 2)
	assert.EqualValues(t, "Brett", items[0].First)
	assert.EqualValues(t, "", items[0].Last)
	assert.EqualValues(t, 40, items[0].Age)

	assert.EqualValues(t, "Fred", items[1].First)
	assert.EqualValues(t, "", items[1].Last)
	assert.EqualValues(t, 50, items[1].Age)
}

func TestRowScansToPrimitiveType(t *testing.T) {
	expected := "Bob"
	rows := fakeRowsWithRecords(t, []string{"name"},
		[]any{expected},
	)
	defer rows.Close()

	name, err := scan.Row[string](rows)
	assert.NoError(t, err)
	assert.Equal(t, expected, name)
}

func TestReturnsScannerError(t *testing.T) {
	scanErr := errors.New("broken")

	rows := fakeRowsWithColumns(t, 1, "Name")
	rows.ErrReturns(scanErr)
	defer rows.Close()

	_, err := scan.Rows[struct{ Name string }](rows)
	assert.EqualValues(t, scanErr, err)
}

func TestScansPrimitiveSlices(t *testing.T) {
	table := [][]any{
		{1, 2, 3},
		{"brett", "fred", "geoff"},
		{true, false},
		{1.0, 1.1, 1.2},
	}

	for _, items := range table {
		// each item in items is a single value which needs to be converted
		// to a single row with a scalar value
		dbrows := make([][]any, len(items))
		for i, item := range items {
			dbrows[i] = []any{item}
		}
		rows := fakeRowsWithRecords(t, []string{"a"}, dbrows...)

		scanned, err := scan.Rows[any](rows)
		require.NoError(t, err)
		assert.EqualValues(t, items, scanned)
	}
}

func TestErrorsWhenMoreThanOneColumnForPrimitiveSlice(t *testing.T) {
	rows := fakeRowsWithColumns(t, 1, "fname", "lname")
	defer rows.Close()

	_, err := scan.Rows[string](rows)
	assert.EqualValues(t, scan.ErrTooManyColumns, err)
}

func TestErrorsWhenScanRowToSlice(t *testing.T) {
	rows := &FakeRowsScanner{}
	defer rows.Close()

	_, err := scan.Row[[]struct{ ID int }](rows)
	assert.EqualValues(t, sql.ErrNoRows, err)
}

func TestRowReturnsErrNoRowsWhenQueryHasNoRows(t *testing.T) {
	rows := fakeRowsWithColumns(t, 0, "First")
	defer rows.Close()

	type Item struct {
		First string
	}
	_, err := scan.Row[Item](rows)
	assert.EqualValues(t, sql.ErrNoRows, err)
}

func TestRowErrorsWhenItemIsNotAPointer(t *testing.T) {
	rows := &FakeRowsScanner{}
	defer rows.Close()

	_, err := scan.Row[struct{ First string }](rows)
	assert.EqualValues(t, sql.ErrNoRows, err)
}

func TestRowScansNestedFields(t *testing.T) {
	rows := fakeRowsWithRecords(t, []string{"p.First", "p.Last"},
		[]any{"Brett", "Jones"},
	)
	defer rows.Close()

	type Item struct {
		First string `db:"p.First"`
		Last  string `db:"p.Last"`
	}
	res, err := scan.Row[struct{ Item Item }](rows)
	require.NoError(t, err)
	assert.Equal(t, "Brett", res.Item.First)
	assert.Equal(t, "Jones", res.Item.Last)
}

func TestRowClosesEarly(t *testing.T) {
	rows := fakeRowsWithRecords(t, []string{"name"},
		[]any{"Bob"},
	)

	_, _ = scan.Row[string](rows)
	rows.Close()
	assert.EqualValues(t, 1, rows.CloseCallCount())
}

func setValue(ptr, val any) {
	if s, ok := ptr.(sql.Scanner); ok {
		_ = s.Scan(val)
		return
	}
	reflect.ValueOf(ptr).Elem().Set(reflect.ValueOf(val))
}
