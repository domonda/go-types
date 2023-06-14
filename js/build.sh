#!/bin/bash

SCRIPT_DIR=$(cd -P -- $(dirname -- "$0") && pwd -P)
cd $SCRIPT_DIR

go install github.com/gopherjs/gopherjs@v1.18.0-beta3
GOPHERJS_GOROOT="$(go1.18.10 env GOROOT)" gopherjs build vat/vat.go
