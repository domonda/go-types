package types

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestReflectCompare(t *testing.T) {
	tests := []struct {
		a    reflect.Value
		b    reflect.Value
		want int
	}{
		{reflect.ValueOf(0), reflect.ValueOf(1), -1},
		{reflect.ValueOf(0), reflect.ValueOf(0), 0},
		{reflect.ValueOf(1), reflect.ValueOf(0), +1},
		{reflect.ValueOf(0.0), reflect.ValueOf(1.0), -1},
		{reflect.ValueOf(0.0), reflect.ValueOf(0.0), 0},
		{reflect.ValueOf(1.0), reflect.ValueOf(0.0), +1},
		{reflect.ValueOf("0"), reflect.ValueOf("1"), -1},
		{reflect.ValueOf("0"), reflect.ValueOf("0"), 0},
		{reflect.ValueOf("1"), reflect.ValueOf("0"), +1},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%#v_%#v", tt.a.Interface(), tt.b.Interface()), func(t *testing.T) {
			if got := ReflectCompare(tt.a, tt.b); got != tt.want {
				t.Errorf("ReflectCompare(%#v, %#v) = %d, want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

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
			wantPaths: []string{"elememt [0]", "elememt [1]"},
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
