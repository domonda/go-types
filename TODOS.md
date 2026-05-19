# TODOS

Open issues and cleanup items before tagging `v1.0.0`. Grouped by what blocks a v1 release versus what's nice to land afterward.

## Blockers for v1.0 (API or semantic)

- [ ] **Tag `v1.0.0`.** No git tags exist yet. Pick semver baseline first.

## Functional gaps with explicit TODOs

- [ ] **`language.Code.Normalized()` doesn't handle ISO 639-2/3 or BCP-47** (`en-US`, `sr-Latn`). Either implement, or scope the limit in godoc.
  - `language/code.go:43-44`
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

- [ ] Audit thin coverage. 52 test files for 192 source files. Packages without tests at all worth a smoke pass: `deref`, `set` (root-level), `queue`, `notnull`, `charset` (partial), `language` (partial).

## Docs & release hygiene

- [ ] **Top-level `README.md`** still narrates the package list; add a TOC linking to the new per-package READMEs.
- [ ] **Module-level godoc** on the root `types` package is minimal; consider expanding for `pkg.go.dev` landing.
- [ ] **No `CHANGELOG.md`.** Add one for the v1 cut.
