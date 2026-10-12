package types

import (
	"encoding"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
)

type jsonMarshalerStub struct{ s string }

func (j jsonMarshalerStub) MarshalJSON() ([]byte, error) { return []byte(`"` + j.s + `"`), nil }

type jsonMarshalerPtrStub struct{ s string }

func (j *jsonMarshalerPtrStub) MarshalJSON() ([]byte, error) { return []byte(`"` + j.s + `"`), nil }

type textMarshalerStub struct{ s string }

func (t textMarshalerStub) MarshalText() ([]byte, error) { return []byte(t.s), nil }

var (
	_ json.Marshaler         = jsonMarshalerStub{}
	_ json.Marshaler         = (*jsonMarshalerPtrStub)(nil)
	_ encoding.TextMarshaler = textMarshalerStub{}
)

func TestCanMarshalJSON(t *testing.T) {
	tests := []struct {
		name string
		t    reflect.Type
		want bool
	}{
		// Primitives all marshal fine.
		{"bool", reflect.TypeFor[bool](), true},
		{"int", reflect.TypeFor[int](), true},
		{"int8", reflect.TypeFor[int8](), true},
		{"int64", reflect.TypeFor[int64](), true},
		{"uint", reflect.TypeFor[uint](), true},
		{"uintptr", reflect.TypeFor[uintptr](), true},
		{"float32", reflect.TypeFor[float32](), true},
		{"float64", reflect.TypeFor[float64](), true},
		{"string", reflect.TypeFor[string](), true},

		// Sequences.
		{"[3]int", reflect.TypeFor[[3]int](), true},
		{"[]int", reflect.TypeFor[[]int](), true},
		{"[]byte", reflect.TypeFor[[]byte](), true},

		// Maps — keys must be string, integer, or TextMarshaler.
		{"map[string]int", reflect.TypeFor[map[string]int](), true},
		{"map[int]string", reflect.TypeFor[map[int]string](), true},
		{"map[uint64]string", reflect.TypeFor[map[uint64]string](), true},
		{"map[textMarshalerStub]int", reflect.TypeFor[map[textMarshalerStub]int](), true},
		{"map[float64]string", reflect.TypeFor[map[float64]string](), false},
		{"map[any]int", reflect.TypeFor[map[any]int](), false},
		{"map[bool]int", reflect.TypeFor[map[bool]int](), false},
		{"map[complex128]int", reflect.TypeFor[map[complex128]int](), false},

		// Structs (field types are not inspected).
		{"struct{X int}", reflect.TypeFor[struct{ X int }](), true},
		{"struct{}", reflect.TypeFor[struct{}](), true},

		// Pointers.
		{"*int", reflect.TypeFor[*int](), true},
		{"*string", reflect.TypeFor[*string](), true},
		{"**int", reflect.TypeFor[**int](), true},
		{"*chan int", reflect.TypeFor[*chan int](), false},
		{"*func()", reflect.TypeFor[*func()](), false},

		// Interfaces.
		{"any", reflect.TypeFor[any](), true},
		{"json.Marshaler", reflect.TypeFor[json.Marshaler](), true},
		// encoding/json marshals the dynamic value of any interface.
		{"fmt.Stringer", reflect.TypeFor[fmt.Stringer](), true},

		// Marshalers.
		{"jsonMarshalerStub", reflect.TypeFor[jsonMarshalerStub](), true},
		{"jsonMarshalerPtrStub", reflect.TypeFor[jsonMarshalerPtrStub](), true},
		{"*jsonMarshalerPtrStub", reflect.TypeFor[*jsonMarshalerPtrStub](), true},
		{"textMarshalerStub", reflect.TypeFor[textMarshalerStub](), true},

		// Unsupported kinds.
		{"chan int", reflect.TypeFor[chan int](), false},
		{"func()", reflect.TypeFor[func()](), false},
		{"complex64", reflect.TypeFor[complex64](), false},
		{"complex128", reflect.TypeFor[complex128](), false},

		// Nil reflect.Type.
		{"nil", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CanMarshalJSON(tt.t); got != tt.want {
				t.Errorf("CanMarshalJSON(%v) = %v, want %v", tt.t, got, tt.want)
			}
		})
	}
}

// TestCanMarshalJSON_AgreesWithEncodingJSON verifies that for every type
// CanMarshalJSON returns true, encoding/json successfully marshals a zero
// value of that type. False negatives are surfaced as test failures; false
// positives caused by uninspected struct/array fields are documented in
// CanMarshalJSON's godoc and excluded from this check.
func TestCanMarshalJSON_AgreesWithEncodingJSON(t *testing.T) {
	cases := []any{
		int(0), int64(0), uint(0), float64(0), "", false,
		[3]int{}, []int(nil), []byte(nil),
		map[string]int(nil), map[int]string(nil), map[uint64]string(nil),
		struct{ X int }{},
		(*int)(nil), (**int)(nil),
		jsonMarshalerStub{}, textMarshalerStub{},
	}
	for _, c := range cases {
		ty := reflect.TypeOf(c)
		t.Run(ty.String(), func(t *testing.T) {
			if !CanMarshalJSON(ty) {
				t.Fatalf("CanMarshalJSON returned false for %s", ty)
			}
			if _, err := json.Marshal(c); err != nil {
				t.Errorf("CanMarshalJSON said yes but json.Marshal failed: %v", err)
			}
		})
	}
}

// textMarshalerPtrFunc is a func type, which encoding/json can't marshal,
// that becomes marshalable through its pointer receiver MarshalText.
type textMarshalerPtrFunc func()

func (*textMarshalerPtrFunc) MarshalText() ([]byte, error) { return []byte("text"), nil }

// TestCanMarshalJSON_PointerReceiverTextMarshaler checks that a type is
// reported as marshalable when only its pointer implements
// encoding.TextMarshaler, because encoding/json uses the pointer method
// for addressable values like the fields of a struct passed by pointer.
func TestCanMarshalJSON_PointerReceiverTextMarshaler(t *testing.T) {
	if !CanMarshalJSON(reflect.TypeFor[textMarshalerPtrFunc]()) {
		t.Fatal("CanMarshalJSON(textMarshalerPtrFunc) = false, want true")
	}
	j, err := json.Marshal(&struct{ F textMarshalerPtrFunc }{F: func() {}})
	if err != nil {
		t.Fatalf("json.Marshal error: %v", err)
	}
	if want := `{"F":"text"}`; string(j) != want {
		t.Errorf("json.Marshal = %s, want %s", j, want)
	}
}

// textMarshalerPtrStub implements encoding.TextMarshaler
// only with a pointer receiver.
type textMarshalerPtrStub struct{ s string }

func (t *textMarshalerPtrStub) MarshalText() ([]byte, error) { return []byte(t.s), nil }

// textMarshalerPtrString is of string kind and implements
// encoding.TextMarshaler only with a pointer receiver.
type textMarshalerPtrString string

func (t *textMarshalerPtrString) MarshalText() ([]byte, error) { return []byte(*t), nil }

// TestCanMarshalJSON_MapKeys checks that CanMarshalJSON accepts exactly the
// map key types that encoding/json accepts, because callers use it to decide
// up front whether json.Marshal can encode a value. encoding/json accepts a
// key of string or integer kind, or a key type that implements
// encoding.TextMarshaler itself. It never calls a pointer receiver
// MarshalText for a key, because map keys are not addressable, so a key type
// whose pointer type alone implements the interface fails, while the pointer
// type works as key.
//
// Float and interface keys are not compared with json.Marshal here: the
// encoding/json of Go 1.27, which is built on encoding/json/v2, also encodes
// them, unlike its documented rule and Go 1.26. TestCanMarshalJSON pins the
// documented rule for them.
func TestCanMarshalJSON_MapKeys(t *testing.T) {
	tests := []struct {
		name string
		m    any // map with one entry
		want bool
	}{
		{name: "string", m: map[string]int{"a": 1}, want: true},
		{name: "int8", m: map[int8]int{1: 1}, want: true},
		{name: "uintptr", m: map[uintptr]int{1: 1}, want: true},
		// The string kind is used, the pointer method is not needed
		{name: "string kind with pointer receiver MarshalText", m: map[textMarshalerPtrString]int{"a": 1}, want: true},
		{name: "value receiver MarshalText", m: map[textMarshalerStub]int{{s: "a"}: 1}, want: true},
		{name: "pointer receiver MarshalText", m: map[textMarshalerPtrStub]int{{s: "a"}: 1}, want: false},
		{name: "pointer to pointer receiver MarshalText", m: map[*textMarshalerPtrStub]int{{s: "a"}: 1}, want: true},
		{name: "encoding.TextMarshaler interface", m: map[encoding.TextMarshaler]int{textMarshalerStub{s: "a"}: 1}, want: true},
		{name: "bool", m: map[bool]int{true: 1}, want: false},
		{name: "array", m: map[[2]int]int{{1, 2}: 1}, want: false},
		{name: "struct", m: map[struct{ A int }]int{{A: 1}: 1}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := json.Marshal(tt.m)
			if jsonOK := err == nil; jsonOK != tt.want {
				t.Fatalf("test case is wrong: json.Marshal(%T) error = %v, want success %t", tt.m, err, tt.want)
			}
			if got := CanMarshalJSON(reflect.TypeOf(tt.m)); got != tt.want {
				t.Errorf("CanMarshalJSON(%T) = %t, but json.Marshal error = %v", tt.m, got, err)
			}
		})
	}
}
