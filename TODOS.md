# TODOS

Open issues and cleanup items before tagging `v1.0.0`. Grouped by what blocks a v1 release versus what's nice to land afterward.

## Blockers for v1.0 (API or semantic)

- [ ] **Tag `v1.0.0`.** No git tags exist yet. Pick semver baseline first.

## Functional gaps with explicit TODOs

- [x] **`language.Code.Normalized()` now handles ISO 639-2/B+T, ISO 639-3 (where a 639-1 mapping exists), and BCP-47 tags** by extracting the primary language subtag. Added `language/iso6393.go` with the mapping and table-driven tests in `language/code_test.go`. Languages without a 639-1 assignment still error.
- [x] **Norwegian VAT checksum (`vat/checkSumNO`)** now implements the official MOD-11 algorithm (weights 3,2,7,6,5,4,3,2 on the first 8 digits of the 9-digit organisation number) instead of the previous `// TODO full check-sum implementation` stub that returned `true` for any 11-char input. The "No."/"no." prefix guard against confusion with the English "number" abbreviation is preserved.
- [x] **Spanish VAT validation** rewritten in `vat/checkSumES`. The single permissive regex `^ES[0-9A-Z]\d{7}[0-9A-Z]$` was replaced with six precise sub-patterns covering CIF (numeric-only, letter-only, and either-form variants), DNI/NIF, NIE (Y/Z form), and historical NIE (X) plus K/L/M. Each sub-format now runs its specific control-character algorithm (mod-23 table for DNI/NIE, Luhn-style mod-10 for CIF). `vat/id_test.go` covers all sub-formats with positive and negative cases.
- [ ] **`date` parser doesn't accept many common English date formats.** ~60 commented-out test cases in `date/date_test.go` document the gap (month names, abbreviated months, ordinal suffixes, year-only, CJK, US short forms). Decide per-format: in scope for v1, or explicit non-goal documented in godoc.
  - `date/date_test.go:50-168, 451`

## European VAT formats with loose, broken, or missing regex coverage

Audited against the official spec for each member state's VAT identification number. Only `AT`, `DE`, `ES`, and `NO` currently have checksum validation; the rest rely solely on the regex below for shape checks.

### Regex precedence bugs (`^` and `$` only bind to the first/last alternative)

- [ ] **GB** — `^GB(?:\d{9})|(?:\d{12})|(?:GD\d{3})|(?:HA\d{3})$` accepts strings like `XGD123` or `HA999Y` because only the first alternative is anchored to `^GB` and only the last is anchored to `$`. Wrap the alternation: `^GB(?:\d{9}|\d{12}|GD\d{3}|HA\d{3})$`.
- [ ] **IE** — `^IE(?:\d[0-9A-Z]\d{5}[A-Z])|(?:\d{7}[A-W][A-I])$` has the same problem.

### Over-permissive (matches strings the official spec rejects)

- [ ] **BE** — `^BE\d{10}$`. The 10-digit company number must start with `0` or `1`. Tighten to `^BE[01]\d{9}$`.
- [ ] **HU** — `^HU\d{8,9}$`. The Hungarian community VAT number is exactly 8 digits; 9 is not valid. Tighten to `^HU\d{8}$`.
- [ ] **SE** — `^SE\d{12}$`. The last two digits must be `01` (the rest is the 10-digit organisation number). Tighten to `^SE\d{10}01$`.

### Dead regex branches (normalization strips them before matching)

- [ ] **CH** — `^CHE-?(?:\d{9}|(?:\d{3}\.\d{3}\.\d{3}))$`. `ID.Normalized` uppercases and strips spaces + punctuation, so `-?` and the dotted alternative can never match. Simplify to `^CHE\d{9}$`.

### Missing European VAT regimes

- [ ] **XI** — Northern Ireland's post-Brexit VAT prefix. Same shape as GB; share the regex.
- [ ] **IS** — Iceland (EFTA). 5- or 6-digit VSK number.
- [ ] **LI** — Liechtenstein (EFTA). 5 digits.
- [ ] **SM** — San Marino. 5-digit VAT (used for EU-facing invoices via Italian intermediary).

### Checksum coverage

Only `AT`, `DE`, `ES`, `NO` have a check-digit algorithm wired up. Documented algorithms exist for at least: `BE`, `BG`, `CY`, `CZ`, `DK`, `EE`, `EL`, `FI`, `FR`, `GB`, `HR`, `HU`, `IE`, `IT`, `LT`, `LU`, `LV`, `MT`, `NL`, `PL`, `PT`, `RO`, `SE`, `SI`, `SK`. The BMF PDF linked in `vat/formats.go` is the authoritative reference. Decide which countries warrant a real checksum for v1 (Domonda's primary markets first), and which can stay shape-only.

## Defensive panics worth reviewing

- [ ] **`validator.go:234, 246`** — `MaxValue` / `MinValue` panic on type mismatch. More idiomatic to return an error in v1.
- [ ] **`bank/bicfinder.go:36`** — defensive panic on regex submatch count. The regex is fixed; the assertion can be removed.

## Performance TODOs (defer)

- [ ] `strutil/strutil.go:595,622` — "TODO optimized version" on two helpers.
- [ ] `float/format.go:90` — fast path for non-default decimal separator.
- [ ] `uu/nullableid.go:344` — JSON unmarshal optimization.

## Test coverage

- [x] Smoke pass for previously untested packages: added `deref/deref_test.go`, `set/set_test.go`, `queue/queue_test.go`, `notnull/stringarray_test.go`, `notnull/trimmedstring_test.go`, `charset/bom_test.go`, `charset/encoding_test.go`. `language` was extended with the BCP-47 / 639-3 work above and its test file grown to cover the new paths.
- [ ] Still thin or absent: deeper coverage for `notnull/intarray`, `notnull/nullboolarray`, `notnull/arrays` helpers, and `notnull/json` helpers (only `stringarray`, `floatarray`, `trimmedstring` are smoke-covered today).
- [ ] `language/iso6393macro.go` (macrolanguage → macrolanguage map) and `language/iso6393names.go` (English names of every 639-3 code) together contain ~8500 lines of commented-out data with no live symbols. Independent of the 639-1 mapping that now lives in `language/iso6393.go`. Decide whether to revive (different use cases — dialect rollup and language-name display) or delete.

## Docs & release hygiene

- [ ] **No `CHANGELOG.md`.** Add one for the v1 cut.
