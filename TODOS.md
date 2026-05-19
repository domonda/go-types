# TODOS

Open issues and cleanup items before tagging `v1.0.0`. Grouped by what blocks a v1 release versus what's nice to land afterward.

## Blockers for v1.0 (API or semantic)

- [ ] **Tag `v1.0.0`.** No git tags exist yet. Pick semver baseline first.

## European VAT formats — checksum coverage

Audited against the official spec for each member state's VAT identification number. Most checksums now wired up — only a handful remain shape-only because their official algorithms are letter-bearing, sub-format-dependent, or otherwise too ambiguous to ship without an authoritative reference.

### Checksum coverage

- [x] Wired up: `AT`, `BE`, `BG` (9-digit EIK and 10-digit EGN), `CY`, `CZ` (8-digit legal entity only), `DE`, `DK`, `EE`, `EL`, `ES`, `FI`, `FR` (digit-key SIREN form), `GB` (and the shared `XI`), `HR`, `HU`, `IT`, `LT` (9- and 12-digit), `LU`, `LV` (legal entity only), `MT`, `NL`, `NO`, `PL`, `PT`, `RO`, `SE`, `SI`, `SK`.
- [ ] Remaining shape-only: `IE` (old vs new format use different letter-mapping algorithms that need cross-referencing against the Revenue spec before shipping). `CH`, `IS`, `LI`, `SM` have no published public checksum. `FR` letter-key form, `CZ` 9- and 10-digit personal IDs, and `LV` natural-person IDs likewise pass shape-only (sub-format-specific algorithms documented in the wired-up functions' godoc).
- [ ] Real-world validation: the per-country test vectors in `vat/id_test.go` come from public registers (Shell NL, BNP Paribas FR, HEP HR, OTE EL, Nokia FI, Škoda CZ, etc.) but coverage is one valid + one tampered-checksum case each. Consider expanding with a corpus of real anonymized customer VATs before v1 if regressions surface.

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
