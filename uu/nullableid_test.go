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
