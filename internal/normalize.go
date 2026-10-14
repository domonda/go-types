package internal

import (
	"bytes"
	"database/sql/driver"
	"encoding"
	"encoding/json"
	"encoding/xml"
	"errors"
	"reflect"

	"github.com/domonda/go-errs"

	"github.com/domonda/go-types/strutil"
)

// The helpers in this file implement the reading and writing contract
// shared by the normalizing string types of the domain packages,
// like country.Code, vat.ID or date.Date:
//
//   - Reading (Scan, UnmarshalJSON, UnmarshalXML, UnmarshalXMLAttr)
//     stores a valid value normalized, a blank value as "" and an
//     invalid value unchanged, without an error, so that one bad value
//     doesn't fail a whole query or document and Validate can report it.
//   - Value writes the normalized value, SQL NULL for a blank value,
//     and returns the error of normalize for an invalid value,
//     because the database is where an invalid value would stay.
//   - MarshalJSON and MarshalText write the normalized value,
//     "" for a blank value and an invalid value unchanged.
//     MarshalXMLAttr writes MarshalText, and no attribute for a nil pointer.
//
// Each helper takes the Normalized method of the type as normalize,
// which must return its receiver unchanged together with an error
// for a value that can't be normalized. isBlank reports values that
// mean "no value", a nil isBlank uses IsBlank. nullable is true for
// a nullable type, whose blank value is null.

// IsBlank reports whether s is empty after strutil.TrimSpace.
func IsBlank[S ~string](s S) bool {
	return strutil.TrimSpace(s) == ""
}

// NormalizedOrRaw returns "" for a blank s, the normalized s if it can be
// normalized, or else s unchanged.
func NormalizedOrRaw[S ~string](s S, normalize func(S) (S, error), isBlank func(S) bool) S {
	if isBlank == nil {
		isBlank = IsBlank[S]
	}
	if isBlank(s) {
		return ""
	}
	if norm, err := normalize(s); err == nil {
		return norm
	}
	return s
}

// ScanNormalized implements database/sql.Scanner for a normalizing
// string type: a string or []byte is stored with NormalizedOrRaw,
// SQL NULL as "". Other types are an error.
func ScanNormalized[S ~string](dst *S, value any, normalize func(S) (S, error), isBlank func(S) bool) error {
	switch x := value.(type) {
	case string:
		*dst = NormalizedOrRaw(S(x), normalize, isBlank)
	case []byte:
		*dst = NormalizedOrRaw(S(x), normalize, isBlank)
	case nil:
		*dst = ""
	default:
		return errs.Errorf("can't scan SQL value of type %T as %s", value, reflect.TypeFor[S]())
	}
	return nil
}

// ScanStringNormalized implements the ScanString method of a normalizing
// string type used by the strfmt package. The strings "NULL", "null" and
// "nil" set a nullable type to null. Without validate it stores like
// NormalizedOrRaw, with validate it stores the normalized value or
// returns the error of normalize, also for a blank value of a type
// whose normalize rejects blank values.
func ScanStringNormalized[S ~string](dst *S, source string, validate, nullable bool, normalize func(S) (S, error), isBlank func(S) bool) error {
	if nullable {
		switch source {
		case "NULL", "null", "nil":
			*dst = ""
			return nil
		}
	}
	if !validate {
		*dst = NormalizedOrRaw(S(source), normalize, isBlank)
		return nil
	}
	norm, err := normalize(S(source))
	if err != nil {
		return err
	}
	*dst = norm
	return nil
}

// ValueNormalized implements database/sql/driver.Valuer for a normalizing
// string type: SQL NULL for a blank value, the normalized value as string,
// or the error of normalize for a value that can't be normalized.
func ValueNormalized[S ~string](s S, normalize func(S) (S, error), isBlank func(S) bool) (driver.Value, error) {
	if isBlank == nil {
		isBlank = IsBlank[S]
	}
	if isBlank(s) {
		return nil, nil
	}
	norm, err := normalize(s)
	if err != nil {
		return nil, err
	}
	return string(norm), nil
}

// MarshalJSONNormalized implements encoding/json.Marshaler for a normalizing
// string type by writing NormalizedOrRaw as JSON string, or JSON null
// instead of "" for a blank value if nullable is true.
func MarshalJSONNormalized[S ~string](s S, nullable bool, normalize func(S) (S, error), isBlank func(S) bool) ([]byte, error) {
	str := NormalizedOrRaw(s, normalize, isBlank)
	if str == "" && nullable {
		return []byte(`null`), nil
	}
	return json.Marshal(string(str))
}

// MarshalTextNormalized implements encoding.TextMarshaler for a normalizing
// string type by returning NormalizedOrRaw.
func MarshalTextNormalized[S ~string](s S, normalize func(S) (S, error), isBlank func(S) bool) ([]byte, error) {
	return []byte(NormalizedOrRaw(s, normalize, isBlank)), nil
}

// MarshalXMLAttrText implements encoding/xml.MarshalerAttr for a pointer
// to a type with a value receiver MarshalText: it writes the text of
// MarshalText like encoding/xml does, and no attribute for a nil v.
// encoding/xml calls a value receiver MarshalText through a nil pointer
// attribute field and panics, but checks for MarshalerAttr first.
func MarshalXMLAttrText[T encoding.TextMarshaler](v *T, name xml.Name) (xml.Attr, error) {
	if v == nil {
		return xml.Attr{}, nil
	}
	text, err := (*v).MarshalText()
	if err != nil {
		return xml.Attr{}, err
	}
	return xml.Attr{Name: name, Value: string(text)}, nil
}

// UnmarshalJSONNormalized implements encoding/json.Unmarshaler for a
// normalizing string type by storing a JSON string with NormalizedOrRaw.
// JSON null sets "" if nullable is true, which a nullable type needs to
// reset a field that holds a value, else it is a no-op like encoding/json
// does for null and a non-pointer value. Any other JSON value is an error
// that names the type S.
func UnmarshalJSONNormalized[S ~string](dst *S, data []byte, nullable bool, normalize func(S) (S, error), isBlank func(S) bool) error {
	if bytes.Equal(data, []byte(`null`)) {
		if nullable {
			*dst = ""
		}
		return nil
	}
	str, err := unmarshalJSONString[S](data)
	if err != nil {
		return err
	}
	*dst = NormalizedOrRaw(S(str), normalize, isBlank)
	return nil
}

// UnmarshalXMLNormalized implements encoding/xml.Unmarshaler for a
// normalizing string type by storing the text of the element
// with NormalizedOrRaw.
func UnmarshalXMLNormalized[S ~string](dst *S, d *xml.Decoder, start xml.StartElement, normalize func(S) (S, error), isBlank func(S) bool) error {
	var str string
	err := d.DecodeElement(&str, &start)
	if err != nil {
		return err
	}
	*dst = NormalizedOrRaw(S(str), normalize, isBlank)
	return nil
}

// UnmarshalXMLAttrNormalized implements encoding/xml.UnmarshalerAttr for a
// normalizing string type by storing the attribute value
// with NormalizedOrRaw.
func UnmarshalXMLAttrNormalized[S ~string](dst *S, attr xml.Attr, normalize func(S) (S, error), isBlank func(S) bool) error {
	*dst = NormalizedOrRaw(S(attr.Value), normalize, isBlank)
	return nil
}

// unmarshalJSONString decodes a JSON string, reporting any other JSON value
// with the type S instead of string like encoding/json does for a string
// type without UnmarshalJSON.
func unmarshalJSONString[S ~string](data []byte) (string, error) {
	var str string
	err := json.Unmarshal(data, &str)
	if err != nil {
		var typeErr *json.UnmarshalTypeError
		if errors.As(err, &typeErr) {
			return "", &json.UnmarshalTypeError{Value: typeErr.Value, Type: reflect.TypeFor[S](), Offset: typeErr.Offset}
		}
		return "", err
	}
	return str, nil
}
