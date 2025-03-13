module github.com/domonda/go-types/js

go 1.24

// Parent module in same repo
replace github.com/domonda/go-types => ..

require github.com/domonda/go-types v0.0.0-00010101000000-000000000000 // replaced

// External
require github.com/gopherjs/gopherjs v1.18.0-beta3

require (
	github.com/domonda/go-errs v0.0.0-20240702051036-0e696c849b5f // indirect
	github.com/domonda/go-pretty v0.0.0-20240110134850-17385799142f // indirect
)
