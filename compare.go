package types

import (
	"bytes"
	"cmp"
	"reflect"
)

// CompareReflectValue compares two reflect.Values.
// Returns 0 if the values are not of the same type or not of an orderable kind,
// so the function is safe to use as the less function for [slices.SortFunc]
// on slices of mixed or non-orderable [reflect.Value]s (the order of such
// elements is left unchanged).
//
// Orderable kinds are:
//   - all integer, unsigned integer, and float kinds (compared by value)
//   - strings (compared lexicographically)
//   - bool (false < true)
//   - complex (compared by real part, then imaginary part)
//   - pointer, channel, and unsafe pointer (compared by address)
//   - array and slice (compared lexicographically by element, shorter first on tie;
//     []byte uses [bytes.Compare])
//   - struct (compared lexicographically by field)
//   - interface (compared by dynamic type name, then by underlying value)
//
// This is used for sorting map keys in DeepValidate.
func CompareReflectValue(a, b reflect.Value) int {
	if a.Type() != b.Type() {
		return 0
	}

	switch a.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return cmp.Compare(a.Int(), b.Int())

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return cmp.Compare(a.Uint(), b.Uint())

	case reflect.Float32, reflect.Float64:
		return cmp.Compare(a.Float(), b.Float())

	case reflect.String:
		return cmp.Compare(a.String(), b.String())

	case reflect.Bool:
		ab, bb := a.Bool(), b.Bool()
		switch {
		case ab == bb:
			return 0
		case !ab:
			return -1
		default:
			return 1
		}

	case reflect.Complex64, reflect.Complex128:
		ac, bc := a.Complex(), b.Complex()
		if c := cmp.Compare(real(ac), real(bc)); c != 0 {
			return c
		}
		return cmp.Compare(imag(ac), imag(bc))

	case reflect.Pointer, reflect.UnsafePointer:
		return cmp.Compare(a.Pointer(), b.Pointer())

	case reflect.Array:
		for i := range a.Len() {
			if c := CompareReflectValue(a.Index(i), b.Index(i)); c != 0 {
				return c
			}
		}
		return 0

	case reflect.Slice:
		if a.Type().Elem().Kind() == reflect.Uint8 {
			return bytes.Compare(a.Bytes(), b.Bytes())
		}
		n := min(a.Len(), b.Len())
		for i := range n {
			if c := CompareReflectValue(a.Index(i), b.Index(i)); c != 0 {
				return c
			}
		}
		return cmp.Compare(a.Len(), b.Len())

	case reflect.Struct:
		for i := range a.NumField() {
			if c := CompareReflectValue(a.Field(i), b.Field(i)); c != 0 {
				return c
			}
		}
		return 0

	case reflect.Interface:
		aNil, bNil := a.IsNil(), b.IsNil()
		if aNil || bNil {
			switch {
			case aNil && bNil:
				return 0
			case aNil:
				return -1
			default:
				return 1
			}
		}
		return CompareReflectValue(a.Elem(), b.Elem())

	default:
		return 0
	}
}
