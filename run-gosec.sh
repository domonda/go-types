#!/bin/bash
SCRIPT_DIR=$(cd -P -- "$(dirname -- "$0")" && pwd -P)
cd $SCRIPT_DIR

go tool gosec ./...