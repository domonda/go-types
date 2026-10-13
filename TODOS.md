# TODOS

Open issues and cleanup items before tagging `v1.0.0`. Grouped by what blocks a v1 release versus what's nice to land afterward.

## Blockers for v1.0 (API or semantic)

Changing any of these after `v1.0.0` would break callers, so each needs a
decision first. They are inconsistencies between packages, not bugs in one
package: every listed behaviour is documented in the package READMEs.

- [ ] **Tag `v1.0.0`.** No git tags exist yet. `VERSION` holds the first
      pre-release, `v1.0.0-beta.1`, to tag before the final release.
- [ ] **One rule for empty input to a nullable type.**
      `Scan("")` returns an error for `nullable.NonEmptyString` but null for
      `nullable.TrimmedString`, `email.NullableAddress`,
      `email.NullableAddressList` and `account.NullableNumber`. JSON `""` is
      a present zero for `money.NullableAmount`, `NullableDecimalAmount` and
      `NullableRate` but an error for `NullableCentAmount` and
      `uu.NullableID`. `uu.IDSlice.Scan("")` gives a nil slice,
      `uu.IDSet.Scan("")` an empty non-nil set. Whitespace-only values are
      null for the country, vat, bank, money, account number, email address
      and date nullable types, but a real value for `date.NullableYearMonth`,
      `date.NullableYearQuarter` and `nullable.NonEmptyString`.
      `email.NullableAddressList.Scan("  ")` is null, but
      `NullableAddressList("  ").IsNull()` is false. `date.NullableDate` JSON
      strings are stored as given while `Scan` normalizes them.
      Decided: empty and whitespace-only input is null for every nullable
      type, see Design decisions in the [README](README.md).
- [ ] **One rule for the empty value of a non-nullable type.**
      `country.Code("")` writes SQL NULL and JSON `""`, `bank.IBAN("")`
      writes `""` to both, and `account.Number("")` writes SQL NULL and JSON
      `null` but reads neither back (`Scan(nil)` and JSON `null` are errors).
      `date.Date`, `language.Code`, `money.Currency`, `vat.ID`, `uu.ID`,
      `notnull.TrimmedString` and `money.DecimalAmount` differ in the same
      way; [docs/explanation-null-conventions.md](docs/explanation-null-conventions.md)
      has the full table.
- [ ] **`language.NullableCode`.** `language.Code` has no nullable variant,
      and `language.Null.Validate()` fails, so an optional language field
      fails `types.DeepValidate`.
- [ ] **Naming:** `money.Currency.NullableCurrency()` is called `Nullable()`
      on every other type, and `money.CurrencyNull` is the only untyped null
      constant; every other package declares its null value as a typed
      constant.
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

## Decisions before v1.0: unfinished or misplaced API

Found in a code audit on 2026-10-12. Each item removes or changes exported
API or behaviour, so it should be decided before `v1.0.0`. "No caller" means
none in this module or in the local checkouts of the domonda repos.

- [ ] **`GetOr` of the nullable types.** `money.NullableCurrency.GetOr`
      returns the default for null and for an invalid currency, while
      `GetOr` of `bank.NullableIBAN`, `bank.NullableBIC`,
      `country.NullableCode`, `vat.NullableID`, `date.NullableDate` and
      `email.NullableAddress` returns it only for null, and the invalid
      value otherwise. Pick one meaning, or give the validating variant
      another name.
- [ ] **`country.NullableCode.NormalizedWithAltCodes`.** It returns the
      same result as `Normalized`, which already maps the `AltCodes`, and
      `country.Code` has no such method. Remove it or deprecate it.

- [ ] **`internal/pq`, a vendored copy of `lib/pq` from 2022.** Its README
      says "until refactored for go-types". Only `StringArray` and
      `BoolArray` are used, as the aliases `nullable.StringArray` and
      `nullable.BoolArray` and inside `notnull.StringArray`. The other array
      types, `NullTime`, the timestamp functions, most of `encode.go` and
      the `oid` package are dead. Its parser also disagrees with
      `internal.SplitArray`: `nullable.StringArray.Scan` reads `{a\,b}` as
      the two elements `a\` and `b`, while `nullable.SplitArrayValues` and
      PostgreSQL read the single element `a,b`. Options: delete the dead
      code and keep both parsers, or implement the two types on
      `internal.SplitArray` and delete `internal/pq`. Recommendation: the
      second, so every array type parses like PostgreSQL. The parsers also
      accept different booleans: `nullable.BoolArray` rejects
      `{true,false}`, which `nullable.NullBoolArray` and
      `notnull.NullBoolArray` accept like PostgreSQL does.
- [ ] **`email.Rule`, `RuleFunc`, `BoolRule`, `AllRule` and `AnyRule`.**
      Nothing accepts a `Rule` or calls `AppliesToMessage`, and the message
      filter that `email/doc.go` implies doesn't exist. No caller. Remove
      them, or add the filter.
- [ ] **`email.ProviderDomains`.** A hardcoded list of 112 domains,
      including some that aren't email providers, like `wu.ac.at` and
      `guest.booking.com`. Its only caller,
      `email.Address.IsFromKnownEmailProvider`, was removed in fbce477, and
      domonda-service pulls its own list. No caller. Recommendation: remove.
- [ ] **`strfmt.ScanConfig.AcceptedMoneyAmountDecimals`.** It defaults to
      `{0, 2, 4}`, but `Scan` never reads it, and the docs tell callers to
      pass it to `money.ParseAmount` themselves. Apply it in `Scan` for
      money destinations, which changes what the default config accepts,
      or remove it like `FormatConfig.Percent`.
- [ ] **`types.LenString`.** No caller. Its zero value rejects every
      non-empty string because min and max are 0, it has no SQL `Scan` and
      `Value` unlike every other type, and with only pointer receivers a
      struct field marshalled by value encodes as `{}`. Remove it, or
      finish it.
- [ ] **`strfmt.Parser`.** `strfmt/doc.go` calls it "the older interface".
      Nothing consumes it, ten types document that they implement it.
      Remove the interface and those doc claims (the `Parse` methods can
      stay), or give it a consumer.
- [ ] **`date.Format` and `date.Formatter`.** Both have the same `Format`
      and `Parse` methods, `date.Format` adds `NilString`, `ZeroString` and
      `AssignString`. Keep one.
- [ ] **`strutil` helpers without a caller.** `EqualJSON`,
      `HTMLEscapeSpecialRunes`, `ParseDomainName` and `RandomStringBytes`
      aren't used anywhere, and the first two aren't string manipulation.
      Remove them, or keep them as public API. `DerefPtr` duplicates
      `types.FromPtr` and `IndexInStrings` duplicates `slices.Index`, but
      both have callers.
- [ ] **Two HTML to text converters in `email`.** The hand-written
      `HTMLToPlaintext`, which domonda-service calls, and the
      `inbucket/html2text` library used for TNEF bodies. Keep one.

## Cleanup decisions (internal or additive)

From the same audit. These don't remove exported API, so they can land any
time.

- [ ] **Date formats marked TODO in `date/date_test.go`.** `8/8/71`,
      `2014-04`, `2014.05`, `2014`, `201412` and `20140601` are
      commented-out cases that `Normalize` still rejects, while `08/08/71`
      works. `date.Finder` uses `Normalize`, so accepting digit-only forms
      would make it report plain numbers like `2014` as dates.
      Recommendation: support `8/8/71`, decide the digit-only forms with
      the finder in mind, and drop the year-month cases, which
      `date.YearMonth` covers.
- [ ] **TNEF attachments** in `email/parsetnef.go`. `Inline: false, // ??`
      is an open question, attachments get no `ContentType` although
      `email.NewAttachment` detects it, and the TNEF attributes holding
      subject and sender are ignored. Fill them in, or keep the gaps
      documented in the README.
- [ ] **Address list with quoted commas.** In
      `email/parseaddress_test.go`, a commented-out case that
      `ParseAddressList` splits wrongly is marked "TODO resolve this case",
      and `numRandomCombinations` is 0, so the random list tests never run.
- [ ] **`date.Format.AssignString` dead code.** It has a commented-out
      `loc` parameter and time zone conversion, and `tPtr` starts non-nil,
      so its three `nil` branches never run (staticcheck SA4031).
      `date/date.go` also assigns `val0` twice without reading it
      (SA4006). Implement the time zone or delete the dead parts.
- [ ] **`charset` `encodingImpl.bom`.** No constructor sets the field, so
      `BOM()` always returns `NoBOM` as documented, also for
      `GetEncoding("UTF-16")`, which writes a BOM. Delete the field, or set
      it for UTF-16.
- [ ] **Error constructors.** Decided: go-errs, see Design decisions in
      the [README](README.md). The module still calls `errors.New` and
      `fmt.Errorf` about 340 times and go-errs `errs.New` four times, in
      `email/message.go`, `email/parsetnef.go`, and the whitespace-only
      errors of `AddressList.Scan` and `AddressSet.Scan`. Convert the
      standard library calls to `errs.New` and `errs.Errorf`.
- [ ] **Commented-out code.** An old implementation after the `return` in
      `money.Amount.String`, old `From`/`ReplyTo` assignments in
      `email/parsemime.go`, `email/utils.go`, a debug print in
      `email/parseaddress.go`, and blocks in `email/parseaddress_test.go`
      and `uu/nullableid_test.go`.
- [ ] **Test leftovers.** `TestStandardComparison` in
      `email/addresslist_test.go` only tests `net/mail`. Bare `// TODO`
      lines end the tables in `strfmt/scan_test.go` and the `AddMonths`
      test in `date/date_test.go`. Test comments in
      `language/code_test.go`, `money/currency_test.go`,
      `country/code_test.go` and `email/utils_test.go` carry review-tool
      metadata (`protects=… why_new=… seam=none`, "this branch changed").
      `charset/encoding_test.go` holds the tests of `utf16.go`, `utf32.go`
      and `bom.go`, some of them twice.
- [ ] **Quadratic `email` address parsing.** `ParseAddressList` copies the
      whole list into every per-element error and keeps all of them
      (`email/parseaddress.go`, the two `errList = append` lines), and
      `parseAddress` recurses once per address-like token, matching
      `nameAddressRegexp` against the whole rest each time. Measured in the
      2026-10-12 review: an invalid 40 KB list builds a 481 MB error in 9 s,
      a 32 KB address without commas takes 3.5 s, and it is reachable from
      `Address.Validate`, `AddressList.Validate` and the `Normalized` and
      `Split` methods. Cap and truncate the collected errors, make the
      duplicated-address case a single second attempt instead of a
      recursion, and reject inputs beyond a few KB.
- [ ] **Fast paths for the exact money and date math.** The exact
      rewrites made hot paths slower, measured against master in the
      2026-10-12 review: `CentAmount.MultipliedByRate`, `DividedByRate` and
      `Percentage` 2.4 ns to about 630 ns with about 20 allocations
      (`big.Rat`), `float.RoundToDecimals` and so `Amount.RoundToCents`,
      `Cents` and `WithinOneCent` 2.5 ns to about 185 ns (a
      `FormatFloat`/`ParseFloat` round trip). Keep the exact results:
      int128 arithmetic with `math/bits` for the rates like
      `DecimalAmount.Mul`, and `math.Round(f*pow)/pow` for rounding unless
      the scaled value is within a few ulps of a tie.
      The /ship review of 2026-10-13 measured more:
      `Amount.DecimalAmount` 9 ns to 95 ns (a string round trip),
      `internal.UnmarshalJSONStringOrNull` about twice as slow as decoding a
      plain string (a nested `json.Unmarshal`, a fast path for unescaped
      strings would do), `CurrencyAmount.ScanString` parses twice,
      `types.DeepValidate` formats a path label for every slice element
      (about 4 allocations per byte of a `[]byte` field, skip element types
      that can't hold a validator), uncached reflection per value in
      `strfmt.FormatValue`, and `vat.ID.Normalized` copies a huge invalid
      input several times before its length check rejects it.

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
- [ ] **Time zones in the `strfmt` presets.** The English and German
      presets' time layouts end in `MST` without a numeric offset, so a
      formatted time only scans back where its zone abbreviation is known
      (`UTC`, `GMT` or one of `time.Local`); elsewhere `ParseTime` rejects
      it. The English layout is also day first and reads US month-first
      input with a known zone as another date. A numeric offset in the
      layouts would make the round trip lossless.
- [ ] **`email` dates.** A `Date` header with a named zone other than
      UT/GMT, like `EST`, gets a zero offset.
- [ ] **`bank.IBANFinder`** only finds IBANs written compact and upper case.
- [ ] **`queue.Queue`** has no way to drain the queue gracefully before
      `Close`, which discards undelivered items.
- [ ] **`strutil.TransliterateSpecialCharacters`** drops a valid U+FFFD
      together with invalid bytes.
- [ ] **Bump `gosec` in `tools/go.mod`.** The pinned version cannot read the
      build data of Go 1.27 and then runs without type information, so its
      "0 issues" is meaningless locally. CI on Go 1.26 is not affected.
- [ ] **Go 1.27 `encoding/json`.** With json/v2 as the default, `float64` and
      `any` map keys marshal, unlike in Go 1.26. `types.CanMarshalJSON`
      follows the documented Go 1.26 rule and reports false for them.

## Completed

- [x] **One rule for what `Value` writes.** The types with a `Normalized`
      method write the normalized form in `Value` and `MarshalJSON`, and
      their `Validate` normalizes first, also for `language.Code` and
      `bank.BIC`. `money.CurrencyAmount` and `CurrencyDecimalAmount` write
      their currency normalized. Normalized email addresses keep non-ASCII
      names as UTF-8, only `BuildRawMessage` encodes them with RFC 2047.
      **Completed:** 2026-10-13
- [x] **`strfmt` and other `nullable.Type[T]`.** A non-null
      `nullable.Type[T]` without a registered formatter, like
      `money.NullableRate`, is formatted like its `T` instead of with
      `fmt.Sprint`.
      **Completed:** 2026-10-13
- [x] **Fixes from the code review of PR #28.** `date`: full "August"
      dates parse, a number with an ordinal suffix is never the month,
      year 0 dates give the zero time, `NullableDate.Format(Layout)` is
      `""` for every null date, comparisons parse each date once.
      `language.Code.Validate` was strict like `Valid` (since normalizing
      like the other types, see the entry above). UTF-16 text starting
      with the UTF-8 BOM bytes decodes. `account.NullableNumber` treats a
      whitespace-only value as null in every method.
      `float.ParseDetails("1,5e3")` reports `,`. `email.NullableAddress`
      and `NullableAddressList` scan `""` and whitespace as null, and
      `NewReplyMessage` falls back to the parent's `In-Reply-To`.
      `strfmt.ScanConfig.ParseTime` rejects zone abbreviations with an
      unknown offset. The copied `UnmarshalJSON` of the nullable string
      types and the number array parsers share `internal` helpers.
      **Completed:** 2026-10-13
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
