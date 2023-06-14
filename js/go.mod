module github.com/domonda/go-types/js

go 1.18

// Parent module in same repo
replace github.com/domonda/go-types => ../

require github.com/domonda/go-types v0.0.0-20230607122810-d591578dc741

// External
require github.com/gopherjs/gopherjs v1.18.0-beta3

// Indirect
require (
	github.com/domonda/go-pretty v0.0.0-20220317123925-dd9e6bef129a // indirect
	golang.org/x/exp v0.0.0-20230522175609-2e198f4a06a1 // indirect
)
