module github.com/domonda/go-types/js

go 1.24.0

// Parent module in same repo
replace github.com/domonda/go-types => ..

require github.com/domonda/go-types v0.0.0-00010101000000-000000000000 // replaced

// External
require github.com/gopherjs/gopherjs v1.18.0-beta3

require (
	github.com/bahlo/generic-list-go v0.2.0 // indirect
	github.com/buger/jsonparser v1.1.1 // indirect
	github.com/domonda/go-errs v0.0.0-20251216095446-6397ee531051 // indirect
	github.com/domonda/go-pretty v0.0.0-20251216083144-fe0d655ca8d7 // indirect
	github.com/invopop/jsonschema v0.13.0 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/mailru/easyjson v0.9.1 // indirect
	github.com/rogpeppe/go-internal v1.14.1 // indirect
	github.com/wk8/go-ordered-map/v2 v2.1.8 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
