// Package normtest provides a reusable conformance test for the reading and
// writing contract of the string types of this module, like country.Code,
// vat.ID or date.Date.
//
// With Prettify set, reading and writing trim, and uppercase when the
// function does, including a value Normalized would reject or rewrite.
// Value does not return an error for that value. A whitespace-only value
// is still empty, and SQL NULL from Value. ScanString with validate still
// stores Canonical or returns an error.
//
// With SQLNormalized set, Scan and Value keep using Normalized: a variant
// becomes Canonical, the extra Blank entries are empty, and Value returns
// an error for an invalid value. JSON, XML and MarshalText use Prettify.
//
// Without Prettify the strict contract is checked: every path stores
// Canonical, an invalid value stays unchanged, and Value returns an error
// for it.
//
// fmt prints the stored value. The types have no normalizing String.
//
// It is only imported from _test.go files.
package normtest

import (
	"bytes"
	"database/sql"
	"database/sql/driver"
	"encoding"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/domonda/go-types/strutil"
)

// Trim is the prettify function of the types that only remove surrounding
// whitespace while reading and writing.
func Trim[T ~string](v T) T {
	return T(strutil.TrimSpace(string(v)))
}

// TrimUpper is the prettify function of the types that also uppercase.
func TrimUpper[T ~string](v T) T {
	return T(strings.ToUpper(strutil.TrimSpace(string(v))))
}

// Type is a normalizing string type with its writing methods.
type Type[T any] interface {
	~string
	Normalized() (T, error)
	Valid() bool
	Validate() error
	ValidAndNormalized() bool
	driver.Valuer
	json.Marshaler
	encoding.TextMarshaler
}

// Ptr is the pointer type of a Type with its reading methods.
type Ptr[T any] interface {
	*T
	sql.Scanner
	json.Unmarshaler
	xml.Unmarshaler
	xml.UnmarshalerAttr
}

// Spec lists the spellings of one type that Run checks.
type Spec[T any] struct {
	// Canonical is a valid value in normalized form.
	Canonical T
	// Variants are other spellings that normalize to Canonical.
	Variants []T
	// Invalid values can't be normalized and are not blank.
	Invalid []T
	// Blank are spellings of "no value" in addition to "", "   " and "\t".
	Blank []T
	// Nullable is true for a type whose blank value is its valid null value.
	Nullable bool
	// InvalidErr, if not nil, must be wrapped by the error that Value
	// returns for an invalid value, so errors.Is keeps working.
	InvalidErr error
	// StoredString is true for a type with a String method that returns the
	// stored value, like go-enum generates it for the enum types and like
	// date.YearMonth and date.YearQuarter have it. All other types have no
	// String.
	StoredString bool
	// Prettify is what JSON, XML, MarshalText and ScanString(false) apply,
	// and what Scan and Value apply unless SQLNormalized is set. It must
	// not fail. Nil selects the strict contract described on the package.
	Prettify func(T) T
	// ZeroBlank stores Blank values as empty on every path, including
	// JSON, XML and MarshalText.
	ZeroBlank bool
	// SQLNormalized makes Scan and Value use Normalized. Prettify must be
	// set. JSON, XML and MarshalText still use Prettify, unless ZeroBlank
	// is set, in which case every Blank value is empty on those paths too.
	SQLNormalized bool
}

type xmlDoc[T any] struct {
	XMLName xml.Name `xml:"doc"`
	Attr    T        `xml:"attr,attr"`
	Elem    T        `xml:"elem"`
}

type xmlCharData[T any] struct {
	XMLName xml.Name `xml:"doc"`
	Text    T        `xml:",chardata"`
}

type xmlPtrCharData[T any] struct {
	XMLName xml.Name `xml:"doc"`
	Text    *T       `xml:",chardata"`
}

type xmlPtrAttr[T any] struct {
	XMLName   xml.Name `xml:"doc"`
	Attr      *T       `xml:"attr,attr"`
	OmitEmpty *T       `xml:"omitempty,attr,omitempty"`
}

type stringScanner interface {
	ScanString(source string, validate bool) error
}

// Run checks the reading and writing contract of T against spec.
func Run[T Type[T], P Ptr[T]](t *testing.T, spec Spec[T]) {
	t.Helper()

	valid := append([]T{spec.Canonical}, spec.Variants...)
	blank := append([]T{"", "   ", "\t"}, spec.Blank...)
	typeName := reflect.TypeFor[T]().String()
	spaceBlank := func(v T) bool {
		return strutil.TrimSpace(string(v)) == ""
	}
	// textStored is what JSON, XML, MarshalText and ScanString(false) store.
	textStored := func(v T) T {
		if spec.Prettify != nil {
			return spec.Prettify(v)
		}
		return spec.Canonical
	}
	// sqlStored is what Scan and Value store for a variant.
	sqlStored := func(v T) T {
		if spec.Prettify != nil && !spec.SQLNormalized {
			return spec.Prettify(v)
		}
		return spec.Canonical
	}
	sqlRead := func(path string) bool {
		return strings.HasPrefix(path, "Scan(")
	}

	// read stores input through every reading path and returns
	// what each path stored, keyed by the path name.
	read := func(t *testing.T, input T) map[string]T {
		t.Helper()
		got := make(map[string]T)

		var scanned T
		if err := P(&scanned).Scan(string(input)); err != nil {
			t.Errorf("Scan(%q) error: %v", input, err)
		}
		got["Scan(string)"] = scanned

		var scannedBytes T
		if err := P(&scannedBytes).Scan([]byte(input)); err != nil {
			t.Errorf("Scan([]byte(%q)) error: %v", input, err)
		}
		got["Scan([]byte)"] = scannedBytes

		var unmarshalled T
		if err := json.Unmarshal(jsonString(string(input)), P(&unmarshalled)); err != nil {
			t.Errorf("json.Unmarshal(%q) error: %v", input, err)
		}
		got["UnmarshalJSON"] = unmarshalled

		var doc xmlDoc[T]
		data := fmt.Sprintf(`<doc attr="%s"><elem>%s</elem></doc>`, xmlEscape(string(input)), xmlEscape(string(input)))
		if err := xml.Unmarshal([]byte(data), &doc); err != nil {
			t.Errorf("xml.Unmarshal(%s) error: %v", data, err)
		}
		got["XML attribute"] = doc.Attr
		got["XML element"] = doc.Elem

		var scannedString T
		if s, ok := any(P(&scannedString)).(stringScanner); ok {
			if err := s.ScanString(string(input), false); err != nil {
				t.Errorf("ScanString(%q, false) error: %v", input, err)
			}
			got["ScanString(validate=false)"] = scannedString
		}
		return got
	}

	t.Run("reading a valid value", func(t *testing.T) {
		for _, input := range valid {
			for path, got := range read(t, input) {
				want := textStored(input)
				if sqlRead(path) {
					want = sqlStored(input)
				}
				if got != want {
					t.Errorf("%s of %q stored %q, want %q", path, input, got, want)
				}
			}
		}
	})

	t.Run("reading an invalid value without error", func(t *testing.T) {
		for _, input := range spec.Invalid {
			for path, got := range read(t, input) {
				want := input
				if spec.Prettify != nil && (!sqlRead(path) || !spec.SQLNormalized) {
					want = spec.Prettify(input)
				}
				if got != want {
					t.Errorf("%s of invalid %q stored %q, want %q", path, input, got, want)
				}
			}
		}
	})

	t.Run("reading a blank value", func(t *testing.T) {
		for _, input := range blank {
			for path, got := range read(t, input) {
				want := T("")
				if spec.Prettify != nil && !spec.ZeroBlank && !spaceBlank(input) && (!sqlRead(path) || !spec.SQLNormalized) {
					want = spec.Prettify(input)
				}
				if got != want {
					t.Errorf("%s of blank %q stored %q, want %q", path, input, got, want)
				}
			}
		}
	})

	t.Run("reading null and wrong types", func(t *testing.T) {
		v := spec.Canonical
		if err := P(&v).Scan(nil); err != nil || v != "" {
			t.Errorf("Scan(nil) = %q, %v; want \"\", nil", v, err)
		}

		v = spec.Canonical
		if err := P(&v).Scan(int64(1)); err == nil {
			t.Errorf("Scan(int64) error = nil, want an error for a wrong SQL type")
		}

		// encoding/json leaves a non-pointer value unchanged for null;
		// a nullable type must reset it, or a field that held a value
		// would keep it when the JSON says there is none.
		v = spec.Canonical
		if err := json.Unmarshal([]byte(`null`), P(&v)); err != nil {
			t.Errorf("json.Unmarshal(null) error: %v", err)
		}
		want := spec.Canonical
		if spec.Nullable {
			want = ""
		}
		if v != want {
			t.Errorf("json.Unmarshal(null) into %q stored %q, want %q", spec.Canonical, v, want)
		}

		v = spec.Canonical
		if err := json.Unmarshal([]byte(`123`), P(&v)); err == nil {
			t.Errorf("json.Unmarshal(123) error = nil, want an error for a JSON number")
		} else if !strings.Contains(err.Error(), typeName) {
			t.Errorf("json.Unmarshal(123) error %q doesn't name %s", err, typeName)
		}
	})

	t.Run("ScanString with validate", func(t *testing.T) {
		if _, ok := any(P(new(T))).(stringScanner); !ok {
			t.Skipf("%s has no ScanString method", typeName)
		}
		scanString := func(input T) (v T, err error) {
			return v, any(P(&v)).(stringScanner).ScanString(string(input), true)
		}
		for _, input := range valid {
			if v, err := scanString(input); err != nil || v != spec.Canonical {
				t.Errorf("ScanString(%q, true) = %q, %v; want %q, nil", input, v, err, spec.Canonical)
			}
		}
		for _, input := range spec.Invalid {
			if _, err := scanString(input); err == nil {
				t.Errorf("ScanString(%q, true) error = nil, want an error for an invalid value", input)
			}
		}
		// A required value can't be missing, a nullable one can
		v, err := scanString("")
		if spec.Nullable && (err != nil || v != "") {
			t.Errorf("ScanString(\"\", true) = %q, %v; want null", v, err)
		}
		if !spec.Nullable && err == nil {
			t.Errorf("ScanString(\"\", true) error = nil, want an error for a missing required value")
		}
		// strfmt writes null as one of these words,
		// which a nullable type reads back as null
		if spec.Nullable {
			for _, null := range []T{"NULL", "null", "nil"} {
				if v, err := scanString(null); err != nil || v != "" {
					t.Errorf("ScanString(%q, true) = %q, %v; want null", null, v, err)
				}
			}
		}
	})

	t.Run("Value", func(t *testing.T) {
		for _, input := range valid {
			got, err := input.Value()
			want := string(sqlStored(input))
			if err != nil || got != want {
				t.Errorf("Value() of %q = %#v, %v; want %q, nil", input, got, err, want)
			}
		}
		for _, input := range spec.Invalid {
			got, err := input.Value()
			if spec.Prettify != nil && !spec.SQLNormalized {
				want := string(spec.Prettify(input))
				if err != nil || got != want {
					t.Errorf("Value() of invalid %q = %#v, %v; want %q, nil", input, got, err, want)
				}
				continue
			}
			if err == nil {
				t.Errorf("Value() of invalid %q = %#v, nil; want an error", input, got)
			}
			if spec.InvalidErr != nil && !errors.Is(err, spec.InvalidErr) {
				t.Errorf("Value() error of %q = %v, want it to wrap %v", input, err, spec.InvalidErr)
			}
		}
		for _, input := range blank {
			got, err := input.Value()
			if spec.Prettify != nil && !spec.SQLNormalized && !spaceBlank(input) {
				want := string(spec.Prettify(input))
				if err != nil || got != want {
					t.Errorf("Value() of blank %q = %#v, %v; want %q, nil", input, got, err, want)
				}
				continue
			}
			if err != nil || got != nil {
				t.Errorf("Value() of blank %q = %#v, %v; want SQL NULL", input, got, err)
			}
		}
	})

	t.Run("MarshalJSON, MarshalText and XML", func(t *testing.T) {
		check := func(input T, want string, wantJSON string) {
			t.Helper()
			gotJSON, err := json.Marshal(input)
			if err != nil || string(gotJSON) != wantJSON {
				t.Errorf("json.Marshal(%q) = %s, %v; want %s", input, gotJSON, err, wantJSON)
			}
			gotText, err := input.MarshalText()
			if err != nil || string(gotText) != want {
				t.Errorf("MarshalText() of %q = %q, %v; want %q", input, gotText, err, want)
			}
			gotXML, err := xml.Marshal(xmlDoc[T]{Attr: input, Elem: input})
			wantXML := fmt.Sprintf(`<doc attr="%s"><elem>%s</elem></doc>`, xmlEscape(want), xmlEscape(want))
			if err != nil || string(gotXML) != wantXML {
				t.Errorf("xml.Marshal of %q = %s, %v; want %s", input, gotXML, err, wantXML)
			}
			// A pointer attribute is written by MarshalXMLAttr instead of
			// MarshalText, see below, which must not change the text
			gotXML, err = xml.Marshal(xmlPtrAttr[T]{Attr: &input})
			wantXML = fmt.Sprintf(`<doc attr="%s"></doc>`, xmlEscape(want))
			if err != nil || string(gotXML) != wantXML {
				t.Errorf("xml.Marshal of *%q attribute = %s, %v; want %s", input, gotXML, err, wantXML)
			}
		}
		for _, input := range valid {
			want := spec.Canonical
			if spec.Prettify != nil {
				want = spec.Prettify(input)
			}
			check(input, string(want), string(jsonString(string(want))))
		}
		for _, input := range spec.Invalid {
			want := input
			if spec.Prettify != nil {
				want = spec.Prettify(input)
			}
			check(input, string(want), string(jsonString(string(want))))
		}
		for _, input := range blank {
			if spec.Prettify != nil && !spec.ZeroBlank && !spaceBlank(input) {
				want := spec.Prettify(input)
				check(input, string(want), string(jsonString(string(want))))
				continue
			}
			wantJSON := `""`
			if spec.Nullable {
				wantJSON = `null`
			}
			check(input, "", wantJSON)
		}

		// encoding/xml calls the value receiver MarshalText through a nil
		// pointer attribute field without omitempty and panics, like for a
		// *time.Time, unless the pointer implements xml.MarshalerAttr,
		// which encoding/xml checks first. A nil pointer attribute must be
		// left out, with and without omitempty.
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("xml.Marshal of nil *%s attribute panicked: %v", typeName, r)
				}
			}()
			gotXML, err := xml.Marshal(xmlPtrAttr[T]{})
			if err != nil || string(gotXML) != `<doc></doc>` {
				t.Errorf("xml.Marshal of nil *%s attributes = %s, %v; want <doc></doc>", typeName, gotXML, err)
			}
		}()
	})

	t.Run("round trips", func(t *testing.T) {
		for _, input := range valid {
			sqlValue, err := input.Value()
			if err != nil {
				t.Fatalf("Value() of %q error: %v", input, err)
			}
			var scanned T
			if err := P(&scanned).Scan(sqlValue); err != nil || scanned != sqlStored(input) {
				t.Errorf("Scan(Value()) of %q = %q, %v; want %q", input, scanned, err, sqlStored(input))
			}
		}
		// JSON and XML write the prettified value and read it back,
		// so they round-trip every value that prettify does not change further.
		roundTrip := func(input, want T) {
			t.Helper()
			data, err := json.Marshal(input)
			if err != nil {
				t.Fatalf("json.Marshal(%q) error: %v", input, err)
			}
			var unmarshalled T
			if err := json.Unmarshal(data, P(&unmarshalled)); err != nil || unmarshalled != want {
				t.Errorf("JSON round trip of %q = %q, %v; want %q", input, unmarshalled, err, want)
			}
			data, err = xml.Marshal(xmlDoc[T]{Attr: input, Elem: input})
			if err != nil {
				t.Fatalf("xml.Marshal(%q) error: %v", input, err)
			}
			var doc xmlDoc[T]
			if err := xml.Unmarshal(data, &doc); err != nil || doc.Attr != want || doc.Elem != want {
				t.Errorf("XML round trip of %q = %q, %q, %v; want %q", input, doc.Attr, doc.Elem, err, want)
			}
		}
		for _, input := range valid {
			want := spec.Canonical
			if spec.Prettify != nil {
				want = spec.Prettify(input)
			}
			roundTrip(input, want)
		}
		for _, input := range spec.Invalid {
			want := input
			if spec.Prettify != nil {
				want = spec.Prettify(input)
			}
			roundTrip(input, want)
		}
	})

	t.Run("validation", func(t *testing.T) {
		for _, input := range valid {
			if !input.Valid() || input.Validate() != nil {
				t.Errorf("%q: Valid() = %v, Validate() = %v; want valid", input, input.Valid(), input.Validate())
			}
			if got := input.ValidAndNormalized(); got != (input == spec.Canonical) {
				t.Errorf("%q: ValidAndNormalized() = %v, want %v", input, got, input == spec.Canonical)
			}
		}
		for _, input := range spec.Invalid {
			if input.Valid() || input.Validate() == nil || input.ValidAndNormalized() {
				t.Errorf("%q: Valid() = %v, Validate() = %v, ValidAndNormalized() = %v; want invalid",
					input, input.Valid(), input.Validate(), input.ValidAndNormalized())
			}
			// The writing methods rely on Normalized returning
			// its receiver for a value it can't normalize
			if norm, err := input.Normalized(); err == nil || norm != input {
				t.Errorf("%q: Normalized() = %q, %v; want the receiver and an error", input, norm, err)
			}
		}
		for _, input := range blank {
			if input.Valid() != spec.Nullable {
				t.Errorf("blank %q: Valid() = %v, want %v", input, input.Valid(), spec.Nullable)
			}
		}
	})

	t.Run("fmt prints the stored value", func(t *testing.T) {
		for _, input := range append(append([]T{}, valid...), spec.Invalid...) {
			if got := fmt.Sprint(input); got != string(input) {
				t.Errorf("fmt.Sprint(%q) = %q, want the stored value", string(input), got)
			}
			if got := fmt.Sprintf("%q", input); got != fmt.Sprintf("%q", string(input)) {
				t.Errorf("fmt %%q of %q = %s, want the stored value", string(input), got)
			}
		}
		_, isStringer := any(spec.Canonical).(fmt.Stringer)
		if isStringer != spec.StoredString {
			t.Errorf("%s implements fmt.Stringer: %v, want %v", typeName, isStringer, spec.StoredString)
		}
	})

	t.Run("JSON map keys stay as given", func(t *testing.T) {
		// encoding/json reads and writes the keys of a map with a string
		// key type as given, as long as the type has no UnmarshalText,
		// so two spellings of the same value stay two keys and a map
		// round-trips with all its entries.
		//
		// The Go 1.26 encoding/json decodes the key of a type with
		// UnmarshalText through its UnmarshalJSON, which normalizes and
		// merges the keys. Go 1.27 builds encoding/json on json/v2,
		// which doesn't, so the round trip below only fails on Go 1.26
		// and this check catches an UnmarshalText on every Go version.
		if _, ok := any(new(T)).(encoding.TextUnmarshaler); ok {
			t.Errorf("%s implements encoding.TextUnmarshaler, so Go 1.26 encoding/json normalizes its map keys", typeName)
		}
		for _, variant := range spec.Variants {
			if variant == spec.Canonical {
				continue
			}
			m := map[T]int{variant: 1, spec.Canonical: 2}
			data, err := json.Marshal(m)
			if err != nil {
				t.Fatalf("json.Marshal(%v) error: %v", m, err)
			}
			var back map[T]int
			if err := json.Unmarshal(data, &back); err != nil {
				t.Fatalf("json.Unmarshal(%s) error: %v", data, err)
			}
			if !reflect.DeepEqual(back, m) {
				t.Errorf("JSON map round trip of %v = %v, want the keys unchanged", m, back)
			}
		}
	})

	t.Run("XML chardata stays as given", func(t *testing.T) {
		// encoding/xml reads a ,chardata field only through
		// encoding.TextUnmarshaler, which the types don't implement
		// with normalization, so such a field keeps the text as given.
		for _, input := range spec.Variants {
			var doc xmlCharData[T]
			data := fmt.Sprintf(`<doc>%s</doc>`, xmlEscape(string(input)))
			if err := xml.Unmarshal([]byte(data), &doc); err != nil || doc.Text != input {
				t.Errorf("xml.Unmarshal(%s) into ,chardata = %q, %v; want %q", data, doc.Text, err, input)
			}
		}
	})

	t.Run("nil pointer chardata panics", func(t *testing.T) {
		// encoding/xml calls the value receiver MarshalText through a nil
		// pointer ,chardata field and has no hook like MarshalXMLAttr to
		// skip it, so the README says to use a non-pointer field. This pins
		// that limitation: it fails once encoding/xml skips a nil pointer,
		// and the README note can go.
		panicked := func() (panicked bool) {
			defer func() { panicked = recover() != nil }()
			_, _ = xml.Marshal(xmlPtrCharData[T]{})
			return false
		}()
		if !panicked {
			t.Errorf("xml.Marshal of a nil *%s ,chardata field doesn't panic anymore, remove the README note about it", typeName)
		}
	})
}

func jsonString(s string) []byte {
	data, err := json.Marshal(s)
	if err != nil {
		panic(err)
	}
	return data
}

func xmlEscape(s string) string {
	var b bytes.Buffer
	if err := xml.EscapeText(&b, []byte(s)); err != nil {
		panic(err)
	}
	return b.String()
}
