package uu

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

// And returns result of binary AND of two UUIDs.
func And(u1 ID, u2 ID) ID {
	u := ID{}
	for i := 0; i < 16; i++ {
		u[i] = u1[i] & u2[i]
	}
	return u
}

// Or returns result of binary OR of two UUIDs.
func Or(u1 ID, u2 ID) ID {
	u := ID{}
	for i := 0; i < 16; i++ {
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

func TestEqual(t *testing.T) {
	if NamespaceDNS != NamespaceDNS {
		t.Errorf("Incorrect comparison of %s and %s", NamespaceDNS, NamespaceDNS)
	}

	if NamespaceDNS == NamespaceURL {
		t.Errorf("Incorrect comparison of %s and %s", NamespaceDNS, NamespaceURL)
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
	b1 := []byte{0x6b, 0xa7, 0xb8, 0x10, 0x9d, 0xad, 0x11, 0xd1, 0x80, 0xb4, 0x00, 0xc0, 0x4f, 0xd4, 0x30, 0xc8}

	u1, err := IDFromBytes(b1)
	if err != nil {
		t.Errorf("Error parsing UUID from bytes: %s", err)
	}

	if u != u1 {
		t.Errorf("UUIDs should be equal: %s and %s", u, u1)
	}

	b2 := []byte{}

	_, err = IDFromBytes(b2)
	if err == nil {
		t.Errorf("Should return error parsing from empty byte slice, got %s", err)
	}

	u3, err := IDFromBytes(u.StringBytes())
	assert.NoError(t, err)
	assert.Equal(t, u, u3)

	u4, err := IDFromBytes([]byte("6ba7b810-9dad-11d1-80b4-00c04fd430c8"))
	assert.NoError(t, err)
	assert.Equal(t, u, u4)

	u5, err := IDFromBytes([]byte("6ba7b8109dad11d180b400c04fd430c8"))
	assert.NoError(t, err)
	assert.Equal(t, u, u5)
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

	s0 := "6ba7b8109dad11d180b400c04fd430c8"
	s1 := "6ba7b810-9dad-11d1-80b4-00c04fd430c8"
	s2 := "{6ba7b810-9dad-11d1-80b4-00c04fd430c8}"
	s3 := "urn:uuid:6ba7b810-9dad-11d1-80b4-00c04fd430c8"

	_, err := IDFromString("")
	if err == nil {
		t.Errorf("Should return error trying to parse empty string, got %s", err)
	}

	u0, err := IDFromString(s0)
	if err != nil {
		t.Errorf("Error parsing UUID from string: %s", err)
	}

	if u != u0 {
		t.Errorf("UUIDs should be equal: %s and %s", u, u0)
	}

	u1, err := IDFromString(s1)
	if err != nil {
		t.Errorf("Error parsing UUID from string: %s", err)
	}

	if u != u1 {
		t.Errorf("UUIDs should be equal: %s and %s", u, u1)
	}

	u2, err := IDFromString(s2)
	if err != nil {
		t.Errorf("Error parsing UUID from string: %s", err)
	}

	if u != u2 {
		t.Errorf("UUIDs should be equal: %s and %s", u, u2)
	}

	u3, err := IDFromString(s3)
	if err != nil {
		t.Errorf("Error parsing UUID from string: %s", err)
	}

	if u != u3 {
		t.Errorf("UUIDs should be equal: %s and %s", u, u3)
	}
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
			if l == 22 {
				// 22 characters substring of UUID is valid base64
				// so we won't get a "too short" error
				continue
			}

			id, err := IDFromString(str[:l])
			if err == nil {
				t.Errorf("Should return error trying to parse too short string, got UUID %s from string %q", id, str[:l])
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
	b1 := []byte("6ba7b810-9dad-11d1-80b4-00c04fd430c8")

	u1 := ID{}
	err := u1.UnmarshalText(b1)
	if err != nil {
		t.Errorf("Error unmarshaling UUID: %s", err)
	}

	if u != u1 {
		t.Errorf("UUIDs should be equal: %s and %s", u, u1)
	}

	b2 := []byte("")
	u2 := ID{}

	err = u2.UnmarshalText(b2)
	if err == nil {
		t.Errorf("Should return error trying to unmarshal from empty string")
	}
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
	u := IDv3(NamespaceDNS, "www.example.com")

	if u.Version() != 3 {
		t.Errorf("UUIDv3 generated with incorrect version: %d", u.Version())
	}

	if u.Variant() != IDVariantRFC4122 {
		t.Errorf("UUIDv3 generated with incorrect variant: %d", u.Variant())
	}

	if u.String() != "5df41881-3aed-3515-88a7-2f4a814cf09e" {
		t.Errorf("UUIDv3 generated incorrectly: %s", u.String())
	}

	u = IDv3(NamespaceDNS, "python.org")

	if u.String() != "6fa459ea-ee8a-3ca4-894e-db77e160355e" {
		t.Errorf("UUIDv3 generated incorrectly: %s", u.String())
	}

	u1 := IDv3(NamespaceDNS, "golang.org")
	u2 := IDv3(NamespaceDNS, "golang.org")
	if u1 != u2 {
		t.Errorf("UUIDv3 generated different UUIDs for same namespace and name: %s and %s", u1, u2)
	}

	u3 := IDv3(NamespaceDNS, "example.com")
	if u1 == u3 {
		t.Errorf("UUIDv3 generated same UUIDs for different names in same namespace: %s and %s", u1, u2)
	}

	u4 := IDv3(NamespaceURL, "golang.org")
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
	u := IDv5(NamespaceDNS, "www.example.com")

	if u.Version() != 5 {
		t.Errorf("UUIDv5 generated with incorrect version: %d", u.Version())
	}

	if u.Variant() != IDVariantRFC4122 {
		t.Errorf("UUIDv5 generated with incorrect variant: %d", u.Variant())
	}

	u = IDv5(NamespaceDNS, "python.org")

	if u.String() != "886313e1-3b8a-5372-9b90-0c9aee199e5d" {
		t.Errorf("UUIDv5 generated incorrectly: %s", u.String())
	}

	u1 := IDv5(NamespaceDNS, "golang.org")
	u2 := IDv5(NamespaceDNS, "golang.org")
	if u1 != u2 {
		t.Errorf("UUIDv5 generated different UUIDs for same namespace and name: %s and %s", u1, u2)
	}

	u3 := IDv5(NamespaceDNS, "example.com")
	if u1 == u3 {
		t.Errorf("UUIDv5 generated same UUIDs for different names in same namespace: %s and %s", u1, u2)
	}

	u4 := IDv5(NamespaceURL, "golang.org")
	if u1 == u4 {
		t.Errorf("UUIDv3 generated same UUIDs for sane names in different namespaces: %s and %s", u1, u4)
	}
}

func TestID_GoString(t *testing.T) {
	tests := []struct {
		name string
		id   ID
		want string
	}{
		{"9256978d-18e6-4435-ad16-d7046d41b71a", IDMustFromString("9256978d-18e6-4435-ad16-d7046d41b71a"), "uu.IDMust(`9256978d-18e6-4435-ad16-d7046d41b71a`)"},
		{"a92bb308-f0f9-43d2-b31d-c0962198c31c", IDMustFromString("a92bb308-f0f9-43d2-b31d-c0962198c31c"), "uu.IDMust(`a92bb308-f0f9-43d2-b31d-c0962198c31c`)"},
		{"59a268e5-d820-4884-b120-10d1a9b0dd00", IDMustFromString("59a268e5-d820-4884-b120-10d1a9b0dd00"), "uu.IDMust(`59a268e5-d820-4884-b120-10d1a9b0dd00`)"},
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
	for i := 0; i < 100; i++ {
		id := IDv4()
		t.Run(id.String(), func(t *testing.T) {
			b := id.Base64()
			assert.Len(t, b, 22, "always 22 characters")
			parsed, err := IDFromString(b)
			assert.NoError(t, err, "can parse")
			assert.Equal(t, id, parsed, "parsed base64 UUID equal to original")
		})
	}
}

func TestIDFrom(t *testing.T) {
	refID := ID{0x6b, 0xa7, 0xb8, 0x10, 0x9d, 0xad, 0x11, 0xd1, 0x80, 0xb4, 0x00, 0xc0, 0x4f, 0xd4, 0x30, 0xc8}
	tests := []struct {
		name string
		val  interface{}
		want ID
	}{
		{name: "nil", val: nil, want: IDNil},
		{name: "passthrough", val: refID, want: refID},
		{name: "nullable", val: refID.Nullable(), want: refID},
		{name: "6ba7b8109dad11d180b400c04fd430c8", val: "6ba7b8109dad11d180b400c04fd430c8", want: refID},
		{name: "6ba7b810-9dad-11d1-80b4-00c04fd430c8", val: "6ba7b810-9dad-11d1-80b4-00c04fd430c8", want: refID},
		{name: "{6ba7b810-9dad-11d1-80b4-00c04fd430c8}", val: "{6ba7b810-9dad-11d1-80b4-00c04fd430c8}", want: refID},
		{name: "urn:uuid:6ba7b810-9dad-11d1-80b4-00c04fd430c8", val: "urn:uuid:6ba7b810-9dad-11d1-80b4-00c04fd430c8", want: refID},
		{name: "6BA7B8109DAD11D180B400C04FD430C8", val: "6BA7B8109DAD11D180B400C04FD430C8", want: refID},
		{name: "6BA7B810-9DAD-11D1-80B4-00C04FD430C8", val: "6BA7B810-9DAD-11D1-80B4-00C04FD430C8", want: refID},
		{name: "byte slice", val: []byte{0x6b, 0xa7, 0xb8, 0x10, 0x9d, 0xad, 0x11, 0xd1, 0x80, 0xb4, 0x00, 0xc0, 0x4f, 0xd4, 0x30, 0xc8}, want: refID},
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.id.Less(tt.rhs); got != tt.want {
				t.Errorf("ID.Less() = %v, want %v", got, tt.want)
			}
		})
	}
}
