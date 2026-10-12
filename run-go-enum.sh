#!/bin/bash
SCRIPT_DIR=$(cd -P -- "$(dirname -- "$0")" && pwd -P)
cd "$SCRIPT_DIR" || exit 1

# go-enum is pinned as a tool dependency in tools/go.mod so that this module
# does not depend on it. -modfile resolves it from there, while go-enum runs
# from the repo root so it finds the //#enum types of this module.
#
# Pass -validate to only check that the generated methods are up to date
# without modifying any file (exits non-zero if they are not).
go tool -modfile=tools/go.mod go-enum "$@" ./...
