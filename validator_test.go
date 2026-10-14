package types

import (
	"errors"
	"fmt"
	"math"
	"slices"
	"strings"
	"testing"
)

// Test types for DeepValidate tests

// validString is always valid
type validString string

func (v validString) Valid() bool {
	return true
}

// invalidString is always invalid
type invalidString string

func (v invalidString) Valid() bool {
	return false
}

// nonEmptyString validates that it's not empty
type nonEmptyString string

func (v nonEmptyString) Validate() error {
	if v == "" {
		return errors.New("string cannot be empty")
	}
	return nil
}

// positiveInt validates that it's greater than 0
type positiveInt int

func (v positiveInt) Validate() error {
	if v <= 0 {
		return fmt.Errorf("value %d must be positive", v)
	}
	return nil
}

// rangeValue validates that it's within a range
type rangeValue struct {
	Value int
	Min   int
	Max   int
}

func (r rangeValue) Valid() bool {
	return r.Value >= r.Min && r.Value <= r.Max
}

func TestDeepValidate(t *testing.T) {
	tests := []struct {
		name      string
		input     any
		wantErrs  int
		wantPaths []string // Expected error path substrings
	}{
		{
			name:     "nil value",
			input:    nil,
			wantErrs: 0,
		},
		{
			name:     "valid simple type",
			input:    validString("test"),
			wantErrs: 0,
		},
		{
			name:     "invalid simple type",
			input:    invalidString("test"),
			wantErrs: 1,
		},
		{
			name:     "valid ValidatErr type",
			input:    nonEmptyString("test"),
			wantErrs: 0,
		},
		{
			name:     "invalid ValidatErr type",
			input:    nonEmptyString(""),
			wantErrs: 1,
		},
		{
			name:     "valid positiveInt",
			input:    positiveInt(10),
			wantErrs: 0,
		},
		{
			name:     "invalid positiveInt",
			input:    positiveInt(-5),
			wantErrs: 1,
		},
		{
			name: "struct with valid fields",
			input: struct {
				Name  validString
				Value positiveInt
			}{
				Name:  "test",
				Value: 10,
			},
			wantErrs: 0,
		},
		{
			name: "struct with invalid fields",
			input: struct {
				Name  invalidString
				Value positiveInt
			}{
				Name:  "test",
				Value: -5,
			},
			wantErrs:  2,
			wantPaths: []string{"struct field Name", "struct field Value"},
		},
		{
			name: "struct with mixed valid/invalid fields",
			input: struct {
				Valid   validString
				Invalid invalidString
			}{
				Valid:   "valid",
				Invalid: "invalid",
			},
			wantErrs:  1,
			wantPaths: []string{"struct field Invalid"},
		},
		{
			name:     "slice with all valid elements",
			input:    []validString{"a", "b", "c"},
			wantErrs: 0,
		},
		{
			name:     "slice with invalid elements",
			input:    []invalidString{"a", "b", "c"},
			wantErrs: 3,
		},
		{
			name:     "slice with mixed valid/invalid",
			input:    []any{validString("valid"), invalidString("invalid")},
			wantErrs: 1,
		},
		{
			name:     "array with all valid elements",
			input:    [3]validString{"a", "b", "c"},
			wantErrs: 0,
		},
		{
			name:     "array with invalid elements",
			input:    [3]invalidString{"a", "b", "c"},
			wantErrs: 3,
		},
		{
			name: "map with valid values",
			input: map[string]validString{
				"key1": "value1",
				"key2": "value2",
			},
			wantErrs: 0,
		},
		{
			name: "map with invalid values",
			input: map[string]invalidString{
				"key1": "value1",
				"key2": "value2",
			},
			wantErrs: 2,
		},
		{
			name: "map with mixed valid/invalid values",
			input: map[string]any{
				"valid":   validString("ok"),
				"invalid": invalidString("bad"),
			},
			wantErrs:  1,
			wantPaths: []string{"map value"},
		},
		{
			name: "nested struct",
			input: struct {
				Outer struct {
					Inner invalidString
				}
			}{
				Outer: struct {
					Inner invalidString
				}{
					Inner: "test",
				},
			},
			wantErrs:  1,
			wantPaths: []string{"struct field Outer", "struct field Inner"},
		},
		{
			name: "deeply nested struct",
			input: struct {
				Level1 struct {
					Level2 struct {
						Level3 invalidString
					}
				}
			}{
				Level1: struct {
					Level2 struct {
						Level3 invalidString
					}
				}{
					Level2: struct {
						Level3 invalidString
					}{
						Level3: "test",
					},
				},
			},
			wantErrs:  1,
			wantPaths: []string{"struct field Level1", "struct field Level2", "struct field Level3"},
		},
		{
			name: "slice of structs",
			input: []struct {
				Value invalidString
			}{
				{Value: "a"},
				{Value: "b"},
			},
			wantErrs:  2,
			wantPaths: []string{"element [0]", "element [1]"},
		},
		{
			name: "map of structs",
			input: map[string]struct {
				Value invalidString
			}{
				"first":  {Value: "a"},
				"second": {Value: "b"},
			},
			wantErrs: 2,
		},
		{
			name: "pointer to valid value",
			input: func() *validString {
				v := validString("test")
				return &v
			}(),
			wantErrs: 0,
		},
		{
			name: "pointer to invalid value",
			input: func() *invalidString {
				v := invalidString("test")
				return &v
			}(),
			wantErrs: 1,
		},
		{
			name:     "nil pointer",
			input:    (*invalidString)(nil),
			wantErrs: 0, // nil pointers should not cause validation errors
		},
		{
			name: "struct with nil pointer field",
			input: struct {
				Value *invalidString
			}{
				Value: nil,
			},
			wantErrs: 0,
		},
		{
			name: "struct with non-nil pointer to invalid value",
			input: struct {
				Value *invalidString
			}{
				Value: func() *invalidString {
					v := invalidString("test")
					return &v
				}(),
			},
			wantErrs:  1,
			wantPaths: []string{"struct field Value"},
		},
		{
			name: "complex nested structure",
			input: struct {
				Users []struct {
					Name  nonEmptyString
					Age   positiveInt
					Tags  map[string]invalidString
					Valid validString
				}
			}{
				Users: []struct {
					Name  nonEmptyString
					Age   positiveInt
					Tags  map[string]invalidString
					Valid validString
				}{
					{
						Name:  "Alice",
						Age:   30,
						Tags:  map[string]invalidString{"tag1": "value1"},
						Valid: "ok",
					},
					{
						Name:  "", // Invalid: empty string
						Age:   -5, // Invalid: negative
						Tags:  map[string]invalidString{"tag2": "value2"},
						Valid: "ok",
					},
				},
			},
			wantErrs:  4, // 2 invalid tags + 1 empty name + 1 negative age
			wantPaths: []string{"struct field Users"},
		},
		{
			name:     "empty slice",
			input:    []invalidString{},
			wantErrs: 0,
		},
		{
			name:     "empty map",
			input:    map[string]invalidString{},
			wantErrs: 0,
		},
		{
			name: "empty struct",
			input: struct {
			}{},
			wantErrs: 0,
		},
		{
			name: "struct with non-validatable fields",
			input: struct {
				Name  string
				Value int
				Flag  bool
			}{
				Name:  "test",
				Value: 42,
				Flag:  true,
			},
			wantErrs: 0,
		},
		{
			name: "rangeValue valid",
			input: rangeValue{
				Value: 5,
				Min:   0,
				Max:   10,
			},
			wantErrs: 0,
		},
		{
			name: "rangeValue invalid",
			input: rangeValue{
				Value: 15,
				Min:   0,
				Max:   10,
			},
			wantErrs: 1,
		},
		{
			name: "slice of slices",
			input: [][]invalidString{
				{"a", "b"},
				{"c", "d"},
			},
			wantErrs: 4, // All 4 strings are invalid
		},
		{
			name: "map with int keys",
			input: map[int]invalidString{
				1: "a",
				2: "b",
				3: "c",
			},
			wantErrs: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := DeepValidate(tt.input)

			if len(errs) != tt.wantErrs {
				t.Errorf("DeepValidate() returned %d errors, want %d", len(errs), tt.wantErrs)
				for i, err := range errs {
					t.Logf("  Error %d: %v", i, err)
				}
			}

			// Check that expected path fragments are in the errors
			if tt.wantPaths != nil {
				for _, wantPath := range tt.wantPaths {
					found := false
					for _, err := range errs {
						if strings.Contains(err.Error(), wantPath) {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Expected error path containing %q not found in errors", wantPath)
						for i, err := range errs {
							t.Logf("  Error %d: %v", i, err)
						}
					}
				}
			}
		})
	}
}

func TestDeepValidate_ErrorPaths(t *testing.T) {
	// Test that error paths are correctly constructed
	input := struct {
		Outer struct {
			Inner struct {
				Value invalidString
			}
		}
	}{
		Outer: struct {
			Inner struct {
				Value invalidString
			}
		}{
			Inner: struct {
				Value invalidString
			}{
				Value: "test",
			},
		},
	}

	errs := DeepValidate(input)
	if len(errs) != 1 {
		t.Fatalf("Expected 1 error, got %d", len(errs))
	}

	errMsg := errs[0].Error()
	expectedParts := []string{"struct field Outer", "struct field Inner", "struct field Value"}
	for _, part := range expectedParts {
		if !strings.Contains(errMsg, part) {
			t.Errorf("Error message %q does not contain expected part %q", errMsg, part)
		}
	}
}

func TestDeepValidate_MapKeySorting(t *testing.T) {
	// Test that map validation is deterministic (keys are sorted)
	input := map[string]invalidString{
		"z": "last",
		"a": "first",
		"m": "middle",
	}

	errs1 := DeepValidate(input)
	errs2 := DeepValidate(input)

	if len(errs1) != len(errs2) {
		t.Fatalf("Inconsistent error count: %d vs %d", len(errs1), len(errs2))
	}

	// Errors should be in the same order due to key sorting
	for i := range errs1 {
		if errs1[i].Error() != errs2[i].Error() {
			t.Errorf("Error order differs:\n  Run 1: %v\n  Run 2: %v", errs1[i], errs2[i])
		}
	}
}

// Coverage for the bug-fix cases: unexported fields, multi-level pointers,
// interface fields, cross-kind unwrapping, and the path/element wording.

func TestDeepValidate_SkipsUnexportedFields(t *testing.T) {
	type withPrivate struct {
		Pub     invalidString
		private invalidString
	}
	errs := DeepValidate(withPrivate{Pub: "x", private: "y"})
	if len(errs) != 1 {
		t.Fatalf("expected 1 error (only the exported field), got %d: %v", len(errs), errs)
	}
	if !strings.Contains(errs[0].Error(), "Pub") {
		t.Errorf("expected error to reference Pub, got %v", errs[0])
	}
}

func TestDeepValidate_OnlyUnexportedFields(t *testing.T) {
	type allPrivate struct {
		a invalidString
		b invalidString
	}
	errs := DeepValidate(allPrivate{a: "x", b: "y"})
	if len(errs) != 0 {
		t.Errorf("expected 0 errors (all fields unexported), got %d: %v", len(errs), errs)
	}
}

func TestDeepValidate_MultiLevelPointers(t *testing.T) {
	v := invalidString("x")
	pv := &v
	ppv := &pv
	pppv := &ppv
	errs := DeepValidate(pppv)
	if len(errs) != 1 {
		t.Errorf("expected 1 error for ***invalidString, got %d: %v", len(errs), errs)
	}
}

func TestDeepValidate_NilPointerChain(t *testing.T) {
	var ppv **invalidString
	errs := DeepValidate(ppv)
	if len(errs) != 0 {
		t.Errorf("expected 0 errors for nil **invalidString, got %d: %v", len(errs), errs)
	}
}

func TestDeepValidate_InterfaceFieldWithValidator(t *testing.T) {
	type wrap struct {
		V any
	}
	errs := DeepValidate(wrap{V: invalidString("x")})
	if len(errs) != 1 {
		t.Fatalf("expected 1 error for any-wrapped validator, got %d: %v", len(errs), errs)
	}
	if !strings.Contains(errs[0].Error(), "field V") {
		t.Errorf("expected path to include field V, got %v", errs[0])
	}
}

func TestDeepValidate_InterfaceFieldRecursesIntoStruct(t *testing.T) {
	type leaf struct {
		Bad invalidString
	}
	type wrap struct {
		V any
	}
	errs := DeepValidate(wrap{V: leaf{Bad: "x"}})
	if len(errs) != 1 {
		t.Fatalf("expected 1 error reaching through interface into struct, got %d: %v", len(errs), errs)
	}
	msg := errs[0].Error()
	for _, part := range []string{"field V", "field Bad"} {
		if !strings.Contains(msg, part) {
			t.Errorf("expected path to contain %q, got %v", part, msg)
		}
	}
}

func TestDeepValidate_NilInterfaceField(t *testing.T) {
	type wrap struct {
		V any
	}
	errs := DeepValidate(wrap{V: nil})
	if len(errs) != 0 {
		t.Errorf("expected 0 errors for nil interface, got %d: %v", len(errs), errs)
	}
}

func TestDeepValidate_InterfaceHoldingNilPointer(t *testing.T) {
	type wrap struct {
		V any
	}
	var p *invalidString
	errs := DeepValidate(wrap{V: p})
	if len(errs) != 0 {
		t.Errorf("expected 0 errors for interface holding typed-nil pointer, got %d: %v", len(errs), errs)
	}
}

func TestDeepValidate_SliceOfInterfaces(t *testing.T) {
	errs := DeepValidate([]any{
		invalidString("a"),
		validString("ok"),
		invalidString("b"),
	})
	if len(errs) != 2 {
		t.Errorf("expected 2 errors in []any, got %d: %v", len(errs), errs)
	}
}

func TestDeepValidate_MapWithInterfaceValues(t *testing.T) {
	errs := DeepValidate(map[string]any{
		"a": invalidString("x"),
		"b": validString("ok"),
		"c": invalidString("y"),
	})
	if len(errs) != 2 {
		t.Errorf("expected 2 errors in map[string]any, got %d: %v", len(errs), errs)
	}
}

func TestDeepValidate_ElementWordingFixed(t *testing.T) {
	// Regression: error paths spelled "elememt"; should now be "element".
	errs := DeepValidate([]invalidString{"a"})
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errs))
	}
	msg := errs[0].Error()
	if strings.Contains(msg, "elememt") {
		t.Errorf("error still contains typo 'elememt': %v", msg)
	}
	if !strings.Contains(msg, "element [0]") {
		t.Errorf("expected 'element [0]' in path, got %v", msg)
	}
}

func TestDeepValidate_PointerInsideSlice(t *testing.T) {
	v1 := invalidString("a")
	v2 := invalidString("b")
	errs := DeepValidate([]*invalidString{&v1, nil, &v2})
	if len(errs) != 2 {
		t.Errorf("expected 2 errors (nil skipped), got %d: %v", len(errs), errs)
	}
}

func TestDeepValidate_PointerToStructWithValidator(t *testing.T) {
	type s struct {
		Bad invalidString
	}
	val := &s{Bad: "x"}
	errs := DeepValidate(val)
	if len(errs) != 1 {
		t.Errorf("expected 1 error for *struct field, got %d: %v", len(errs), errs)
	}
}

func TestDeepValidate_TopLevelValidatErr(t *testing.T) {
	errs := DeepValidate(nonEmptyString(""))
	if len(errs) != 1 {
		t.Fatalf("expected 1 error from ValidatErr, got %d", len(errs))
	}
	if !strings.Contains(errs[0].Error(), "empty") {
		t.Errorf("expected ValidatErr's own message, got %v", errs[0])
	}
}

func TestDeepValidate_DeterministicIntKeyOrder(t *testing.T) {
	input := map[int]invalidString{3: "c", 1: "a", 2: "b"}
	var paths []string
	for _, err := range DeepValidate(input) {
		paths = append(paths, err.Error())
	}
	if len(paths) != 3 {
		t.Fatalf("expected 3 errors, got %d", len(paths))
	}
	// Keys sorted ascending: 1, 2, 3
	if !strings.Contains(paths[0], "[1]") || !strings.Contains(paths[1], "[2]") || !strings.Contains(paths[2], "[3]") {
		t.Errorf("expected int keys in ascending order, got:\n  %s\n  %s\n  %s", paths[0], paths[1], paths[2])
	}
}

func TestDeepValidate_NonNilableLeafTypes(t *testing.T) {
	// Channels/funcs/etc. are reachable but not validatable; should not crash.
	type s struct {
		Ch  chan int
		Fn  func()
		Map map[string]int
	}
	errs := DeepValidate(s{Ch: make(chan int), Fn: func() {}, Map: map[string]int{"x": 1}})
	if len(errs) != 0 {
		t.Errorf("expected 0 errors for non-validatable leaf kinds, got %d: %v", len(errs), errs)
	}
}

func TestDeepValidate_JoinErrors(t *testing.T) {
	// Test the suggested usage: errors.Join(DeepValidate(v)...)
	input := struct {
		Field1 invalidString
		Field2 invalidString
	}{
		Field1: "a",
		Field2: "b",
	}

	errs := DeepValidate(input)
	if len(errs) != 2 {
		t.Fatalf("Expected 2 errors, got %d", len(errs))
	}

	joinedErr := errors.Join(errs...)
	if joinedErr == nil {
		t.Fatal("Expected non-nil joined error")
	}

	errMsg := joinedErr.Error()
	if !strings.Contains(errMsg, "Field1") || !strings.Contains(errMsg, "Field2") {
		t.Errorf("Joined error missing field names: %v", errMsg)
	}
}

// errNegative is the specific error of signedInt.Validate.
var errNegative = errors.New("negative number")

// signedInt implements ValidatErr with its own specific error.
type signedInt int

func (v signedInt) Validate() error {
	if v < 0 {
		return errNegative
	}
	return nil
}

// strictValidLenientValidate implements both ValidatErr and Validator
// with Valid deliberately disagreeing with Validate, to pin which of
// the two methods decides.
type strictValidLenientValidate int

func (v strictValidLenientValidate) Valid() bool { return false }

func (v strictValidLenientValidate) Validate() error {
	if v < 0 {
		return errNegative
	}
	return nil
}

// validateTestCases are shared by TestValidate and TestTryValidate because
// both functions must report the same error for the same value.
//
// A non-nil error is what marks a value as invalid, there is no sentinel
// error to check with errors.Is. The error of a ValidatErr is the better
// message and has to reach the caller unchanged, not wrapped or replaced.
// A Validator's Valid can't say why a value is invalid, so the error at
// least has to name the type and value to be diagnosable.
var validateTestCases = []struct {
	name              string
	value             any
	wantErr           string // expected error message prefix, empty for no error
	wantUnchanged     error  // if set, the exact error value expected, not a wrapper of it
	wantIsValidatable bool
}{
	{name: "nil", value: nil, wantIsValidatable: false},
	{name: "not validatable", value: 1, wantIsValidatable: false},
	{name: "valid Validator", value: validString("x"), wantIsValidatable: true},
	{name: "invalid Validator names its type and value", value: invalidString("x"), wantErr: `invalid types.invalidString value: "x"`, wantIsValidatable: true},
	// %#v of a pointer is its address, so only the prefix is checked
	{name: "invalid pointer Validator names its pointer type", value: new(invalidString("x")), wantErr: "invalid *types.invalidString value: (*types.invalidString)(0x", wantIsValidatable: true},
	{name: "valid ValidatErr", value: signedInt(1), wantIsValidatable: true},
	{name: "invalid ValidatErr returns its error unchanged", value: signedInt(-1), wantErr: "negative number", wantUnchanged: errNegative, wantIsValidatable: true},
	// Validate wins over Valid, so a type can offer a lenient Validate
	// next to a strict Valid without types.Validate rejecting its values.
	{name: "ValidatErr takes precedence over valid=false Validator", value: strictValidLenientValidate(1), wantIsValidatable: true},
	{name: "ValidatErr error takes precedence over Validator error", value: strictValidLenientValidate(-1), wantErr: "negative number", wantUnchanged: errNegative, wantIsValidatable: true},
	// A nil pointer has no value that a value receiver method could be
	// called on, Go panics with a nil pointer dereference instead.
	// Like DeepValidate skips nil pointers, the absent value is not
	// validated, but its type implements the interface so it is validatable.
	{name: "nil pointer to value receiver Validator is not validated", value: (*invalidString)(nil), wantIsValidatable: true},
	{name: "nil pointer to value receiver ValidatErr is not validated", value: (*signedInt)(nil), wantIsValidatable: true},
	{name: "nil pointer to value receiver ValidatErr and Validator is not validated", value: (*strictValidLenientValidate)(nil), wantIsValidatable: true},
	// A pointer receiver method can handle nil itself, so it is still called
	{name: "nil pointer receiver ValidatErr is called with nil", value: (*nilCheckingValidatErr)(nil), wantErr: "nil pointer", wantUnchanged: errNilPointer, wantIsValidatable: true},
	{name: "nil pointer receiver Validator is called with nil", value: (*nilCheckingValidator)(nil), wantErr: "invalid *types.nilCheckingValidator value: (*types.nilCheckingValidator)(nil)", wantIsValidatable: true},
	// A nil func has no function that its Valid or Validate method
	// could call, Go panics with a nil func call instead.
	// Like DeepValidate skips nil funcs, the absent func is not
	// validated, but its type implements the interface so it is validatable.
	{name: "nil ValidatorFunc is not called", value: ValidatorFunc(nil), wantIsValidatable: true},
	{name: "nil ValidatErrFunc is not called", value: ValidatErrFunc(nil), wantIsValidatable: true},
	// Only the nil func is skipped, a func is still called
	{name: "non-nil ValidatorFunc is called", value: ValidatorFunc(func() bool { return false }), wantErr: "invalid types.ValidatorFunc value: (types.ValidatorFunc)(0x", wantIsValidatable: true},
	{name: "non-nil ValidatErrFunc is called", value: ValidatErrFunc(func() error { return errNegative }), wantErr: "negative number", wantUnchanged: errNegative, wantIsValidatable: true},
}

// errNilPointer is the specific error of nilCheckingValidatErr.Validate.
var errNilPointer = errors.New("nil pointer")

// nilCheckingValidatErr implements ValidatErr with a pointer receiver
// that handles being called with a nil pointer.
type nilCheckingValidatErr struct{}

func (p *nilCheckingValidatErr) Validate() error {
	if p == nil {
		return errNilPointer
	}
	return nil
}

// nilCheckingValidator implements Validator with a pointer receiver
// that reports a nil pointer as invalid.
type nilCheckingValidator struct{}

func (p *nilCheckingValidator) Valid() bool { return p != nil }

// errString returns the message of err or an empty string for nil.
func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

// errHasPrefix reports if err is nil for an empty prefix,
// or if err is not nil and its message starts with prefix.
func errHasPrefix(err error, prefix string) bool {
	if prefix == "" {
		return err == nil
	}
	return err != nil && strings.HasPrefix(err.Error(), prefix)
}

func TestValidate(t *testing.T) {
	for _, tt := range validateTestCases {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.value)
			if !errHasPrefix(err, tt.wantErr) {
				t.Errorf("Validate(%#v) = %q, want prefix %q", tt.value, errString(err), tt.wantErr)
			}
			if tt.wantUnchanged != nil && err != tt.wantUnchanged {
				t.Errorf("Validate(%#v) = %#v, want the ValidatErr's error %#v unchanged", tt.value, err, tt.wantUnchanged)
			}
		})
	}
}

func TestTryValidate(t *testing.T) {
	for _, tt := range validateTestCases {
		t.Run(tt.name, func(t *testing.T) {
			err, isValidatable := TryValidate(tt.value)
			if !errHasPrefix(err, tt.wantErr) {
				t.Errorf("TryValidate(%#v) error = %q, want prefix %q", tt.value, errString(err), tt.wantErr)
			}
			if tt.wantUnchanged != nil && err != tt.wantUnchanged {
				t.Errorf("TryValidate(%#v) error = %#v, want the ValidatErr's error %#v unchanged", tt.value, err, tt.wantUnchanged)
			}
			if isValidatable != tt.wantIsValidatable {
				t.Errorf("TryValidate(%#v) isValidatable = %t, want %t", tt.value, isValidatable, tt.wantIsValidatable)
			}
		})
	}
}

// TestValidatorAsValidatErr checks that adapting a Validator to a ValidatErr
// gives the same error as Validate gives for the Validator itself,
// so wrapping a Validator does not lose the type name of the message.
func TestValidatorAsValidatErr(t *testing.T) {
	if err := (ValidatorAsValidatErr{validString("x")}).Validate(); err != nil {
		t.Errorf("Validate() of valid Validator = %v, want nil", err)
	}
	err := ValidatorAsValidatErr{invalidString("x")}.Validate()
	if got, want := errString(err), errString(Validate(invalidString("x"))); got != want {
		t.Errorf("Validate() of invalid Validator = %q, want %q", got, want)
	}
}

// TestValidators checks that Validators and CombinedValidator are valid only
// if every element is valid, using the Valid method of each element even if
// it also implements ValidatErr. An absent validator has nothing to validate,
// like Validate returns nil for it, so a nil element, a nil func, and a nil
// pointer to a value receiver Validator are skipped instead of panicking.
func TestValidators(t *testing.T) {
	tests := []struct {
		name       string
		validators []Validator
		want       bool
	}{
		{name: "nil", validators: nil, want: true},
		{name: "all valid", validators: []Validator{validString("a"), StaticValidator(true)}, want: true},
		{name: "one invalid", validators: []Validator{validString("a"), invalidString("b")}, want: false},
		// Valid is used, not the lenient Validate of the same value
		{name: "Valid of a ValidatErr is used", validators: []Validator{strictValidLenientValidate(1)}, want: false},
		{name: "nil elements are skipped", validators: []Validator{nil, ValidatorFunc(nil), (*invalidString)(nil), validString("a")}, want: true},
		{name: "nil elements don't hide an invalid one", validators: []Validator{nil, ValidatorFunc(nil), invalidString("b")}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name+"/Validators", func(t *testing.T) {
			if got := Validators(tt.validators).Valid(); got != tt.want {
				t.Errorf("Valid() = %t, want %t", got, tt.want)
			}
		})
		t.Run(tt.name+"/CombinedValidator", func(t *testing.T) {
			if got := CombinedValidator(tt.validators...).Valid(); got != tt.want {
				t.Errorf("Valid() = %t, want %t", got, tt.want)
			}
		})
	}
}

// TestValidatErrs checks that ValidatErrs and CombinedValidatErr report all
// invalid elements at once instead of only the first one, so that a caller
// can show every problem of a value together. Every element has to be
// validated, also the ones after an error, the message has to list the
// errors in element order without gaps for valid elements, and errors.Is
// has to find each joined error, including a single one that the join wraps.
func TestValidatErrs(t *testing.T) {
	errA := errors.New("error a")
	errB := errors.New("error b")
	var called []string
	validatErr := func(name string, err error) ValidatErr {
		return ValidatErrFunc(func() error {
			called = append(called, name)
			return err
		})
	}

	tests := []struct {
		name        string
		validatErrs []ValidatErr
		wantCalled  []string
		wantErrs    []error // joined in this order, none for a nil error
	}{
		{
			name:        "nil",
			validatErrs: nil,
		},
		{
			name:        "all valid",
			validatErrs: []ValidatErr{validatErr("1", nil), validatErr("2", nil)},
			wantCalled:  []string{"1", "2"},
		},
		{
			name:        "single error",
			validatErrs: []ValidatErr{validatErr("1", nil), validatErr("2", errA), validatErr("3", nil)},
			wantCalled:  []string{"1", "2", "3"},
			wantErrs:    []error{errA},
		},
		{
			name:        "all errors joined",
			validatErrs: []ValidatErr{validatErr("1", errA), validatErr("2", nil), validatErr("3", errB)},
			wantCalled:  []string{"1", "2", "3"},
			wantErrs:    []error{errA, errB},
		},
		{
			// An absent validator has nothing to validate,
			// like Validate returns nil for each of these values,
			// so it must neither panic nor hide the other errors.
			name: "nil elements are skipped",
			validatErrs: []ValidatErr{
				nil,
				validatErr("1", errA),
				ValidatErrFunc(nil),
				(*signedInt)(nil),
				validatErr("2", errB),
			},
			wantCalled: []string{"1", "2"},
			wantErrs:   []error{errA, errB},
		},
	}
	for _, tt := range tests {
		for _, combined := range []struct {
			name string
			v    ValidatErr
		}{
			{"ValidatErrs", ValidatErrs(tt.validatErrs)},
			{"CombinedValidatErr", CombinedValidatErr(tt.validatErrs...)},
		} {
			t.Run(tt.name+"/"+combined.name, func(t *testing.T) {
				called = nil
				err := combined.v.Validate()

				if !slices.Equal(called, tt.wantCalled) {
					t.Errorf("called validators %q, want %q", called, tt.wantCalled)
				}
				if len(tt.wantErrs) == 0 {
					if err != nil {
						t.Errorf("Validate() = %q, want nil", err)
					}
					return
				}
				if err == nil {
					t.Fatalf("Validate() = nil, want %q", tt.wantErrs)
				}
				for _, want := range tt.wantErrs {
					if !errors.Is(err, want) {
						t.Errorf("errors.Is(%q, %q) = false, want true", err, want)
					}
				}
				wantMsg := errors.Join(tt.wantErrs...).Error()
				if err.Error() != wantMsg {
					t.Errorf("Validate() message = %q, want %q", err, wantMsg)
				}
			})
		}
	}
}

// TestDeepValidate_PathWrapsOriginalError checks that the path prefix of
// DeepValidate keeps the error of the value's validator reachable:
// callers that classify errors with errors.Is or errors.As must still
// see the ValidatErr's own error behind the path.
func TestDeepValidate_PathWrapsOriginalError(t *testing.T) {
	input := struct {
		Number  signedInt
		Invalid invalidString
	}{
		Number:  -1,
		Invalid: "x",
	}
	errs := DeepValidate(input)
	if len(errs) != 2 {
		t.Fatalf("expected 2 errors, got %d: %v", len(errs), errs)
	}
	if got, want := errs[0].Error(), "struct field Number: negative number"; got != want {
		t.Errorf("DeepValidate error = %q, want %q", got, want)
	}
	if errors.Unwrap(errs[0]) != errNegative {
		t.Errorf("expected the path to wrap the ValidatErr's error unchanged, got %#v", errors.Unwrap(errs[0]))
	}
	if got, want := errs[1].Error(), `struct field Invalid: invalid types.invalidString value: "x"`; got != want {
		t.Errorf("DeepValidate error = %q, want %q", got, want)
	}
}

// ptrValidatErr implements ValidatErr only with a pointer receiver, like LenString.
type ptrValidatErr string

func (p *ptrValidatErr) Validate() error {
	if *p == "" {
		return errors.New("empty ptrValidatErr")
	}
	return nil
}

// ptrValidator implements Validator only with a pointer receiver.
type ptrValidator string

func (p *ptrValidator) Valid() bool { return *p != "" }

// mixedReceivers implements Validator with a value receiver
// and ValidatErr with a pointer receiver.
type mixedReceivers string

func (m mixedReceivers) Valid() bool { return false }

func (m *mixedReceivers) Validate() error {
	if *m == "" {
		return errors.New("empty mixedReceivers")
	}
	return nil
}

// errorStrings returns the messages of errs for comparing them in one go.
func errorStrings(errs []error) []string {
	strs := make([]string, len(errs))
	for i, err := range errs {
		strs[i] = err.Error()
	}
	return strs
}

// TestDeepValidate_PointerReceiverValidators checks that a validator with a
// pointer receiver is called for every non-pointer value of its type.
// Such values are struct fields, slice, array, and map elements, or the
// dynamic value of an interface, and a LenString field would otherwise
// never be validated at all, because its own fields are unexported.
// Values that are not addressable, like the fields of a struct passed by
// value or map values, have to be validated through a pointer to a copy.
func TestDeepValidate_PointerReceiverValidators(t *testing.T) {
	type fields struct {
		F ptrValidatErr
	}
	tests := []struct {
		name  string
		input any
		want  []string
	}{
		{"field of struct by value", fields{}, []string{"struct field F: empty ptrValidatErr"}},
		{"field of struct by pointer", &fields{}, []string{"struct field F: empty ptrValidatErr"}},
		{"pointer to top-level value", new(ptrValidatErr), []string{"empty ptrValidatErr"}},
		// Like Validate, a value passed by value is validated with the
		// methods of its own type, see TestDeepValidate_PointerReceiverCallingDeepValidate
		{"top-level value", ptrValidatErr(""), []string{}},
		{"slice element", []ptrValidatErr{"x", ""}, []string{"element [1]: empty ptrValidatErr"}},
		{"array element of struct by value", struct{ A [1]ptrValidatErr }{}, []string{"struct field A -> element [0]: empty ptrValidatErr"}},
		{"map value", map[string]ptrValidatErr{"k": ""}, []string{`map value ["k"]: empty ptrValidatErr`}},
		{"dynamic value of interface", struct{ I any }{I: ptrValidatErr("")}, []string{"struct field I: empty ptrValidatErr"}},
		{"valid value", fields{F: "x"}, []string{}},
		// Like Validate, a ValidatErr is preferred over a Validator, even
		// if only the pointer type implements ValidatErr
		{"pointer receiver ValidatErr preferred over Validator", struct{ M mixedReceivers }{M: "x"}, []string{}},
		{"pointer receiver ValidatErr error preferred", struct{ M mixedReceivers }{}, []string{"struct field M: empty mixedReceivers"}},
		{"LenString field", struct{ S LenString }{S: LenString{str: "abc", min: 5, max: 1}}, []string{`struct field S: minimum length 5 is greater than maximum length 1 of LenString "abc"`}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := errorStrings(DeepValidate(tt.input))
			if !slices.Equal(got, tt.want) {
				t.Errorf("DeepValidate() errors = %q, want %q", got, tt.want)
			}
		})
	}

	t.Run("pointer receiver Validator", func(t *testing.T) {
		errs := DeepValidate(struct{ V ptrValidator }{})
		// The error names the pointer type because that's what implements
		// Validator, %#v of the pointer is an address so only the prefix is checked
		want := "struct field V: invalid *types.ptrValidator value: (*types.ptrValidator)(0x"
		if len(errs) != 1 || !errHasPrefix(errs[0], want) {
			t.Errorf("DeepValidate() errors = %q, want one with prefix %q", errorStrings(errs), want)
		}
	})
}

// selfDeepValidating validates its fields with DeepValidate(*s) in a
// pointer receiver Validate, the usual way to combine DeepValidate
// with the own Validate method of a type.
type selfDeepValidating struct {
	Name nonEmptyString
}

func (s *selfDeepValidating) Validate() error {
	return errors.Join(DeepValidate(*s)...)
}

// TestDeepValidate_PointerReceiverCallingDeepValidate checks that a pointer
// receiver Validate calling DeepValidate on its dereferenced receiver
// terminates. If DeepValidate validated the value passed to it through a
// pointer to a copy, it would call that Validate method again for every
// copy, until the stack overflows and the process crashes unrecoverably.
func TestDeepValidate_PointerReceiverCallingDeepValidate(t *testing.T) {
	type parent struct {
		Child selfDeepValidating
	}
	// DeepValidate also descends into the fields of Child after calling
	// its Validate method, so only the error of the method is checked
	want := "struct field Child: struct field Name: string cannot be empty"
	if got := errorStrings(DeepValidate(&parent{})); !slices.Contains(got, want) {
		t.Errorf("DeepValidate(&parent{}) errors = %q, want them to contain %q", got, want)
	}
	if got := errorStrings(DeepValidate(parent{})); !slices.Contains(got, want) {
		t.Errorf("DeepValidate(parent{}) errors = %q, want them to contain %q", got, want)
	}
	if err := (&selfDeepValidating{}).Validate(); err == nil || err.Error() != "struct field Name: string cannot be empty" {
		t.Errorf("Validate() = %v, want struct field Name: string cannot be empty", err)
	}
}

// TestDeepValidate_PointerToInterfaceWithPointerReceiver checks that only
// the value passed to DeepValidate itself is validated without a copy.
// A value behind a passed pointer to an interface is not addressable,
// it must still be validated through a pointer to a copy.
func TestDeepValidate_PointerToInterfaceWithPointerReceiver(t *testing.T) {
	var v any = ptrValidatErr("")
	want := []string{"empty ptrValidatErr"}
	if got := errorStrings(DeepValidate(&v)); !slices.Equal(got, want) {
		t.Errorf("DeepValidate(&v) errors = %q, want %q", got, want)
	}
}

// TestDeepValidate_NaNMapKey checks that the value of a NaN map key is
// validated like every other map value. A NaN key is not equal to itself,
// so looking the value up with the key can't find it.
func TestDeepValidate_NaNMapKey(t *testing.T) {
	errs := DeepValidate(map[float64]nonEmptyString{math.NaN(): "", 1: ""})
	if len(errs) != 2 {
		t.Errorf("DeepValidate() errors = %q, want 2 errors", errorStrings(errs))
	}
	errs = DeepValidate(map[any]nonEmptyString{math.NaN(): "", "x": ""})
	if len(errs) != 2 {
		t.Errorf("DeepValidate() errors = %q, want 2 errors", errorStrings(errs))
	}
}

// countingValidator counts the calls of its value receiver Valid method.
type countingValidator struct{ calls *int }

func (c countingValidator) Valid() bool {
	*c.calls++
	return true
}

// countingPtrValidatErr counts the calls of its pointer receiver Validate method.
type countingPtrValidatErr struct{ calls int }

func (c *countingPtrValidatErr) Validate() error {
	c.calls++
	return nil
}

// TestDeepValidate_ValidValueValidatedOnce checks that a valid value reached
// through a pointer is validated only once. The pointer type has the methods
// of its element type, so validating again after dereferencing would call
// the same method twice and repeat the side effects of stateful validators.
func TestDeepValidate_ValidValueValidatedOnce(t *testing.T) {
	var valueCalls int
	value := countingValidator{calls: &valueCalls}
	DeepValidate(&value)
	if valueCalls != 1 {
		t.Errorf("value receiver Valid called %d times, want 1", valueCalls)
	}

	var ptr countingPtrValidatErr
	DeepValidate(&ptr)
	if ptr.calls != 1 {
		t.Errorf("pointer receiver Validate called %d times, want 1", ptr.calls)
	}
}

// TestDeepValidate_PointerReceiverDoesNotModifyMapValue checks that
// validating a map value through a pointer to a copy can't write to the map,
// which a pointer receiver Validate that normalizes in place would otherwise do.
func TestDeepValidate_PointerReceiverDoesNotModifyMapValue(t *testing.T) {
	m := map[string]LenString{"k": {str: "abc", min: 0, max: 5}}
	if errs := DeepValidate(m); len(errs) != 0 {
		t.Fatalf("expected no errors, got %v", errs)
	}
	if m["k"] != (LenString{str: "abc", min: 0, max: 5}) {
		t.Errorf("map value changed to %#v", m["k"])
	}
}

type embeddedFields struct {
	Invalid    invalidString
	unexported invalidString
}

type invalidStrings []invalidString

// TestDeepValidate_UnexportedEmbeddedStruct checks that the exported fields
// of an unexported embedded struct are validated: they are promoted to the
// embedding struct and accessible from outside of the package like its own
// fields, so skipping them because the embedded field itself is unexported
// left them unvalidated.
func TestDeepValidate_UnexportedEmbeddedStruct(t *testing.T) {
	tests := []struct {
		name  string
		input any
		want  []string
	}{
		{
			name:  "embedded struct",
			input: struct{ embeddedFields }{embeddedFields{Invalid: "x", unexported: "x"}},
			want:  []string{`struct field embeddedFields -> struct field Invalid: invalid types.invalidString value: "x"`},
		},
		{
			name:  "embedded struct pointer",
			input: struct{ *embeddedFields }{&embeddedFields{Invalid: "x", unexported: "x"}},
			want:  []string{`struct field embeddedFields -> struct field Invalid: invalid types.invalidString value: "x"`},
		},
		{
			name:  "nil embedded struct pointer",
			input: struct{ *embeddedFields }{},
			want:  []string{},
		},
		{
			// Only struct fields are promoted, the elements of an unexported
			// embedded slice are not accessible from outside of the package
			name:  "embedded non-struct type",
			input: struct{ invalidStrings }{invalidStrings{"x"}},
			want:  []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := errorStrings(DeepValidate(tt.input))
			if !slices.Equal(got, tt.want) {
				t.Errorf("DeepValidate() errors = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestDeepValidate_MixedTypeMapKeysDeterministic checks that the errors of a
// map with keys of different dynamic types are returned in the same order
// every time. The keys are ordered by type name first, so the float64 key
// comes before the int keys, which come before the string keys.
func TestDeepValidate_MixedTypeMapKeysDeterministic(t *testing.T) {
	want := []string{
		"map value [1.5]: string cannot be empty",
		"map value [1]: string cannot be empty",
		"map value [2]: string cannot be empty",
		`map value ["a"]: string cannot be empty`,
		`map value ["b"]: string cannot be empty`,
	}
	// Map iteration order is random, so one run could pass by chance
	for range 50 {
		input := map[any]nonEmptyString{2: "", "b": "", 1.5: "", "a": "", 1: ""}
		got := errorStrings(DeepValidate(input))
		if !slices.Equal(got, want) {
			t.Fatalf("DeepValidate() errors = %q, want %q", got, want)
		}
	}
}

// TestDeepValidate_NilFunc checks that a nil func whose type implements
// Validator or ValidatErr is skipped like a nil pointer: its Valid or
// Validate method would call the nil func and panic.
func TestDeepValidate_NilFunc(t *testing.T) {
	input := struct {
		Validator  ValidatorFunc
		ValidatErr ValidatErrFunc
		Interface  any
	}{
		Interface: ValidatErrFunc(nil),
	}
	if errs := DeepValidate(input); len(errs) != 0 {
		t.Errorf("expected no errors for nil funcs, got %v", errs)
	}

	// A non-nil func is still called,
	// %#v of a func is an address so only the prefix is checked
	input.Validator = func() bool { return false }
	errs := DeepValidate(input)
	want := "struct field Validator: invalid types.ValidatorFunc value: (types.ValidatorFunc)(0x"
	if len(errs) != 1 || !errHasPrefix(errs[0], want) {
		t.Errorf("DeepValidate() errors = %q, want one with prefix %q", errorStrings(errs), want)
	}
}

type cycleNode struct {
	Value invalidString
	Next  *cycleNode
}

// invalidCycleNode is always invalid itself,
// so revisiting it on a cycle would report it again.
type invalidCycleNode struct {
	Next *invalidCycleNode
}

func (*invalidCycleNode) Validate() error { return errors.New("invalid node") }

type innerStruct struct {
	Value invalidString
}

type holder struct {
	Inner innerStruct
	Ptr   *innerStruct
}

// TestDeepValidate_Cycles checks that cyclic data is validated once per
// value on the cycle instead of recursing until the stack overflows,
// which can't be recovered from and crashes the process.
// Data shared by multiple paths without a cycle must still be validated
// and reported under each of its paths.
func TestDeepValidate_Cycles(t *testing.T) {
	t.Run("pointer to itself", func(t *testing.T) {
		n := &cycleNode{Value: "x"}
		n.Next = n
		got := errorStrings(DeepValidate(n))
		want := []string{`struct field Value: invalid types.invalidString value: "x"`}
		if !slices.Equal(got, want) {
			t.Errorf("DeepValidate() errors = %q, want %q", got, want)
		}
	})

	t.Run("invalid value on a cycle is reported once", func(t *testing.T) {
		n := &invalidCycleNode{}
		n.Next = n
		got := errorStrings(DeepValidate(n))
		want := []string{"invalid node"}
		if !slices.Equal(got, want) {
			t.Errorf("DeepValidate() errors = %q, want %q", got, want)
		}
	})

	t.Run("two pointers to each other", func(t *testing.T) {
		a := &cycleNode{Value: "a"}
		a.Next = &cycleNode{Value: "b", Next: a}
		got := errorStrings(DeepValidate(a))
		want := []string{
			`struct field Value: invalid types.invalidString value: "a"`,
			`struct field Next -> struct field Value: invalid types.invalidString value: "b"`,
		}
		if !slices.Equal(got, want) {
			t.Errorf("DeepValidate() errors = %q, want %q", got, want)
		}
	})

	t.Run("map containing itself", func(t *testing.T) {
		m := map[string]any{"v": invalidString("x")}
		m["self"] = m
		got := errorStrings(DeepValidate(m))
		want := []string{`map value ["v"]: invalid types.invalidString value: "x"`}
		if !slices.Equal(got, want) {
			t.Errorf("DeepValidate() errors = %q, want %q", got, want)
		}
	})

	t.Run("slice containing itself", func(t *testing.T) {
		s := []any{invalidString("x"), nil}
		s[1] = s
		got := errorStrings(DeepValidate(s))
		want := []string{`element [0]: invalid types.invalidString value: "x"`}
		if !slices.Equal(got, want) {
			t.Errorf("DeepValidate() errors = %q, want %q", got, want)
		}
	})

	t.Run("shorter slice of the same array is no cycle", func(t *testing.T) {
		s := []any{invalidString("x"), nil}
		s[1] = s[:1]
		got := errorStrings(DeepValidate(s))
		want := []string{
			`element [0]: invalid types.invalidString value: "x"`,
			`element [1] -> element [0]: invalid types.invalidString value: "x"`,
		}
		if !slices.Equal(got, want) {
			t.Errorf("DeepValidate() errors = %q, want %q", got, want)
		}
	})

	t.Run("pointer shared by multiple paths is no cycle", func(t *testing.T) {
		p := new(nonEmptyString(""))
		got := errorStrings(DeepValidate([]*nonEmptyString{p, p}))
		want := []string{
			"element [0]: string cannot be empty",
			"element [1]: string cannot be empty",
		}
		if !slices.Equal(got, want) {
			t.Errorf("DeepValidate() errors = %q, want %q", got, want)
		}
	})

	t.Run("pointer to own first field is no cycle", func(t *testing.T) {
		// h and h.Ptr have the same address but different types
		h := &holder{Inner: innerStruct{Value: "x"}}
		h.Ptr = &h.Inner
		got := errorStrings(DeepValidate(h))
		want := []string{
			`struct field Inner -> struct field Value: invalid types.invalidString value: "x"`,
			`struct field Ptr -> struct field Value: invalid types.invalidString value: "x"`,
		}
		if !slices.Equal(got, want) {
			t.Errorf("DeepValidate() errors = %q, want %q", got, want)
		}
	})
}

type validatedByValue string

func (v validatedByValue) Validate() error {
	if v == "" {
		return errors.New("empty")
	}
	return nil
}

type embedsNilPointer struct {
	*validatedByValue
	Name string
}

type declaresOwnValidate struct {
	*validatedByValue
}

func (declaresOwnValidate) Validate() error { return errors.New("own") }

// TestValidate_NilEmbeddedPointer protects against a panic: a struct that
// embeds a nil *T, where T implements Validate with a value receiver,
// promotes a method that dereferences the nil pointer. Such a value is
// not validated, like a nil pointer, but a method the struct declares
// itself still runs.
func TestValidate_NilEmbeddedPointer(t *testing.T) {
	v := embedsNilPointer{Name: "x"}
	if err := Validate(v); err != nil {
		t.Errorf("Validate(%#v) = %v, want nil", v, err)
	}
	if err := Validate(&v); err != nil {
		t.Errorf("Validate(&%#v) = %v, want nil", v, err)
	}
	if errs := DeepValidate(v); len(errs) != 0 {
		t.Errorf("DeepValidate(%#v) = %v, want none", v, errs)
	}
	// Promoted through two embedded fields, by a chain of generated wrappers
	if err := Validate(struct{ embedsNilPointer }{v}); err != nil {
		t.Errorf("Validate of a struct embedding %#v = %v, want nil", v, err)
	}

	empty := validatedByValue("")
	if err := Validate(embedsNilPointer{validatedByValue: &empty}); err == nil {
		t.Error("a non-nil embedded value must be validated")
	}
	if err := Validate(declaresOwnValidate{}); err == nil || err.Error() != "own" {
		t.Errorf("a declared Validate must run, got %v", err)
	}
}

type derefsNilField struct {
	*validatedByValue
	Field *validatedByValue
}

func (d derefsNilField) Validate() error { return d.Field.Validate() }

func (d derefsNilField) Valid() bool { return *d.Field != "" }

// TestValidate_NilEmbeddedPointerKeepsMethodPanics protects against hiding
// a bug as a valid value: only the panic of the method wrapper that the
// compiler generates to promote a method from a nil embedded pointer is
// recovered. A nil pointer dereference in a method that a struct declares
// itself is a bug of that method and must still panic, also if the struct
// happens to embed a nil pointer, instead of reporting the value as valid.
func TestValidate_NilEmbeddedPointerKeepsMethodPanics(t *testing.T) {
	v := derefsNilField{}
	calls := map[string]func(){
		"Validate":              func() { _ = Validate(v) },
		"TryValidate":           func() { _, _ = TryValidate(v) },
		"DeepValidate":          func() { _ = DeepValidate(v) },
		"Validators":            func() { _ = Validators{v}.Valid() },
		"ValidatorAsValidatErr": func() { _ = ValidatorAsValidatErr{v}.Validate() },
	}
	for name, call := range calls {
		t.Run(name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Error("the nil pointer dereference of the declared method was recovered")
				}
			}()
			call()
		})
	}
}

// TestValidatorAsValidatErr_SkipsNil checks that wrapping a Validator
// that Validate skips, because its Valid method can't be called,
// returns nil like Validate and Validators.Valid instead of panicking.
func TestValidatorAsValidatErr_SkipsNil(t *testing.T) {
	for _, validator := range []Validator{
		nil,
		ValidatorFunc(nil),
		(*invalidString)(nil),
		struct{ *invalidString }{},
	} {
		if err := (ValidatorAsValidatErr{validator}).Validate(); err != nil {
			t.Errorf("ValidatorAsValidatErr{%#v}.Validate() = %v, want nil", validator, err)
		}
		if err := CombinedValidatErr(ValidatorAsValidatErr{validator}).Validate(); err != nil {
			t.Errorf("CombinedValidatErr(ValidatorAsValidatErr{%#v}).Validate() = %v, want nil", validator, err)
		}
	}
}
