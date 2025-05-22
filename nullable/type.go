package nullable

import (
	"bytes"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"time"

	"github.com/invopop/jsonschema"
)

// Implemented interfaces
var (
	_ driver.Valuer    = Type[int]{}
	_ sql.Scanner      = &Type[int]{}
	_ json.Marshaler   = Type[int]{}
	_ json.Unmarshaler = &Type[int]{}
)

// Type wraps a type T to support null values without resorting to pointers.
//
// The zero value represents null.
//
// Type implements the following interfaces:
//   - database/sql/driver.Valuer
//   - database/sql.Scanner
//   - encoding/json.Marshaler
//   - encoding/json.Unmarshaler
//
// It also provides a JSONSchema method to generate a JSON Schema for the type
// using the github.com/invopop/jsonschema package.
//
// Use TypeFromPtr to create a valid nullable type from a pointer.
// Use Type.Ptr to get a pointer to the value.
// Use Type.Get to get a non-null value.
// Use Type.GetOr to get a non-null value or a default value if the value is null.
// Use Type.Set to set the value.
// Use Type.SetNull to set the value to null.
type Type[T any] struct {
	value T
	valid bool
}

// TypeFromPtr returns a nullable type from a pointer
// using nil as the null value.
//
// See Type[T].Ptr for the inverse.
func TypeFromPtr[T any](ptr *T) Type[T] {
	if ptr == nil {
		return Type[T]{}
	}
	return Type[T]{value: *ptr, valid: true}
}

// Ptr returns a pointer to the value
// or nil if the value is null.
//
// See TypeFromPtr[T] for the inverse.
func (t Type[T]) Ptr() *T {
	if !t.valid {
		return nil
	}
	return &t.value
}

// IsNull returns true if the value is null.
func (t Type[T]) IsNull() bool {
	return !t.valid
}

// IsNotNull returns true if the value is not null.
func (t Type[T]) IsNotNull() bool {
	return t.valid
}

// Get returns the non-null value
// or panics if the value is null.
//
// Use IsNotNull first to check if the value is not null.
func (t Type[T]) Get() T {
	if !t.valid {
		panic(fmt.Sprintf("Get() called on NULL %T", t.value))
	}
	return t.value
}

// GetOr returns the non-null value
// or a default value if the value is null.
func (t Type[T]) GetOr(defaultValue T) T {
	if !t.valid {
		return defaultValue
	}
	return t.value
}

// Set sets a non-null value.
func (t *Type[T]) Set(value T) {
	t.value = value
	t.valid = true
}

// SetNull sets the value to null.
func (t *Type[T]) SetNull() {
	t.value = *new(T)
	t.valid = false
}

// Value implements the driver database/sql/driver.Valuer interface.
func (t Type[T]) Value() (driver.Value, error) {
	if !t.valid {
		return nil, nil
	}
	return t.value, nil
}

// Scan implements the database/sql.Scanner interface.
func (t *Type[T]) Scan(value any) error {
	if value == nil {
		t.SetNull()
		return nil
	}
	err := convertAssign(&t.value, value)
	if err != nil {
		return err
	}
	t.valid = true
	return nil
}

// UnarshalJSON implements encoding/json.Unmarshaler.
// Interprets []byte(nil), []byte(""), []byte("null") as null.
func (t *Type[T]) UnmarshalJSON(sourceJSON []byte) error {
	if len(sourceJSON) == 0 || bytes.Equal(sourceJSON, []byte("null")) {
		t.SetNull()
		return nil
	}
	err := json.Unmarshal(sourceJSON, &t.value)
	if err != nil {
		return err
	}
	t.valid = true
	return nil
}

// MarshalJSON implements encoding/json.Marshaler
func (t Type[T]) MarshalJSON() ([]byte, error) {
	if !t.valid {
		return []byte("null"), nil
	}
	return json.Marshal(t.value)
}

// JSONSchema returns a JSON Schema for the type
// using the github.com/invopop/jsonschema package.
func (t Type[T]) JSONSchema() *jsonschema.Schema {
	baseTypeSchema := jsonschema.Reflect(t.value)
	if baseTypeSchema.Title != "" {
		baseTypeSchema.Title = "Nullable " + baseTypeSchema.Title
	}
	if len(baseTypeSchema.OneOf) > 0 {
		// Check if base type already has null as oneOf type option
		if slices.ContainsFunc(baseTypeSchema.OneOf, func(s *jsonschema.Schema) bool {
			return s.Type == "null"
		}) {
			if baseTypeSchema.Default == nil {
				baseTypeSchema.Default = Type[T]{} // null
			}
			return baseTypeSchema
		}
		// Base type already has oneOf type options,
		// add null as another oneOf type option
		baseTypeSchema.OneOf = append(baseTypeSchema.OneOf, &jsonschema.Schema{Type: "null"})
		baseTypeSchema.Default = Type[T]{} // null
		return baseTypeSchema
	}
	// Copy version, title and description for new nullable type schema
	version := baseTypeSchema.Version
	title := baseTypeSchema.Title
	description := baseTypeSchema.Description
	// Clear version, title and description for base type schema
	baseTypeSchema.Version = ""
	baseTypeSchema.Title = ""
	baseTypeSchema.Description = ""
	return &jsonschema.Schema{
		Version:     version,
		Title:       title,
		Description: description,
		// Add null as oneOf type option
		OneOf: []*jsonschema.Schema{
			baseTypeSchema,
			{Type: "null"},
		},
		Default: Type[T]{}, // null
	}
}

///////////////////////////////////////////////////////////////////////////////
// The following code is copied and slightly simplified from
// database/sql/convert.go

var errNilPtr = errors.New("destination pointer is nil") // embedded in descriptive error

type decimalDecompose interface {
	// Decompose returns the internal decimal state in parts.
	// If the provided buf has sufficient capacity, buf may be returned as the coefficient with
	// the value set and length set as appropriate.
	Decompose(buf []byte) (form byte, negative bool, coefficient []byte, exponent int32)
}

type decimalCompose interface {
	// Compose sets the internal decimal value from parts. If the value cannot be
	// represented then an error should be returned.
	Compose(form byte, negative bool, coefficient []byte, exponent int32) error
}

// convertAssign copies to dest the value in src, converting it if possible.
// An error is returned if the copy would result in loss of information.
// dest should be a pointer type.
func convertAssign(dest any, src driver.Value) error {
	// Common cases, without reflect.
	switch s := src.(type) {
	case string:
		switch d := dest.(type) {
		case *string:
			if d == nil {
				return errNilPtr
			}
			*d = s
			return nil
		case *[]byte:
			if d == nil {
				return errNilPtr
			}
			*d = []byte(s)
			return nil
		case *sql.RawBytes:
			if d == nil {
				return errNilPtr
			}
			*d = sql.RawBytes(s)
			return nil
		}
	case []byte:
		switch d := dest.(type) {
		case *string:
			if d == nil {
				return errNilPtr
			}
			*d = string(s)
			return nil
		case *any:
			if d == nil {
				return errNilPtr
			}
			*d = bytes.Clone(s)
			return nil
		case *[]byte:
			if d == nil {
				return errNilPtr
			}
			*d = bytes.Clone(s)
			return nil
		case *sql.RawBytes:
			if d == nil {
				return errNilPtr
			}
			*d = bytes.Clone(s)
			return nil
		}
	case time.Time:
		switch d := dest.(type) {
		case *time.Time:
			*d = s
			return nil
		case *string:
			*d = s.Format(time.RFC3339Nano)
			return nil
		case *[]byte:
			if d == nil {
				return errNilPtr
			}
			*d = []byte(s.Format(time.RFC3339Nano))
			return nil
		case *sql.RawBytes:
			if d == nil {
				return errNilPtr
			}
			*d = sql.RawBytes(s.Format(time.RFC3339Nano))
			return nil
		}
	case decimalDecompose:
		switch d := dest.(type) {
		case decimalCompose:
			return d.Compose(s.Decompose(nil))
		}
	case nil:
		switch d := dest.(type) {
		case *any:
			if d == nil {
				return errNilPtr
			}
			*d = nil
			return nil
		case *[]byte:
			if d == nil {
				return errNilPtr
			}
			*d = nil
			return nil
		case *sql.RawBytes:
			if d == nil {
				return errNilPtr
			}
			*d = nil
			return nil
		}
	}

	var sv reflect.Value

	switch d := dest.(type) {
	case *string:
		sv = reflect.ValueOf(src)
		switch sv.Kind() {
		case reflect.Bool,
			reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Float32, reflect.Float64:
			*d = asString(src)
			return nil
		}
	case *[]byte:
		sv = reflect.ValueOf(src)
		if b, ok := asBytes(nil, sv); ok {
			*d = b
			return nil
		}
	case *sql.RawBytes:
		sv = reflect.ValueOf(src)
		if b, ok := asBytes(nil, sv); ok {
			*d = b
			return nil
		}
	case *bool:
		bv, err := driver.Bool.ConvertValue(src)
		if err == nil {
			*d = bv.(bool)
		}
		return err
	case *any:
		*d = src
		return nil
	}

	if scanner, ok := dest.(sql.Scanner); ok {
		return scanner.Scan(src)
	}

	dpv := reflect.ValueOf(dest)
	if dpv.Kind() != reflect.Pointer {
		return errors.New("destination not a pointer")
	}
	if dpv.IsNil() {
		return errNilPtr
	}

	if !sv.IsValid() {
		sv = reflect.ValueOf(src)
	}

	dv := reflect.Indirect(dpv)
	if sv.IsValid() && sv.Type().AssignableTo(dv.Type()) {
		switch b := src.(type) {
		case []byte:
			dv.Set(reflect.ValueOf(bytes.Clone(b)))
		default:
			dv.Set(sv)
		}
		return nil
	}

	if dv.Kind() == sv.Kind() && sv.Type().ConvertibleTo(dv.Type()) {
		dv.Set(sv.Convert(dv.Type()))
		return nil
	}

	// The following conversions use a string value as an intermediate representation
	// to convert between various numeric types.
	//
	// This also allows scanning into user defined types such as "type Int int64".
	// For symmetry, also check for string destination types.
	switch dv.Kind() {
	case reflect.Pointer:
		if src == nil {
			dv.SetZero()
			return nil
		}
		dv.Set(reflect.New(dv.Type().Elem()))
		return convertAssign(dv.Interface(), src)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if src == nil {
			return fmt.Errorf("converting NULL to %s is unsupported", dv.Kind())
		}
		s := asString(src)
		i64, err := strconv.ParseInt(s, 10, dv.Type().Bits())
		if err != nil {
			err = strconvErr(err)
			return fmt.Errorf("converting driver.Value type %T (%q) to a %s: %v", src, s, dv.Kind(), err)
		}
		dv.SetInt(i64)
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if src == nil {
			return fmt.Errorf("converting NULL to %s is unsupported", dv.Kind())
		}
		s := asString(src)
		u64, err := strconv.ParseUint(s, 10, dv.Type().Bits())
		if err != nil {
			err = strconvErr(err)
			return fmt.Errorf("converting driver.Value type %T (%q) to a %s: %v", src, s, dv.Kind(), err)
		}
		dv.SetUint(u64)
		return nil
	case reflect.Float32, reflect.Float64:
		if src == nil {
			return fmt.Errorf("converting NULL to %s is unsupported", dv.Kind())
		}
		s := asString(src)
		f64, err := strconv.ParseFloat(s, dv.Type().Bits())
		if err != nil {
			err = strconvErr(err)
			return fmt.Errorf("converting driver.Value type %T (%q) to a %s: %v", src, s, dv.Kind(), err)
		}
		dv.SetFloat(f64)
		return nil
	case reflect.String:
		if src == nil {
			return fmt.Errorf("converting NULL to %s is unsupported", dv.Kind())
		}
		switch v := src.(type) {
		case string:
			dv.SetString(v)
			return nil
		case []byte:
			dv.SetString(string(v))
			return nil
		}
	}

	return fmt.Errorf("unsupported Scan, storing driver.Value type %T into type %T", src, dest)
}

func strconvErr(err error) error {
	if ne, ok := err.(*strconv.NumError); ok {
		return ne.Err
	}
	return err
}

func asString(src any) string {
	switch v := src.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	}
	rv := reflect.ValueOf(src)
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(rv.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(rv.Uint(), 10)
	case reflect.Float64:
		return strconv.FormatFloat(rv.Float(), 'g', -1, 64)
	case reflect.Float32:
		return strconv.FormatFloat(rv.Float(), 'g', -1, 32)
	case reflect.Bool:
		return strconv.FormatBool(rv.Bool())
	}
	return fmt.Sprintf("%v", src)
}

func asBytes(buf []byte, rv reflect.Value) (b []byte, ok bool) {
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.AppendInt(buf, rv.Int(), 10), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.AppendUint(buf, rv.Uint(), 10), true
	case reflect.Float32:
		return strconv.AppendFloat(buf, rv.Float(), 'g', -1, 32), true
	case reflect.Float64:
		return strconv.AppendFloat(buf, rv.Float(), 'g', -1, 64), true
	case reflect.Bool:
		return strconv.AppendBool(buf, rv.Bool()), true
	case reflect.String:
		s := rv.String()
		return append(buf, s...), true
	}
	return
}
