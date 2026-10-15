# How to store optional values

This guide shows how to choose the Go type for a nullable SQL column or an
optional JSON field, and how to check that the value survives a write, a read
and a JSON round trip. For why the types behave as they do, see
[Null conventions](explanation-null-conventions.md).

## Prerequisites

- Go 1.26 or later and `github.com/domonda/go-types` in your `go.mod`.
- `database/sql` with a PostgreSQL driver, for example the `stdlib` package of
  pgx or lib/pq. The array types read and write the PostgreSQL array text
  format (`{"a","b"}`), which databases without array types don't understand.
- The table definition at hand: for every column you need its type and whether
  it is `NULL` or `NOT NULL`.

## Steps

### 1. Answer two questions per column

1. Can the column be NULL, or the JSON field be absent or `null`?
2. If it can, is "empty" a different value from NULL? An empty tag list is not
   the same as an unknown one, an amount of `0.00` is not the same as no
   amount, but an empty nickname usually means no nickname.

### 2. Pick the Go type

| Column                            | Go type                                          | Zero value writes                             |
|-----------------------------------|--------------------------------------------------|-----------------------------------------------|
| `uuid NOT NULL`                   | `uu.ID`                                          | the Nil UUID, which `Validate` rejects        |
| `uuid NULL`                       | `uu.NullableID`                                  | NULL                                          |
| `uuid[] NULL`                     | `uu.IDSlice` (ordered) or `uu.IDSet`             | NULL; an empty value writes `{}`              |
| `uuid[] NOT NULL DEFAULT '{}'`    | `uu.IDSlice` or `uu.IDSet`, never nil            | NULL, so initialize with `uu.IDSlice{}`       |
| `text NOT NULL`                   | `string` or `notnull.TrimmedString`              | `''`                                          |
| `text NULL`, `''` never stored    | `nullable.NonEmptyString`                        | NULL                                          |
| `text NULL`, blank means NULL     | `nullable.TrimmedString`                         | NULL                                          |
| `text NULL`, `''` and NULL differ | `nullable.Type[string]`                          | NULL                                          |
| `text[] NULL`                     | `nullable.StringArray`                           | NULL                                          |
| `text[] NOT NULL DEFAULT '{}'`    | `notnull.StringArray`                            | `{}`                                          |
| `bigint[]`, `float8[]`            | `nullable` or `notnull` `IntArray`, `FloatArray` | NULL or `{}`                                  |
| `date NOT NULL`                   | `date.Date`                                      | NULL, so the constraint catches an unset date |
| `date NULL`                       | `date.NullableDate`                              | NULL                                          |
| `timestamptz NULL`                | `nullable.Time`                                  | NULL                                          |
| `numeric NOT NULL`                | `money.DecimalAmount`                            | `0`                                           |
| `numeric NULL`                    | `money.NullableDecimalAmount`                    | NULL; a present zero writes a zero            |
| `bigint NULL` counting cents      | `money.NullableCentAmount`                       | NULL                                          |
| `float8 NULL` amount              | `money.NullableAmount`                           | NULL                                          |
| `char(3) NULL` currency           | `money.NullableCurrency`                         | NULL                                          |
| `char(2) NULL` country            | `country.NullableCode`                           | NULL                                          |
| `text NULL` IBAN or BIC           | `bank.NullableIBAN`, `bank.NullableBIC`          | NULL                                          |
| `text NULL` VAT ID                | `vat.NullableID`                                 | NULL                                          |
| `text NULL` email address         | `email.NullableAddress`                          | NULL                                          |
| `text[] NULL` email addresses     | `email.AddressSet`                               | NULL; an empty set writes `{}`                |
| `text NULL` account number        | `account.NullableNumber`                         | NULL                                          |
| `jsonb NULL`                      | `nullable.JSON`                                  | NULL                                          |
| `jsonb NOT NULL DEFAULT '{}'`     | `notnull.JSON`                                   | `{}`                                          |

The module has no `notnull` array of UUIDs. For a `uuid[] NOT NULL` column,
assign `uu.IDSlice{}` or `uu.IDSet{}` when you create the struct, because a
nil value writes NULL. Reading such a column never gives nil: `{}` scans as an
empty, non-nil value.

The same types work for optional JSON fields. Each nullable type in the
table writes `null` for its null value, and decoding `null` gives null, also
into a field that already holds a value.

### 3. Declare the struct and tag optional fields with `omitzero`

Use `omitzero`, not `omitempty`. `omitempty` never omits `uu.NullableID` or
the `nullable.Type[T]` money types, and it drops an empty `uu.IDSlice`, which
then reads back as nil:

```go
type Customer struct {
	ID       uu.ID                       `json:"id"`
	ParentID uu.NullableID               `json:"parentId,omitzero"`
	TagIDs   uu.IDSlice                  `json:"tagIds"`
	Nickname nullable.TrimmedString      `json:"nickname,omitzero"`
	Birthday date.NullableDate           `json:"birthday,omitzero"`
	Credit   money.NullableDecimalAmount `json:"credit,omitzero"`
	Country  country.NullableCode        `json:"country,omitzero"`
}
```

`TagIDs` has no `omitzero` on purpose: the column is `NOT NULL`, so the field
is always sent, as `[]` when empty.

### 4. Write the row

Pass the fields as query arguments. `database/sql` calls their `Value`
methods, so null values become SQL NULL without extra code:

```go
const insertCustomer = /*sql*/ `
	INSERT INTO customer (id, parent_id, tag_ids, nickname, birthday, credit, country)
	VALUES ($1, $2, $3, $4, $5, $6, $7)`

func insert(ctx context.Context, db *sql.DB, c *Customer) error {
	_, err := db.ExecContext(ctx, insertCustomer, c.ID, c.ParentID, c.TagIDs, c.Nickname, c.Birthday, c.Credit, c.Country)
	return err
}
```

### 5. Read the row

Scan into pointers to the fields. `database/sql` calls their `Scan` methods,
so SQL NULL becomes the null value of each type:

```go
const selectCustomer = /*sql*/ `
	SELECT id, parent_id, tag_ids, nickname, birthday, credit, country
	FROM customer
	WHERE id = $1`

func load(ctx context.Context, db *sql.DB, id uu.ID) (*Customer, error) {
	c := new(Customer)
	err := db.QueryRowContext(ctx, selectCustomer, id).Scan(c.columns()...)
	if err != nil {
		return nil, err
	}
	return c, nil
}
```

`columns` returns the field pointers in the column order of the query; the
complete program below defines it.

### 6. Work with the values in Go

- Test for null with `IsNull()` or compare with the null value:
  `c.ParentID == uu.IDNull`.
- Read a value with `GetOr(def)`, or with `Get()` after checking `IsNull()`;
  `Get()` panics on null.
- Set a value with `Set(v)` or `SetNull()`, or convert: `id.Nullable()`,
  `date.Date("2024-02-29").Nullable()`, `nullable.TypeFrom(amount)`.

### The complete program

The program replaces the database with direct `Value` and `Scan` calls, which
is what `database/sql` does with every argument and destination. `insert` and
`load` are the functions from steps 4 and 5 for use with a real database.

```go
package main

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"github.com/domonda/go-types/country"
	"github.com/domonda/go-types/date"
	"github.com/domonda/go-types/money"
	"github.com/domonda/go-types/nullable"
	"github.com/domonda/go-types/uu"
)

// Customer maps a row of this table:
//
//	CREATE TABLE customer (
//	    id        uuid    NOT NULL PRIMARY KEY,
//	    parent_id uuid    NULL,
//	    tag_ids   uuid[]  NOT NULL DEFAULT '{}',
//	    nickname  text    NULL,
//	    birthday  date    NULL,
//	    credit    numeric NULL,
//	    country   char(2) NULL
//	);
type Customer struct {
	ID       uu.ID                       `json:"id"`
	ParentID uu.NullableID               `json:"parentId,omitzero"`
	TagIDs   uu.IDSlice                  `json:"tagIds"`
	Nickname nullable.TrimmedString      `json:"nickname,omitzero"`
	Birthday date.NullableDate           `json:"birthday,omitzero"`
	Credit   money.NullableDecimalAmount `json:"credit,omitzero"`
	Country  country.NullableCode        `json:"country,omitzero"`
}

// columns returns pointers to the fields in the column order of the queries.
func (c *Customer) columns() []any {
	return []any{&c.ID, &c.ParentID, &c.TagIDs, &c.Nickname, &c.Birthday, &c.Credit, &c.Country}
}

const insertCustomer = /*sql*/ `
	INSERT INTO customer (id, parent_id, tag_ids, nickname, birthday, credit, country)
	VALUES ($1, $2, $3, $4, $5, $6, $7)`

func insert(ctx context.Context, db *sql.DB, c *Customer) error {
	_, err := db.ExecContext(ctx, insertCustomer, c.ID, c.ParentID, c.TagIDs, c.Nickname, c.Birthday, c.Credit, c.Country)
	return err
}

const selectCustomer = /*sql*/ `
	SELECT id, parent_id, tag_ids, nickname, birthday, credit, country
	FROM customer
	WHERE id = $1`

func load(ctx context.Context, db *sql.DB, id uu.ID) (*Customer, error) {
	c := new(Customer)
	err := db.QueryRowContext(ctx, selectCustomer, id).Scan(c.columns()...)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func main() {
	c := Customer{
		ID:       uu.IDMust("0e5f8a45-5f2c-4d6c-9e57-6a8d1e1b2c3d"),
		TagIDs:   uu.IDSlice{}, // never nil, the column is NOT NULL
		Nickname: nullable.TrimmedStringFrom("  Ann  "),
		Birthday: "2024-02-29",
		Credit:   nullable.TypeFrom(money.NewDecimalAmount(0, 2)),
		// ParentID and Country stay null
	}

	// What database/sql sends to the driver for insert
	var row []driver.Value
	for _, col := range c.columns() {
		v, err := col.(driver.Valuer).Value()
		if err != nil {
			panic(err)
		}
		row = append(row, v)
		fmt.Printf("%#v\n", v)
	}

	// What database/sql does with the values the driver returns for load
	var scanned Customer
	for i, col := range scanned.columns() {
		err := col.(sql.Scanner).Scan(row[i])
		if err != nil {
			panic(err)
		}
	}

	j, err := json.Marshal(scanned)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(j))

	var back Customer
	err = json.Unmarshal(j, &back)
	if err != nil {
		panic(err)
	}
	fmt.Println(back.ParentID.IsNull(), back.TagIDs != nil, back.Credit.IsNull(), back.Country.IsNull())

	// Output:
	// "0e5f8a45-5f2c-4d6c-9e57-6a8d1e1b2c3d"
	// <nil>
	// "{}"
	// "Ann"
	// "2024-02-29"
	// "0.00"
	// <nil>
	// {"id":"0e5f8a45-5f2c-4d6c-9e57-6a8d1e1b2c3d","tagIds":[],"nickname":"Ann","birthday":"2024-02-29","credit":0.00}
	// true true false true
}
```

## Verification

Run the program with `go run .`. Its output shows each rule at work:

- `<nil>` for `ParentID` and `Country`: their null values write SQL NULL.
- `"{}"` for `TagIDs`: the empty, non-nil slice writes the empty array that
  the `NOT NULL` column needs.
- `"Ann"`: `nullable.TrimmedString` stores trimmed text.
- `"0.00"` and `"credit":0.00`: a present zero amount stays a value, unlike a
  null one.
- The JSON has no `parentId` and no `country` key: `omitzero` dropped the null
  values, and reading the JSON back gives null for both (`true` in the last
  line) and a non-nil `TagIDs` (`true`).

Against a real database, insert a row with `insert`, read it with `load`, and
compare the two structs field by field.

## Troubleshooting

**`cannot convert <nil> to uu.ID`**, **`invalid VAT ID: can't scan SQL NULL as
vat.ID`**, **`can't scan <nil> as Number`**, **`can't scan <nil> as
notnull.TrimmedString`**, **`can't scan value of type <nil> as
money.DecimalAmount`**, or `database/sql`'s **`converting NULL to float64 is
unsupported`**.
The column contains NULL but the field type has no null. Use the nullable
type from the table: `uu.NullableID`, `vat.NullableID`,
`account.NullableNumber`, `nullable.TrimmedString`,
`money.NullableDecimalAmount` or `money.NullableAmount`.

**`can't scan empty string as nullable.NonEmptyString`**.
The column contains `''`, which this type can't hold apart from NULL. Use
`nullable.TrimmedString`, which reads `''` as NULL, or `nullable.Type[string]`
if `''` is a real value. Or clean the data: set the column to NULL where it is
`''`.

**The insert fails with a not-null constraint violation on an array column.**
A nil `uu.IDSlice`, `uu.IDSet`, `email.AddressSet` or `nullable` array wrote
NULL. Initialize the field with an empty value, or use the `notnull` array
type.

**`pq: unable to parse array; expected '{' at offset 0`**.
An array type scanned text that is not an array literal, often the empty
string. Check that the column is an array type. `nullable.StringArray` and
`notnull.StringArray` reject `""`.

**An empty list arrives as a missing field and reads back as nil.**
The field has `omitempty`, which drops empty slices and maps. Use `omitzero`.

**A null amount arrives as `0` after a JSON round trip.**
The client sent `""`. For `money.NullableAmount`, `NullableDecimalAmount` and
`NullableRate`, the JSON string `""` decodes as a present zero, not as null.
Send `null` or leave the field out.

**`invalid UUID length 0: ""` when decoding JSON.**
The client sent `""` for a `uu.NullableID`. Send `null` or leave the field out.

**An optional `language.Code` field fails validation with `invalid
language.Code: ""`.**
`language.Code` has no nullable type. `language.Null` writes and reads SQL
NULL, but validation rejects it. Store the column in a
`nullable.NonEmptyString` and convert it to `language.Code` where it is not
null.

## Related

- [Null conventions](explanation-null-conventions.md): the zero value to SQL
  and JSON table for every type, and the known inconsistencies.
- [Sets](explanation-sets.md): `uu.IDSet` and `email.AddressSet` encodings.
- Reference: [nullable](../nullable/README.md), [notnull](../notnull/README.md),
  [uu](../uu/README.md), [money](../money/README.md),
  [date](../date/README.md), [country](../country/README.md).
