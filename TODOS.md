# TODOS

Open issues and cleanup items before tagging `v1.0.0`. Grouped by what blocks a v1 release versus what's nice to land afterward.

## Blockers for v1.0 (API or semantic)

- [ ] **Tag `v1.0.0`.** No git tags exist yet. Pick semver baseline first.

## Functional gaps with explicit TODOs

- [x] **`language.Code.Normalized()` now handles ISO 639-2/B+T, ISO 639-3 (where a 639-1 mapping exists), and BCP-47 tags** by extracting the primary language subtag. Added `language/iso6393.go` with the mapping and table-driven tests in `language/code_test.go`. Languages without a 639-1 assignment still error.
- [ ] **VAT checksum implementation is incomplete for at least one country.** Finish or remove.
  - `vat/formats.go:133`
- [ ] **Spanish VAT validation needs improvement.** See pointer in code.
  - `vat/formats.go:34`
- [ ] **`date` parser doesn't accept many common English date formats.** ~60 commented-out test cases in `date/date_test.go` document the gap (month names, abbreviated months, ordinal suffixes, year-only, CJK, US short forms). Decide per-format: in scope for v1, or explicit non-goal documented in godoc.
  - `date/date_test.go:50-168, 451`

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
