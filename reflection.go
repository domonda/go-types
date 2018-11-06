package types

import (
	"reflect"
)

// DerefType dereferences a pointer type to its base type
func DerefType(t reflect.Type) reflect.Type {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t
}

// DerefValue dereferences a pointer type to its base type
func DerefValue(v reflect.Value) reflect.Value {
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	return v
}

// FlatStructFieldCount returns the number of flattened struct fields,
// meaning that the fields of anonoymous embedded fields are flattened
// to the top level of the struct.
func FlatStructFieldCount(t reflect.Type) int {
	t = DerefType(t)
	count := 0
	numField := t.NumField()
	for i := 0; i < numField; i++ {
		f := t.Field(i)
		if f.Anonymous {
			count += FlatStructFieldCount(f.Type)
		} else {
			count++
		}
	}
	return count
}

// FlatStructFieldNames returns the names of flattened struct fields,
// meaning that the fields of anonoymous embedded fields are flattened
// to the top level of the struct.
func FlatStructFieldNames(t reflect.Type) (names []string) {
	t = DerefType(t)
	numField := t.NumField()
	names = make([]string, 0, numField)
	for i := 0; i < numField; i++ {
		f := t.Field(i)
		if f.Anonymous {
			names = append(names, FlatStructFieldNames(f.Type)...)
		} else {
			names = append(names, f.Name)
		}
	}
	return names
}

// FlatStructFieldValues returns the values of flattened struct fields,
// meaning that the fields of anonoymous embedded fields are flattened
// to the top level of the struct.
func FlatStructFieldValues(v reflect.Value) (values []reflect.Value) {
	v = DerefValue(v)
	t := v.Type()
	numField := t.NumField()
	values = make([]reflect.Value, 0, numField)
	for i := 0; i < numField; i++ {
		ft := t.Field(i)
		fv := v.Field(i)
		if ft.Anonymous {
			values = append(values, FlatStructFieldValues(fv)...)
		} else {
			values = append(values, fv)
		}
	}
	return values
}
