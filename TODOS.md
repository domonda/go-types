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
      and date nullable types, but a real value for
      `nullable.NonEmptyString`.
      Decided: empty and whitespace-only input is null for every nullable
      type, see Design decisions in the [README](README.md).
- [ ] **One rule for the empty value of a non-nullable type.**
      The normalizing types (`country.Code`, `money.Currency`,
      `language.Code`, `vat.ID`, `bank.IBAN`, `bank.BIC`, `date.Date`,
      `date.YearMonth`, `date.YearQuarter`, `email.Address`,
      `email.AddressList`) write it as SQL NULL and JSON
      `""` and read SQL NULL back as it. `account.Number("")` writes SQL NULL
      and JSON `null` but reads neither back (`Scan(nil)` and JSON `null` are
      errors). `uu.ID`, `notnull.TrimmedString` and `money.DecimalAmount`
      differ too; [docs/explanation-null-conventions.md](docs/explanation-null-conventions.md)
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

- [ ] **`types.ValidatErr` users outside this module.** The rename to
      `types.Validatable` breaks `go-function/htmlform` v0.7.0, whose public
      `SetArgValidator` uses `types.ValidatErr`. Either add deprecated
      aliases (`type ValidatErr = Validatable` and the helpers) for one
      release, or release go-function in lockstep. The CHANGELOG `Fixed`
      entries name the new types for bugs that the old-named types had.
- [ ] **Unquoted email display names that mail stacks read as recipients.**
      `email.AddressList`, `NullableAddressList` and `AddressSet` store an
      unquoted name as written, and Python's `EmailMessage` with
      `smtplib.send_message` reads its words as recipients:
      `admin, John <ceo@corp.com>` goes to `admin` and `ceo@corp.com`, and
      `John Doe [External] <ceo@corp.com>` only to `John`. Decided for PR #28
      to keep accepting them as a documented risk. Options: refuse an
      unquoted name with a comma or one of `;:)>[]` in the collections'
      `Validate`, or have `Value` write the normalized, quoted form.
- [ ] **Read US dates with dashes month first.** Dots, dashes and spaces
      are read day first with every language hint, so a US document that
      writes `05-06-2024` can no longer be read as May 6. `ScanStringWithLang`
      reports `monthMustBeFirst` but has no flag for "must be day first", so a
      caller can't detect the order of a whole column. A document-level order
      detection, or a hint that also decides dashes, would cover it.

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
      the [README](README.md). Code new in PR #28 uses `errs.New` and
      `errs.Errorf`, but the older code still calls `errors.New` and
      `fmt.Errorf` in most places. Convert the standard library calls to
      `errs.New` and `errs.Errorf`.
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
- [ ] **Bugs from the PR #28 adversarial review that predate it.**
      - `email.ParseAddressList` doesn't read group syntax, so
        `Team: x@example.com, y@example.com;` gives only `y@example.com`
        and `undisclosed-recipients: a@b.com;` is an error, while
        `net/mail` reads both addresses.
      - `ParseMIMEMessage` turns `To: undisclosed-recipients:;` into an
        empty `To`, so a stored message can't tell it from a message
        without a To header; set the placeholder when a To header was
        present without addresses.
      - `date` language hints are languages only, so `"en-GB"` reads
        ambiguous dates with slashes month first like US English.
      - `types.DeepValidate` calls a `Validate` method that a struct
        promotes from an embedded field twice, for the struct and for the
        field.
      - `types.KeyMutex` gives `NaN` keys no mutual exclusion, because
        `NaN != NaN`, and `Unlock` then panics.
      - `money.Amount.WithinOneCent(1e20, 5e20)` is true, because the
        float64 difference loses the cents beyond 2^53.
      - `vat` accepts `" No.974760673"` but rejects `"No.974760673"`: the
        Norwegian "No." guard depends on the spacing.
- [ ] **Follow-ups from the PR #28 /ship review (2026-10-15).**
      - `email`: the anti-spoofing differential against Python's
        `getaddresses` and its default header parser ran offline and is not
        committed; commit a golden corpus of lists with the recipients each
        parser reads, so CI keeps checking the agreement.
      - `email`: after Python writes a message and reads it back, it drops
        the parentheses of a comment that holds a non-ASCII or dropped rune
        or a nested comment, which glues the comment to the domain or
        exposes an address in it: `<o'brien@müller.de ((ne<U+00A0>sted) x)`
        goes to `o'brien@müller.denested`. Spaces in a quoted local part
        are normalized away (same domain): `" Kate"@evil.com` goes to
        `Kate@evil.com`. A bare word with `;` next to a non-ASCII quoted
        name in another entry exposes that name's quoted address.
      - `email`: `Address.AddressPart()` returns a quoted local part that
        holds `<` without angle brackets, which doesn't parse back.
      - `email`: Python folds a long quoted UTF-8 name that holds an `@`, in
        a list with more entries, into an ASCII-only encoded-word, which a
        second Python write turns into a recipient like `"ä x"@evil.com`.
        The input holds no encoded-word, so the parser's encoded-word rules
        can't see it. Python's SMTPUTF8 policy also writes a decoded
        non-ASCII encoded-word with `, x@evil.com` raw, which then reads as
        another recipient.
      - `email`: an encoded-word whose text is ASCII with an `@` outside
        a comment is refused, also where Python joins it with the word
        before it: `mail.Address.String` writes the name
        `Ivan Petrović [via Mailchimp] (ivan.petrovic@example.org)` with
        the last word `@example.org)`, which Python reads back correctly,
        but `ParseMIMEMessage` leaves such a `From` empty with an error.
        Accepting it needs a model of when Python joins encoded-words.
      - `email`: `ASCII` writes a long name as one encoded-word longer than
        75 characters where a word of at most 75 would hold only ASCII text.
      - `email`: a non-ASCII name holding a literal encoded-word, like
        `Jürgen =?utf-8?q?x?=`, is valid as quoted UTF-8, but its `ASCII`
        form is refused. Python reads both forms as `Jürgenx` after
        writing them, so no ASCII form keeps the name: refuse both or
        accept both.
      - `email`: an encoded-word within a word, like the one of
        `x=?utf-8?q?=3Cevil=40evil.com=3E?= <ceo@corp.com>`, is decoded by
        the parser to the name `x<evil@evil.com>`, while Python shows it as
        written. The recipient is the same for both.
      - `email`: Python's lenient `getaddresses` reads a garbled duplicate of
        the same mailbox, like `<_x@x.de ""<_x@X.DE >`, as an extra entry.
        Same mailbox, so not a spoofing path, but `Validate` could refuse it.
      - `email`: `AddressList.MarshalText`, `MarshalJSON` and those of
        `NullableAddressList` write the stored text untrimmed, so a CR/LF at
        the edges that `Validate` skips reaches the output.
      - `email`: the "unexpected remaining characters" error of
        `ParseAddress` embeds the whole input twice instead of a truncated
        part, so a 1 MB element gives a 2 MB error.
      - `types.Validate`: an embedded interface that holds a nil `*T`, and an
        embedded nil func type like `ValidatorFunc`, still panic. A nil
        embedded pointer or interface is skipped.
      - `date`: `YearMonth.AddMonths` and the `AddYears` methods add the
        whole count in one sum, which wraps the year for counts near
        `math.MaxInt`; the result is invalid and `Value` refuses it.
        `"0001-Q2".AddQuarters(-1)` and `"0001-02".AddMonths(-1)` give the
        zero value, which `Value` writes as NULL instead of an error.
      - `strfmt.Format` panics for a float when `FormatConfig.Float` has a
        zero `DecimalSep`; `float.FormatDef.Format` already reads it as `.`.
      - `strfmt`: a `Precision` of -1 writes a whole KWD 1234 as
        `"KWD 1,234"`, which `money.CurrencyAmount.ScanString` reads as
        1.234. Scanning with the decimal separator of the config instead of
        guessing would remove the ambiguity.
      - `strfmt.EnglishFloatFormat` and `EnglishMoneyFormat` kept their
        names when `NewEnglishFormatConfig` became `NewUSFormatConfig`;
        rename them too if "English" should only mean UK conventions.
      - Tests proposed by the coverage audits: `Date.Scan` of a
        `time.Time` on 0001-01-01 that is not `time.Time{}`,
        `NullableDate.NormalizedOrNull` with a language hint,
        `AddressList.UnmarshalJSON(null)`, `AddressSet.Scan` of a
        placeholder list, and `NullableAddressList("  ").MarshalText()`.
- [ ] **Fast paths for the remaining exact math and parsing.** The
      `CentAmount` rate methods, `float.RoundToDecimals` and `float.Format`
      have float64 fast paths, and `internal.UnmarshalJSONNormalized`
      decodes a plain string without `json.Unmarshal`. Still slower than
      on master, measured in the /ship review of 2026-10-13:
      `Amount.DecimalAmount` 9 ns to 95 ns (a string round trip),
      `CurrencyAmount.ScanString` parses twice,
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
- [ ] **Move the `NullableYearQuarter` tests** from
      `date/yearquarter_test.go` into `date/nullableyearquarter_test.go`.
- [ ] **Multi-word currency names in `money.ParseCurrencyAmount`.** Only a
      single word is passed to `Currency.Normalized`, so
      `"100 Canadian Dollar"` does not parse.
- [ ] **Time zones in the `strfmt` presets.** The US and German presets'
      time layouts end in `MST` without a numeric offset, so a formatted
      time only scans back where its zone abbreviation is known (`UTC`,
      `GMT` or one of `time.Local`); elsewhere `ParseTime` rejects it. The
      US layout is month first and reads British day-first input with a
      known zone as another date. A numeric offset in the layouts would make
      the round trip lossless.
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
- [ ] **Performance of the `email` address checks.** `AddressList.Validate`
      allocates about 35 to 40 times the size of the list, partly for the
      rune offset map, and `AddressListJoinStrings` grows its builder
      without `Grow`. `NullableYearMonth.Compare` and
      `NullableYearQuarter.Compare` normalize on every call, unlike the
      non-nullable types.
- [ ] **Go 1.27 `encoding/json`.** With json/v2 as the default, `float64` and
      `any` map keys marshal, unlike in Go 1.26. `types.CanMarshalJSON`
      follows the documented Go 1.26 rule and reports false for them.

## Completed

- [x] **`date.YearMonth` and `date.YearQuarter` work like `date.Date`.**
      `Normalized` trims the whitespace and upper-cases the `q` that
      `PeriodRange` accepts, `Valid` is lenient like for the other types,
      the zero values are blank, and SQL, JSON, XML and `strfmt` read and
      write them like `Date`, with null as SQL NULL and JSON `null`.
- [x] **Quadratic `email` address parsing.** `ParseAddressList` keeps
      the errors short and stops after 100 of them, so an invalid list of
      10,000 entries takes milliseconds instead of seconds and hundreds of
      megabytes, and a second address is only taken from angle brackets.
      **Completed:** 2026-10-14
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
      **Completed:** 2026-09-01 (no version was tagged then)
- [x] **Require a Go version that accepts the collections proposal's
      self-referential constraints.** Raised the floor to Go 1.26.0.
      **Completed:** 2026-09-01
