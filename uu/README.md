# uu

UUID types with `database/sql`, `encoding/json` and text support. `ID` holds a required UUID, `NullableID` an optional one that maps its zero value to SQL `NULL` and JSON `null`, and `IDSlice` and `IDSet` map to PostgreSQL `uuid[]` columns. `NewID` generates version 7 UUIDs by default, which sort by creation time, and lets tests inject deterministic IDs through the context.

```
import "github.com/domonda/go-types/uu"
```

The package is derived from [github.com/satori/go.uuid](https://github.com/satori/go.uuid), see [Origin and license](#origin-and-license).

## Generating IDs

Application code should call `NewID(ctx)`, so tests can replace the generator per context (see [Context functions](#context-functions)).

| Function or variable                             | Description                                        |
|--------------------------------------------------|----------------------------------------------------|
| `NewID(ctx)`                                     | New ID from the function stored with `ContextWithIDFunc` without keys, or `IDvDefault()` if there is none. |
| `IDvDefault`                                     | Variable holding the default generator used by `NewID` and `IDFromContext`. It is `IDv7`. |
| `IDv7()`                                         | Version 7 ID for `time.Now()`.                     |
| `IDv7WithTime(t)`                                | Version 7 ID: the big-endian Unix time of `t` in milliseconds in the first 48 bits as specified by RFC 9562, followed by 74 random bits. A time before 1970 or after the year 10889 keeps only its lower 48 bits, so `V7Time` returns another time. |
| `IDv7Deterministic(unixMilli)`                   | Version 7 ID with the passed timestamp and all random bits zero, for tests. `IDv7Deterministic(1)` is `00000000-0001-7000-8000-000000000000`. |
| `IDv7DeterministicFunc(startAtUnixMilli)`        | Returns a goroutine-safe `func() ID` that returns `IDv7Deterministic(start)`, then `start+1`, and so on. |
| `IDv4()`                                         | Version 4 ID with 122 random bits from `crypto/rand`. |
| `IDv1()`                                         | Version 1 ID from the current time, a clock sequence and the MAC address of the first network interface (random if there is none). |
| `IDv2(domain)`                                   | Version 2 DCE Security ID: like version 1 with the POSIX UID (`IDDomainPerson`) or GID (`IDDomainGroup`) in the first 4 bytes. |
| `IDv3(namespace, data)`                          | Version 3 ID from the MD5 hash of `namespace` and `data`. |
| `IDv5(namespace, data)`                          | Version 5 ID from the SHA-1 hash of `namespace` and `data`. |
| `IDHash(h, namespace, data, version)`            | ID from hashing `namespace` and `data` with `h`, with `version` and the RFC variant set. |
| `NamespaceDNS`, `NamespaceURL`                   | Predefined namespaces for `IDv3` and `IDv5` from RFC 4122. |
| `NamespaceOID`, `NamespaceX500`                  | Predefined namespaces for `IDv3` and `IDv5` from RFC 4122. |
| `IDDomainPerson`, `IDDomainGroup`, `IDDomainOrg` | Domain values for `IDv2`. `IDDomainOrg` leaves the first 4 bytes zero. |

## Context functions

A context can carry a `func() ID` per key. `NewID` uses a package-internal default key, other code can use its own keys with `IDFromContext`. A key must have a unique address, so use a pointer to a package-level variable of non-zero size like `var invoiceIDKey int` and pass `&invoiceIDKey`. Pointers to zero-size values like `struct{}` are not guaranteed to be unique.

| Function                                  | Description                                        |
|-------------------------------------------|----------------------------------------------------|
| `ContextWithIDFunc(ctx, f, keys...)`      | Stores `f` for every passed key, or for the default key of `NewID` if no key is passed. |
| `ContextWithID(ctx, id, keys...)`         | Shortcut for `ContextWithIDFunc` with a function that always returns `id`. |
| `ContextWithIDSequence(ctx, key, ids...)` | Stores a goroutine-safe function for `key` that returns `ids` in order and panics when called after the last one. |
| `IDFromContext(ctx, key)`                 | Calls the function stored for `key`, or returns `IDvDefault()` if there is none. |
| `IDFuncFromContext(ctx)`                  | Returns the function stored for the default key of `NewID`, or nil. Functions stored for other keys are not returned. |

## ID

`type ID [16]byte` is a UUID that is always serialized as a value. Its zero value `IDNil` is not valid, see [Serialization](#serialization) for how it is written.

### Parsing and conversion

| Function                     | Description                                        |
|------------------------------|----------------------------------------------------|
| `IDFromString(s)`            | Parses the text formats below. Does not validate version or variant, so the Nil UUID parses without an error. |
| `IDFromStringOrNil(s)`       | Like `IDFromString`, but returns `IDNil` instead of an error. |
| `IDMustFromString(s)`        | Like `IDFromString`, but panics on a parse error. Does not validate. |
| `IDFromBytes(b)`             | Parses the text formats below, and also exactly 16 bytes as raw binary UUID. Does not validate. |
| `IDFromBytesOrNil(b)`        | Like `IDFromBytes`, but returns `IDNil` instead of an error. |
| `IDFromPtr(ptr, defaultVal)` | `*ptr`, or `defaultVal` if `ptr` is nil.           |
| `IDFrom[T IDSource](val)`    | Converts a `string`, `[]byte`, `ID`, `NullableID` or `[16]byte`. Returns `IDNil` if parsing fails. Does not validate. |
| `IDMust[T IDSource](val)`    | Like `IDFrom`, but validates and panics on any error, including the Nil UUID. |
| `IDFromAny(val)`             | Like `IDMust` for an `any` value, but returns the error. A nil value returns `ErrNilID`. |
| `IDSource`                   | Type constraint `string \| []byte \| ID \| NullableID \| [16]byte`. |

Accepted text formats:

| Format                                      | Example                                         |
|---------------------------------------------|-------------------------------------------------|
| Canonical, 36 characters, any letter case   | `6ba7b810-9dad-11d1-80b4-00c04fd430c8`          |
| Hex without dashes, 32 characters           | `6ba7b8109dad11d180b400c04fd430c8`              |
| Unpadded base64 URL encoding, 22 characters | `a6e4EJ2tEdGAtADAT9QwyA`                        |
| URN with lower case prefix                  | `urn:uuid:6ba7b810-9dad-11d1-80b4-00c04fd430c8` |
| The first three wrapped in braces or quotes | `{6ba7b810-9dad-11d1-80b4-00c04fd430c8}`        |

### Methods

| Method                      | Description                                        |
|-----------------------------|----------------------------------------------------|
| `String()`, `StringBytes()` | Canonical lower case form `xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx`. |
| `StringUpper()`             | Canonical form in upper case.                      |
| `Hex()`                     | 32 hex characters without dashes.                  |
| `Base64()`                  | 22 characters of unpadded base64 URL encoding.     |
| `Bytes()`                   | The 16 bytes as a slice.                           |
| `GoString()`                | ``uu.IDFrom(`6ba7b810-9dad-11d1-80b4-00c04fd430c8`)``, used by `%#v`. |
| `PrettyString()`            | Same as `String`, for `github.com/domonda/go-pretty`. |
| `Version()`                 | The version number from the 4 version bits.        |
| `Variant()`                 | `IDVariantNCS`, `IDVariantRFC4122`, `IDVariantMicrosoft` or `IDVariantInvalid`. |
| `Valid()`                   | True for versions 1 to 8 with a variant other than `IDVariantInvalid`. False for `IDNil`. |
| `Validate()`                | Returns `ErrNilID`, `ErrInvalidVersion(v)` or `ErrInvalidVariant`, or nil if `Valid()` is true. |
| `IsNil()`, `IsZero()`       | True for `IDNil`. `IsZero` makes the `omitzero` JSON option work. |
| `IsNotNil()`                | False for `IDNil`.                                 |
| `Less(rhs)`                 | Byte-wise order, the same as `bytes.Compare` and the order of the canonical strings. |
| `V7Time()`                  | The millisecond timestamp and true for a version 7 ID, the zero time and false for other versions. |
| `Nullable()`                | Converts to `NullableID`. `IDNil` becomes `IDNull`. |
| `SetVersion(v)`             | Overwrites the version bits in place.              |
| `SetVariant()`              | Sets the RFC 4122 variant bits in place.           |
| `JSONSchema()`              | String schema with format `uuid` for `github.com/invopop/jsonschema`. |

Implements: `fmt.Stringer`, `fmt.GoStringer`, `encoding.TextMarshaler`, `encoding.TextUnmarshaler`, `encoding.BinaryMarshaler`, `encoding.BinaryUnmarshaler`, `driver.Valuer`, `sql.Scanner`, `pretty.Stringer`. `ID` has no `MarshalJSON` method: `encoding/json` uses the text methods, so an `ID` is a JSON string and works as a map key.

### Constants and errors

| Identifier                                         | Description                                        |
|----------------------------------------------------|----------------------------------------------------|
| `IDNil`                                            | The Nil UUID with all 128 bits zero, the zero value of `ID`. |
| `IDRegex`                                          | Unanchored regular expression string for the canonical lower case form. |
| `IDVariantNCS`, `IDVariantRFC4122`, `IDVariantMicrosoft`, `IDVariantInvalid` | Results of `Variant()`.                            |
| `ErrNilID`                                         | `errs.Sentinel` returned for the Nil UUID, test with `errors.Is`. |
| `ErrInvalidVariant`                                | `errs.Sentinel` returned for an unknown variant.   |
| `ErrInvalidVersion`                                | `int` error type holding an unsupported version, test with `errors.As`. |

## NullableID

`type NullableID [16]byte` is a UUID whose zero value `IDNull` is SQL `NULL` and JSON `null`. `IDNull` and the Nil UUID are the same 16 zero bytes, so parsing `"00000000-0000-0000-0000-000000000000"` returns `IDNull`, and `Valid()` and `Validate()` accept it.

| Function                          | Description                                        |
|-----------------------------------|----------------------------------------------------|
| `IDNull`                          | The null value, the zero value of `NullableID`.    |
| `NullableIDFromString(s)`         | Parses like `IDFromString`. Does not validate.     |
| `NullableIDFromStringOrNull(s)`   | Like `NullableIDFromString`, but returns `IDNull` instead of an error. |
| `NullableIDFromBytes(b)`          | Parses like `IDFromBytes`, including 16 raw bytes. |
| `NullableIDFromPtr(ptr)`          | `IDNull` for a nil pointer, otherwise `*ptr` as `NullableID`. |
| `NullableIDMust[T IDSource](val)` | Converts and validates, panics on an error. The Nil UUID is valid and returns `IDNull`. |
| `NullableIDFromAny(val)`          | Like `NullableIDMust` for an `any` value, but returns the error. A nil value returns `IDNull` without an error. |
| `NullableIDCompare(a, b)`         | `bytes.Compare` of `a` and `b`, for `slices.SortFunc`. |

| Method                               | Description                                        |
|--------------------------------------|----------------------------------------------------|
| `IsNull()`, `IsNotNull()`            | Whether the value is `IDNull`.                     |
| `Get()`                              | The `ID`, panics if null.                          |
| `GetOr(defaultID)`                   | The `ID`, or `defaultID` if null.                  |
| `GetOrNil()`                         | The `ID`, which is `IDNil` if null. Never panics.  |
| `Set(id)`, `SetNull()`               | Assign an `ID` or `IDNull`.                        |
| `Ptr()`                              | Pointer to a copy of the `ID`, or nil if null.     |
| `String()`, `StringUpper()`          | Canonical form, or `NULL` if null.                 |
| `StringOr(nullStr)`                  | Canonical form, or `nullStr` if null.              |
| `PrettyString()`                     | Canonical form, or `NULL` if null.                 |
| `GoString()`                         | ``uu.NullableID(`...`)``, or `uu.IDNull` if null.  |
| `StringBytes()`, `Hex()`, `Base64()` | Like the `ID` methods. A null value returns the zero UUID in that format. |
| `Version()`, `Variant()`             | Like the `ID` methods.                             |
| `Valid()`, `Validate()`              | Like the `ID` methods, but `IDNull` is valid.      |
| `JSONSchema()`                       | One of a `uuid` string or `null`.                  |

Implements: `fmt.Stringer`, `fmt.GoStringer`, `json.Marshaler`, `json.Unmarshaler`, `encoding.TextMarshaler`, `encoding.TextUnmarshaler`, `encoding.BinaryMarshaler`, `encoding.BinaryUnmarshaler`, `driver.Valuer`, `sql.Scanner`, `pretty.Stringer`, `nullable.Nullable`, and with a pointer receiver `nullable.NullSetable[ID]`.

## IDSlice

`type IDSlice []ID` is an ordered list of IDs. Use it instead of `[]ID`. A nil slice is SQL `NULL` and JSON `null`, an empty non-nil slice is the empty array.

| Function                           | Description                                        |
|------------------------------------|----------------------------------------------------|
| `IDSliceFromString(s)`             | Parses the `String()` format `[id1,id2]`, also without brackets. Trims the elements. `""`, `[]`, `null` and `NULL` return nil. |
| `IDSliceFromStrings(strs)`         | Parses every string with `IDFromString`. Returns nil for no strings. |
| `IDSliceMustFromStrings(strs...)`  | Like `IDSliceFromStrings`, but panics on an error. |
| `IDSliceMust[T IDSource](vals...)` | Converts every value with `IDMust`, so it panics on invalid IDs. Returns nil for no values. |

| Method                              | Description                                        |
|-------------------------------------|----------------------------------------------------|
| `Len()`, `Less(i, j)`, `Swap(i, j)` | `sort.Interface` in `ID.Less` order.               |
| `Sort()`                            | Sorts in place in `ID.Less` order.                 |
| `SortedClone()`                     | Sorted copy.                                       |
| `Clone()`                           | Copy. A nil slice returns nil.                     |
| `Contains(id)`                      | Whether `id` is in the slice.                      |
| `IndexOf(id)`                       | Index of the first `id`, or -1.                    |
| `ContainsAny(other)`                | Whether any ID of the `IDSlice` `other` is in the slice. |
| `ContainsAnyFromSet(set)`           | Whether any ID of the slice is in `set`.           |
| `Equal(other)`                      | Same IDs in the same order. nil and empty are equal. |
| `RemoveFirst(id)`                   | Removes the first `id`, returns its index or -1.   |
| `RemoveAll(id)`                     | Removes every `id`, returns the count.             |
| `RemoveAt(index)`                   | Removes the element at `index`, panics if out of range. |
| `AsSet()`                           | New `IDSet` with the IDs. Never nil, also not for a nil slice. |
| `AsSlice()`                         | The slice unchanged, for the `IDs` interface.      |
| `Raw()`                             | The same memory as `[][16]byte`, not a copy.       |
| `ForEach(callback)`                 | Calls `callback` for every ID and returns its first error. |
| `Strings()`                         | Canonical strings. Returns nil for an empty slice. |
| `String()`, `PrettyString()`        | `[id1,id2]`. nil and empty both return `[]`.       |

Implements: `driver.Valuer`, `sql.Scanner`, `json.Marshaler`, `json.Unmarshaler`, `encoding.TextMarshaler`, `encoding.TextUnmarshaler`, `fmt.Stringer`, `pretty.Stringer`, `sort.Interface`, `IDs`. It has no `IsNull` method.

## IDSet

`type IDSet map[ID]struct{}` is a set of IDs. Its set methods delegate to [`mapset`](../mapset/) and implement the abstract set interface of the Go collections proposal ([go.dev/issue/80590](https://go.dev/issue/80590)), like `types.Set`, `strutil.StringSet` and `email.AddressSet`.

| Function                            | Description                                        |
|-------------------------------------|----------------------------------------------------|
| `MakeIDSet(ids...)`                 | New set with the IDs. Never nil, `MakeIDSet()` is an empty set. |
| `MakeIDSetFromStrings(strs)`        | Parses every string with `IDFromString`.           |
| `MakeIDSetMustFromStrings(strs...)` | Like `MakeIDSetFromStrings`, but panics on an error. |
| `IDSetMust[T IDSource](vals...)`    | Converts every value with `IDMust`, so it panics on invalid IDs. Returns nil for no values. |
| `IDSetFromString(s)`                | Parses the `String()` format `set[id1,id2]`, also without `set` or brackets. `""`, `set[]`, `[]`, `null` and `NULL` return nil. |

| Method                           | Description                                        |
|----------------------------------|----------------------------------------------------|
| `Insert(id)`                     | Adds `id`, reports whether the set changed. Panics on a nil set. |
| `InsertAll(seq)`                 | Adds every ID of an `iter.Seq[ID]`, reports whether the set changed. Panics on a nil set. |
| `Delete(id)`                     | Removes `id`, reports whether the set changed.     |
| `DeleteAll(seq)`                 | Removes every ID of `seq`, reports whether the set changed. |
| `DeleteFunc(del)`                | Removes every ID for which `del` returns true, reports whether the set changed. |
| `Clear()`                        | Removes all IDs.                                   |
| `Contains(id)`                   | Whether `id` is in the set.                        |
| `ContainsAll(seq)`               | Whether every ID of `seq` is in the set. True for an empty `seq`. |
| `ContainsAny(ids...)`            | Whether any of the variadic `ids` is in the set. False for no IDs. |
| `All()`                          | `iter.Seq[ID]` over the IDs in undefined order.    |
| `Union(other)`                   | New set with the IDs of both.                      |
| `UnionWith(other)`               | Adds the IDs of `other`. Panics on a nil set if `other` has new IDs. |
| `Intersection(other)`            | New set with the IDs in both.                      |
| `IntersectionWith(other)`        | Removes the IDs that are not in `other`.           |
| `Intersects(other)`              | Whether the sets have at least one ID in common.   |
| `Difference(other)`              | New set with the IDs that are not in `other`. Asymmetric. |
| `DifferenceWith(other)`          | Removes the IDs of `other`.                        |
| `SymmetricDifference(other)`     | New set with the IDs that are in exactly one of the sets. |
| `SymmetricDifferenceWith(other)` | In-place `SymmetricDifference`. Panics on a nil set if `other` has new IDs. |
| `Equal(other)`                   | Same IDs. nil and empty are equal.                 |
| `GetOne()`                       | Any one ID, or `IDNil` for an empty set.           |
| `Len()`, `IsEmpty()`             | Number of IDs, whether it is zero.                 |
| `IsNull()`                       | Whether the set is nil.                            |
| `Clone()`                        | Copy. A nil set returns nil.                       |
| `Sorted()`                       | New `IDSlice` in `ID.Less` order. Returns nil for an empty set. |
| `AsSlice()`                      | New `IDSlice` in undefined order. Returns nil for an empty set. |
| `AsSet()`                        | The set unchanged, for the `IDs` interface.        |
| `Strings()`                      | Sorted canonical strings. Returns nil for an empty set. |
| `String()`                       | `set[id1,id2]` in sorted order. nil and empty both return `set[]`. |
| `PrettyString()`                 | `[id1,id2]` in sorted order, without `set`. nil and empty both return `[]`. |
| `ForEach(callback)`              | Calls `callback` for every ID in undefined order and returns its first error. |

All reading methods, `Delete`, `DeleteAll`, `DeleteFunc`, `Clear`, `IntersectionWith` and `DifferenceWith` accept a nil set.

Implements: `driver.Valuer`, `sql.Scanner`, `json.Marshaler`, `json.Unmarshaler`, `encoding.TextMarshaler`, `encoding.TextUnmarshaler`, `fmt.Stringer`, `pretty.Stringer`, `nullable.Nullable`, `IDs`.

## IDs

`IDs` is the interface implemented by `IDSet` and `IDSlice` for passing a collection of IDs as an argument. It embeds `fmt.Stringer`, `pretty.Stringer`, `driver.Valuer` and `json.Marshaler` and has the methods `Len`, `AsSet`, `AsSlice`, `Strings` and `ForEach`. It has no scan or unmarshal methods.

## IDMutex

`IDMutex` holds one mutex per locked ID, so goroutines working on different IDs don't block each other. The zero value is ready to use. Like `sync.Mutex`, it must not be copied after first use.

| Function or method                          | Description                                        |
|---------------------------------------------|----------------------------------------------------|
| `NewIDMutex()`                              | New `IDMutex`, the same as `new(IDMutex)`.         |
| `NewIDMutexWithCallbacks(onLock, onUnlock)` | New `IDMutex` that calls `onLock` and `onUnlock` first in `Lock` and `Unlock`, before any locking. Panics in the callbacks are recovered and ignored. |
| `Lock(id)`                                  | Locks the mutex of `id`, blocks while another goroutine holds it. |
| `Unlock(id)`                                | Unlocks the mutex of `id`. Panics with `uu.IDMutex.Unlock called for non locked key <id>` if `id` is not locked. |
| `IsLocked(id)`                              | Whether `id` is locked or a `Lock(id)` call is waiting. Only a snapshot. |

## Serialization

`<id>` stands for the canonical lower case form, like `6ba7b810-9dad-11d1-80b4-00c04fd430c8`.

Writing:

| Go value              | SQL `Value()`                            | JSON                                     | Text `MarshalText()`                   |
|-----------------------|------------------------------------------|------------------------------------------|----------------------------------------|
| `ID` `IDNil`          | `'00000000-0000-0000-0000-000000000000'` | `"00000000-0000-0000-0000-000000000000"` | `00000000-0000-0000-0000-000000000000` |
| `ID` valid            | `'<id>'`                                 | `"<id>"`                                 | `<id>`                                 |
| `NullableID` `IDNull` | `NULL`                                   | `null`                                   | empty                                  |
| `NullableID` valid    | `'<id>'`                                 | `"<id>"`                                 | `<id>`                                 |
| `IDSlice(nil)`        | `NULL`                                   | `null`                                   | `[]`                                   |
| `IDSlice{}`           | `'{}'`                                   | `[]`                                     | `[]`                                   |
| `IDSlice` with IDs    | `'{"<id>","<id>"}'` in slice order       | `["<id>","<id>"]` in slice order         | `[<id>,<id>]`                          |
| `IDSet(nil)`          | `NULL`                                   | `null`                                   | `set[]`                                |
| `IDSet{}`             | `'{}'`                                   | `[]`                                     | `set[]`                                |
| `IDSet` with IDs      | `'{"<id>","<id>"}'` sorted               | `["<id>","<id>"]` sorted                 | `set[<id>,<id>]` sorted                |

Reading:

| Input                                 | `ID`                | `NullableID`                | `IDSlice`            | `IDSet`         |
|---------------------------------------|---------------------|-----------------------------|----------------------|-----------------|
| SQL `NULL`                            | error               | `IDNull`                    | nil                  | nil             |
| SQL `''`                              | error               | error                       | nil                  | empty set       |
| SQL `'{}'`                            | error               | error                       | empty slice          | empty set       |
| SQL `'{<id>,"<id>"}'`                 | error               | error                       | IDs, quotes optional | IDs             |
| SQL 16-byte `[]byte`                  | raw binary UUID     | raw binary UUID             | error                | error           |
| JSON `null`                           | no change, no error | `IDNull`                    | nil                  | nil             |
| JSON `""`                             | error               | error                       | error                | error           |
| JSON `"<id>"`                         | ID                  | ID                          | error                | error           |
| JSON `{"String":"<id>","Valid":true}` | error               | ID (`sql.NullString` shape) | error                | error           |
| JSON `[]`                             | error               | error                       | empty slice          | empty set       |
| Text empty                            | error               | `IDNull`                    | nil                  | nil             |
| Nil UUID string                       | `IDNil`, no error   | `IDNull`                    | element `IDNil`      | element `IDNil` |

The text and JSON parsers of a single ID require at least 22 bytes (an empty text is `IDNull` for `NullableID`), so 16 characters are never read as raw binary there. Only `IDFromBytes`, `NullableIDFromBytes`, `UnmarshalBinary` and `Scan` of a 16-byte `[]byte` read raw binary. `Value` and `Scan` of `IDSlice` and `IDSet` use the [PostgreSQL array text format](https://www.postgresql.org/docs/current/arrays.html), which PostgreSQL, CockroachDB and YugabyteDB understand. Databases without array types, like MySQL, MariaDB, SQLite, SQL Server and Oracle, are not supported.

## Gotchas

- `IDNil` is not SQL `NULL`. `IDNil.Value()` writes `'00000000-0000-0000-0000-000000000000'`, and `ID.Scan(nil)` returns `cannot convert <nil> to uu.ID`. Use `NullableID` for nullable columns.
- Parsing does not validate. `IDFromString("00000000-0000-0000-0000-000000000000")` returns `IDNil` and a nil error, and a string with version 0 or 9 also parses without an error. Call `Validate()` on untrusted input, or use `IDMust` or `IDFromAny`, which validate.
- `IDFromString("ABCDEFGHIJKLMNOP")` is an error, but `IDFromBytes([]byte("ABCDEFGHIJKLMNOP"))` returns `41424344-4546-4748-494a-4b4c4d4e4f50`, because 16 bytes are read as raw binary.
- `NullableID` treats an empty text as null, but JSON `""` and `Scan("")` return an error. Send JSON `null` for a missing ID.
- `omitempty` never omits an `ID` or `NullableID`, because they are arrays of length 16. Use `omitzero`: `IDNil` and `IDNull` are then left out.
- A nil and an empty `IDSlice` or `IDSet` are different in SQL and JSON, but not as text. Both print as `[]` or `set[]`, and both parse back to nil.
- Conversions lose the nil versus empty distinction. `IDSet{}.AsSlice()` and `IDSet{}.Sorted()` return a nil `IDSlice`, whose `Value()` is `NULL`. `IDSlice(nil).AsSet()` returns an empty set, whose `Value()` is `'{}'`.
- `IDSet.Scan("")` returns an empty set, but `IDSlice.Scan("")` returns nil.
- `Insert`, `InsertAll`, `UnionWith` and `SymmetricDifferenceWith` panic on a nil `IDSet`, like an assignment to a nil map. SQL `NULL` and JSON `null` produce a nil set, so check `IsNull()` or start from `MakeIDSet()` before inserting.
- `IDSet.ContainsAny` takes variadic IDs, `IDSlice.ContainsAny` takes an `IDSlice`.
- `IDSet.Difference` is the asymmetric difference. The removed `Diff` method returned the symmetric difference, which is now `SymmetricDifference`. The removed `AsSortedSlice` is now `Sorted`.
- PostgreSQL array elements must not contain whitespace or `NULL`: `Scan("{NULL}")` returns `invalid UUID length 4: "NULL"`.
- Version 7 IDs created by versions before `v1.0.0-beta.1`, which switched to RFC 9562 byte order, stored the timestamp little-endian. `V7Time()` returns wrong times for them, and they don't sort by creation time.
- Version 7 IDs have millisecond precision. IDs from the same millisecond are ordered by their random bits, not by call order. Use `IDv7DeterministicFunc` where order matters in tests.
- `ContextWithIDSequence` needs a key and so can't feed `NewID`. For `NewID` pass your own function to `ContextWithIDFunc(ctx, f)` without keys, for example `IDv7DeterministicFunc`.
- A function from `ContextWithIDSequence` panics after the last ID instead of returning `IDNil`.
- `IDMutex` is not reentrant: calling `Lock(id)` twice from the same goroutine deadlocks.
- `IDv1` and `IDv2` contain the MAC address of the host.

## Examples

Parse and validate:

```go
package main

import (
	"errors"
	"fmt"

	"github.com/domonda/go-types/uu"
)

func main() {
	for _, s := range []string{
		"6ba7b810-9dad-11d1-80b4-00c04fd430c8",
		"6BA7B810-9DAD-11D1-80B4-00C04FD430C8",
		"{6ba7b810-9dad-11d1-80b4-00c04fd430c8}",
		"urn:uuid:6ba7b810-9dad-11d1-80b4-00c04fd430c8",
		"6ba7b8109dad11d180b400c04fd430c8",
		"a6e4EJ2tEdGAtADAT9QwyA",
	} {
		id, err := uu.IDFromString(s)
		fmt.Println(id, err)
	}

	// IDFromString does not validate, IDFromAny does
	nilID, err := uu.IDFromString("00000000-0000-0000-0000-000000000000")
	fmt.Println(nilID.IsNil(), err)
	_, err = uu.IDFromAny("00000000-0000-0000-0000-000000000000")
	fmt.Println(err, errors.Is(err, uu.ErrNilID))

	id := uu.IDv5(uu.NamespaceDNS, []byte("www.example.com"))
	fmt.Println(id, id.Version(), id.Hex(), id.Base64())

	// Output:
	// 6ba7b810-9dad-11d1-80b4-00c04fd430c8 <nil>
	// 6ba7b810-9dad-11d1-80b4-00c04fd430c8 <nil>
	// 6ba7b810-9dad-11d1-80b4-00c04fd430c8 <nil>
	// 6ba7b810-9dad-11d1-80b4-00c04fd430c8 <nil>
	// 6ba7b810-9dad-11d1-80b4-00c04fd430c8 <nil>
	// 6ba7b810-9dad-11d1-80b4-00c04fd430c8 <nil>
	// true <nil>
	// Nil UUID true
	// 2ed6657d-e927-568b-95e1-2665a8aea6a2 5 2ed6657de927568b95e12665a8aea6a2 LtZlfeknVouV4SZlqK6mog
}
```

Deterministic IDs from the context, and JSON of null and empty values:

```go
package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/domonda/go-types/uu"
)

type Invoice struct {
	ID       uu.ID         `json:"id"`
	ParentID uu.NullableID `json:"parentId"`
	Tags     uu.IDSlice    `json:"tags"`
}

func main() {
	// A test puts a deterministic generator into the context,
	// the code under test calls uu.NewID(ctx) without knowing about it.
	ctx := uu.ContextWithIDFunc(context.Background(), uu.IDv7DeterministicFunc(1700000000000))

	a := uu.NewID(ctx)
	b := uu.NewID(ctx)
	fmt.Println(a)
	fmt.Println(b)
	fmt.Println(a.Version(), a.Less(b))
	t, ok := a.V7Time()
	fmt.Println(t.UTC(), ok)

	inv := Invoice{ID: a}
	j, _ := json.Marshal(inv)
	fmt.Println(string(j))

	inv.ParentID = b.Nullable()
	inv.Tags = uu.IDSlice{}
	j, _ = json.Marshal(inv)
	fmt.Println(string(j))

	// Output:
	// 018bcfe5-6800-7000-8000-000000000000
	// 018bcfe5-6801-7000-8000-000000000000
	// 7 true
	// 2023-11-14 22:13:20 +0000 UTC true
	// {"id":"018bcfe5-6800-7000-8000-000000000000","parentId":null,"tags":null}
	// {"id":"018bcfe5-6800-7000-8000-000000000000","parentId":"018bcfe5-6801-7000-8000-000000000000","tags":[]}
}
```

SQL values of zero, null and empty values:

```go
package main

import (
	"database/sql/driver"
	"fmt"

	"github.com/domonda/go-types/uu"
)

func main() {
	show := func(name string, v driver.Valuer) {
		val, err := v.Value()
		fmt.Printf("%-13s %#v %v\n", name, val, err)
	}
	show("uu.IDNil", uu.IDNil)
	show("uu.IDNull", uu.IDNull)
	show("IDSlice(nil)", uu.IDSlice(nil))
	show("IDSlice{}", uu.IDSlice{})
	show("IDSet(nil)", uu.IDSet(nil))
	show("MakeIDSet()", uu.MakeIDSet())

	var id uu.ID
	fmt.Println(id.Scan(nil))
	var n uu.NullableID
	fmt.Println(n.Scan(nil), n.IsNull())

	// Output:
	// uu.IDNil      "00000000-0000-0000-0000-000000000000" <nil>
	// uu.IDNull     <nil> <nil>
	// IDSlice(nil)  <nil> <nil>
	// IDSlice{}     "{}" <nil>
	// IDSet(nil)    <nil> <nil>
	// MakeIDSet()   "{}" <nil>
	// cannot convert <nil> to uu.ID
	// <nil> true
}
```

## Related

- [nullable](../nullable/): `NullableID` follows its zero-is-null convention and implements `nullable.NullSetable[ID]`.
- [mapset](../mapset/): the set operations behind `IDSet`.
- [Null conventions](../docs/explanation-null-conventions.md): `IDNil` versus `IDNull`, nil versus empty collections.
- [How to store optional values](../docs/howto-store-optional-values.md)
- [How to use deterministic IDs in tests](../docs/howto-deterministic-ids-in-tests.md)
- [Sets in go-types](../docs/explanation-sets.md)
- [RFC 9562](https://www.rfc-editor.org/rfc/rfc9562) (UUID versions 1 to 8, obsoletes RFC 4122), [RFC 4122](https://www.rfc-editor.org/rfc/rfc4122), [DCE 1.1: Authentication and Security Services](https://pubs.opengroup.org/onlinepubs/9696989899/chap5.htm#tagcjh_08_02_01_01)
- API reference: [pkg.go.dev/github.com/domonda/go-types/uu](https://pkg.go.dev/github.com/domonda/go-types/uu)

## Origin and license

This package is derived from [github.com/satori/go.uuid](https://github.com/satori/go.uuid), Copyright (C) 2013-2016 by Maxim Bublis <b@codemonkey.ru>, released under the MIT License. See the [satori/go.uuid LICENSE](https://github.com/satori/go.uuid/blob/master/LICENSE) for the license text.
