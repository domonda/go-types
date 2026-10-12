# How to use deterministic IDs in tests

This guide shows how to make code that generates UUIDs produce known IDs in
tests, so that tests can assert on them and golden files stay stable. The
mechanism is a function in the `context.Context` that `uu.NewID` calls instead
of generating a random ID.

## Prerequisites

- `github.com/domonda/go-types` in your `go.mod`, Go 1.26 or later.
- Code under test that receives a `context.Context` where it creates IDs.

## Steps

### 1. Generate IDs from the context in production code

Call `uu.NewID(ctx)` wherever the code creates an ID, instead of `uu.IDv7()`
or `uu.IDv4()`:

```go
package invoice

import (
	"context"

	"github.com/domonda/go-types/uu"
)

type Invoice struct {
	ID    uu.ID
	Lines []Line
}

type Line struct {
	ID   uu.ID
	Text string
}

// New creates an invoice with one line per text.
func New(ctx context.Context, texts ...string) *Invoice {
	inv := &Invoice{ID: uu.NewID(ctx)}
	for _, text := range texts {
		inv.Lines = append(inv.Lines, Line{ID: uu.NewID(ctx), Text: text})
	}
	return inv
}
```

Without a function in the context, `uu.NewID` calls `uu.IDvDefault`, which is
`uu.IDv7`, so production behaviour doesn't change.

### 2. Add a deterministic ID function to the test context

`uu.IDv7DeterministicFunc(start)` returns a function that yields the version 7
IDs `uu.IDv7Deterministic(start)`, `uu.IDv7Deterministic(start+1)` and so on.
`uu.ContextWithIDFunc` without keys installs it for `uu.NewID`:

```go
ctx := uu.ContextWithIDFunc(context.Background(), uu.IDv7DeterministicFunc(1))
```

### 3. Assert on the generated IDs

The n-th call returns `uu.IDv7Deterministic(start + n - 1)`. Compare with that,
or with the string, which encodes the number as the millisecond timestamp of
the version 7 layout:

```go
func TestNew(t *testing.T) {
	ctx := uu.ContextWithIDFunc(context.Background(), uu.IDv7DeterministicFunc(1))

	inv := invoice.New(ctx, "first", "second")

	if want := uu.IDv7Deterministic(1); inv.ID != want {
		t.Errorf("invoice ID = %s, want %s", inv.ID, want)
	}
	if got, want := inv.Lines[1].ID.String(), "00000000-0003-7000-8000-000000000000"; got != want {
		t.Errorf("second line ID = %s, want %s", got, want)
	}
}
```

If the code under test reads the timestamp of the IDs with `ID.V7Time`, start
the series at a realistic time instead of `1`, like
`time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC).UnixMilli()`.

### 4. Return one fixed ID

`uu.ContextWithID(ctx, id)` installs a function that returns `id` on every
call. Use it when the code creates exactly one ID; a second call returns the
same ID again.

### 5. Give single call sites their own IDs

When a test needs to fix one ID among many, give that call site its own
context key and read the ID with `uu.IDFromContext(ctx, key)`:

```go
// IDKeys holds the context keys for IDs that tests may want to set
// one by one. Each byte field has its own address.
var IDKeys struct {
	Payment byte
}

// Pay returns the ID of a new payment for the invoice.
func Pay(ctx context.Context, inv *Invoice) uu.ID {
	return uu.IDFromContext(ctx, &IDKeys.Payment)
}
```

The key is compared by address, so use a pointer to a variable of non-zero
size, like the `byte` fields above. The Go specification allows pointers to
distinct zero-size variables, like two `struct{}` values, to be equal.

In the test, `uu.ContextWithIDSequence` returns the passed IDs in order for
that key:

```go
func TestPay(t *testing.T) {
	first := uu.IDMust("6fb0f2bc-cba7-4cdb-ac36-87404e7c94cc")
	second := uu.IDMust("edb4b646-2989-4128-bcdc-daf48a56c7fa")
	ctx := uu.ContextWithIDSequence(context.Background(), &invoice.IDKeys.Payment, first, second)
	inv := invoice.New(ctx, "line") // random IDs, no NewID function in ctx

	if got := invoice.Pay(ctx, inv); got != first {
		t.Errorf("first payment ID = %s, want %s", got, first)
	}
	if got := invoice.Pay(ctx, inv); got != second {
		t.Errorf("second payment ID = %s, want %s", got, second)
	}
}
```

`uu.ContextWithID(ctx, id, keys...)` and `uu.ContextWithIDFunc(ctx, f, keys...)`
accept keys too. Without a function for its key, `uu.IDFromContext` generates
a random ID with `uu.IDvDefault`; it never returns `uu.IDNil`.

### 6. Assert on sets when the code runs concurrently

The functions of `uu.IDv7DeterministicFunc` and `uu.ContextWithIDSequence` are
safe for concurrent use: every call gets a different ID of the series. Which
goroutine gets which ID depends on scheduling, so compare the set of IDs:

```go
func TestNew_Concurrent(t *testing.T) {
	ctx := uu.ContextWithIDFunc(context.Background(), uu.IDv7DeterministicFunc(1))

	invoices := make([]*invoice.Invoice, 10)
	var wg sync.WaitGroup
	for i := range invoices {
		wg.Go(func() { invoices[i] = invoice.New(ctx, "line") })
	}
	wg.Wait()

	// Which goroutine got which ID depends on scheduling,
	// so compare the set of IDs, not their order.
	got := make(uu.IDSet)
	for _, inv := range invoices {
		got.Insert(inv.ID)
		got.Insert(inv.Lines[0].ID)
	}
	want := make(uu.IDSet)
	for n := range int64(20) {
		want.Insert(uu.IDv7Deterministic(n + 1))
	}
	if !got.Equal(want) {
		t.Errorf("IDs = %s, want %s", got, want)
	}
}
```

### 7. Define enough IDs for a sequence

A sequence from `uu.ContextWithIDSequence` panics when the code asks for more
IDs than it holds, and stays exhausted:

```
panic: uu.ContextWithIDSequence: all 2 IDs of the sequence have already been used
```

It panics instead of returning `uu.IDNil`, because using more IDs than the
test defined is a bug in the test, and the invalid `uu.IDNil` must not end up
as a primary key. The function of `uu.IDv7DeterministicFunc` has no such
limit, it keeps counting.

### Behaviour at a glance

This program prints what the functions of the steps return:

```go
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/domonda/go-types/uu"
)

var idKeys struct {
	payment byte
}

func main() {
	ctx := uu.ContextWithIDFunc(context.Background(), uu.IDv7DeterministicFunc(1))
	fmt.Println(uu.NewID(ctx))
	fmt.Println(uu.NewID(ctx))
	fmt.Println(uu.NewID(ctx) == uu.IDv7Deterministic(3))

	// A realistic start time for code that reads the timestamp
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC).UnixMilli()
	id := uu.IDv7Deterministic(start)
	t, ok := id.V7Time()
	fmt.Println(id, t.UTC(), ok)

	// A fixed ID comes back on every call
	fixed := uu.IDMust("6fb0f2bc-cba7-4cdb-ac36-87404e7c94cc")
	ctx = uu.ContextWithID(context.Background(), fixed)
	fmt.Println(uu.NewID(ctx) == fixed, uu.NewID(ctx) == fixed)

	// A function added for a key does not affect NewID
	ctx = uu.ContextWithID(context.Background(), fixed, &idKeys.payment)
	fmt.Println(uu.IDFromContext(ctx, &idKeys.payment) == fixed)
	fmt.Println(uu.IDFuncFromContext(ctx) == nil, uu.NewID(ctx) == fixed)

	// Output:
	// 00000000-0001-7000-8000-000000000000
	// 00000000-0002-7000-8000-000000000000
	// true
	// 018cc251-f400-7000-8000-000000000000 2024-01-01 00:00:00 +0000 UTC true
	// true true
	// true
	// true false
}
```

## Verification

Run the tests with the race detector, which also checks that concurrent
calls of the ID functions are synchronized:

```sh
go test -race -count=1 ./...
```

Run them twice: assertions on deterministic IDs pass on every run, while an
assertion that accidentally depends on a random ID fails.

## Troubleshooting

**The IDs are random although the test installed a function.** Check, in this
order:

- The code under test calls `uu.IDv7()`, `uu.IDv4()` or another generator
  directly instead of `uu.NewID(ctx)`.
- The code creates IDs with a context that doesn't derive from the test
  context, for example `context.Background()` deeper in the call chain.
- The test passed keys to `uu.ContextWithIDFunc` or `uu.ContextWithID`. With
  keys, the function is only installed for those keys, not for `uu.NewID`.
  `uu.IDFuncFromContext(ctx)` returns nil in that case.
- The code calls `uu.IDFromContext(ctx, key)` with a key the test didn't set.
  A keyed lookup doesn't fall back to the `uu.NewID` function; it generates a
  random ID.

**`panic: uu.ContextWithIDSequence: all 2 IDs of the sequence have already
been used`.**
The code under test created more IDs for that key than the test defined. Add
IDs to the sequence, or find the extra call.

**Every row gets the same ID, or the database reports a duplicate key.**
`uu.ContextWithID` returns its ID on every call. Use
`uu.IDv7DeterministicFunc` or a sequence when the code creates more than one
ID.

**Two keyed sequences hand out each other's IDs.**
The keys share an address, which happens with pointers to zero-size variables
like `struct{}`. Use `byte` fields, as in step 5.

**Parallel tests against a shared database collide on primary keys.**
Every `uu.IDv7DeterministicFunc(1)` starts the same series. Give each test its
own start value.

**Golden files contain `01000000-0000-7000-8000-000000000000` where
`uu.IDv7Deterministic(1)` now returns `00000000-0001-7000-8000-000000000000`.**
The golden values were recorded before version 7 timestamps were written in
the big-endian byte order RFC 9562 requires. Earlier versions of the module
wrote them in native byte order, which is little-endian on amd64 and arm64.
Regenerate the golden files.

## Related

- [uu reference](../uu/README.md)
- [How to store optional values](howto-store-optional-values.md): `uu.ID`
  versus `uu.NullableID` columns.
