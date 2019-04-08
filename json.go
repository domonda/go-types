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

// CanMarshalJSON returns if a type can be marshalled as JSON
func CanMarshalJSON(t reflect.Type) bool {
	if t == emptyInterfaceTye {
		return true
	}

	if t.Implements(jsonMarshalerType) {
		return true
	}
	kind := t.Kind()
	if kind != reflect.Ptr && reflect.PtrTo(t).Implements(jsonMarshalerType) {
		return true
	}

	if t.Implements(textMarshalerType) {
		return true
	}
	if kind != reflect.Ptr && reflect.PtrTo(t).Implements(textMarshalerType) {
		return true
	}

	if kind == reflect.Ptr {
		t = t.Elem()
		kind = t.Kind()
	}
	return kind == reflect.Struct || kind == reflect.Map || kind == reflect.Slice
}
