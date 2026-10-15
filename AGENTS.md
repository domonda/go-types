# AGENTS.md

Instructions for coding agents working in this repository.

## Skill routing

When the user's request matches an available skill, invoke it via the Skill tool. When in doubt, invoke the skill.

Key routing rules:

- Product ideas/brainstorming → invoke /office-hours
- Strategy/scope → invoke /plan-ceo-review
- Architecture → invoke /plan-eng-review
- Design system/plan review → invoke /design-consultation or /plan-design-review
- Full review pipeline → invoke /autoplan
- Bugs/errors → invoke /investigate
- QA/testing site behavior → invoke /qa or /qa-only
- Code review/diff check → invoke /review
- Visual polish → invoke /design-review
- Ship/deploy/PR → invoke /ship or /land-and-deploy
- Save progress → invoke /context-save
- Resume context → invoke /context-restore
- Author a backlog-ready spec/issue → invoke /spec

This repository is a Go library with no UI and no running service, so the
browser-driven entries (/qa, /qa-only, /design-review, /design-consultation,
/plan-design-review) rarely apply here.

## gstack /ship overrides

These rules override conflicting /ship steps:

- **Rename first**: before Step 1, rename the branch to its final name. gstack keys review records by branch name; a later rename orphans them and the next /ship reviews everything again.
- **Committed start**: commit pending work on the feature branch before Step 1.
- **Version** (Step 12): `gstack-version-bump` rejects Go versions, so write `VERSION` directly as `vX.Y.Z`: keep a version already bumped on the branch, else take the version or level from the /ship arguments, else the next patch. It must be above the latest `git tag`. CHANGELOG headings use `[X.Y.Z]`.
- **Branch scope** (Steps 9 and 11): fix findings in code the branch changed; collect problems that predate the branch into one opt-in batch. After a fix round, re-verify the fixed findings with targeted tests rather than re-running every specialist.
- **Read-only subagents**: every subagent prompt states the worktree is read-only: no edits, no `git checkout`/`restore`/`stash`, no `go.mod`/`go.sum` changes. Mutation tests run in a temp copy (`rsync -a` or `go test -overlay`).
- **Codex** adversarial pass: only on diffs under ~750 changed lines; larger ones time out at 9 minutes.
- **Greptile** is not used in any of my repos: skip Step 6.5 (early Greptile PR) and Step 10 (Greptile triage) entirely.
- **Pregate**: "stray" and "no test lane" findings on `*_test.go`, and PII hits in test fixtures, are Go false positives: note them in the PR body and continue.
- **Generated tests** carry the comments a maintainer would write; gstack `Value:` cards go in the PR body.
- **TODOS.md**: skip creating one.
- **PR body**: open with a summary of at most five sentences; omit sections that don't apply to a Go library (Design Review, Eval Results, Greptile).
- **Release**: when the /ship arguments ask to merge, tag or release, continue after the PR: squash-merge only on green CI (red: stop and ask), tag every released module in lockstep on the base branch (with the repo's tag script if it has one), then create the GitHub release from the CHANGELOG entry.

## Test Coverage

Minimum: 60%
Target: 75%
Base control: off
Star rating: off
