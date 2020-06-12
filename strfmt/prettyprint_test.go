package strfmt

import (
	"testing"

	"github.com/domonda/go-types/uu"
)

func TestPrettySprint(t *testing.T) {
	type Parent struct {
		Map map[int]string
	}
	type Struct struct {
		Parent
		Int int
		Str string
		Sub struct {
			Map map[string]struct{}
		}
	}

	tests := []struct {
		name  string
		value interface{}
		want  string
	}{
		{name: "nil", value: nil, want: `nil`},
		{name: "nilPtr", value: (*int)(nil), want: `nil`},
		{name: "empty string", value: "", want: `""`},
		{name: "empty bytes string", value: []byte{}, want: `""`},
		{name: "bytes string", value: []byte("Hello World"), want: `"Hello World"`},
		{name: "int", value: 666, want: `666`},
		{name: "struct no sub-init", value: Struct{Int: -1, Str: "xxx"}, want: `Struct{Parent{Map:nil},Int:-1,Str:"xxx",Sub:{Map:nil}}`},
		{name: "struct sub-init", value: Struct{Sub: struct{ Map map[string]struct{} }{Map: map[string]struct{}{"key": {}}}}, want: `Struct{Parent{Map:nil},Int:0,Str:"",Sub:{Map:{"key":{}}}}`},
		{name: "string slice", value: []string{"", `"quoted"`, "hello\nworld"}, want: `["","\"quoted\"","hello\nworld"]`},
		{name: "Nil UUID", value: uu.IDNil, want: `[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0]`},

		// TODO
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PrettySprint(tt.value); got != tt.want {
				t.Errorf("PrettySprint() = %v, want %v", got, tt.want)
			}
		})
	}
}
