package uu

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/invopop/jsonschema"
	"github.com/stretchr/testify/require"

	"github.com/domonda/go-pretty"
)

// And returns result of binary AND of two UUIDs.
func And(u1 ID, u2 ID) ID {
	u := ID{}
	for i := range 16 {
		u[i] = u1[i] & u2[i]
	}
	return u
}

// Or returns result of binary OR of two UUIDs.
func Or(u1 ID, u2 ID) ID {
	u := ID{}
	for i := range 16 {
		u[i] = u1[i] | u2[i]
	}
	return u
}

func TestBytes(t *testing.T) {
	u := ID{0x6b, 0xa7, 0xb8, 0x10, 0x9d, 0xad, 0x11, 0xd1, 0x80, 0xb4, 0x00, 0xc0, 0x4f, 0xd4, 0x30, 0xc8}

	bytes1 := []byte{0x6b, 0xa7, 0xb8, 0x10, 0x9d, 0xad, 0x11, 0xd1, 0x80, 0xb4, 0x00, 0xc0, 0x4f, 0xd4, 0x30, 0xc8}

	if !bytes.Equal(u.Bytes(), bytes1) {
		t.Errorf("Incorrect bytes representation for UUID: %s", u)
	}
}

func TestString(t *testing.T) {
	if NamespaceDNS.String() != "6ba7b810-9dad-11d1-80b4-00c04fd430c8" {
		t.Errorf("Incorrect string representation for UUID: %s", NamespaceDNS.String())
	}
}

func TestOr(t *testing.T) {
	u1 := ID{0x00, 0xff, 0x00, 0xff, 0x00, 0xff, 0x00, 0xff, 0x00, 0xff, 0x00, 0xff, 0x00, 0xff, 0x00, 0xff}
	u2 := ID{0xff, 0x00, 0xff, 0x00, 0xff, 0x00, 0xff, 0x00, 0xff, 0x00, 0xff, 0x00, 0xff, 0x00, 0xff, 0x00}

	u := ID{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}

	if u != Or(u1, u2) {
		t.Errorf("Incorrect bitwise OR result %s", Or(u1, u2))
	}
}

func TestAnd(t *testing.T) {
	u1 := ID{0x00, 0xff, 0x00, 0xff, 0x00, 0xff, 0x00, 0xff, 0x00, 0xff, 0x00, 0xff, 0x00, 0xff, 0x00, 0xff}
	u2 := ID{0xff, 0x00, 0xff, 0x00, 0xff, 0x00, 0xff, 0x00, 0xff, 0x00, 0xff, 0x00, 0xff, 0x00, 0xff, 0x00}

	u := ID{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}

	if u != And(u1, u2) {
		t.Errorf("Incorrect bitwise AND result %s", And(u1, u2))
	}
}

func TestVersion(t *testing.T) {
	u := ID{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x10, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}

	if u.Version() != 1 {
		t.Errorf("Incorrect version for UUID: %d", u.Version())
	}
}

func TestSetVersion(t *testing.T) {
	u := ID{}
	u.SetVersion(4)

	if u.Version() != 4 {
		t.Errorf("Incorrect version for UUID after u.setVersion(4): %d", u.Version())
	}
}

func TestVariant(t *testing.T) {
	u1 := ID{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}

	if u1.Variant() != IDVariantNCS {
		t.Errorf("Incorrect variant for UUID variant %d: %d", IDVariantNCS, u1.Variant())
	}

	u2 := ID{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}

	if u2.Variant() != IDVariantRFC4122 {
		t.Errorf("Incorrect variant for UUID variant %d: %d", IDVariantRFC4122, u2.Variant())
	}

	u3 := ID{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xc0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}

	if u3.Variant() != IDVariantMicrosoft {
		t.Errorf("Incorrect variant for UUID variant %d: %d", IDVariantMicrosoft, u3.Variant())
	}

	u4 := ID{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xe0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}

	if u4.Variant() != IDVariantInvalid {
		t.Errorf("Incorrect variant for UUID variant %d: %d", IDVariantInvalid, u4.Variant())
	}
}

func TestSetVariant(t *testing.T) {
	u := new(ID)
	u.SetVariant()

	if u.Variant() != IDVariantRFC4122 {
		t.Errorf("Incorrect variant for UUID after u.setVariant(): %d", u.Variant())
	}
}

func TestIDFromBytes(t *testing.T) {
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
			parsed, err := IDFromBytes(tt.input)
			require.NoError(t, err, "parsing should succeed")
			require.Equal(t, u, parsed, "parsed UUID should equal expected")
		})
	}

	// Test error cases
	t.Run("empty byte slice", func(t *testing.T) {
		_, err := IDFromBytes([]byte{})
		require.Error(t, err, "should return error for empty slice")
	})

	t.Run("too short", func(t *testing.T) {
		_, err := IDFromBytes([]byte("too-short"))
		require.Error(t, err, "should return error for too short input")
	})
}

func TestMarshalBinary(t *testing.T) {
	u := ID{0x6b, 0xa7, 0xb8, 0x10, 0x9d, 0xad, 0x11, 0xd1, 0x80, 0xb4, 0x00, 0xc0, 0x4f, 0xd4, 0x30, 0xc8}
	b1 := []byte{0x6b, 0xa7, 0xb8, 0x10, 0x9d, 0xad, 0x11, 0xd1, 0x80, 0xb4, 0x00, 0xc0, 0x4f, 0xd4, 0x30, 0xc8}

	b2, err := u.MarshalBinary()
	if err != nil {
		t.Errorf("Error marshaling UUID: %s", err)
	}

	if !bytes.Equal(b1, b2) {
		t.Errorf("Marshaled UUID should be %s, got %s", b1, b2)
	}
}

func TestUnmarshalBinary(t *testing.T) {
	u := ID{0x6b, 0xa7, 0xb8, 0x10, 0x9d, 0xad, 0x11, 0xd1, 0x80, 0xb4, 0x00, 0xc0, 0x4f, 0xd4, 0x30, 0xc8}
	b1 := []byte{0x6b, 0xa7, 0xb8, 0x10, 0x9d, 0xad, 0x11, 0xd1, 0x80, 0xb4, 0x00, 0xc0, 0x4f, 0xd4, 0x30, 0xc8}

	u1 := ID{}
	err := u1.UnmarshalBinary(b1)
	if err != nil {
		t.Errorf("Error unmarshaling UUID: %s", err)
	}

	if u != u1 {
		t.Errorf("UUIDs should be equal: %s and %s", u, u1)
	}

	b2 := []byte{}
	u2 := ID{}

	err = u2.UnmarshalBinary(b2)
	if err == nil {
		t.Errorf("Should return error unmarshalling from empty byte slice, got %s", err)
	}
}

func TestIDFromString(t *testing.T) {
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
		{
			name:  "quoted braces",
			input: `"{6ba7b810-9dad-11d1-80b4-00c04fd430c8}"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, err := IDFromString(tt.input)
			require.NoError(t, err, "parsing should succeed")
			require.Equal(t, u, parsed, "parsed UUID should equal expected")
		})
	}

	// Test error cases
	t.Run("empty string", func(t *testing.T) {
		_, err := IDFromString("")
		require.Error(t, err, "should return error for empty string")
	})
}

func TestIDFromStringShort(t *testing.T) {
	strs := []string{
		"1a9a39a3-cc4b-4fef-bf59-975667a93b3e",
		"d33c40a4-b8d4-4229-b0bc-56b08cd925da",
		"fdd3e1a3-4e11-4bcb-b89a-35adb170a132",
		"0810befa-8992-4ee5-8bea-f34d8c288b6d",
		"44f5163a-cb73-45f9-a7e9-785bc3753014",
		"d8d13d23-60e0-423d-b2e5-49d553519f49",
	}
	for _, str := range strs {
		for l := len(str) - 1; l >= 0; l-- {
			// Skip valid format lengths:
			// - 22 characters: valid base64 format
			// - 32 characters: valid hex without dashes format
			// Note: 36 is the full length (tested separately)
			if l == 22 || l == 32 {
				continue
			}

			id, err := IDFromString(str[:l])
			if err == nil {
				t.Errorf("Should return error trying to parse too short string (length %d), got UUID %s from string %q", l, id, str[:l])
			}
		}
	}
}

func TestIDFromStringLong(t *testing.T) {
	// Invalid 37+ character UUID string
	s := []string{
		"6ba7b810-9dad-11d1-80b4-00c04fd430c8=",
		"6ba7b810-9dad-11d1-80b4-00c04fd430c8}",
		"{6ba7b810-9dad-11d1-80b4-00c04fd430c8}f",
		"6ba7b810-9dad-11d1-80b4-00c04fd430c800c04fd430c8",
	}

	for _, str := range s {
		_, err := IDFromString(str)
		if err == nil {
			t.Errorf("Should return error trying to parse too long string, passed %s", str)
		}
	}
}

func TestIDFromStringInvalid(t *testing.T) {
	// Invalid UUID string formats
	s := []string{
		"6ba7b8109dad11d180b400c04fd430c86ba7b8109dad11d180b400c04fd430c8",
		"urn:uuid:{6ba7b810-9dad-11d1-80b4-00c04fd430c8}",
		"6ba7b8109-dad-11d1-80b4-00c04fd430c8",
		"6ba7b810-9dad1-1d1-80b4-00c04fd430c8",
		"6ba7b810-9dad-11d18-0b4-00c04fd430c8",
		"6ba7b810-9dad-11d1-80b40-0c04fd430c8",
		"6ba7b810+9dad+11d1+80b4+00c04fd430c8",
		"6ba7b810-9dad11d180b400c04fd430c8",
		"6ba7b8109dad-11d180b400c04fd430c8",
		"6ba7b8109dad11d1-80b400c04fd430c8",
		"6ba7b8109dad11d180b4-00c04fd430c8",
	}

	for _, str := range s {
		_, err := IDFromString(str)
		if err == nil {
			t.Errorf("Should return error trying to parse invalid string, passed %s", str)
		}
	}
}

func TestIDFromStringOrNil(t *testing.T) {
	u := IDFromStringOrNil("")
	if u != IDNil {
		t.Errorf("Should return Nil UUID on parse failure, got %s", u)
	}
}

func TestIDFromBytesOrNil(t *testing.T) {
	b := []byte{}
	u := IDFromBytesOrNil(b)
	if u != IDNil {
		t.Errorf("Should return Nil UUID on parse failure, got %s", u)
	}
}

func TestMarshalText(t *testing.T) {
	u := ID{0x6b, 0xa7, 0xb8, 0x10, 0x9d, 0xad, 0x11, 0xd1, 0x80, 0xb4, 0x00, 0xc0, 0x4f, 0xd4, 0x30, 0xc8}
	b1 := []byte("6ba7b810-9dad-11d1-80b4-00c04fd430c8")

	b2, err := u.MarshalText()
	if err != nil {
		t.Errorf("Error marshaling UUID: %s", err)
	}

	if !bytes.Equal(b1, b2) {
		t.Errorf("Marshaled UUID should be %s, got %s", b1, b2)
	}
}

func TestUnmarshalText(t *testing.T) {
	u := ID{0x6b, 0xa7, 0xb8, 0x10, 0x9d, 0xad, 0x11, 0xd1, 0x80, 0xb4, 0x00, 0xc0, 0x4f, 0xd4, 0x30, 0xc8}

	tests := []struct {
		name  string
		input []byte
	}{
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
			name:  "uppercase hex",
			input: []byte("6BA7B8109DAD11D180B400C04FD430C8"),
		},
		{
			name:  "uppercase dashed",
			input: []byte("6BA7B810-9DAD-11D1-80B4-00C04FD430C8"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var parsed ID
			err := parsed.UnmarshalText(tt.input)
			require.NoError(t, err, "unmarshaling should succeed")
			require.Equal(t, u, parsed, "unmarshaled UUID should equal expected")
		})
	}

	// Test error case
	t.Run("empty string", func(t *testing.T) {
		var id ID
		err := id.UnmarshalText([]byte(""))
		require.Error(t, err, "should return error for empty string")
	})
}

// TestUnmarshalText_RejectsRawBinary checks that text and JSON are only
// parsed in the textual UUID formats. The raw 16 byte binary format that
// IDFromBytes also accepts is not text: accepting it would turn every
// 16 character string, like the JSON value "abcdefghijklmnop", into an ID
// without an error. Raw binary stays valid where binary data is expected,
// in UnmarshalBinary and in Scan of a 16 byte []byte, which is how SQL
// drivers return binary UUID columns.
func TestUnmarshalText_RejectsRawBinary(t *testing.T) {
	const text16 = "abcdefghijklmnop"
	ref := IDMust("6ba7b810-9dad-11d1-80b4-00c04fd430c8")

	id := ref
	require.Error(t, id.UnmarshalText([]byte(text16)), "UnmarshalText of 16 characters")
	require.Error(t, id.UnmarshalText(ref.Bytes()), "UnmarshalText of raw binary")
	require.Error(t, json.Unmarshal([]byte(`"`+text16+`"`), &id), "json.Unmarshal of 16 characters")
	require.Error(t, id.Scan(text16), "Scan of a 16 character string")
	require.Equal(t, ref, id, "ID must be unchanged after errors")

	var binID ID
	require.NoError(t, binID.UnmarshalBinary(ref.Bytes()))
	require.Equal(t, ref, binID)
	var scanned ID
	require.NoError(t, scanned.Scan(ref.Bytes()), "Scan of raw binary []byte")
	require.Equal(t, ref, scanned)
}

func TestValue(t *testing.T) {
	u, err := IDFromString("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	if err != nil {
		t.Errorf("Error parsing UUID from string: %s", err)
	}

	val, err := u.Value()
	if err != nil {
		t.Errorf("Error getting UUID value: %s", err)
	}

	if val != u.String() {
		t.Errorf("Wrong value returned, should be equal: %s and %s", val, u)
	}
}

func TestValueNil(t *testing.T) {
	u := ID{}

	val, err := u.Value()
	if err != nil {
		t.Errorf("Error getting UUID value: %s", err)
	}

	if val != IDNil.String() {
		t.Errorf("Wrong value returned, should be equal to UUID.Nil: %s", val)
	}
}

func TestScanBinary(t *testing.T) {
	u := ID{0x6b, 0xa7, 0xb8, 0x10, 0x9d, 0xad, 0x11, 0xd1, 0x80, 0xb4, 0x00, 0xc0, 0x4f, 0xd4, 0x30, 0xc8}
	b1 := []byte{0x6b, 0xa7, 0xb8, 0x10, 0x9d, 0xad, 0x11, 0xd1, 0x80, 0xb4, 0x00, 0xc0, 0x4f, 0xd4, 0x30, 0xc8}

	u1 := ID{}
	err := u1.Scan(b1)
	if err != nil {
		t.Errorf("Error unmarshaling UUID: %s", err)
	}

	if u != u1 {
		t.Errorf("UUIDs should be equal: %s and %s", u, u1)
	}

	b2 := []byte{}
	u2 := ID{}

	err = u2.Scan(b2)
	if err == nil {
		t.Errorf("Should return error unmarshalling from empty byte slice, got %s", err)
	}
}

func TestScanString(t *testing.T) {
	u := ID{0x6b, 0xa7, 0xb8, 0x10, 0x9d, 0xad, 0x11, 0xd1, 0x80, 0xb4, 0x00, 0xc0, 0x4f, 0xd4, 0x30, 0xc8}
	s1 := "6ba7b810-9dad-11d1-80b4-00c04fd430c8"

	u1 := ID{}
	err := u1.Scan(s1)
	if err != nil {
		t.Errorf("Error unmarshaling UUID: %s", err)
	}

	if u != u1 {
		t.Errorf("UUIDs should be equal: %s and %s", u, u1)
	}

	s2 := ""
	u2 := ID{}

	err = u2.Scan(s2)
	if err == nil {
		t.Errorf("Should return error trying to unmarshal from empty string")
	}
}

func TestScanText(t *testing.T) {
	u := ID{0x6b, 0xa7, 0xb8, 0x10, 0x9d, 0xad, 0x11, 0xd1, 0x80, 0xb4, 0x00, 0xc0, 0x4f, 0xd4, 0x30, 0xc8}
	b1 := []byte("6ba7b810-9dad-11d1-80b4-00c04fd430c8")

	u1 := ID{}
	err := u1.Scan(b1)
	if err != nil {
		t.Errorf("Error unmarshaling UUID: %s", err)
	}

	if u != u1 {
		t.Errorf("UUIDs should be equal: %s and %s", u, u1)
	}

	b2 := []byte("")
	u2 := ID{}

	err = u2.Scan(b2)
	if err == nil {
		t.Errorf("Should return error trying to unmarshal from empty string")
	}
}

func TestScanUnsupported(t *testing.T) {
	u := ID{}

	err := u.Scan(true)
	if err == nil {
		t.Errorf("Should return error trying to unmarshal from bool")
	}
}

func TestScanNil(t *testing.T) {
	u := ID{0x6b, 0xa7, 0xb8, 0x10, 0x9d, 0xad, 0x11, 0xd1, 0x80, 0xb4, 0x00, 0xc0, 0x4f, 0xd4, 0x30, 0xc8}

	err := u.Scan(nil)
	if err == nil {
		t.Errorf("Error UUID shouldn't allow unmarshalling from nil")
	}
}

func TestIDv1(t *testing.T) {
	u := IDv1()

	if u.Version() != 1 {
		t.Errorf("UUIDv1 generated with incorrect version: %d", u.Version())
	}

	if u.Variant() != IDVariantRFC4122 {
		t.Errorf("UUIDv1 generated with incorrect variant: %d", u.Variant())
	}

	u1 := IDv1()
	u2 := IDv1()

	if u1 == u2 {
		t.Errorf("UUIDv1 generated two equal UUIDs: %s and %s", u1, u2)
	}

	oldFunc := epochFunc
	epochFunc = func() uint64 { return 0 }

	u3 := IDv1()
	u4 := IDv1()

	if u3 == u4 {
		t.Errorf("UUIDv1 generated two equal UUIDs: %s and %s", u3, u4)
	}

	epochFunc = oldFunc
}

func TestIDv2(t *testing.T) {
	u1 := IDv2(IDDomainPerson)

	if u1.Version() != 2 {
		t.Errorf("UUIDv2 generated with incorrect version: %d", u1.Version())
	}

	if u1.Variant() != IDVariantRFC4122 {
		t.Errorf("UUIDv2 generated with incorrect variant: %d", u1.Variant())
	}

	u2 := IDv2(IDDomainGroup)

	if u2.Version() != 2 {
		t.Errorf("UUIDv2 generated with incorrect version: %d", u2.Version())
	}

	if u2.Variant() != IDVariantRFC4122 {
		t.Errorf("UUIDv2 generated with incorrect variant: %d", u2.Variant())
	}
}

func TestIDv3(t *testing.T) {
	u := IDv3(NamespaceDNS, []byte("www.example.com"))

	if u.Version() != 3 {
		t.Errorf("UUIDv3 generated with incorrect version: %d", u.Version())
	}

	if u.Variant() != IDVariantRFC4122 {
		t.Errorf("UUIDv3 generated with incorrect variant: %d", u.Variant())
	}

	if u.String() != "5df41881-3aed-3515-88a7-2f4a814cf09e" {
		t.Errorf("UUIDv3 generated incorrectly: %s", u.String())
	}

	u = IDv3(NamespaceDNS, []byte("python.org"))

	if u.String() != "6fa459ea-ee8a-3ca4-894e-db77e160355e" {
		t.Errorf("UUIDv3 generated incorrectly: %s", u.String())
	}

	u1 := IDv3(NamespaceDNS, []byte("golang.org"))
	u2 := IDv3(NamespaceDNS, []byte("golang.org"))
	if u1 != u2 {
		t.Errorf("UUIDv3 generated different UUIDs for same namespace and name: %s and %s", u1, u2)
	}

	u3 := IDv3(NamespaceDNS, []byte("example.com"))
	if u1 == u3 {
		t.Errorf("UUIDv3 generated same UUIDs for different names in same namespace: %s and %s", u1, u2)
	}

	u4 := IDv3(NamespaceURL, []byte("golang.org"))
	if u1 == u4 {
		t.Errorf("UUIDv3 generated same UUIDs for sane names in different namespaces: %s and %s", u1, u4)
	}
}

func TestIDv4(t *testing.T) {
	u := IDv4()

	if u.Version() != 4 {
		t.Errorf("UUIDv4 generated with incorrect version: %d", u.Version())
	}

	if u.Variant() != IDVariantRFC4122 {
		t.Errorf("UUIDv4 generated with incorrect variant: %d", u.Variant())
	}
}

func TestIDv5(t *testing.T) {
	u := IDv5(NamespaceDNS, []byte("www.example.com"))

	if u.Version() != 5 {
		t.Errorf("UUIDv5 generated with incorrect version: %d", u.Version())
	}

	if u.Variant() != IDVariantRFC4122 {
		t.Errorf("UUIDv5 generated with incorrect variant: %d", u.Variant())
	}

	u = IDv5(NamespaceDNS, []byte("python.org"))

	if u.String() != "886313e1-3b8a-5372-9b90-0c9aee199e5d" {
		t.Errorf("UUIDv5 generated incorrectly: %s", u.String())
	}

	u1 := IDv5(NamespaceDNS, []byte("golang.org"))
	u2 := IDv5(NamespaceDNS, []byte("golang.org"))
	if u1 != u2 {
		t.Errorf("UUIDv5 generated different UUIDs for same namespace and name: %s and %s", u1, u2)
	}

	u3 := IDv5(NamespaceDNS, []byte("example.com"))
	if u1 == u3 {
		t.Errorf("UUIDv5 generated same UUIDs for different names in same namespace: %s and %s", u1, u2)
	}

	u4 := IDv5(NamespaceURL, []byte("golang.org"))
	if u1 == u4 {
		t.Errorf("UUIDv3 generated same UUIDs for sane names in different namespaces: %s and %s", u1, u4)
	}
}

func TestIDv7(t *testing.T) {
	id := IDMust("019222e8-1ec3-7e4f-97e2-919670df6d6b")
	require.NoError(t, id.Validate(), "validating UUID")
	require.True(t, id.Valid(), "valid UUID")
	require.Equal(t, 7, id.Version(), "detecting version 7")

	id = IDv7()
	require.NoError(t, id.Validate(), "validating UUID")
	require.True(t, id.Valid(), "valid UUID")
	require.Equal(t, 7, id.Version(), "detecting version 7")

	time.Sleep(time.Millisecond)
	id2 := IDv7()
	require.NotEqual(t, id, id2, "different timestamps should produce different UUIDs")
}

func TestIDv7DeterministicFunc(t *testing.T) {
	require.Equal(t, IDv7Deterministic(666), IDv7Deterministic(666), "deterministic function is stable")

	id := IDv7Deterministic(0)
	require.NoError(t, id.Validate(), "validating UUID")
	require.Equal(t, 7, id.Version(), "detecting version 7")

	id = IDv7Deterministic(time.Now().UnixMilli())
	require.NoError(t, id.Validate(), "validating UUID")
	require.Equal(t, 7, id.Version(), "detecting version 7")
}

func TestID_GoString(t *testing.T) {
	tests := []struct {
		name string
		id   ID
		want string
	}{
		{"9256978d-18e6-4435-ad16-d7046d41b71a", IDMustFromString("9256978d-18e6-4435-ad16-d7046d41b71a"), "uu.IDFrom(`9256978d-18e6-4435-ad16-d7046d41b71a`)"},
		{"a92bb308-f0f9-43d2-b31d-c0962198c31c", IDMustFromString("a92bb308-f0f9-43d2-b31d-c0962198c31c"), "uu.IDFrom(`a92bb308-f0f9-43d2-b31d-c0962198c31c`)"},
		{"59a268e5-d820-4884-b120-10d1a9b0dd00", IDMustFromString("59a268e5-d820-4884-b120-10d1a9b0dd00"), "uu.IDFrom(`59a268e5-d820-4884-b120-10d1a9b0dd00`)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.id.GoString(); got != tt.want {
				t.Errorf("ID.GoString() = %v, want %v", got, tt.want)
			}
			if got := fmt.Sprintf("%#v", tt.id); got != tt.want {
				t.Errorf("ID.GoString() = %v, want %v", got, tt.want)
			}
			if got := fmt.Sprintf("%v", tt.id); got != tt.name {
				t.Errorf("ID.String() = %v, want %v", got, tt.name)
			}
		})
	}

}

func TestID_Base64(t *testing.T) {
	for range 100 {
		id := IDv4()
		t.Run(id.String(), func(t *testing.T) {
			b := id.Base64()
			require.Len(t, b, 22, "always 22 characters")
			parsed, err := IDFromString(b)
			require.NoError(t, err, "can parse")
			require.Equal(t, id, parsed, "parsed base64 UUID equal to original")
		})
	}
}

func TestIDFrom(t *testing.T) {
	refID := ID{0x6b, 0xa7, 0xb8, 0x10, 0x9d, 0xad, 0x11, 0xd1, 0x80, 0xb4, 0x00, 0xc0, 0x4f, 0xd4, 0x30, 0xc8}

	{
		tests := []struct {
			name string
			val  string
			want ID
		}{
			{name: "6ba7b8109dad11d180b400c04fd430c8", val: "6ba7b8109dad11d180b400c04fd430c8", want: refID},
			{name: "6ba7b810-9dad-11d1-80b4-00c04fd430c8", val: "6ba7b810-9dad-11d1-80b4-00c04fd430c8", want: refID},
			{name: "{6ba7b810-9dad-11d1-80b4-00c04fd430c8}", val: "{6ba7b810-9dad-11d1-80b4-00c04fd430c8}", want: refID},
			{name: "urn:uuid:6ba7b810-9dad-11d1-80b4-00c04fd430c8", val: "urn:uuid:6ba7b810-9dad-11d1-80b4-00c04fd430c8", want: refID},
			{name: "6BA7B8109DAD11D180B400C04FD430C8", val: "6BA7B8109DAD11D180B400C04FD430C8", want: refID},
			{name: "6BA7B810-9DAD-11D1-80B4-00C04FD430C8", val: "6BA7B810-9DAD-11D1-80B4-00C04FD430C8", want: refID},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if got := IDFrom(tt.val); !reflect.DeepEqual(got, tt.want) {
					t.Errorf("IDFrom() = %v, want %v", got, tt.want)
				}
				if got := IDMust(tt.val); !reflect.DeepEqual(got, tt.want) {
					t.Errorf("IDMust() = %v, want %v", got, tt.want)
				}
			})
		}
	}
	{
		tests := []struct {
			name string
			val  []byte
			want ID
		}{
			{name: "byte slice", val: []byte{0x6b, 0xa7, 0xb8, 0x10, 0x9d, 0xad, 0x11, 0xd1, 0x80, 0xb4, 0x00, 0xc0, 0x4f, 0xd4, 0x30, 0xc8}, want: refID},
			{name: "6ba7b8109dad11d180b400c04fd430c8", val: []byte("6ba7b8109dad11d180b400c04fd430c8"), want: refID},
			{name: "6ba7b810-9dad-11d1-80b4-00c04fd430c8", val: []byte("6ba7b810-9dad-11d1-80b4-00c04fd430c8"), want: refID},
			{name: "{6ba7b810-9dad-11d1-80b4-00c04fd430c8}", val: []byte("{6ba7b810-9dad-11d1-80b4-00c04fd430c8}"), want: refID},
			{name: "urn:uuid:6ba7b810-9dad-11d1-80b4-00c04fd430c8", val: []byte("urn:uuid:6ba7b810-9dad-11d1-80b4-00c04fd430c8"), want: refID},
			{name: "6BA7B8109DAD11D180B400C04FD430C8", val: []byte("6BA7B8109DAD11D180B400C04FD430C8"), want: refID},
			{name: "6BA7B810-9DAD-11D1-80B4-00C04FD430C8", val: []byte("6BA7B810-9DAD-11D1-80B4-00C04FD430C8"), want: refID},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if got := IDFrom(tt.val); !reflect.DeepEqual(got, tt.want) {
					t.Errorf("IDFrom() = %v, want %v", got, tt.want)
				}
				if got := IDMust(tt.val); !reflect.DeepEqual(got, tt.want) {
					t.Errorf("IDMust() = %v, want %v", got, tt.want)
				}
			})
		}
	}
	{
		tests := []struct {
			name string
			val  [16]byte
			want ID
		}{
			{name: "passthrough", val: refID, want: refID},
			{name: "nullable", val: refID.Nullable(), want: refID},
			{name: "byte array", val: [16]byte{0x6b, 0xa7, 0xb8, 0x10, 0x9d, 0xad, 0x11, 0xd1, 0x80, 0xb4, 0x00, 0xc0, 0x4f, 0xd4, 0x30, 0xc8}, want: refID},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if got := IDFrom(tt.val); !reflect.DeepEqual(got, tt.want) {
					t.Errorf("IDFrom() = %v, want %v", got, tt.want)
				}
				if got := IDMust(tt.val); !reflect.DeepEqual(got, tt.want) {
					t.Errorf("IDMust() = %v, want %v", got, tt.want)
				}
			})
		}
	}
}

func TestID_Less(t *testing.T) {
	tests := []struct {
		name string
		id   ID
		rhs  ID
		want bool
	}{
		{
			name: "equal",
			id:   ID{0x6b, 0xa7, 0xb8, 0x10, 0x9d, 0xad, 0x11, 0xd1, 0x80, 0xb4, 0x00, 0xc0, 0x4f, 0xd4, 0x30, 0xc8},
			rhs:  ID{0x6b, 0xa7, 0xb8, 0x10, 0x9d, 0xad, 0x11, 0xd1, 0x80, 0xb4, 0x00, 0xc0, 0x4f, 0xd4, 0x30, 0xc8},
			want: false,
		},
		{
			name: "less least significant byte",
			id:   ID{0x6b, 0xa7, 0xb8, 0x10, 0x9d, 0xad, 0x11, 0xd1, 0x80, 0xb4, 0x00, 0xc0, 0x4f, 0xd4, 0x30, 0xc8},
			rhs:  ID{0x6b, 0xa7, 0xb8, 0x10, 0x9d, 0xad, 0x11, 0xd1, 0x80, 0xb4, 0x00, 0xc0, 0x4f, 0xd4, 0x30, 0xc9},
			want: true,
		},
		{
			name: "greater least significant byte",
			id:   ID{0x6b, 0xa7, 0xb8, 0x10, 0x9d, 0xad, 0x11, 0xd1, 0x80, 0xb4, 0x00, 0xc0, 0x4f, 0xd4, 0x30, 0xc9},
			rhs:  ID{0x6b, 0xa7, 0xb8, 0x10, 0x9d, 0xad, 0x11, 0xd1, 0x80, 0xb4, 0x00, 0xc0, 0x4f, 0xd4, 0x30, 0xc8},
			want: false,
		},
		{
			name: "less most significant byte",
			id:   ID{0x6b, 0xa7, 0xb8, 0x10, 0x9d, 0xad, 0x11, 0xd1, 0x80, 0xb4, 0x00, 0xc0, 0x4f, 0xd4, 0x30, 0xc8},
			rhs:  ID{0x6c, 0xa7, 0xb8, 0x10, 0x9d, 0xad, 0x11, 0xd1, 0x80, 0xb4, 0x00, 0xc0, 0x4f, 0xd4, 0x30, 0xc9},
			want: true,
		},
		{
			name: "greater most significant byte",
			id:   ID{0x6c, 0xa7, 0xb8, 0x10, 0x9d, 0xad, 0x11, 0xd1, 0x80, 0xb4, 0x00, 0xc0, 0x4f, 0xd4, 0x30, 0xc8},
			rhs:  ID{0x6b, 0xa7, 0xb8, 0x10, 0x9d, 0xad, 0x11, 0xd1, 0x80, 0xb4, 0x00, 0xc0, 0x4f, 0xd4, 0x30, 0xc8},
			want: false,
		},
		// The cases above also pass with a byte order where the
		// last byte is the most significant, because the
		// most significant and least significant byte change together
		// or alone. The following cases differ in two bytes
		// where only the first one must decide the order.
		{
			name: "first byte outweighs last byte",
			id:   ID{0x00, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0xff},
			rhs:  ID{0x01, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0x00},
			want: true,
		},
		{
			name: "first byte outweighs second byte",
			id:   ID{0x01, 0x00, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			rhs:  ID{0x00, 0x01, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			want: false,
		},
		{
			name: "byte 7 outweighs byte 8",
			id:   ID{0, 0, 0, 0, 0, 0, 0, 0x00, 0xff, 0, 0, 0, 0, 0, 0, 0},
			rhs:  ID{0, 0, 0, 0, 0, 0, 0, 0x01, 0x00, 0, 0, 0, 0, 0, 0, 0},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.id.Less(tt.rhs); got != tt.want {
				t.Errorf("ID.Less() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestID_Less_CanonicalOrder checks that ID.Less sorts like the
// canonical byte representation and the canonical string format,
// so sorted IDs come out in the same order as from a database index
// or from sorting their strings. IDSlice.Sort and IDSet.String use Less,
// while IDSet.Strings and NullableIDCompare use the string and byte order.
func TestID_Less_CanonicalOrder(t *testing.T) {
	ids := IDSlice{IDNil, {0xff}, {15: 0xff}, {7: 0xff}, {8: 0xff}}
	for range 100 {
		ids = append(ids, IDv4(), IDv7())
	}
	for _, a := range ids {
		for _, b := range ids {
			wantLess := bytes.Compare(a[:], b[:]) < 0
			require.Equal(t, wantLess, a.Less(b), "%s.Less(%s)", a, b)
			require.Equal(t, wantLess, a.String() < b.String(), "string order of %s and %s", a, b)
			require.Equal(t, wantLess, NullableIDCompare(a.Nullable(), b.Nullable()) < 0, "NullableIDCompare(%s, %s)", a, b)
		}
	}
}

func TestID_PrettyPrint(t *testing.T) {
	tests := []struct {
		id   ID
		want string
	}{
		{id: IDNil, want: "00000000-0000-0000-0000-000000000000"},
		{id: IDMust("78c08786-f18d-442e-8598-30c9c59cc424"), want: "78c08786-f18d-442e-8598-30c9c59cc424"},
	}
	for _, tt := range tests {
		t.Run(tt.id.String(), func(t *testing.T) {
			got := pretty.Sprint(tt.id)
			if got != tt.want {
				t.Errorf("ID.PrettyPrint() = %q, want %q", got, tt.want)
			}
		})
	}
	// Test with pointer to ID
	for _, tt := range tests {
		t.Run(tt.id.String(), func(t *testing.T) {
			got := pretty.Sprint(&tt.id)
			if got != tt.want {
				t.Errorf("ID.PrettyPrint() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestID_V7Time(t *testing.T) {
	timestamp := time.UnixMilli(1743076688815)
	now := time.UnixMilli(time.Now().UnixMilli())
	tests := []struct {
		name   string
		id     ID
		want   time.Time
		wantOK bool
	}{
		{name: "nil UUID", id: IDNil, want: time.Time{}, wantOK: false},
		{name: "version 1 UUID", id: IDv1(), want: time.Time{}, wantOK: false},
		{name: "version 4 UUID", id: IDv4(), want: time.Time{}, wantOK: false},
		{name: "version 7 UUID fixed timestamp", id: IDv7WithTime(timestamp), want: timestamp, wantOK: true},
		{name: "version 7 UUID current timestamp", id: IDv7WithTime(now), want: now, wantOK: true},
		// RFC 9562 stores unsigned 48 bit milliseconds since 1970,
		// so the first and the last representable time round trip
		{name: "version 7 UUID Unix epoch", id: IDv7WithTime(time.UnixMilli(0)), want: time.UnixMilli(0), wantOK: true},
		{name: "version 7 UUID last 48 bit time", id: IDv7WithTime(time.UnixMilli(1<<48 - 1)), want: time.UnixMilli(1<<48 - 1), wantOK: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := tt.id.V7Time()
			// ok reports whether the UUID carries a timestamp at all,
			// so callers don't have to compare against a sentinel
			// zero time to detect a non-version-7 UUID.
			require.Equal(t, tt.wantOK, ok)
			require.Equal(t, tt.want, got)
		})
	}
}

// A time outside of what RFC 9562 can store, before 1970 or after the
// year 10889, keeps only its lower 48 bits and so doesn't round trip.
// The README documents it, this pins that it is a documented limit
// of the format, not a time that silently comes back correct.
func TestIDv7WithTime_OutOfRange(t *testing.T) {
	before1970 := time.Date(1960, 1, 1, 0, 0, 0, 0, time.UTC)
	got, ok := IDv7WithTime(before1970).V7Time()
	require.True(t, ok)
	require.False(t, got.Equal(before1970), "%s must not round trip", before1970)
}

// TestIDv7_RFC9562Layout checks the byte layout of the 48 bit
// v7 timestamp. RFC 9562 stores it big-endian in the first 6 bytes,
// which is what makes v7 IDs sortable by creation time and
// interoperable with v7 IDs generated by other implementations.
func TestIDv7_RFC9562Layout(t *testing.T) {
	// Test vector of RFC 9562 Appendix A.6:
	// unix_ts_ms 0x017F22E279B0 = 2022-02-22T19:22:22Z
	rfcVector := IDMust("017F22E2-79B0-7CC3-98C4-DC0C0C07398F")
	rfcTime := time.Date(2022, 2, 22, 19, 22, 22, 0, time.UTC)

	got, ok := rfcVector.V7Time()
	require.True(t, ok)
	require.True(t, rfcTime.Equal(got), "V7Time decodes the RFC test vector: got %s, want %s", got, rfcTime)

	// A v7 ID generated by another implementation
	got, ok = IDMust("019222e8-1ec3-7e4f-97e2-919670df6d6b").V7Time()
	require.True(t, ok)
	require.Equal(t, int64(0x019222e81ec3), got.UnixMilli(), "V7Time of foreign v7 ID")

	id := IDv7WithTime(rfcTime)
	require.Equal(t, rfcVector[:6], id[:6], "IDv7WithTime writes the timestamp big-endian")
	require.Equal(t, "017f22e2-79b0-7", id.String()[:15])

	id = IDv7Deterministic(rfcTime.UnixMilli())
	require.Equal(t, "017f22e2-79b0-7000-8000-000000000000", id.String(), "IDv7Deterministic writes the timestamp big-endian")
}

// TestIDv7_TimeSortable checks the main reason to use v7 IDs:
// IDs created later sort after IDs created earlier,
// in byte order (database indexes), string order and with ID.Less.
// The time steps cross the byte boundaries of the timestamp,
// where a wrong byte order breaks the sort order.
func TestIDv7_TimeSortable(t *testing.T) {
	start := time.Date(2024, 9, 24, 7, 21, 30, 0, time.UTC)
	var ids IDSlice
	for _, offset := range []int64{0, 1, 2, 255, 256, 257, 1 << 16, 1<<16 + 1, 1 << 24, 1 << 32, 1 << 40} {
		ts := start.Add(time.Duration(offset) * time.Millisecond)
		ids = append(ids, IDv7WithTime(ts))
		ids = append(ids, IDv7Deterministic(ts.UnixMilli()))
	}
	for i := 2; i < len(ids); i++ {
		prev, id := ids[i-2], ids[i]
		require.Negative(t, bytes.Compare(prev[:], id[:]), "byte order: %s before %s", prev, id)
		require.Less(t, prev.String(), id.String(), "string order")
		require.True(t, prev.Less(id), "ID.Less order: %s before %s", prev, id)
	}
}

// TestIDv7_V7TimeRoundTrip checks that V7Time returns the time
// passed to IDv7WithTime and IDv7Deterministic with millisecond precision
// for times that use all 6 bytes of the timestamp.
func TestIDv7_V7TimeRoundTrip(t *testing.T) {
	for _, ts := range []time.Time{
		time.UnixMilli(0),
		time.UnixMilli(1),
		time.UnixMilli(1<<48 - 1), // largest 48 bit timestamp
		time.Date(2022, 2, 22, 19, 22, 22, 123_456_789, time.UTC),
		time.Now(),
	} {
		want := time.UnixMilli(ts.UnixMilli())

		got, ok := IDv7WithTime(ts).V7Time()
		require.True(t, ok)
		require.True(t, want.Equal(got), "IDv7WithTime(%s).V7Time() = %s", want, got)

		got, ok = IDv7Deterministic(ts.UnixMilli()).V7Time()
		require.True(t, ok)
		require.True(t, want.Equal(got), "IDv7Deterministic(%d).V7Time() = %s", ts.UnixMilli(), got)
	}
}

func ExampleID_JSONSchema() {
	reflector := jsonschema.Reflector{DoNotReference: true}
	schema, _ := json.MarshalIndent(reflector.Reflect(ID{}), "", "  ")
	fmt.Println(string(schema))

	// Output:
	// {
	//   "$schema": "https://json-schema.org/draft/2020-12/schema",
	//   "$id": "https://github.com/domonda/go-types/uu/id",
	//   "type": "string",
	//   "format": "uuid",
	//   "title": "UUID"
	// }
}

func TestIDv7DeterministicFunc_Series(t *testing.T) {
	// The returned function is meant to replace random IDs in tests
	// (e.g. via ContextWithIDFunc), so it must produce the same
	// strictly increasing series of valid v7 IDs on every run.
	next := IDv7DeterministicFunc(1000)
	other := IDv7DeterministicFunc(1000)

	prev := IDNil
	for i := range int64(5) {
		id := next()
		require.NoError(t, id.Validate())
		require.Equal(t, IDv7Deterministic(1000+i), id, "deterministic series")
		require.Equal(t, id, other(), "independent series with same start are equal")
		ts, ok := id.V7Time()
		require.True(t, ok)
		require.Equal(t, 1000+i, ts.UnixMilli(), "timestamp counts up per call")
		require.True(t, prev.Less(id) || prev.IsNil(), "series is sortable")
		prev = id
	}
}

func TestIDv7DeterministicFunc_Concurrent(t *testing.T) {
	// Documented as safe for concurrent use: no ID may be
	// handed out twice and none may be skipped.
	const goroutines, perGoroutine = 8, 100
	next := IDv7DeterministicFunc(0)

	results := make(chan ID, goroutines*perGoroutine)
	done := make(chan struct{})
	for range goroutines {
		go func() {
			for range perGoroutine {
				results <- next()
			}
			done <- struct{}{}
		}()
	}
	for range goroutines {
		<-done
	}
	close(results)

	seen := make(IDSet)
	for id := range results {
		require.True(t, seen.Insert(id), "duplicate ID %s", id)
	}
	require.Len(t, seen, goroutines*perGoroutine)
	for i := range int64(goroutines * perGoroutine) {
		require.True(t, seen.Contains(IDv7Deterministic(i)), "missing ID number %d", i)
	}
}

func TestIDFromPtr(t *testing.T) {
	id := IDMust("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	def := IDMust("de8482f2-164a-40f9-b4e0-c349b78c22f6")

	require.Equal(t, id, IDFromPtr(&id, def))
	require.Equal(t, def, IDFromPtr(nil, def), "nil pointer returns default")
	require.Equal(t, IDNil, IDFromPtr(nil, IDNil))
}

func TestIDFromAny(t *testing.T) {
	refStr := "6ba7b810-9dad-11d1-80b4-00c04fd430c8"
	ref := IDMust(refStr)
	invalidVersion := ref
	invalidVersion.SetVersion(0)

	tests := []struct {
		name    string
		val     any
		want    ID
		wantErr error
	}{
		{name: "string", val: refStr, want: ref},
		{name: "hex string", val: ref.Hex(), want: ref},
		{name: "[]byte text", val: []byte(refStr), want: ref},
		{name: "[]byte binary", val: ref.Bytes(), want: ref},
		{name: "ID", val: ref, want: ref},
		{name: "NullableID", val: ref.Nullable(), want: ref},
		{name: "[16]byte", val: [16]byte(ref), want: ref},
		// Unlike NullableIDFromAny, a missing value is an error for a required ID
		{name: "nil", val: nil, want: IDNil, wantErr: ErrNilID},
		{name: "Nil ID", val: IDNil, want: IDNil, wantErr: ErrNilID},
		{name: "NULL NullableID", val: IDNull, want: IDNil, wantErr: ErrNilID},
		{name: "invalid version", val: invalidVersion, want: invalidVersion, wantErr: ErrInvalidVersion(0)},
		// Parsed strings and bytes must be validated like the typed IDs,
		// a well formed but invalid UUID is still not a usable ID
		{name: "Nil UUID string", val: IDNil.String(), want: IDNil, wantErr: ErrNilID},
		{name: "Nil UUID []byte text", val: []byte(IDNil.String()), want: IDNil, wantErr: ErrNilID},
		{name: "Nil UUID []byte binary", val: IDNil.Bytes(), want: IDNil, wantErr: ErrNilID},
		{name: "invalid version string", val: invalidVersion.String(), want: invalidVersion, wantErr: ErrInvalidVersion(0)},
		{name: "invalid version []byte text", val: []byte(invalidVersion.String()), want: invalidVersion, wantErr: ErrInvalidVersion(0)},
		{name: "invalid version []byte binary", val: invalidVersion.Bytes(), want: invalidVersion, wantErr: ErrInvalidVersion(0)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := IDFromAny(tt.val)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, tt.want, got)
		})
	}

	t.Run("errors", func(t *testing.T) {
		for _, val := range []any{"", "not-a-uuid", []byte("short"), 42, &ref} {
			_, err := IDFromAny(val)
			require.Error(t, err, "%#v", val)
		}
	})
}

func TestIDFrom_IDTypes(t *testing.T) {
	// IDFrom does not validate, so typed IDs must pass through unchanged
	ref := IDMust("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	require.Equal(t, ref, IDFrom(ref))
	require.Equal(t, ref, IDFrom(ref.Nullable()))
	require.Equal(t, IDNil, IDFrom(IDNil))
	require.Equal(t, IDNil, IDFrom(IDNull))
	require.Equal(t, IDNil, IDFrom("not-a-uuid"), "unparseable string returns IDNil")
	require.Equal(t, IDNil, IDFrom([]byte("short")), "unparseable bytes return IDNil")
}

func TestIDMust_Panics(t *testing.T) {
	// IDMust is for IDs that must be valid, so every invalid
	// input including the Nil UUID must panic instead of
	// silently producing an unusable ID.
	invalidVersion := IDMust("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	invalidVersion.SetVersion(0)

	require.Panics(t, func() { IDMust("not-a-uuid") })
	require.Panics(t, func() { IDMust([]byte("short")) })
	require.Panics(t, func() { IDMust(IDNil) })
	require.Panics(t, func() { IDMust([16]byte{}) })
	require.Panics(t, func() { IDMust(invalidVersion) })
	require.Panics(t, func() { IDMust(invalidVersion.Nullable()) })
	require.Panics(t, func() { IDMustFromString("not-a-uuid") })

	// Parsed strings and bytes are no exception
	require.Panics(t, func() { IDMust("00000000-0000-0000-0000-000000000000") })
	require.Panics(t, func() { IDMust([]byte("00000000-0000-0000-0000-000000000000")) })
	require.Panics(t, func() { IDMust(make([]byte, 16)) })
	require.Panics(t, func() { IDMust(invalidVersion.String()) })
	require.Panics(t, func() { IDMust([]byte(invalidVersion.String())) })
	require.Panics(t, func() { IDMust(invalidVersion.Bytes()) })
}

func TestID_IsZeroIsNotNil(t *testing.T) {
	id := IDMust("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	require.True(t, IDNil.IsZero())
	require.False(t, IDNil.IsNotNil())
	require.False(t, id.IsZero())
	require.True(t, id.IsNotNil())
}

func TestID_StringUpperAndHex(t *testing.T) {
	id := IDMust("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	require.Equal(t, "6BA7B810-9DAD-11D1-80B4-00C04FD430C8", id.StringUpper())
	require.Equal(t, "6ba7b8109dad11d180b400c04fd430c8", id.Hex())

	// Both representations must be parseable back to the same ID
	for _, s := range []string{id.StringUpper(), id.Hex()} {
		parsed, err := IDFromString(s)
		require.NoError(t, err)
		require.Equal(t, id, parsed)
	}

	require.Equal(t, "00000000000000000000000000000000", IDNil.Hex(), "Hex is always 32 characters")
}

func TestIDFromString_InvalidEncodings(t *testing.T) {
	// The content must be validated for every supported encoding,
	// not only the overall length and dash positions.
	for _, s := range []string{
		"6ba7b8109dad11d180b400c04fd430cx", // hex without dashes
		"a3fW6RxSR_amqRvJxqIi2!",           // base64 URL
		"a3fW6RxSR+amqRvJxqIi2g",           // standard base64 is not URL encoding
		"xba7b810-9dad-11d1-80b4-00c04fd430c8",
		"6ba7b810-xdad-11d1-80b4-00c04fd430c8",
		"6ba7b810-9dad-x1d1-80b4-00c04fd430c8",
		"6ba7b810-9dad-11d1-x0b4-00c04fd430c8",
		"6ba7b810-9dad-11d1-80b4-x0c04fd430c8",
		"urn:uuid:6ba7b810-9dad-11d1-80b4-00c04fd430cx",
		"urn:uuid:6ba7b810-9dad-11d1-80b4-00c04fd430c",
	} {
		_, err := IDFromString(s)
		require.Error(t, err, s)
	}
}
