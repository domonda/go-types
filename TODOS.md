# TODOS

Open issues and cleanup items before tagging `v1.0.0`. Grouped by what blocks a v1 release versus what's nice to land afterward.

## Blockers for v1.0 (API or semantic)

Changing any of these after `v1.0.0` would break callers, so each needs a
decision first. They are inconsistencies between packages, not bugs in one
package: every listed behaviour is documented in the package READMEs.

- [ ] **Tag `v1.0.0`.** No git tags exist yet. Pick semver baseline first.
- [ ] **One rule for empty input to a nullable type.**
      `Scan("")` returns an error for `nullable.NonEmptyString` and
      `email.NullableAddress` but null for `nullable.TrimmedString` and
      `account.NullableNumber`. JSON `""` is a present zero for
      `money.NullableAmount`, `NullableDecimalAmount` and `NullableRate` but
      an error for `NullableCentAmount` and `uu.NullableID`.
      `uu.IDSlice.Scan("")` gives a nil slice, `uu.IDSet.Scan("")` an empty
      non-nil set. Whitespace-only values are null for the country, vat,
      bank, money, email address and date nullable types, but a real value
      for `email.NullableAddressList`, `date.NullableYearMonth`,
      `date.NullableYearQuarter` and `nullable.NonEmptyString`, and
      `account.NullableNumber` reports one as not null while writing JSON
      `null`. `email.NullableAddress.Scan("")` is an error while
      `Scan("  ")` is null. `date.NullableDate` JSON strings are stored as
      given while `Scan` normalizes them. Recommendation: empty and
      whitespace-only input is null for every nullable type, because an
      empty CSV cell or text column means "no value".
- [ ] **One rule for what `Value` writes.** `country.Code` and `vat.ID` write
      their normalized form, `money.Currency`, `language.Code`, `bank.IBAN`
      and `bank.BIC` write the raw string. `String()` normalizes for all of
      them, which hides the difference in logs.
- [ ] **One rule for the empty value of a non-nullable type.**
      `country.Code("")` writes SQL NULL and JSON `""`, `bank.IBAN("")`
      writes `""` to both, and `account.Number("")` writes SQL NULL and JSON
      `null` but reads neither back (`Scan(nil)` and JSON `null` are errors).
- [ ] **`language.NullableCode`.** `language.Code` has no nullable variant,
      and `language.Null.Validate()` fails, so an optional language field
      fails `types.DeepValidate`.
- [ ] **Naming:** `money.Currency.NullableCurrency()` is called `Nullable()`
      on every other type.
- [ ] **`ValidAndNotNull` on every nullable type.** It is missing on
      `uu.NullableID`, `account.NullableNumber` and `email.NullableAddress`,
      and no package asserts `nullable.NullableValidator` at compile time.
- [ ] **SQL and JSON methods for `date.NullableYearMonth` and
      `date.NullableYearQuarter`.** Without them null is written as `""`
      instead of SQL NULL and JSON `null`; adding them later changes stored
      values.
- [ ] **Consolidate the three per-key mutexes.** `types.KeyMutex[T]`,
      `uu.IDMutex` and `strutil.StrMutex` implement the same algorithm.
      Keep the generic `KeyMutex` (newest, has `IsLocked`), and make the other
      two thin wrappers or remove them; `uu.IDMutex` additionally has lock
      and unlock callbacks.
- [ ] **`bank` CAMT.053 structs.** The `time.Time` fields require RFC 3339,
      so an `ISODateTime` without an offset, as many banks export, fails the
      whole decode. Only the camt.053.001.02 element names are mapped (`BIC`,
      not `BICFI`; `Dt`, not `DtTm`), and only the structured remittance
      reference is captured (no `Ustrd`, `EndToEndId` or `BkTxCd`). Fixing
      this changes field types.
- [ ] **Values in validation errors.** `types.Validate` names the invalid
      value with `%#v` (commit 50a091b), so error messages and logs contain
      pointer addresses and possibly secrets. Decide whether to keep it,
      truncate it, or make it opt-in.
- [ ] **Message content in `email` parse errors.** `ParseMessage`,
      `ParseMIMEMessage` and `ParseTNEFMessageBytes` wrap errors with
      go-errs function parameters, so up to 5000 bytes of the raw message
      per wrap level end up in error strings and logs.

## After v1.0

- [ ] **Replace `mapset` with the standard library `container/mapset`** once
      Go 1.28 ships. The package mirrors the proposed API function for function,
      so it should be an import path change. `internal/collections/settest` runs
      the abstract set specification against every set type to catch behaviour
      drift when it is swapped.
- [ ] **Reconsider `types.Set` and the `set` package** at the same time.
      `container/set.Set` covers both, and neither has many callers
      (`types.Set` 3 call sites, `set` 5 across all known consumers).
- [ ] **Relax `cmp.Ordered` to `comparable`** in `set`, `types.ReduceSet` and
      `types.ReduceSlice`, which never compare order. This is backward
      compatible.
- [ ] **`date.YearQuarter.Validate`** requires an upper case `Q` while
      `date.PeriodRange` accepts `q`.
- [ ] **Move the `NullableYearQuarter` tests** from
      `date/yearquarter_test.go` into `date/nullableyearquarter_test.go`.
- [ ] **Multi-word currency names in `money.ParseCurrencyAmount`.** Only a
      single word is passed to `Currency.Normalized`, so
      `"100 Canadian Dollar"` does not parse.
- [ ] **`strfmt` and other `nullable.Type[T]`.** `money.NullableRate` and
      other `nullable.Type[T]` without a registered formatter are written
      with `fmt.Sprint`.
- [ ] **Time zones in the `strfmt` presets.** The presets' time layouts end
      in `MST`, so a non-UTC offset is lost when scanning back.
- [ ] **`email` dates and threading.** A `Date` header with a named zone
      other than UT/GMT, like `EST`, gets a zero offset.
      `NewReplyMessage` does not fall back to the parent's `In-Reply-To` when
      it has no `References`.
- [ ] **`bank.IBANFinder`** only finds IBANs written compact and upper case.
- [ ] **`queue.Queue`** has no way to drain the queue gracefully before
      `Close`, which discards undelivered items.
- [ ] **`float.ParseDetails("1,5e3")`** reports `.` as the decimal separator.
- [ ] **`strutil.TransliterateSpecialCharacters`** drops a valid U+FFFD
      together with invalid bytes.
- [ ] **Bump `gosec` in `tools/go.mod`.** The pinned version cannot read the
      build data of Go 1.27 and then runs without type information, so its
      "0 issues" is meaningless locally. CI on Go 1.26 is not affected.
- [ ] **Go 1.27 `encoding/json`.** With json/v2 as the default, `float64` and
      `any` map keys marshal, unlike in Go 1.26. `types.CanMarshalJSON`
      follows the documented Go 1.26 rule and reports false for them.

## Completed

- [x] **Bug sweep and documentation pass.** Audited every package against
      its README, fixed the bugs found (see the Changed and Fixed sections of
      the CHANGELOG, many of them breaking), rewrote all package READMEs as
      complete references with verified examples, added package doc
      comments, and added the tutorial, how-to and explanation docs in
      `docs/`.
      **Completed:** 2026-10-12
- [x] **Remove `uu.IDSet.AsSortedSlice` and `uu.IDSet.AddIDs`,** both
      deprecated in favour of `Sorted` and `InsertAll`/`UnionWith`.
      **Completed:** 2026-10-09
- [x] **Remove the deprecated set methods.** `Add`, `AddSlice`, `AddSet`,
      `DeleteSlice` and `DeleteSet` on `types.Set`, `uu.IDSet` and
      `strutil.StringSet`, plus `DeleteSlice` and `DeleteSet` on
      `email.AddressSet`.
      **Completed:** 2026-10-09
- [x] **No `CHANGELOG.md`.** Add one for the v1 cut.
      **Completed:** 2026-09-01 (unreleased section; no version tagged yet)
- [x] **Require a Go version that accepts the collections proposal's
      self-referential constraints.** Raised the floor to Go 1.26.0.
      **Completed:** 2026-09-01
