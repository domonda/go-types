# TODOS

Open issues and cleanup items before tagging `v1.0.0`. Grouped by what blocks a v1 release versus what's nice to land afterward.

## Blockers for v1.0 (API or semantic)

- [ ] **Tag `v1.0.0`.** No git tags exist yet. Pick semver baseline first.

## Defensive panics worth reviewing

- [ ] **`bank/bicfinder.go:36`** — defensive panic on regex submatch count. The regex is fixed; the assertion can be removed.

## Test coverage

- [ ] Still thin or absent: deeper coverage for `notnull/intarray`, `notnull/nullboolarray`, `notnull/arrays` helpers, and `notnull/json` helpers (only `stringarray`, `floatarray`, `trimmedstring` are smoke-covered today).
- [ ] `language/iso6393macro.go` (macrolanguage → macrolanguage map) and `language/iso6393names.go` (English names of every 639-3 code) together contain ~8500 lines of commented-out data with no live symbols. Independent of the 639-1 mapping that now lives in `language/iso6393.go`. Decide whether to revive (different use cases — dialect rollup and language-name display) or delete.

## Docs & release hygiene

- [ ] **No `CHANGELOG.md`.** Add one for the v1 cut.
