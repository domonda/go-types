package uu

import (
	"bytes"
	"crypto/md5"  //#nosec G501 -- Needed for standard conform IDv3
	"crypto/sha1" //#nosec G505 -- Needed for standard conform IDv5
	"database/sql/driver"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"io"
	"strings"
	"unsafe"
)

// The nil UUID is special form of UUID that is specified to have all
// 128 bits set to zero: "00000000-0000-0000-0000-000000000000"
var IDNil ID

// ID is a UUID representation compliant with specification
// described in RFC 4122.
type ID [16]byte

// IDv1 returns an ID based on current timestamp and MAC address.
func IDv1() (id ID) {
	timeNow, clockSeq, hardwareAddr := getStorage()

	binary.BigEndian.PutUint32(id[0:], uint32(timeNow))
	binary.BigEndian.PutUint16(id[4:], uint16(timeNow>>32))
	binary.BigEndian.PutUint16(id[6:], uint16(timeNow>>48))
	binary.BigEndian.PutUint16(id[8:], clockSeq)

	copy(id[10:], hardwareAddr)

	id.SetVersion(1)
	id.SetVariant()

	return id
}

// IDv2 returns DCE Security UUID based on POSIX UID/GID.
func IDv2(domain byte) (id ID) {
	timeNow, clockSeq, hardwareAddr := getStorage()

	switch domain {
	case IDDomainPerson:
		binary.BigEndian.PutUint32(id[0:], posixUID)
	case IDDomainGroup:
		binary.BigEndian.PutUint32(id[0:], posixGID)
	}

	binary.BigEndian.PutUint16(id[4:], uint16(timeNow>>32))
	binary.BigEndian.PutUint16(id[6:], uint16(timeNow>>48))
	binary.BigEndian.PutUint16(id[8:], clockSeq)
	id[9] = domain

	copy(id[10:], hardwareAddr)

	id.SetVersion(2)
	id.SetVariant()

	return id
}

// IDv3 returns an ID based on MD5 hash of namespace UUID and name.
func IDv3(ns ID, name string) ID {
	//#nosec G401 -- Needed for standard conformity
	id := idFromHash(md5.New(), ns, name)
	id.SetVersion(3)
	id.SetVariant()
	return id
}

// IDv4 returns random generated UUID.
func IDv4() (id ID) {
	safeRandom(id[:])
	id.SetVersion(4)
	id.SetVariant()
	return id
}

// IDv5 returns an ID based on SHA-1 hash of namespace UUID and name.
func IDv5(ns ID, name string) ID {
	//#nosec G401 -- Needed for standard conformity
	id := idFromHash(sha1.New(), ns, name)
	id.SetVersion(5)
	id.SetVariant()
	return id
}

// IDFromBytes parses a byte slice as UUID.
// If the slice has a length of 16, it will be interpred as a binary UUID,
// if the length is 22, 32, or 36, it will be parsed as string.
func IDFromBytes(b []byte) (ID, error) {
	if len(b) < 16 {
		return IDNil, fmt.Errorf("uu.ID %q is too short", b)
	}
	if len(b) == 16 {
		var id ID
		copy(id[:], b)
		return id, nil
	}

	if bytes.HasPrefix(b, []byte("urn:uuid:")) {
		text := b[9:]
		if len(text) != 36 {
			return IDNil, fmt.Errorf("uu.ID string has wrong length: %q", b)
		}
		return parseDashedFormat(text, b)
	}

	text := b
	if text[0] == '"' && text[len(text)-1] == '"' {
		text = text[1 : len(text)-1]
	}
	if text[0] == '{' && text[len(text)-1] == '}' {
		text = text[1 : len(text)-1]
	}

	switch len(text) {
	case 22:
		var id ID
		_, err := base64.RawURLEncoding.Decode(id[:], text)
		if err != nil {
			return IDNil, fmt.Errorf("uu.ID string %q base64 decoding error: %w", b, err)
		}
		return id, nil

	case 32:
		var id ID
		_, err := hex.Decode(id[:], text)
		if err != nil {
			return IDNil, fmt.Errorf("uu.ID string %q hex decoding error: %w", b, err)
		}
		return id, nil

	case 36:
		return parseDashedFormat(text, b)

	default:
		return IDNil, fmt.Errorf("uu.IDFromBytes expects 16, 22, 32, or 36 bytes, but got %d", len(b))
	}
}

// IDFromBytesOrNil returns an ID converted from raw byte slice input.
// Same behavior as IDFromBytes, but returns a Nil UUID on error.
func IDFromBytesOrNil(input []byte) ID {
	id, err := IDFromBytes(input)
	if err != nil {
		return IDNil
	}
	return id
}

// IDFromString returns an ID parsed from string input.
// Input is expected in a form accepted by UnmarshalText.
func IDFromString(text string) (ID, error) {
	if len(text) < 22 {
		return IDNil, fmt.Errorf("uu.ID string too short: %q", text)
	}
	return IDFromBytes([]byte(text))
}

// IDFromStringOrNil returns an ID parsed from string input.
// Same behavior as FromString, but returns a Nil UUID on error.
func IDFromStringOrNil(input string) ID {
	id, err := IDFromString(input)
	if err != nil {
		return IDNil
	}
	return id
}

// IDMustFromString returns an ID parsed from string input.
// Panics if there is an error
func IDMustFromString(input string) ID {
	id, err := IDFromString(input)
	if err != nil {
		panic(err)
	}
	return id
}

// IDFromPtr returns the dereferenced value of ptr,
// or defaultVal if ptr is nil.
func IDFromPtr(ptr *ID, defaultVal ID) ID {
	if ptr == nil {
		return defaultVal
	}
	return *ptr
}

// IDFrom converts val to an ID or returns IDNil
// if that's not possible.
// Supported types are string, []byte, [16]byte,
// ID, NullableID, and nil.
// The returned ID is not validated.
func IDFrom(val interface{}) ID {
	switch x := val.(type) {
	case string:
		return IDFromStringOrNil(x)
	case []byte:
		return IDFromBytesOrNil(x)
	case ID:
		return x
	case NullableID:
		return ID(x)
	case [16]byte:
		return ID(x)
	default:
		return IDNil
	}
}

// IDMust converts val to an ID or panics
// if that's not possible or the ID is not valid.
// Supported types are string, []byte, [16]byte,
// ID, NullableID, and nil.
func IDMust(val interface{}) ID {
	switch x := val.(type) {
	case string:
		id, err := IDFromString(x)
		if err != nil {
			panic(err)
		}
		return id
	case []byte:
		id, err := IDFromBytes(x)
		if err != nil {
			panic(err)
		}
		return id
	case ID:
		if err := x.Validate(); err != nil {
			panic(err)
		}
		return x
	case NullableID:
		if err := x.Validate(); err != nil {
			panic(err)
		}
		return ID(x)
	case [16]byte:
		if err := ID(x).Validate(); err != nil {
			panic(err)
		}
		return ID(x)
	case nil:
		return IDNil
	default:
		panic(fmt.Errorf("uu.IDMust type not supported: %T", val))
	}
}

// Version returns algorithm version used to generate UUID.
func (id ID) Version() uint {
	return uint(id[6] >> 4)
}

// Variant returns an ID layout variant or IDVariantInvalid if unknown.
func (id ID) Variant() uint {
	switch {
	case (id[8] & 0x80) == 0x00:
		return IDVariantNCS
	case (id[8]&0xc0)|0x80 == 0x80:
		return IDVariantRFC4122
	case (id[8]&0xe0)|0xc0 == 0xc0:
		return IDVariantMicrosoft
	}
	return IDVariantInvalid
}

// Valid returns if Variant and Version of this UUID are supported.
// A Nil UUID is not valid.
func (id ID) Valid() bool {
	v := id.Version()
	return v >= 1 && v <= 5 && id.Variant() != IDVariantInvalid
}

// Validate returns an error if the Variant and Version of this UUID are not supported.
// A Nil UUID is not valid.
func (id ID) Validate() error {
	if v := id.Version(); v < 1 || v > 5 {
		return fmt.Errorf("invalid UUID version: %d", v)
	}
	if id.Variant() == IDVariantInvalid {
		return errors.New("invalid UUID variant")
	}
	return nil
}

// IsNil returns if the id is the Nil UUID value (all zeros)
func (id ID) IsNil() bool {
	return id == IDNil
}

// IsNotNil returns if the id is not the Nil UUID value (all zeros)
func (id ID) IsNotNil() bool {
	return id != IDNil
}

// Nullable returns the ID as NullableID
func (id ID) Nullable() NullableID {
	return NullableID(id)
}

// Bytes returns bytes slice representation of UUID.
func (id ID) Bytes() []byte {
	return id[:]
}

// StringBytes returns the canonical string representation of the UUID as byte slice:
// xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
func (id ID) StringBytes() []byte {
	b := make([]byte, 36)

	hex.Encode(b[0:8], id[0:4])
	b[8] = dash
	hex.Encode(b[9:13], id[4:6])
	b[13] = dash
	hex.Encode(b[14:18], id[6:8])
	b[18] = dash
	hex.Encode(b[19:23], id[8:10])
	b[23] = dash
	hex.Encode(b[24:36], id[10:16])

	return b
}

// String returns the canonical string representation of the UUID:
//   xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
// String implements the fmt.Stringer interface.
func (id ID) String() string {
	b := make([]byte, 36)

	hex.Encode(b[0:8], id[0:4])
	b[8] = dash
	hex.Encode(b[9:13], id[4:6])
	b[13] = dash
	hex.Encode(b[14:18], id[6:8])
	b[18] = dash
	hex.Encode(b[19:23], id[8:10])
	b[23] = dash
	hex.Encode(b[24:36], id[10:16])

	return string(b)
}

// StringUpper returns the upper case version
// of the canonical string format:
//   XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX
func (id ID) StringUpper() string {
	return strings.ToUpper(id.String())
}

// GoString returns a pseudo Go literal for the ID in the format:
//   uu.IDMust(`xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx`)
func (id ID) GoString() string {
	return "uu.IDMust(`" + id.String() + "`)"
}

// PrettyPrint the ID in the format xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx.
// Implements the pretty.Printable interface.
func (id ID) PrettyPrint(w io.Writer) {
	w.Write(id.StringBytes()) //#nosec G104 -- go-pretty does not check write errors
}

// Hex returns the hex representation without dashes of the UUID
// The returned string is always 32 characters long.
func (id ID) Hex() string {
	return hex.EncodeToString(id[:])
}

// Base64 returns the unpadded base64 URL encoding of the UUID.
// The returned string is always 22 characters long.
func (id ID) Base64() string {
	return base64.RawURLEncoding.EncodeToString(id[:])
}

// SetVersion sets version bits.
func (id *ID) SetVersion(v byte) {
	id[6] = (id[6] & 0x0f) | (v << 4)
}

// SetVariant sets variant bits as described in RFC 4122.
func (id *ID) SetVariant() {
	id[8] = (id[8] & 0xbf) | 0x80
}

// MarshalText implements the encoding.TextMarshaler interface.
// The encoding is the same as returned by String.
func (id ID) MarshalText() (text []byte, err error) {
	return id.StringBytes(), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
// Following formats are supported:
// `6ba7b8109dad11d180b400c04fd430c8`
// `6ba7b810-9dad-11d1-80b4-00c04fd430c8`
// `"6ba7b810-9dad-11d1-80b4-00c04fd430c8"`
// `{6ba7b810-9dad-11d1-80b4-00c04fd430c8}`
// `urn:uuid:6ba7b810-9dad-11d1-80b4-00c04fd430c`
// Surrounding double quotes will be removed before parsing.
func (id *ID) UnmarshalText(text []byte) (err error) {
	if len(text) < 22 {
		return fmt.Errorf("uu.ID string too short: %q", text)
	}
	newID, err := IDFromBytes(text)
	if err != nil {
		return err
	}
	*id = newID
	return nil
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
func (id ID) MarshalBinary() (data []byte, err error) {
	return id[:], nil
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
// It will return error if the slice isn't 16 bytes long,
// but does not check the validity of the UUID.
func (id *ID) UnmarshalBinary(data []byte) (err error) {
	if len(data) != 16 {
		return fmt.Errorf("uu.ID must be exactly 16 bytes long, got %d bytes", len(data))
	}
	copy(id[:], data)
	return nil
}

// Value implements the driver.Valuer interface.
func (id ID) Value() (driver.Value, error) {
	return id.String(), nil
}

// Scan implements the sql.Scanner interface.
// A 16-byte slice is handled by UnmarshalBinary, while
// a longer byte slice or a string is handled by UnmarshalText.
func (id *ID) Scan(src interface{}) error {
	switch src := src.(type) {
	case []byte:
		if len(src) == 16 {
			return id.UnmarshalBinary(src)
		}
		return id.UnmarshalText(src)

	case string:
		return id.UnmarshalText([]byte(src))
	}

	return fmt.Errorf("cannot convert %T to uu.ID", src)
}

// Returns UUID v1/v2 storage state.
// Returns epoch timestamp, clock sequence, and hardware address.
func getStorage() (uint64, uint16, []byte) {
	storageOnce.Do(initStorage)

	storageMutex.Lock()
	defer storageMutex.Unlock()

	timeNow := epochFunc()
	// Clock changed backwards since last UUID generation.
	// Should increase clock sequence.
	if timeNow <= lastTime {
		clockSequence++
	}
	lastTime = timeNow

	return timeNow, clockSequence, hardwareAddr[:]
}

// Returns UUID based on hashing of namespace UUID and name.
func idFromHash(h hash.Hash, ns ID, name string) (id ID) {
	h.Write(ns[:])
	h.Write([]byte(name))
	copy(id[:], h.Sum(nil))
	return id
}

// Less returns true if the 128 bit unsigned integer value
// of the id is less than the passed rhs.
func (id ID) Less(rhs ID) bool {
	l := (*[2]uint64)(unsafe.Pointer(&id[0]))  //#nosec G103 -- unsafe OK
	r := (*[2]uint64)(unsafe.Pointer(&rhs[0])) //#nosec G103 -- unsafe OK
	return l[1] < r[1] || (l[1] == r[1] && l[0] < r[0])
}

func parseDashedFormat(text, original []byte) (newID ID, err error) {
	if text[8] != '-' || text[13] != '-' || text[18] != '-' || text[23] != '-' {
		return IDNil, fmt.Errorf("invalid UUID string format: %q", original)
	}

	_, err = hex.Decode(newID[0:4], text[0:8])
	if err != nil {
		return IDNil, fmt.Errorf("uu.ID string %q hex decoding error: %w", original, err)
	}
	_, err = hex.Decode(newID[4:6], text[9:13])
	if err != nil {
		return IDNil, fmt.Errorf("uu.ID string %q hex decoding error: %w", original, err)
	}
	_, err = hex.Decode(newID[6:8], text[14:18])
	if err != nil {
		return IDNil, fmt.Errorf("uu.ID string %q hex decoding error: %w", original, err)
	}
	_, err = hex.Decode(newID[8:10], text[19:23])
	if err != nil {
		return IDNil, fmt.Errorf("uu.ID string %q hex decoding error: %w", original, err)
	}
	_, err = hex.Decode(newID[10:16], text[24:36])
	if err != nil {
		return IDNil, fmt.Errorf("uu.ID string %q hex decoding error: %w", original, err)
	}

	return newID, nil
}
