package scan_test

import (
	"database/sql"
	"sync"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/goapt/scan"
	"github.com/goapt/scan/internal/assert"
	"github.com/goapt/scan/internal/require"
)

var (
	testDB *sql.DB
	once   sync.Once
	dsn    = "root:123456@tcp(127.0.0.1:3306)/test?charset=utf8&parseTime=True&loc=Asia%2FShanghai"
)

func db(t testing.TB) *sql.DB {
	once.Do(func() {
		var err error
		testDB, err = sql.Open("mysql", dsn)
		if err != nil {
			t.Skip(err.Error())
			return
		}
		if err := testDB.Ping(); err != nil {
			t.Skip(err.Error())
			return
		}
	})
	return testDB
}

func q(t testing.TB, query string, args ...any) *sql.Rows {
	rows, err := db(t).Query(query, args...)
	require.NoError(t, err)
	return rows
}

func TestRowsConvertsColumnNamesToTitleText(t *testing.T) {
	type Item struct {
		First string
	}
	expected := "Brett Jones"
	rows := q(t, "SELECT ? AS First", expected)
	defer rows.Close()
	item, err := scan.Row[Item](rows)
	require.NoError(t, err)
	assert.Equal(t, expected, item.First)
}

func TestRowsUsesTagName(t *testing.T) {
	expected := "Brett Jones"
	rows := q(t, "SELECT ? AS first_and_last_name", expected)
	defer rows.Close()
	type Item struct {
		FirstAndLastName string `db:"first_and_last_name"`
	}
	item, err := scan.Row[Item](rows)
	require.NoError(t, err)
	assert.Equal(t, expected, item.FirstAndLastName)
}

func TestRowsIgnoresUnsetableColumns(t *testing.T) {
	expected := "Brett Jones"
	rows := q(t, "SELECT ? AS first_and_last_name", expected)
	defer rows.Close()
	type Item struct {
		firstAndLastName string `db:"first_and_last_name"`
	}
	item, err := scan.Row[Item](rows)
	require.NoError(t, err)
	assert.NotEqual(t, expected, item.firstAndLastName)
}

func TestRowsErrorsWhenNotGivenAPointer(t *testing.T) {
	rows := q(t, "SELECT 'Bob' AS name")
	defer rows.Close()
	_, err := scan.Rows[string](rows)
	require.NoError(t, err)
}

func TestRowsErrorsWhenNotGivenAPointerToSlice(t *testing.T) {
	rows := q(t, "SELECT 'Bob' AS name")
	defer rows.Close()
	_, err := scan.Rows[struct{}](rows)
	require.NoError(t, err)
}

func TestDoesNothingWhenNoColumns(t *testing.T) {
	rows := q(t, "SELECT 'x' AS a LIMIT 0")
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
	rows := q(t, "SELECT 'x' AS Name LIMIT 0")
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
	rows := q(t, "SELECT 'Brett' AS First, 'Jones' AS Last, 1 AS Age UNION ALL SELECT 'Fred','Jones',2")
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
	rows := q(t, "SELECT 'Brett' AS first, 40 AS age UNION ALL SELECT 'Fred', 50")
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
	rows := q(t, "SELECT ? AS name", expected)
	defer rows.Close()
	name, err := scan.Row[string](rows)
	assert.NoError(t, err)
	assert.Equal(t, expected, name)
}

func TestScansPrimitiveSlices(t *testing.T) {
	table := [][]any{
		{1, 2, 3},
		{"brett", "fred", "geoff"},
		{true, false},
		{1.0, 1.1, 1.2},
	}
	for _, items := range table {
		query := "SELECT ? AS a"
		args := []any{items[0]}
		for i := 1; i < len(items); i++ {
			query += " UNION ALL SELECT ?"
			args = append(args, items[i])
		}
		rows := q(t, query, args...)
		defer rows.Close()
		scanned, err := scan.Rows[any](rows)
		require.NoError(t, err)
		got := make([]any, len(scanned))
		for i := range scanned {
			switch items[i].(type) {
			case int:
				switch v := scanned[i].(type) {
				case int64:
					got[i] = int(v)
				default:
					got[i] = scanned[i]
				}
			case string:
				if b, ok := scanned[i].([]byte); ok {
					got[i] = string(b)
				} else {
					got[i] = scanned[i]
				}
			case bool:
				switch v := scanned[i].(type) {
				case int64:
					got[i] = v != 0
				default:
					got[i] = scanned[i]
				}
			default:
				got[i] = scanned[i]
			}
		}
		assert.EqualValues(t, items, got)
	}
}

func TestErrorsWhenMoreThanOneColumnForPrimitiveSlice(t *testing.T) {
	rows := q(t, "SELECT 'brett' AS fname, 'jones' AS lname")
	defer rows.Close()
	_, err := scan.Rows[string](rows)
	assert.EqualValues(t, scan.ErrTooManyColumns, err)
}

func TestErrorsWhenScanRowToSlice(t *testing.T) {
	rows := q(t, "SELECT 1 AS ID LIMIT 0")
	defer rows.Close()
	_, err := scan.Row[[]struct{ ID int }](rows)
	assert.EqualValues(t, sql.ErrNoRows, err)
}

func TestRowReturnsErrNoRowsWhenQueryHasNoRows(t *testing.T) {
	rows := q(t, "SELECT 'x' AS First LIMIT 0")
	defer rows.Close()
	type Item struct {
		First string
	}
	_, err := scan.Row[Item](rows)
	assert.EqualValues(t, sql.ErrNoRows, err)
}

func TestRowErrorsWhenItemIsNotAPointer(t *testing.T) {
	rows := q(t, "SELECT 'x' AS First LIMIT 0")
	defer rows.Close()
	_, err := scan.Row[struct{ First string }](rows)
	assert.EqualValues(t, sql.ErrNoRows, err)
}

func TestRowClosesEarly(t *testing.T) {
	rows := q(t, "SELECT 'Bob' AS name")
	_, _ = scan.Row[string](rows)
	err := rows.Close()
	assert.NoError(t, err)
}
