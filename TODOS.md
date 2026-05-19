# TODOS

Open issues and cleanup items before tagging `v1.0.0`. Grouped by what blocks a v1 release versus what's nice to land afterward.

## Blockers for v1.0 (API or semantic)

- [ ] **Tag `v1.0.0`.** No git tags exist yet. Pick semver baseline first.

## European VAT formats with loose, broken, or missing regex coverage

Audited against the official spec for each member state's VAT identification number. Only `AT`, `DE`, `ES`, and `NO` currently have checksum validation; the rest rely solely on the regex below for shape checks.


### Checksum coverage

Only `AT`, `DE`, `ES`, `GB` (and the shared `XI`), `NO` have a check-digit algorithm wired up. Documented algorithms exist for at least: `BE`, `BG`, `CY`, `CZ`, `DK`, `EE`, `EL`, `FI`, `FR`, `HR`, `HU`, `IE`, `IT`, `LT`, `LU`, `LV`, `MT`, `NL`, `PL`, `PT`, `RO`, `SE`, `SI`, `SK`. The BMF PDF linked in `vat/formats.go` is the authoritative reference. Decide which countries warrant a real checksum for v1 (Domonda's primary markets first), and which can stay shape-only.

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
