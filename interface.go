package scan

import "database/sql"

// RowsScanner is a database scanner for many rows. It is most commonly the
// result of *sql.DB Query(...).
type RowsScanner interface {
	Close() error
	Scan(dest ...any) error
	Columns() ([]string, error)
	ColumnTypes() ([]*sql.ColumnType, error)
	Err() error
	Next() bool
}

// cache is an interface for a sync.Map that is used for cache internally
type cache interface {
	Delete(key any)
	Load(key any) (value any, ok bool)
	LoadOrStore(key any, value any) (actual any, loaded bool)
	Range(f func(key any, value any) bool)
	Store(key any, value any)
}
