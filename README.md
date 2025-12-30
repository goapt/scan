# Scan

[![GoDoc](https://godoc.org/github.com/goapt/scan?status.svg)](https://godoc.org/github.com/goapt/scan)
[![Build Status](https://github.com/goapt/scan/workflows/go%20test/badge.svg)](https://github.com/goapt/scan/actions)
[![Coverage Status](https://coveralls.io/repos/github/goapt/scan/badge.svg?branch=master)](https://coveralls.io/github/goapt/scan?branch=master)

Scan is a Go package for scanning database rows into structs or slices of primitive types.

> This project is forked from [https://github.com/blockloop/scan](https://github.com/blockloop/scan) and adjusted to be generic, while all dependencies are removed.

## Installation

```sh
go get github.com/goapt/scan
```

## Usage

```go
import "github.com/goapt/scan"
```

## Examples

### Multiple Rows

```go
type Person struct {
    ID   int    `db:"id"`
    Name string
}

db, err := sql.Open("sqlite3", "database.sqlite")
rows, err := db.Query("SELECT * FROM persons")
defer rows.Close()
persons, err := scan.Rows[Person](rows)

fmt.Printf("%#v", persons)
// []Person{
//    {ID: 1, Name: "brett"},
//    {ID: 2, Name: "fred"},
//    {ID: 3, Name: "stacy"},
// }
```

### Multiple rows of primitive type

```go
rows, err := db.Query("SELECT name FROM persons")
defer rows.Close()
names, err := scan.Rows[string](rows)

fmt.Printf("%#v", names)
// []string{
//    "brett",
//    "fred",
//    "stacy",
// }
```

### Single row

```go
rows, err := db.Query("SELECT * FROM persons where name = 'brett' LIMIT 1")
person, err := scan.Row[Person](rows)
defer rows.Close()

fmt.Printf("%#v", person)
// Person{ ID: 1, Name: "brett" }
```

### Scalar value

```go
rows, err := db.Query("SELECT age FROM persons where name = 'brett' LIMIT 1")
defer rows.Close()
age, err := scan.Row[int8](rows)

fmt.Printf("%d", age)
// 100
```

### Custom Column Mapping

By default, column names are mapped to and from database column names using basic title case conversion. You can override this behavior by setting `ScannerMapper` to custom functions.

```go
scan.ScannerMapper = func(columnName string) string {
	return strings.ToLower(columnName)
}
```

## Why

While many other projects support similar features (i.e. [sqlx](https://github.com/jmoiron/sqlx)) scan allows you to use any database lib such as the stdlib to write fluent SQL statements and pass the resulting `rows` to `scan` for scanning.
