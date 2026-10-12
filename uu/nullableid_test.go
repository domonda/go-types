package uu

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/domonda/go-pretty"
	"github.com/invopop/jsonschema"
)

func TestNullableIDFromString(t *testing.T) {
	u := ID{0x6b, 0xa7, 0xb8, 0x10, 0x9d, 0xad, 0x11, 0xd1, 0x80, 0xb4, 0x00, 0xc0, 0x4f, 0xd4, 0x30, 0xc8}

	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "hex without dashes (32 chars)",
			input: "6ba7b8109dad11d180b400c04fd430c8",
		},
		{
			name:  "standard dashed format (36 chars)",
			input: "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
		},
		{
			name:  "base64 URL encoding (22 chars)",
			input: u.Base64(),
		},
		{
			name:  "braces with dashed",
			input: "{6ba7b810-9dad-11d1-80b4-00c04fd430c8}",
		},
		{
			name:  "quoted dashed",
			input: `"6ba7b810-9dad-11d1-80b4-00c04fd430c8"`,
		},
		{
			name:  "URN format",
			input: "urn:uuid:6ba7b810-9dad-11d1-80b4-00c04fd430c8",
		},
		{
			name:  "uppercase hex without dashes",
			input: "6BA7B8109DAD11D180B400C04FD430C8",
		},
		{
			name:  "uppercase dashed",
			input: "6BA7B810-9DAD-11D1-80B4-00C04FD430C8",
		},
		{
			name:  "mixed case hex",
			input: "6Ba7B8109dAd11D180b400C04fD430c8",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, err := NullableIDFromString(tt.input)
			require.NoError(t, err, "parsing should succeed")
			require.Equal(t, NullableID(u), parsed, "parsed UUID should equal expected")
		})
	}

	// Test Nil UUID is interpreted as NULL
	t.Run("nil UUID becomes null", func(t *testing.T) {
		parsed, err := NullableIDFromString("00000000-0000-0000-0000-000000000000")
		require.NoError(t, err, "parsing should succeed")
		require.Equal(t, IDNull, parsed, "Nil UUID should be interpreted as NULL")
		require.True(t, parsed.IsNull(), "should be null")
	})

	// Test error case
	t.Run("empty string", func(t *testing.T) {
		_, err := NullableIDFromString("")
		require.Error(t, err, "should return error for empty string")
	})
}

func TestNullableIDFromStringOrNull(t *testing.T) {
	u := ID{0x6b, 0xa7, 0xb8, 0x10, 0x9d, 0xad, 0x11, 0xd1, 0x80, 0xb4, 0x00, 0xc0, 0x4f, 0xd4, 0x30, 0xc8}

	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "hex without dashes (32 chars)",
			input: "6ba7b8109dad11d180b400c04fd430c8",
		},
		{
			name:  "standard dashed format (36 chars)",
			input: "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
		},
		{
			name:  "base64 URL encoding (22 chars)",
			input: u.Base64(),
		},
		{
			name:  "braces with dashed",
			input: "{6ba7b810-9dad-11d1-80b4-00c04fd430c8}",
		},
		{
			name:  "quoted dashed",
			input: `"6ba7b810-9dad-11d1-80b4-00c04fd430c8"`,
		},
		{
			name:  "URN format",
			input: "urn:uuid:6ba7b810-9dad-11d1-80b4-00c04fd430c8",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed := NullableIDFromStringOrNull(tt.input)
			require.Equal(t, NullableID(u), parsed, "parsed UUID should equal expected")
		})
	}

	// Test error returns null
	t.Run("empty string returns null", func(t *testing.T) {
		parsed := NullableIDFromStringOrNull("")
		require.Equal(t, IDNull, parsed, "should return IDNull on error")
		require.True(t, parsed.IsNull(), "should be null")
	})

	t.Run("invalid string returns null", func(t *testing.T) {
		parsed := NullableIDFromStringOrNull("invalid-uuid")
		require.Equal(t, IDNull, parsed, "should return IDNull on error")
		require.True(t, parsed.IsNull(), "should be null")
	})
}

func TestNullableIDFromBytes(t *testing.T) {
	u := ID{0x6b, 0xa7, 0xb8, 0x10, 0x9d, 0xad, 0x11, 0xd1, 0x80, 0xb4, 0x00, 0xc0, 0x4f, 0xd4, 0x30, 0xc8}

	tests := []struct {
		name  string
		input []byte
	}{
		{
			name:  "binary format (16 bytes)",
			input: []byte{0x6b, 0xa7, 0xb8, 0x10, 0x9d, 0xad, 0x11, 0xd1, 0x80, 0xb4, 0x00, 0xc0, 0x4f, 0xd4, 0x30, 0xc8},
		},
		{
			name:  "standard dashed format (36 bytes)",
			input: []byte("6ba7b810-9dad-11d1-80b4-00c04fd430c8"),
		},
		{
			name:  "hex without dashes (32 bytes)",
			input: []byte("6ba7b8109dad11d180b400c04fd430c8"),
		},
		{
			name:  "base64 URL encoding (22 bytes)",
			input: []byte(u.Base64()),
		},
		{
			name:  "braces with dashed",
			input: []byte("{6ba7b810-9dad-11d1-80b4-00c04fd430c8}"),
		},
		{
			name:  "quoted dashed",
			input: []byte(`"6ba7b810-9dad-11d1-80b4-00c04fd430c8"`),
		},
		{
			name:  "URN format",
			input: []byte("urn:uuid:6ba7b810-9dad-11d1-80b4-00c04fd430c8"),
		},
		{
			name:  "uppercase hex without dashes",
			input: []byte("6BA7B8109DAD11D180B400C04FD430C8"),
		},
		{
			name:  "uppercase dashed",
			input: []byte("6BA7B810-9DAD-11D1-80B4-00C04FD430C8"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, err := NullableIDFromBytes(tt.input)
			require.NoError(t, err, "parsing should succeed")
			require.Equal(t, NullableID(u), parsed, "parsed UUID should equal expected")
		})
	}

	// Test Nil UUID is interpreted as NULL
	t.Run("nil UUID becomes null", func(t *testing.T) {
		parsed, err := NullableIDFromBytes([]byte("00000000-0000-0000-0000-000000000000"))
		require.NoError(t, err, "parsing should succeed")
		require.Equal(t, IDNull, parsed, "Nil UUID should be interpreted as NULL")
		require.True(t, parsed.IsNull(), "should be null")
	})

	// Test error cases
	t.Run("empty byte slice", func(t *testing.T) {
		_, err := NullableIDFromBytes([]byte{})
		require.Error(t, err, "should return error for empty slice")
	})

	t.Run("too short", func(t *testing.T) {
		_, err := NullableIDFromBytes([]byte("too-short"))
		require.Error(t, err, "should return error for too short input")
	})
}

func TestNullableIDFromPtr(t *testing.T) {
	u := ID{0x6b, 0xa7, 0xb8, 0x10, 0x9d, 0xad, 0x11, 0xd1, 0x80, 0xb4, 0x00, 0xc0, 0x4f, 0xd4, 0x30, 0xc8}

	t.Run("non-nil pointer", func(t *testing.T) {
		parsed := NullableIDFromPtr(&u)
		require.Equal(t, NullableID(u), parsed, "should return the dereferenced ID")
		require.False(t, parsed.IsNull(), "should not be null")
	})

	t.Run("nil pointer", func(t *testing.T) {
		parsed := NullableIDFromPtr(nil)
		require.Equal(t, IDNull, parsed, "should return IDNull for nil pointer")
		require.True(t, parsed.IsNull(), "should be null")
	})

	t.Run("pointer to Nil UUID becomes null", func(t *testing.T) {
		nilUUID := IDNil
		parsed := NullableIDFromPtr(&nilUUID)
		require.Equal(t, IDNull, parsed, "pointer to Nil UUID should become null")
		require.True(t, parsed.IsNull(), "should be null")
	})
}

func TestNullableIDValueNil(t *testing.T) {
	u := NullableID{}

	val, err := u.Value()
	if err != nil {
		t.Errorf("Error getting UUID value: %s", err)
	}

	if val != nil {
		t.Errorf("Wrong value returned, should be nil: %s", val)
	}
}

func TestNullableIDScanValid(t *testing.T) {
	u := ID{0x6b, 0xa7, 0xb8, 0x10, 0x9d, 0xad, 0x11, 0xd1, 0x80, 0xb4, 0x00, 0xc0, 0x4f, 0xd4, 0x30, 0xc8}
	s1 := "6ba7b810-9dad-11d1-80b4-00c04fd430c8"

	u1 := NullableID{}
	err := u1.Scan(s1)
	if err != nil {
		t.Errorf("Error unmarshaling NullableID: %s", err)
	}

	if !u1.Valid() {
		t.Errorf("NullableID should be valid")
	}

	if u != u1.Get() {
		t.Errorf("UUIDs should be equal: %s and %s", u, u1.Get())
	}
}

func TestNullableIDScanNil(t *testing.T) {
	u := NullableID{0x6b, 0xa7, 0xb8, 0x10, 0x9d, 0xad, 0x11, 0xd1, 0x80, 0xb4, 0x00, 0xc0, 0x4f, 0xd4, 0x30, 0xc8}

	err := u.Scan(nil)
	if err != nil {
		t.Errorf("Error unmarshaling NullableID: %s", err)
	}

	if !u.Valid() {
		t.Errorf("NullableID should be valid")
	}

	if !u.IsNull() {
		t.Errorf("NullableID value should be equal to Nil: %v", u)
	}
}

func TestNullableID_MarshalUnmarshalJSON(t *testing.T) {
	u := NullableID{0x6b, 0xa7, 0xb8, 0x10, 0x9d, 0xad, 0x11, 0xd1, 0x80, 0xb4, 0x00, 0xc0, 0x4f, 0xd4, 0x30, 0xc8}
	var u2 NullableID

	data, err := json.Marshal(&u)
	if err != nil {
		t.Errorf("Error JSON marshaling NullableID: %s", err)
	}
	err = json.Unmarshal(data, &u2)
	if err != nil {
		t.Errorf("Error JSON unmarshaling NullableID: %s", err)
	}
	if u2 != u {
		t.Errorf("JSON marshalling and unmarshalling produced a different UUID")
	}

	u.SetNull()

	data, err = json.Marshal(&u)
	if err != nil {
		t.Errorf("Error JSON marshaling NullableID: %s", err)
	}
	err = json.Unmarshal(data, &u2)
	if err != nil {
		t.Errorf("Error JSON unmarshaling NullableID: %s", err)
	}
	if u2 != u {
		t.Errorf("JSON marshalling and unmarshalling produced a different UUID")
	}
}

func TestNullableID_MarshalJSON(t *testing.T) {
	var testStruct struct {
		U ID         `json:"u"`
		N NullableID `json:"n"`
	}
	data, err := json.Marshal(&testStruct)
	if err != nil {
		t.Errorf("Error JSON marshaling: %s", err)
	}
	if string(data) != `{"u":"00000000-0000-0000-0000-000000000000","n":null}` {
		t.Errorf("Marshalled wrong JSON: %s", string(data))
	}

	// testStruct.U = ID{0x6b, 0xa7, 0xb8, 0x10, 0x9d, 0xad, 0x11, 0xd1, 0x80, 0xb4, 0x00, 0xc0, 0x4f, 0xd4, 0x30, 0xc8}
	// testStruct.N.ID = ID{0x6b, 0xa7, 0xb8, 0x10, 0x9d, 0xad, 0x11, 0xd1, 0x80, 0xb4, 0x00, 0xc0, 0x4f, 0xd4, 0x30, 0xc8}

	// data, err = json.Marshal(&testStruct)
	// if err != nil {
	// 	t.Errorf("Error JSON marshaling: %s", err)
	// }
	// if string(data) != `{"u":"6ba7b810-9dad-11d1-80b4-00c04fd430c8","n":"6ba7b810-9dad-11d1-80b4-00c04fd430c8"}` {
	// 	t.Errorf("Marshalled wrong JSON: %s", string(data))
	// }
}

func TestNullableID_UnmarshalJSON(t *testing.T) {
	type testStruct struct {
		U ID         `json:"u"`
		N NullableID `json:"n"`
	}
	var out *testStruct
	err := json.Unmarshal([]byte(`{"u":"00000000-0000-0000-0000-000000000000","n":null}`), &out)
	if err != nil {
		t.Errorf("Error JSON unmarshaling: %s", err)
	}
	if out == nil {
		t.Errorf("Error JSON unmarshaling")
	}
	if out.U != IDNil || out.N != IDNull {
		t.Errorf("Error JSON unmarshaling")
	}

	out = nil
	err = json.Unmarshal([]byte(`{"u":"6ba7b810-9dad-11d1-80b4-00c04fd430c8","n":"6ba7b810-9dad-11d1-80b4-00c04fd430c8"}`), &out)
	if err != nil {
		t.Errorf("Error JSON unmarshaling: %s", err)
	}
	if out == nil {
		t.Errorf("Error JSON unmarshaling")
	}
	ref := ID{0x6b, 0xa7, 0xb8, 0x10, 0x9d, 0xad, 0x11, 0xd1, 0x80, 0xb4, 0x00, 0xc0, 0x4f, 0xd4, 0x30, 0xc8}
	if out.U != ref || !out.N.Valid() || out.N.Get() != ref {
		t.Errorf("Error JSON unmarshaling")
	}
}

func TestNullableID_UnmarshalJSON_Variations(t *testing.T) {
	ref := ID{0x6b, 0xa7, 0xb8, 0x10, 0x9d, 0xad, 0x11, 0xd1, 0x80, 0xb4, 0x00, 0xc0, 0x4f, 0xd4, 0x30, 0xc8}
	refStr := "6ba7b810-9dad-11d1-80b4-00c04fd430c8"

	tests := []struct {
		name    string
		input   string
		want    NullableID
		wantErr bool
	}{
		{name: "null_literal", input: `null`, want: IDNull},
		{name: "null_with_leading_ws", input: "  \t\nnull", want: IDNull},
		{name: "null_with_trailing_ws", input: "null  \n", want: IDNull},
		{name: "string_dashed", input: `"` + refStr + `"`, want: NullableID(ref)},
		{name: "string_with_ws", input: "  \"" + refStr + "\"", want: NullableID(ref)},
		{name: "string_braces", input: `"{` + refStr + `}"`, want: NullableID(ref)},
		{name: "string_urn", input: `"urn:uuid:` + refStr + `"`, want: NullableID(ref)},
		{name: "string_nil_uuid", input: `"00000000-0000-0000-0000-000000000000"`, want: IDNull},
		{name: "nullstring_valid", input: `{"String":"` + refStr + `","Valid":true}`, want: NullableID(ref)},
		{name: "nullstring_invalid", input: `{"String":"","Valid":false}`, want: IDNull},
		// Malformed input must be reported instead of becoming NULL
		{name: "unterminated_string", input: `"6ba7b810`, wantErr: true},
		{name: "nullstring_bad_uuid", input: `{"String":"not-a-uuid","Valid":true}`, wantErr: true},
		{name: "nullstring_wrong_type", input: `{"String":1,"Valid":true}`, wantErr: true},
		{name: "empty", input: ``, wantErr: true},
		{name: "ws_only", input: "   \t\n", wantErr: true},
		{name: "bad_literal", input: `nope`, wantErr: true},
		{name: "bad_string", input: `"not-a-uuid"`, wantErr: true},
		{name: "number_unsupported", input: `123`, wantErr: true},
		{name: "array_unsupported", input: `["x"]`, wantErr: true},
		{name: "bool_unsupported", input: `true`, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var n NullableID
			err := n.UnmarshalJSON([]byte(tt.input))
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got n=%v", n)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if n != tt.want {
				t.Fatalf("got %v, want %v", n, tt.want)
			}
		})
	}
}

func BenchmarkNullableID_UnmarshalJSON(b *testing.B) {
	cases := map[string][]byte{
		"string":     []byte(`"6ba7b810-9dad-11d1-80b4-00c04fd430c8"`),
		"null":       []byte(`null`),
		"nullstring": []byte(`{"String":"6ba7b810-9dad-11d1-80b4-00c04fd430c8","Valid":true}`),
	}
	for name, data := range cases {
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			var n NullableID
			for i := 0; i < b.N; i++ {
				if err := n.UnmarshalJSON(data); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func TestNullableID_PrettyPrint(t *testing.T) {
	tests := []struct {
		id   NullableID
		want string
	}{
		{id: IDNull, want: "NULL"},
		{id: IDNil.Nullable(), want: "NULL"},
		{id: NullableIDMust("78c08786-f18d-442e-8598-30c9c59cc424"), want: "78c08786-f18d-442e-8598-30c9c59cc424"},
	}
	for _, tt := range tests {
		t.Run(tt.id.String(), func(t *testing.T) {
			got := pretty.Sprint(tt.id)
			if got != tt.want {
				t.Errorf("NullableID.PrettyPrint() = %q, want %q", got, tt.want)
			}
		})
	}
	// Test with pointer to NullableID
	for _, tt := range tests {
		t.Run(tt.id.String(), func(t *testing.T) {
			got := pretty.Sprint(&tt.id)
			if got != tt.want {
				t.Errorf("NullableID.PrettyPrint() = %q, want %q", got, tt.want)
			}
		})
	}
}

func ExampleNullableID_JSONSchema() {
	reflector := jsonschema.Reflector{
		Anonymous:      true,
		DoNotReference: true,
	}
	schema, _ := json.MarshalIndent(reflector.Reflect(NullableID{}), "", "  ")
	fmt.Println(string(schema))

	// Output:
	// {
	//   "$schema": "https://json-schema.org/draft/2020-12/schema",
	//   "oneOf": [
	//     {
	//       "type": "string",
	//       "format": "uuid"
	//     },
	//     {
	//       "type": "null"
	//     }
	//   ],
	//   "title": "Nullable UUID",
	//   "default": null
	// }
}

func TestNullableIDFromAny(t *testing.T) {
	refStr := "6ba7b810-9dad-11d1-80b4-00c04fd430c8"
	ref := IDMust(refStr)
	invalidVersion := ref
	invalidVersion.SetVersion(0)

	tests := []struct {
		name    string
		val     any
		want    NullableID
		wantErr bool
	}{
		{name: "string", val: refStr, want: ref.Nullable()},
		{name: "[]byte text", val: []byte(refStr), want: ref.Nullable()},
		{name: "[]byte binary", val: ref.Bytes(), want: ref.Nullable()},
		{name: "ID", val: ref, want: ref.Nullable()},
		{name: "NullableID", val: ref.Nullable(), want: ref.Nullable()},
		{name: "[16]byte", val: [16]byte(ref), want: ref.Nullable()},
		// Unlike IDFromAny, all representations of a missing
		// value are valid and must result in NULL without error.
		{name: "nil", val: nil, want: IDNull},
		{name: "Nil UUID string", val: "00000000-0000-0000-0000-000000000000", want: IDNull},
		{name: "Nil ID", val: IDNil, want: IDNull},
		{name: "IDNull", val: IDNull, want: IDNull},
		{name: "zero [16]byte", val: [16]byte{}, want: IDNull},
		{name: "invalid version ID", val: invalidVersion, want: invalidVersion.Nullable(), wantErr: true},
		{name: "invalid version NullableID", val: invalidVersion.Nullable(), want: invalidVersion.Nullable(), wantErr: true},
		{name: "invalid version [16]byte", val: [16]byte(invalidVersion), want: invalidVersion.Nullable(), wantErr: true},
		// Parsed strings and bytes must be validated like the typed IDs:
		// only the Nil UUID is a valid NULL, other invalid UUIDs are errors
		{name: "Nil UUID []byte text", val: []byte("00000000-0000-0000-0000-000000000000"), want: IDNull},
		{name: "Nil UUID []byte binary", val: make([]byte, 16), want: IDNull},
		{name: "invalid version string", val: invalidVersion.String(), want: invalidVersion.Nullable(), wantErr: true},
		{name: "invalid version []byte text", val: []byte(invalidVersion.String()), want: invalidVersion.Nullable(), wantErr: true},
		{name: "invalid version []byte binary", val: invalidVersion.Bytes(), want: invalidVersion.Nullable(), wantErr: true},
		{name: "empty string", val: "", want: IDNull, wantErr: true},
		{name: "invalid string", val: "not-a-uuid", want: IDNull, wantErr: true},
		{name: "unsupported type", val: 42, want: IDNull, wantErr: true},
		{name: "pointer not supported", val: &ref, want: IDNull, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NullableIDFromAny(tt.val)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, tt.want, got)
		})
	}
}

func TestNullableIDMust(t *testing.T) {
	refStr := "6ba7b810-9dad-11d1-80b4-00c04fd430c8"
	ref := IDMust(refStr)

	require.Equal(t, ref.Nullable(), NullableIDMust(refStr))
	require.Equal(t, ref.Nullable(), NullableIDMust([]byte(refStr)))
	require.Equal(t, ref.Nullable(), NullableIDMust(ref.Bytes()))
	require.Equal(t, ref.Nullable(), NullableIDMust(ref))
	require.Equal(t, ref.Nullable(), NullableIDMust(ref.Nullable()))
	require.Equal(t, ref.Nullable(), NullableIDMust([16]byte(ref)))

	// NULL is a valid NullableID and must not panic
	require.Equal(t, IDNull, NullableIDMust("00000000-0000-0000-0000-000000000000"))
	require.Equal(t, IDNull, NullableIDMust([]byte("00000000-0000-0000-0000-000000000000")))
	require.Equal(t, IDNull, NullableIDMust(make([]byte, 16)))
	require.Equal(t, IDNull, NullableIDMust(IDNull))
	require.Equal(t, IDNull, NullableIDMust([16]byte{}))

	invalidVersion := ref
	invalidVersion.SetVersion(0)
	require.Panics(t, func() { NullableIDMust("not-a-uuid") })
	require.Panics(t, func() { NullableIDMust([]byte("short")) })
	require.Panics(t, func() { NullableIDMust(invalidVersion) })
	require.Panics(t, func() { NullableIDMust(invalidVersion.Nullable()) })
	require.Panics(t, func() { NullableIDMust([16]byte(invalidVersion)) })
	// A parsed non-NULL UUID must be valid like the typed IDs
	require.Panics(t, func() { NullableIDMust(invalidVersion.String()) })
	require.Panics(t, func() { NullableIDMust([]byte(invalidVersion.String())) })
	require.Panics(t, func() { NullableIDMust(invalidVersion.Bytes()) })
}

func TestNullableID_ValidateAndVersion(t *testing.T) {
	ref := IDMust("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	n := ref.Nullable()

	require.Equal(t, 1, n.Version())
	require.Equal(t, IDVariantRFC4122, n.Variant())
	require.NoError(t, n.Validate())

	// NULL is a valid value for a NullableID (unlike for ID)
	require.NoError(t, IDNull.Validate())
	require.True(t, IDNull.Valid())

	invalidVersion := n
	(*ID)(&invalidVersion).SetVersion(0)
	require.ErrorIs(t, invalidVersion.Validate(), ErrInvalidVersion(0))
	require.False(t, invalidVersion.Valid())
}

func TestNullableID_SetGet(t *testing.T) {
	ref := IDMust("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	def := IDMust("de8482f2-164a-40f9-b4e0-c349b78c22f6")

	var n NullableID
	require.True(t, n.IsNull(), "zero value is NULL")
	require.False(t, n.IsNotNull())
	require.Panics(t, func() { n.Get() }, "Get must not silently return the Nil UUID")
	require.Equal(t, def, n.GetOr(def))
	require.Equal(t, IDNil, n.GetOrNil())

	n.Set(ref)
	require.True(t, n.IsNotNull())
	require.Equal(t, ref, n.Get())
	require.Equal(t, ref, n.GetOr(def))
	require.Equal(t, ref, n.GetOrNil())

	n.SetNull()
	require.True(t, n.IsNull())

	// Setting the Nil UUID results in NULL
	n.Set(IDNil)
	require.True(t, n.IsNull())
}

func TestNullableID_StringRepresentations(t *testing.T) {
	ref := IDMust("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	n := ref.Nullable()

	require.Equal(t, "6ba7b810-9dad-11d1-80b4-00c04fd430c8", n.String())
	require.Equal(t, "6BA7B810-9DAD-11D1-80B4-00C04FD430C8", n.StringUpper())
	require.Equal(t, "6ba7b8109dad11d180b400c04fd430c8", n.Hex())
	require.Equal(t, ref.Base64(), n.Base64())
	require.Equal(t, "uu.NullableID(`6ba7b810-9dad-11d1-80b4-00c04fd430c8`)", n.GoString())
	require.Equal(t, n.GoString(), fmt.Sprintf("%#v", n), "used by fmt %#v")

	// NULL is shown as NULL instead of the Nil UUID
	require.Equal(t, "NULL", IDNull.String())
	require.Equal(t, "NULL", IDNull.StringUpper())
	require.Equal(t, "uu.IDNull", IDNull.GoString())
	require.Equal(t, "uu.IDNull", fmt.Sprintf("%#v", IDNull))
}

func TestNullableID_Ptr(t *testing.T) {
	require.Nil(t, IDNull.Ptr(), "NULL must map to a nil pointer")

	ref := IDMust("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	n := ref.Nullable()
	ptr := n.Ptr()
	require.NotNil(t, ptr)
	require.Equal(t, ref, *ptr)

	// The pointer must reference a copy, not the original value
	*ptr = IDNil
	require.Equal(t, ref.Nullable(), n)

	// Ptr and NullableIDFromPtr are inverse operations
	require.Equal(t, n, NullableIDFromPtr(n.Ptr()))
	require.Equal(t, IDNull, NullableIDFromPtr(IDNull.Ptr()))
}

func TestNullableID_Value(t *testing.T) {
	ref := IDMust("6ba7b810-9dad-11d1-80b4-00c04fd430c8")

	val, err := ref.Nullable().Value()
	require.NoError(t, err)
	require.Equal(t, "6ba7b810-9dad-11d1-80b4-00c04fd430c8", val, "canonical string for the database")

	// NULL must be written as SQL NULL, not as the Nil UUID string
	val, err = IDNull.Value()
	require.NoError(t, err)
	require.Nil(t, val)
}

func TestNullableID_Scan(t *testing.T) {
	ref := IDMust("6ba7b810-9dad-11d1-80b4-00c04fd430c8")

	tests := []struct {
		name    string
		src     any
		want    NullableID
		wantErr bool
	}{
		{name: "SQL NULL", src: nil, want: IDNull},
		{name: "string", src: "6ba7b810-9dad-11d1-80b4-00c04fd430c8", want: ref.Nullable()},
		{name: "text bytes", src: []byte("6ba7b810-9dad-11d1-80b4-00c04fd430c8"), want: ref.Nullable()},
		{name: "binary bytes", src: ref.Bytes(), want: ref.Nullable()},
		{name: "Nil UUID string", src: "00000000-0000-0000-0000-000000000000", want: IDNull},
		{name: "invalid string", src: "not-a-uuid", wantErr: true},
		{name: "unsupported type", src: 42, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := IDMust("de8482f2-164a-40f9-b4e0-c349b78c22f6").Nullable()
			err := n.Scan(tt.src)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, n)
		})
	}
}

func TestNullableID_SQLRoundTrip(t *testing.T) {
	for _, original := range []NullableID{IDNull, NullableIDMust("6ba7b810-9dad-11d1-80b4-00c04fd430c8")} {
		val, err := original.Value()
		require.NoError(t, err)

		var scanned NullableID
		require.NoError(t, scanned.Scan(val))
		require.Equal(t, original, scanned)
	}
}

func TestNullableID_UnmarshalJSON_EmptyString(t *testing.T) {
	// Documented: a blank JSON string does not produce a NULL ID.
	// Only the JSON null literal means NULL, so that empty strings
	// from broken clients are reported instead of silently clearing a reference.
	var n NullableID
	require.Error(t, json.Unmarshal([]byte(`""`), &n))
}

func TestNullableID_MarshalUnmarshalText(t *testing.T) {
	ref := IDMust("6ba7b810-9dad-11d1-80b4-00c04fd430c8")

	text, err := ref.Nullable().MarshalText()
	require.NoError(t, err)
	require.Equal(t, "6ba7b810-9dad-11d1-80b4-00c04fd430c8", string(text))

	// NULL is encoded as empty text, e.g. an empty query or form parameter
	text, err = IDNull.MarshalText()
	require.NoError(t, err)
	require.Empty(t, text)

	var n NullableID
	require.NoError(t, n.UnmarshalText([]byte("6ba7b810-9dad-11d1-80b4-00c04fd430c8")))
	require.Equal(t, ref.Nullable(), n)

	// ... and empty text is decoded as NULL
	require.NoError(t, n.UnmarshalText(nil))
	require.Equal(t, IDNull, n)
	n = ref.Nullable()
	require.NoError(t, n.UnmarshalText([]byte{}))
	require.Equal(t, IDNull, n)

	require.Error(t, n.UnmarshalText([]byte("not-a-uuid")))

	// Round trip for both NULL and non-NULL values
	for _, original := range []NullableID{IDNull, ref.Nullable()} {
		text, err := original.MarshalText()
		require.NoError(t, err)
		var decoded NullableID
		require.NoError(t, decoded.UnmarshalText(text))
		require.Equal(t, original, decoded)
	}
}

// TestNullableID_RejectsRawBinaryText checks that NullableID, like ID,
// parses text and JSON only in the textual UUID formats, so a 16 character
// string is not taken as raw binary UUID bytes. A 16 byte []byte from a
// SQL driver is binary data and must still scan.
func TestNullableID_RejectsRawBinaryText(t *testing.T) {
	const text16 = "abcdefghijklmnop"
	ref := IDMust("6ba7b810-9dad-11d1-80b4-00c04fd430c8").Nullable()

	n := ref
	require.Error(t, n.UnmarshalText([]byte(text16)), "UnmarshalText of 16 characters")
	require.Error(t, json.Unmarshal([]byte(`"`+text16+`"`), &n), "json.Unmarshal of 16 characters")
	require.Error(t, n.Scan(text16), "Scan of a 16 character string")
	require.Equal(t, ref, n, "NullableID must be unchanged after errors")

	var scanned NullableID
	require.NoError(t, scanned.Scan(ref.Get().Bytes()), "Scan of raw binary []byte")
	require.Equal(t, ref, scanned)
}

func TestNullableID_MarshalUnmarshalBinary(t *testing.T) {
	ref := IDMust("6ba7b810-9dad-11d1-80b4-00c04fd430c8")

	for _, original := range []NullableID{IDNull, ref.Nullable()} {
		data, err := original.MarshalBinary()
		require.NoError(t, err)
		require.Len(t, data, 16, "binary encoding is always 16 bytes, also for NULL")

		var decoded NullableID
		require.NoError(t, decoded.UnmarshalBinary(data))
		require.Equal(t, original, decoded)
	}

	var n NullableID
	require.Error(t, n.UnmarshalBinary([]byte{1, 2, 3}), "wrong length")
	require.Error(t, n.UnmarshalBinary(nil), "wrong length")
}

func TestNullableIDCompare(t *testing.T) {
	low := NullableIDMust("0192e8f1-0000-7000-8000-000000000000")
	high := NullableIDMust("0192e8f2-0000-7000-8000-000000000000")

	require.Equal(t, -1, NullableIDCompare(low, high))
	require.Equal(t, 1, NullableIDCompare(high, low))
	require.Equal(t, 0, NullableIDCompare(low, low))
	// NULL sorts before every non-NULL ID
	require.Equal(t, -1, NullableIDCompare(IDNull, low))
	require.Equal(t, 0, NullableIDCompare(IDNull, IDNull))
}
