package types

import (
	"encoding"
	"encoding/json"
	"reflect"
)

var (
	jsonMarshalerType = reflect.TypeFor[json.Marshaler]()
	textMarshalerType = reflect.TypeFor[encoding.TextMarshaler]()
	emptyInterfaceTye = reflect.TypeFor[any]()
)

// CanMarshalJSON reports whether [encoding/json.Marshal] can encode a value
// of type t purely based on t's static structure. It returns true when:
//   - t is the empty interface [any] or any other interface type,
//   - t implements [json.Marshaler] or [encoding.TextMarshaler] (directly or
//     via its pointer receiver),
//   - t is a primitive kind that encoding/json accepts: bool, any integer or
//     unsigned integer kind, float32/float64, or string,
//   - t is an array, slice, or struct (field types are not inspected),
//   - t is a map whose key kind is supported by encoding/json — string,
//     any integer kind, or a key type implementing [encoding.TextMarshaler],
//   - t is a pointer or interface to a marshalable element type.
//
// Kinds rejected by encoding/json (complex, chan, func, unsafe.Pointer) return
// false. Because field/element types are not inspected recursively, a struct
// or array containing an unsupported field type may pass this check and still
// fail at runtime. Value-dependent failures such as float NaN/Inf are likewise
// invisible here.
func CanMarshalJSON(t reflect.Type) bool {
	if t == nil {
		return false
	}
	if t == emptyInterfaceTye {
		return true
	}

	if t.Implements(jsonMarshalerType) {
		return true
	}
	kind := t.Kind()
	if kind != reflect.Pointer && reflect.PointerTo(t).Implements(jsonMarshalerType) {
		return true
	}

	if t.Implements(textMarshalerType) {
		return true
	}
	if kind != reflect.Pointer && reflect.PointerTo(t).Implements(textMarshalerType) {
		return true
	}

	switch kind {
	case reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64,
		reflect.String,
		reflect.Array, reflect.Slice, reflect.Struct:
		return true
	case reflect.Map:
		return canMarshalJSONMapKey(t.Key())
	case reflect.Pointer:
		return CanMarshalJSON(t.Elem())
	case reflect.Interface:
		return true
	default:
		return false
	}
}

// canMarshalJSONMapKey reports whether encoding/json accepts k as a map key
// type. The rules mirror encoding/json: string keys, integer keys, or keys
// implementing [encoding.TextMarshaler].
func canMarshalJSONMapKey(k reflect.Type) bool {
	switch k.Kind() {
	case reflect.String,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return true
	}
	if k.Implements(textMarshalerType) {
		return true
	}
	if k.Kind() != reflect.Pointer && reflect.PointerTo(k).Implements(textMarshalerType) {
		return true
	}
	return false
}
