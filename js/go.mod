module github.com/domonda/go-types/js

go 1.22

toolchain go1.22.0

// Parent module in same repo
replace github.com/domonda/go-types => ..

require github.com/domonda/go-types v0.0.0-00010101000000-000000000000 // replaced

// External
require github.com/gopherjs/gopherjs v1.18.0-beta3
