package scan

import (
	"database/sql"
	"errors"
	"io"
	"reflect"
	"strings"
	"unicode"
)

var (
	// ErrTooManyColumns indicates that a select query returned multiple columns and
	// attempted to bind to a slice of a primitive type. For example, trying to bind
	// `select col1, col2 from mutable` to []string
	ErrTooManyColumns = errors.New("too many columns returned for primitive slice")

	// AutoClose is true when scan should automatically close Scanner when the scan
	// is complete. If you set it to false, then you must defer rows.Close() manually
	AutoClose = true

	// OnAutoCloseError can be used to log errors which are returned from rows.Close()
	// By default this is a NOOP function
	OnAutoCloseError = func(error) {}

	// ScannerMapper transforms database field names into struct/map field names
	// E.g. you can set function for convert snake_case into CamelCase
	ScannerMapper = func(name string) string { return toTitleCase(name) }
)

// toTitleCase converts a string to title case (first letter capitalized)
func toTitleCase(s string) string {
	if s == "" {
		return s
	}

	// Split by underscores to handle snake_case
	parts := strings.Split(s, "_")
	for i, part := range parts {
		if part != "" {
			parts[i] = capitalizeFirst(part)
		}
	}
	return strings.Join(parts, "")
}

// capitalizeFirst capitalizes the first letter of a string
func capitalizeFirst(s string) string {
	if s == "" {
		return s
	}

	runes := []rune(s)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

// Row scans a single row and returns a value of type T.
// It requires that you use db.Query and not db.QueryRow, because QueryRow does not return column names.
func Row[T any](r RowsScanner) (T, error) {
	if AutoClose {
		defer closeRows(r)
	}
	var zero T
	items, err := rowsGeneric[T](r)
	if err != nil {
		return zero, err
	}
	if len(items) == 0 {
		return zero, sql.ErrNoRows
	}
	return items[0], nil
}

// Rows scans sql rows into a slice of T.
func Rows[T any](r RowsScanner) ([]T, error) {
	if AutoClose {
		defer closeRows(r)
	}
	return rowsGeneric[T](r)
}

func rowsGeneric[T any](r RowsScanner) ([]T, error) {
	cols, err := r.Columns()
	if err != nil {
		return nil, err
	}

	var out []T
	var t T
	itemType := reflect.TypeOf(t)
	if itemType == nil {
		itemType = reflect.TypeOf((*any)(nil)).Elem()
	}
	isPrimitive := itemType.Kind() != reflect.Struct

	for r.Next() {
		itemVal := reflect.New(itemType).Elem()

		var pointers []any
		if isPrimitive {
			if len(cols) > 1 {
				return nil, ErrTooManyColumns
			}
			pointers = []any{itemVal.Addr().Interface()}
		} else {
			pointers = structPointers(itemVal, cols)
		}

		if len(pointers) == 0 {
			continue
		}

		if err := r.Scan(pointers...); err != nil {
			return nil, err
		}

		// append scanned item
		out = append(out, itemVal.Interface().(T))
	}
	return out, r.Err()
}

// Initialization the tags from struct.
func initFieldTag(sliceItem reflect.Value, fieldTagMap *map[string]reflect.Value) {
	typ := sliceItem.Type()
	for i := 0; i < sliceItem.NumField(); i++ {
		if typ.Field(i).Anonymous || typ.Field(i).Type.Kind() == reflect.Struct {
			// found an embedded struct
			sliceItemOfAnonymous := sliceItem.Field(i)
			initFieldTag(sliceItemOfAnonymous, fieldTagMap)
		}
		tag, ok := typ.Field(i).Tag.Lookup("db")
		if ok && tag != "" {
			(*fieldTagMap)[tag] = sliceItem.Field(i)
		}
	}
}

func structPointers(sliceItem reflect.Value, cols []string) []any {
	pointers := make([]any, 0, len(cols))
	fieldTag := make(map[string]reflect.Value, len(cols))
	initFieldTag(sliceItem, &fieldTag)

	for _, colName := range cols {
		var fieldVal reflect.Value
		if v, ok := fieldTag[colName]; ok {
			fieldVal = v
		} else {
			fieldVal = sliceItem.FieldByName(ScannerMapper(colName))
		}
		if !fieldVal.IsValid() || !fieldVal.CanSet() {
			// have to add if we found a column because Scan() requires
			// len(cols) arguments or it will error. This way we can scan to
			// a useless pointer
			var nothing any
			pointers = append(pointers, &nothing)
			continue
		}

		pointers = append(pointers, fieldVal.Addr().Interface())
	}
	return pointers
}

func closeRows(c io.Closer) {
	if err := c.Close(); err != nil {
		if OnAutoCloseError != nil {
			OnAutoCloseError(err)
		}
	}
}
