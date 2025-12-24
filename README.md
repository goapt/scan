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

### Nested Struct Fields (as of v2.0.0)

```go
rows, err := db.Query(`
	SELECT person.id,person.name,company.name FROM person
	JOIN company on company.id = person.company_id
	LIMIT 1
`)
defer rows.Close()

type Person struct {
	ID      int    `db:"person.id"`
	Name    string `db:"person.name"`
	Company struct {
		Name string `db:"company.name"`
	}
}

person, err := scan.Row[Person](rows)

err = json.NewEncoder(os.Stdout).Encode(&person)
// Output:
// {"ID":1,"Name":"brett","Company":{"Name":"costco"}}
```

### Custom Column Mapping

By default, column names are mapped to and from database column names using basic title case conversion. You can override this behavior by setting `ColumnsMapper` and `ScannerMapper` to custom functions.

### Columns

`Columns` scans a struct and returns a string slice of the assumed column names based on the `db` tag or the struct field name respectively. To avoid assumptions, use `ColumnsStrict` which will _only_ return the fields tagged with the `db` tag. Both `Columns` and `ColumnsStrict` are variadic. They both accept a string slice of column names to exclude from the list. It is recommended that you cache this slice.

```go
package main

type User struct {
        ID        int64
        Name      string
        Age       int
        BirthDate string `db:"bday"`
        Zipcode   string `db:"-"`
        Store     struct {
                ID int
                // ...
        }
}

var nobody = new(User)
var userInsertCols = scan.Columns(nobody, "ID")
// []string{ "Name", "Age", "bday" }

var userSelectCols = scan.Columns(nobody)
// []string{ "ID", "Name", "Age", "bday" }
```

### Values

`Values` scans a struct and returns the values associated with the provided columns. Values uses a `sync.Map` to cache fields of structs to greatly improve the performance of scanning types. The first time a struct is scanned it's **exported** fields locations are cached. When later retrieving values from the same struct it should be much faster.

```go
user := &User{
        ID: 1,
        Name: "Brett",
        Age: 100,
}

vals := scan.Values([]string{"ID", "Name"}, user)
// []interface{}{ 1, "Brett" }
```

## Why

While many other projects support similar features (i.e. [sqlx](https://github.com/jmoiron/sqlx)) scan allows you to use any database lib such as the stdlib to write fluent SQL statements and pass the resulting `rows` to `scan` for scanning.
