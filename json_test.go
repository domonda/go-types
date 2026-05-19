package types

import (
	"encoding"
	"encoding/json"
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
