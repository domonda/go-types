package types

import (
	"encoding"
	"encoding/json"
	"reflect"
)

var (
	jsonMarshalerType = reflect.TypeOf((*json.Marshaler)(nil)).Elem()
	textMarshalerType = reflect.TypeOf((*encoding.TextMarshaler)(nil)).Elem()
	emptyInterfaceTye = reflect.TypeOf((*interface{})(nil)).Elem()
)

// CanMarshalJSON returns true if a type can be marshalled as JSON.
// It checks if the type implements json.Marshaler or encoding.TextMarshaler,
// or if it's a struct, map, or slice that can be marshalled by the standard
// JSON package.
func CanMarshalJSON(t reflect.Type) bool {
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

	if kind == reflect.Pointer {
		t = t.Elem()
		kind = t.Kind()
	}
	return kind == reflect.Struct || kind == reflect.Map || kind == reflect.Slice
}
