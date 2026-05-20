#!/bin/bash
SCRIPT_DIR=$(cd -P -- "$(dirname -- "$0")" && pwd -P)
cd $SCRIPT_DIR

# gosec is pinned as a tool dependency in tools/go.mod. -modfile resolves
# it from there, while gosec runs from the repo root so it analyzes this
# module (running it inside tools/ breaks cross-package type checking).
go tool -modfile=tools/go.mod gosec ./...
