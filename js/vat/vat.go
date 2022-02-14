package main

//go:generate gopherjs build $GOFILE

import (
	"github.com/domonda/go-types/vat"
	"github.com/gopherjs/gopherjs/js"
)

func main() {
	js.Global.Set("vat", map[string]interface{}{
		"normalizeId": vat.NormalizeVATID,
	})
}
