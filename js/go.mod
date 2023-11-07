module github.com/domonda/go-types/js

go 1.21

toolchain go1.21.0

// Parent module in same repo
replace github.com/domonda/go-types => ../

require github.com/domonda/go-types v0.0.0-20231024131150-e5e3ba4f448e

// External
require github.com/gopherjs/gopherjs v1.18.0-beta3

// Indirect
require github.com/domonda/go-pretty v0.0.0-20230810130018-8920f571470a // indirect
