package strfmt

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/domonda/go-types/country"
	"github.com/domonda/go-types/date"
	"github.com/domonda/go-types/money"
	"github.com/domonda/go-types/nullable"
)

func TestScan(t *testing.T) {
	type test struct {
		Int    int
		IntPtr *int
	}

	var (
		config     = NewScanConfig()
		int666 int = 666
	)

	type args struct {
		destField string
		source    string
		config    *ScanConfig
	}
	tests := []struct {
		name     string
		dest     test
		args     args
		wantErr  bool
		wantDest test
	}{
		{name: `666 Int`, dest: test{}, args: args{destField: "Int", source: "666", config: config}, wantErr: false, wantDest: test{Int: 666}},
		{name: `"" IntPtr`, dest: test{}, args: args{destField: "IntPtr", source: "", config: config}, wantErr: false, wantDest: test{}},
		{name: `null IntPtr`, dest: test{}, args: args{destField: "IntPtr", source: "null", config: config}, wantErr: false, wantDest: test{}},
		{name: `666 IntPtr`, dest: test{}, args: args{destField: "IntPtr", source: "666", config: config}, wantErr: false, wantDest: test{IntPtr: &int666}},
		// TODO
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dest := reflect.ValueOf(&tt.dest).Elem().FieldByName(tt.args.destField)
			err := Scan(dest, tt.args.source, tt.args.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("Scan() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.Equal(t, tt.wantDest, tt.dest)
		})
	}
}

// scanNilStringFails scans every string that config declares as nil into a
// destination prefilled with nonZero and requires an error and an unmodified
// destination, because a type that can't represent "no value" is not optional.
func scanNilStringFails[T comparable](t *testing.T, config *ScanConfig, nonZero T) {
	t.Helper()

	var zero T
	if nonZero == zero {
		t.Fatalf("test needs a non-zero value of type %T to prove that nothing was assigned", zero)
	}
	for _, source := range config.NilStrings {
		dest := nonZero
		err := Scan(reflect.ValueOf(&dest).Elem(), source, config)
		if err == nil {
			t.Errorf("Scan(%T, %q) returned no error for a non-optional destination type", dest, source)
		}
		if dest != nonZero {
			t.Errorf("Scan(%T, %q) modified the destination to %v", dest, source, dest)
		}
	}
}

// TestScanNilStringToNonOptionalDest checks that a nil source string is
// reported as an error for a destination type that can't represent
// "no value" when StrictEmptyStringParsing is enabled.
// An empty cell of a CSV file or a spreadsheet
// must not be scanned as the number zero, which can't be told apart from
// a scanned zero, and it keeps a struct field wired to a wrong column
// failing instead of silently scanning as a valid value.
func TestScanNilStringToNonOptionalDest(t *testing.T) {
	config := NewScanConfig()
	config.StrictEmptyStringParsing = true

	t.Run("bool", func(t *testing.T) { scanNilStringFails(t, config, true) })
	t.Run("int", func(t *testing.T) { scanNilStringFails(t, config, 666) })
	t.Run("int8", func(t *testing.T) { scanNilStringFails(t, config, int8(-8)) })
	t.Run("uint64", func(t *testing.T) { scanNilStringFails(t, config, uint64(666)) })
	t.Run("float32", func(t *testing.T) { scanNilStringFails(t, config, float32(6.66)) })
	t.Run("float64", func(t *testing.T) { scanNilStringFails(t, config, 6.66) })
	t.Run("struct", func(t *testing.T) { scanNilStringFails(t, config, struct{ Int int }{666}) })

	// Types with a scanning method of their own decide what a nil source
	// string means for them and report it as an error when they are not
	// optional, instead of every one of them having to be listed here.
	t.Run("time.Time", func(t *testing.T) {
		scanNilStringFails(t, config, time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC))
	})
	t.Run("time.Duration", func(t *testing.T) { scanNilStringFails(t, config, time.Hour) })
	t.Run("money.Amount", func(t *testing.T) { scanNilStringFails(t, config, money.Amount(6.66)) })
	t.Run("date.Date", func(t *testing.T) { scanNilStringFails(t, config, date.Date("2026-01-02")) })
	t.Run("country.Code", func(t *testing.T) { scanNilStringFails(t, config, country.Code("DE")) })
}

// TestScanNilStringToNilableDest checks that a nil source string sets a
// destination of a kind that can hold nil to nil, because nil is how
// those kinds represent "no value".
func TestScanNilStringToNilableDest(t *testing.T) {
	config := NewScanConfig()

	for _, source := range config.NilStrings {
		ptr := new(int)
		assert.NoError(t, Scan(reflect.ValueOf(&ptr).Elem(), source, config), "Scan(*int, %q)", source)
		assert.Nil(t, ptr, "Scan(*int, %q)", source)

		slice := []int{666}
		assert.NoError(t, Scan(reflect.ValueOf(&slice).Elem(), source, config), "Scan([]int, %q)", source)
		assert.Nil(t, slice, "Scan([]int, %q)", source)

		m := map[string]int{"key": 666}
		assert.NoError(t, Scan(reflect.ValueOf(&m).Elem(), source, config), "Scan(map[string]int, %q)", source)
		assert.Nil(t, m, "Scan(map[string]int, %q)", source)

		var iface any = 666
		assert.NoError(t, Scan(reflect.ValueOf(&iface).Elem(), source, config), "Scan(any, %q)", source)
		assert.Nil(t, iface, "Scan(any, %q)", source)
	}
}

// TestScanNilStringToNullableDest checks that a nil source string sets a
// nullable type to null with its SetNull method. Handling it in Scan means
// that every nullable type behaves the same way for an empty cell of a CSV
// file or a spreadsheet, without its scanning method having to repeat it.
func TestScanNilStringToNullableDest(t *testing.T) {
	config := NewScanConfig()

	for _, source := range config.NilStrings {
		nullTime := nullable.TimeFrom(time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC))
		assert.NoError(t, Scan(reflect.ValueOf(&nullTime).Elem(), source, config), "Scan(nullable.Time, %q)", source)
		assert.True(t, nullTime.IsNull(), "Scan(nullable.Time, %q) sets null", source)

		nullDate := date.NullableDate("2026-01-02")
		assert.NoError(t, Scan(reflect.ValueOf(&nullDate).Elem(), source, config), "Scan(date.NullableDate, %q)", source)
		assert.True(t, nullDate.IsNull(), "Scan(date.NullableDate, %q) sets null", source)

		nullCode := country.NullableCode("DE")
		assert.NoError(t, Scan(reflect.ValueOf(&nullCode).Elem(), source, config), "Scan(country.NullableCode, %q)", source)
		assert.True(t, nullCode.IsNull(), "Scan(country.NullableCode, %q) sets null", source)
	}
}

// TestScanNilStringToStringDest checks that a string destination is assigned
// the source string itself instead of the zero value, because unlike the
// number and bool destinations it can hold a string of any content,
// including the empty one.
func TestScanNilStringToStringDest(t *testing.T) {
	config := NewScanConfig()

	for _, source := range config.NilStrings {
		dest := "prefilled"
		err := Scan(reflect.ValueOf(&dest).Elem(), source, config)
		assert.NoError(t, err, "Scan(string, %q)", source)
		assert.Equal(t, source, dest, "Scan(string, %q) keeps the source string", source)
	}
}

// TestScanNonNilStringUnparsable checks that a non-nil source string which
// can't be parsed is reported as an error without modifying the destination,
// instead of being scanned as "no value" like a nil source string.
func TestScanNonNilStringUnparsable(t *testing.T) {
	config := NewScanConfig()

	i := 666
	assert.Error(t, Scan(reflect.ValueOf(&i).Elem(), "abc", config))
	assert.Equal(t, 666, i, "destination is not modified by a failed scan")

	var u uint64
	assert.Error(t, Scan(reflect.ValueOf(&u).Elem(), "-1", config))

	var f float64
	assert.Error(t, Scan(reflect.ValueOf(&f).Elem(), "abc", config))

	var b bool
	assert.Error(t, Scan(reflect.ValueOf(&b).Elem(), "maybe", config))

	var ti time.Time
	assert.Error(t, Scan(reflect.ValueOf(&ti).Elem(), "not a date", config))

	var d time.Duration
	assert.Error(t, Scan(reflect.ValueOf(&d).Elem(), "1 hour", config))
}

// TestScanPointerDestUsesTypeScanner checks that a pointer destination uses
// the scanner registered for the pointed to type. Without it a *time.Time
// would fall through to the RFC3339-only UnmarshalText method of time.Time
// and a *time.Duration to the plain nanoseconds of its int64 kind,
// so pointer fields would parse fewer strings than their value equivalents.
func TestScanPointerDestUsesTypeScanner(t *testing.T) {
	config := NewScanConfig()

	// "2026-01-02" is one of the configured TimeFormats
	// but not the RFC3339 format required by time.Time.UnmarshalText
	t.Run("*time.Time", func(t *testing.T) {
		var dest *time.Time
		err := Scan(reflect.ValueOf(&dest).Elem(), "2026-01-02", config)
		assert.NoError(t, err)
		if assert.NotNil(t, dest) {
			assert.Equal(t, time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC), *dest)
		}
	})

	t.Run("*time.Duration", func(t *testing.T) {
		var dest *time.Duration
		err := Scan(reflect.ValueOf(&dest).Elem(), "1h30m", config)
		assert.NoError(t, err)
		if assert.NotNil(t, dest) {
			assert.Equal(t, 90*time.Minute, *dest)
		}
	})

	// A scanner registered by the user for a type has to be used
	// for a pointer to that type as well.
	t.Run("*customScannedType", func(t *testing.T) {
		config := NewScanConfig()
		config.SetTypeScanner(
			reflect.TypeFor[customScannedType](),
			ScannerFunc(func(dest reflect.Value, str string, config *ScanConfig) error {
				dest.Set(reflect.ValueOf(customScannedType{Scanned: str}))
				return nil
			}),
		)
		var dest *customScannedType
		err := Scan(reflect.ValueOf(&dest).Elem(), "source", config)
		assert.NoError(t, err)
		if assert.NotNil(t, dest) {
			assert.Equal(t, customScannedType{Scanned: "source"}, *dest)
		}
	})

	// A nil source string still means a nil pointer
	// and must not allocate one to scan into.
	t.Run("nil string", func(t *testing.T) {
		for _, source := range config.NilStrings {
			dest := new(time.Time)
			err := Scan(reflect.ValueOf(&dest).Elem(), source, config)
			assert.NoError(t, err, "Scan(*time.Time, %q)", source)
			assert.Nil(t, dest, "Scan(*time.Time, %q) sets the pointer to nil", source)
		}
	})
}

// customScannedType has no scanning method of its own,
// it can only be scanned by a Scanner registered at a ScanConfig.
type customScannedType struct {
	Scanned string
}

// TestScanInvalidDest checks that an invalid or non-settable destination
// is reported as an error instead of panicking in the reflect package.
// Scan uses dest.Interface(), dest.Addr() and the Set methods, which all
// panic for such a destination.
func TestScanInvalidDest(t *testing.T) {
	config := NewScanConfig()

	assert.Error(t, Scan(reflect.Value{}, "666", config), "invalid dest value")

	assert.Error(t, Scan(reflect.ValueOf(666), "666", config), "non-addressable dest value")

	var strct struct{ unexported int }
	unexportedField := reflect.ValueOf(&strct).Elem().Field(0)
	assert.Error(t, Scan(unexportedField, "666", config), "unexported struct field")

	// Those errors are wrapped with the call parameters
	// like every other error returned by Scan
	err := Scan(reflect.Value{}, "666", config)
	assert.ErrorContains(t, err, "strfmt.Scan(", "error is wrapped with the call parameters")

	i := 666
	ptrStrct := struct{ unexported *int }{unexported: &i}
	unexportedPtrField := reflect.ValueOf(&ptrStrct).Elem().Field(0)
	assert.Error(t, Scan(unexportedPtrField, "666", config), "unexported struct field of pointer type")
}

// TestScanNilConfig checks that a nil ScanConfig is reported as an error
// instead of panicking on the first config field access. Scan reads the
// config before every branch, including in the recursion for a pointer
// destination, so a nil one must be rejected up front.
func TestScanNilConfig(t *testing.T) {
	var i int
	assert.Error(t, Scan(reflect.ValueOf(&i).Elem(), "666", nil), "nil ScanConfig")
	assert.Zero(t, i, "destination is not modified")

	var ptr *int
	assert.Error(t, Scan(reflect.ValueOf(&ptr).Elem(), "666", nil), "nil ScanConfig for a pointer destination")
	assert.Nil(t, ptr, "destination is not allocated")
}

// TestScanNonSettablePointerDest checks that a pointer which is not
// settable itself, like the reflect.Value of a &variable expression,
// is scanned by scanning the settable value it points to. Only the
// pointer itself can't be set, which a nil pointer would need.
func TestScanNonSettablePointerDest(t *testing.T) {
	config := NewScanConfig()
	config.StrictEmptyStringParsing = true

	var i int
	assert.NoError(t, Scan(reflect.ValueOf(&i), "666", config))
	assert.Equal(t, 666, i)

	// The scanner registered for the pointed to type is still used
	var ti time.Time
	assert.NoError(t, Scan(reflect.ValueOf(&ti), "2026-01-02", config))
	assert.Equal(t, time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC), ti)

	// The pointed to value decides what a nil source string means:
	// a pointer can be set to nil, an int is not optional
	ptr := new(int)
	assert.NoError(t, Scan(reflect.ValueOf(&ptr), "", config))
	assert.Nil(t, ptr, "the pointed to pointer is set to nil")

	assert.Error(t, Scan(reflect.ValueOf(&i), "", config), "nil source string into a non-optional int")
	assert.Equal(t, 666, i, "the destination is not modified")

	// A nil pointer can't be allocated without setting the pointer
	var nilPtr *int
	assert.Error(t, Scan(reflect.ValueOf(nilPtr), "666", config), "non-settable nil pointer")
}

// TestScanDurationNanoseconds checks that a duration without a unit
// is scanned as nanoseconds, which is how a time.Duration is represented
// as its underlying int64 kind and how it's marshalled as JSON.
func TestScanDurationNanoseconds(t *testing.T) {
	config := NewScanConfig()

	var d time.Duration
	assert.NoError(t, Scan(reflect.ValueOf(&d).Elem(), "5000", config))
	assert.Equal(t, 5*time.Microsecond, d)

	// A pointer destination has to parse the same strings as its value equivalent
	var ptr *time.Duration
	assert.NoError(t, Scan(reflect.ValueOf(&ptr).Elem(), "5000", config))
	if assert.NotNil(t, ptr) {
		assert.Equal(t, 5*time.Microsecond, *ptr)
	}
}

// TestScanNumberOverflow checks that a number which doesn't fit into the
// destination type is reported as an error, because the Set methods of
// reflect.Value would silently truncate it to a different value.
func TestScanNumberOverflow(t *testing.T) {
	config := NewScanConfig()

	i8 := int8(1)
	assert.Error(t, Scan(reflect.ValueOf(&i8).Elem(), "300", config), "Scan(int8, 300)")
	assert.Equal(t, int8(1), i8, "destination is not modified by a failed scan")

	u8 := uint8(1)
	assert.Error(t, Scan(reflect.ValueOf(&u8).Elem(), "300", config), "Scan(uint8, 300)")
	assert.Equal(t, uint8(1), u8, "destination is not modified by a failed scan")

	f32 := float32(1)
	assert.Error(t, Scan(reflect.ValueOf(&f32).Elem(), "1e40", config), "Scan(float32, 1e40)")
	assert.Equal(t, float32(1), f32, "destination is not modified by a failed scan")

	// The limits of the destination type are still scanned
	assert.NoError(t, Scan(reflect.ValueOf(&i8).Elem(), "-128", config))
	assert.Equal(t, int8(-128), i8)
	assert.NoError(t, Scan(reflect.ValueOf(&u8).Elem(), "255", config))
	assert.Equal(t, uint8(255), u8)
}

// TestScanPointerDestAllocatedAfterScan checks that a pointer destination is
// allocated only after the pointed to value was scanned successfully.
// A caller that logs a scan error and continues must be able to tell a
// column with no value apart from one with an unparsable value.
func TestScanPointerDestAllocatedAfterScan(t *testing.T) {
	config := NewScanConfig()

	var i *int
	assert.Error(t, Scan(reflect.ValueOf(&i).Elem(), "abc", config))
	assert.Nil(t, i, "failed scan doesn't allocate the pointer")

	var ti *time.Time
	assert.Error(t, Scan(reflect.ValueOf(&ti).Elem(), "not a date", config))
	assert.Nil(t, ti, "failed scan doesn't allocate the pointer")
}

// partiallyScanned assigns only its Scanned field, leaving Kept untouched,
// like any scanning method that doesn't own every field of its type.
type partiallyScanned struct {
	Scanned string
	Kept    int
}

func (p *partiallyScanned) ScanString(source string, validate bool) error {
	if source == "refuse" {
		return errors.New("refused by ScanString")
	}
	p.Scanned = source
	return nil
}

// TestScanPointerDestKeepsAllocation checks that an already allocated pointer
// destination is scanned in place instead of being replaced by a newly
// allocated value. Replacing it would silently drop every field that the
// scanning method doesn't assign, and would leave anyone else holding the
// original pointer with a stale value.
func TestScanPointerDestKeepsAllocation(t *testing.T) {
	config := NewScanConfig()

	original := &partiallyScanned{Scanned: "old", Kept: 666}
	dest := original
	assert.NoError(t, Scan(reflect.ValueOf(&dest).Elem(), "new", config))
	assert.Same(t, original, dest, "the already allocated pointer is kept")
	assert.Equal(t, "new", dest.Scanned, "the scanned field is assigned")
	assert.Equal(t, 666, dest.Kept, "a field the scanning method doesn't assign is kept")

	// A nil pointer is still allocated
	var nilDest *partiallyScanned
	assert.NoError(t, Scan(reflect.ValueOf(&nilDest).Elem(), "new", config))
	if assert.NotNil(t, nilDest) {
		assert.Equal(t, "new", nilDest.Scanned)
	}
}

// TestScanFailedScanDoesNotModifyDest checks the documented guarantee that a
// failed scan never modifies dest. A scanning method can assign before it
// returns an error and ValidateFunc runs after the kind conversions have
// assigned, so without restoring the previous value a caller scanning a row
// of cells would be left with a half written field on the first bad cell.
func TestScanFailedScanDoesNotModifyDest(t *testing.T) {
	config := NewScanConfig()

	// A scanning method that assigns before returning an error
	strct := partiallyScanned{Scanned: "old", Kept: 666}
	assert.Error(t, Scan(reflect.ValueOf(&strct).Elem(), "refuse", config))
	assert.Equal(t, partiallyScanned{Scanned: "old", Kept: 666}, strct, "unchanged after a failed scan")

	// The same through an already allocated pointer
	ptr := &partiallyScanned{Scanned: "old", Kept: 666}
	assert.Error(t, Scan(reflect.ValueOf(&ptr).Elem(), "refuse", config))
	assert.Equal(t, partiallyScanned{Scanned: "old", Kept: 666}, *ptr, "pointed to value unchanged")

	// A registered type scanner is restored the same way
	ti := time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)
	assert.Error(t, Scan(reflect.ValueOf(&ti).Elem(), "not a date", config))
	assert.Equal(t, time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC), ti, "unchanged after a failed scan")

	// ValidateFunc rejects a value that the kind conversion already assigned
	validating := NewScanConfig()
	validating.ValidateFunc = func(any) error { return errors.New("invalid value") }
	i := 42
	assert.Error(t, Scan(reflect.ValueOf(&i).Elem(), "7", validating))
	assert.Equal(t, 42, i, "unchanged after failed validation")

	// Including through a pointer that is not settable itself
	j := 42
	assert.Error(t, Scan(reflect.ValueOf(&j), "7", validating))
	assert.Equal(t, 42, j, "unchanged after failed validation through a non-settable pointer")
}

// TestScanMultiLevelPointerDest checks that every level of indirection of a
// pointer destination is allocated and scanned, because FormatValue formats
// multi-level pointers as well, so both sides agree on what a **T field means.
func TestScanMultiLevelPointerDest(t *testing.T) {
	config := NewScanConfig()

	var ti **time.Time
	err := Scan(reflect.ValueOf(&ti).Elem(), "2026-01-02", config)
	assert.NoError(t, err)
	if assert.NotNil(t, ti) && assert.NotNil(t, *ti) {
		assert.Equal(t, time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC), **ti)
	}

	var i **int
	assert.Error(t, Scan(reflect.ValueOf(&i).Elem(), "abc", config))
	assert.Nil(t, i, "failed scan doesn't allocate any level of indirection")

	// A nil source string sets the outermost pointer to nil
	nilable := new(*int)
	assert.NoError(t, Scan(reflect.ValueOf(&nilable).Elem(), "", config))
	assert.Nil(t, nilable)
}

// TestScanNullableTime checks that nullable.Time is scanned with the
// TimeFormats of the config like time.Time is, instead of falling back to
// its UnmarshalText method which only accepts the RFC3339 format.
func TestScanNullableTime(t *testing.T) {
	config := NewScanConfig()

	for _, source := range []string{"2026-01-02", "2026-01-02 15:04:05", "2026-01-02T15:04:05Z"} {
		want, ok := config.ParseTime(source)
		if !ok {
			t.Fatalf("test needs %q to be one of the configured TimeFormats", source)
		}
		var dest nullable.Time
		err := Scan(reflect.ValueOf(&dest).Elem(), source, config)
		assert.NoError(t, err, "Scan(nullable.Time, %q)", source)
		assert.Equal(t, want, dest.Get(), "Scan(nullable.Time, %q)", source)
	}
}

// TestScanWhitespaceOnlySource checks that a source string that only holds
// whitespace is recognized as nil, because an "empty" cell of a CSV file or
// a spreadsheet export often holds a space or a tab. ScanConfig.IsNil itself
// keeps comparing the source string unchanged against the configured
// NilStrings, the whitespace is trimmed by Scan before asking it.
func TestScanWhitespaceOnlySource(t *testing.T) {
	config := NewScanConfig()
	config.StrictEmptyStringParsing = true

	assert.False(t, config.IsNil(" "), `IsNil(" ") compares the unchanged string`)

	for _, source := range []string{" ", "  ", "\t", " NULL ", "\tnull\t"} {
		ptr := new(int)
		assert.NoError(t, Scan(reflect.ValueOf(&ptr).Elem(), source, config), "Scan(*int, %q)", source)
		assert.Nil(t, ptr, "Scan(*int, %q) sets the pointer to nil", source)

		nullTime := nullable.TimeFrom(time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC))
		assert.NoError(t, Scan(reflect.ValueOf(&nullTime).Elem(), source, config), "Scan(nullable.Time, %q)", source)
		assert.True(t, nullTime.IsNull(), "Scan(nullable.Time, %q) sets null", source)

		var i int
		assert.Error(t, Scan(reflect.ValueOf(&i).Elem(), source, config), "Scan(int, %q)", source)
	}
}

// TestScanPointerTypeScanner checks that a scanner registered for a pointer
// type is used for a pointer destination instead of being bypassed by
// scanning the pointed to value, which would make the scanned result depend
// on how the destination reflect.Value was obtained.
func TestScanPointerTypeScanner(t *testing.T) {
	config := NewScanConfig()
	config.SetTypeScanner(
		reflect.TypeFor[*customScannedType](),
		ScannerFunc(func(dest reflect.Value, str string, config *ScanConfig) error {
			dest.Set(reflect.ValueOf(&customScannedType{Scanned: "by the pointer scanner"}))
			return nil
		}),
	)

	var dest *customScannedType
	err := Scan(reflect.ValueOf(&dest).Elem(), "source", config)
	assert.NoError(t, err)
	if assert.NotNil(t, dest) {
		assert.Equal(t, "by the pointer scanner", dest.Scanned)
	}

	// A non-settable pointer can't be passed to the scanner registered
	// for its type, so it's an error instead of silently using a
	// different scanner for the pointed to value
	value := customScannedType{Scanned: "unchanged"}
	assert.Error(t, Scan(reflect.ValueOf(&value), "source", config), "non-settable pointer with a scanner for its type")
	assert.Equal(t, "unchanged", value.Scanned)
}

// scanNilStringZeroes scans every string that config declares as nil into a
// destination prefilled with nonZero and requires the zero value to be
// assigned, the non-strict counterpart of scanNilStringFails.
func scanNilStringZeroes[T comparable](t *testing.T, config *ScanConfig, nonZero T) {
	t.Helper()

	var zero T
	if nonZero == zero {
		t.Fatalf("test needs a non-zero value of type %T to prove that the zero value was assigned", zero)
	}
	for _, source := range config.NilStrings {
		dest := nonZero
		err := Scan(reflect.ValueOf(&dest).Elem(), source, config)
		if err != nil {
			t.Errorf("Scan(%T, %q) returned error: %v", dest, source, err)
			continue
		}
		if dest != zero {
			t.Errorf("Scan(%T, %q) assigned %v, want zero value %v", dest, source, dest, zero)
		}
	}
}

// TestScanNilStringReachesTypeScanner pins which destinations still reach a
// registered Scanner with a nil source string, because Scan resolves one
// before dispatching. A string kind does, since only it can hold the source
// string itself, while a kind that can hold nil never does. Getting this
// wrong silently changes what a user's own Scanner is asked to scan.
func TestScanNilStringReachesTypeScanner(t *testing.T) {
	type stringKind string

	newConfig := func(called *string) *ScanConfig {
		config := NewScanConfig()
		scanner := ScannerFunc(func(dest reflect.Value, str string, config *ScanConfig) error {
			*called = str
			dest.SetString("scanned:" + str)
			return nil
		})
		config.SetTypeScanner(reflect.TypeFor[stringKind](), scanner)
		return config
	}

	// A string kind reaches its scanner with the nil source string
	var called string
	config := newConfig(&called)
	var dest stringKind = "prefilled"
	assert.NoError(t, Scan(reflect.ValueOf(&dest).Elem(), "", config))
	assert.Equal(t, "", called, "scanner is called with the nil source string")
	assert.Equal(t, stringKind("scanned:"), dest)

	// A pointer to it does not: the pointer is set to nil instead
	called = "NOT CALLED"
	config = newConfig(&called)
	ptr := new(stringKind)
	assert.NoError(t, Scan(reflect.ValueOf(&ptr).Elem(), "", config))
	assert.Nil(t, ptr, "a nilable kind is set to nil without asking the scanner")
	assert.Equal(t, "NOT CALLED", called, "scanner is not called for a nilable destination")
}

// TestScanNilStringWithoutStrictEmptyStringParsing checks that a nil source
// string assigns the zero value of a destination type that can't represent
// "no value" when StrictEmptyStringParsing is disabled, so that a row with
// empty cells stays scannable into a column of any type.
func TestScanNilStringWithoutStrictEmptyStringParsing(t *testing.T) {
	config := NewScanConfig() // StrictEmptyStringParsing is off by default

	t.Run("bool", func(t *testing.T) { scanNilStringZeroes(t, config, true) })
	t.Run("int", func(t *testing.T) { scanNilStringZeroes(t, config, 666) })
	t.Run("float64", func(t *testing.T) { scanNilStringZeroes(t, config, 6.66) })
	t.Run("struct", func(t *testing.T) { scanNilStringZeroes(t, config, struct{ Int int }{666}) })

	// Types with a scanning method of their own are zeroed as well,
	// instead of reporting the nil source string as a parse error
	t.Run("time.Time", func(t *testing.T) {
		scanNilStringZeroes(t, config, time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC))
	})
	t.Run("time.Duration", func(t *testing.T) { scanNilStringZeroes(t, config, time.Hour) })
	t.Run("money.Amount", func(t *testing.T) { scanNilStringZeroes(t, config, money.Amount(6.66)) })
	t.Run("date.Date", func(t *testing.T) { scanNilStringZeroes(t, config, date.Date("2026-01-02")) })

	// Destinations that can represent "no value" are not affected
	t.Run("optional destinations", func(t *testing.T) {
		for _, source := range config.NilStrings {
			ptr := new(int)
			assert.NoError(t, Scan(reflect.ValueOf(&ptr).Elem(), source, config), "Scan(*int, %q)", source)
			assert.Nil(t, ptr, "Scan(*int, %q)", source)

			nullTime := nullable.TimeFrom(time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC))
			assert.NoError(t, Scan(reflect.ValueOf(&nullTime).Elem(), source, config), "Scan(nullable.Time, %q)", source)
			assert.True(t, nullTime.IsNull(), "Scan(nullable.Time, %q) sets null", source)

			str := "prefilled"
			assert.NoError(t, Scan(reflect.ValueOf(&str).Elem(), source, config), "Scan(string, %q)", source)
			assert.Equal(t, source, str, "Scan(string, %q) keeps the source string", source)
		}
	})
}

// TestScanNilStringNotValidated checks that a destination set to nil, null or
// its zero value for a nil source string is not passed to ValidateFunc,
// because the absence of a value is not a value to validate. Validating it
// would make an empty cell fail for every type whose zero value is invalid,
// which is what both nil string modes exist to avoid.
// A string destination is the exception: it is assigned the source string
// instead of an absent value, so it is validated like any scanned value.
func TestScanNilStringNotValidated(t *testing.T) {
	config := NewScanConfig()
	config.ValidateFunc = func(any) error { return errors.New("invalid value") }

	i := 666
	assert.NoError(t, Scan(reflect.ValueOf(&i).Elem(), "", config), "zero value assigned for a nil source string")
	assert.Zero(t, i)

	nullTime := nullable.TimeFrom(time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC))
	assert.NoError(t, Scan(reflect.ValueOf(&nullTime).Elem(), "", config), "null assigned for a nil source string")
	assert.True(t, nullTime.IsNull())

	ptr := new(int)
	assert.NoError(t, Scan(reflect.ValueOf(&ptr).Elem(), "", config), "nil assigned for a nil source string")
	assert.Nil(t, ptr)

	// A string destination holds the source string of a nil string
	// instead of an absent value, so it is validated
	str := "prefilled"
	assert.Error(t, Scan(reflect.ValueOf(&str).Elem(), "NULL", config), "string destination is validated")

	// A scanned value is still validated
	assert.Error(t, Scan(reflect.ValueOf(&i).Elem(), "666", config))
}
