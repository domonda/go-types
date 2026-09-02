# TODOS

Open issues and cleanup items before tagging `v1.0.0`. Grouped by what blocks a v1 release versus what's nice to land afterward.

## Blockers for v1.0 (API or semantic)

- [ ] **Tag `v1.0.0`.** No git tags exist yet. Pick semver baseline first.
- [ ] **Remove `uu.IDSet.AsSortedSlice` and `uu.IDSet.AddIDs`,** both
      deprecated in favour of `Sorted` and `InsertAll`/`UnionWith`.
- [ ] **Remove the deprecated set methods.** `Add`, `AddSlice`, `AddSet`,
      `DeleteSlice` and `DeleteSet` on `types.Set`, `uu.IDSet` and
      `strutil.StringSet`, plus `DeleteSlice` and `DeleteSet` on
      `email.AddressSet`, are kept only so consumers keep compiling against the
      new abstract set interface. Migrating `domonda-service` first is 37
      compiler-caught call sites: 32 `Add`→`Insert`, 4 `AddSlice`→`InsertAll`,
      1 `AddSet`→`UnionWith`. `email.AddressSet.Add`, `AddSet`, `AddNormalized`
      and `AddAddressPart` stay — they allocate a nil map and are not deprecated.

## After v1.0

- [ ] **Replace `mapset` with the standard library `container/mapset`** once
      Go 1.28 ships. The package mirrors the proposed API function for function,
      so it should be an import path change. `internal/collections/settest` runs
      the abstract set specification against every set type to catch behaviour
      drift when it is swapped.
- [ ] **Reconsider `types.Set` and the `set` package** at the same time.
      `container/set.Set` covers both, and neither has many callers
      (`types.Set` 3 call sites, `set` 5 across all known consumers).

## Completed

- [x] **No `CHANGELOG.md`.** Add one for the v1 cut.
      **Completed:** 2026-09-01 (unreleased section; no version tagged yet)
- [x] **Require a Go version that accepts the collections proposal's
      self-referential constraints.** Raised the floor to Go 1.26.0.
      **Completed:** 2026-09-01
