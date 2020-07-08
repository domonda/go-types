package uu

import (
	"bytes"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha1"
	"database/sql/driver"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"io"
	"net"
	"os"
	"sync"
	"time"
)

// UUID layout variants.
const (
	IDVariantNCS = iota
	IDVariantRFC4122
	IDVariantMicrosoft
	IDVariantInvalid
)

// UUID DCE domains.
const (
	IDDomainPerson = iota
	IDDomainGroup
	IDDomainOrg
)

const IDRegex = "[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}"

// Difference in 100-nanosecond intervals between
// UUID epoch (October 15, 1582) and Unix epoch (January 1, 1970).
const epochStart = 122192928000000000

// Used in string method conversion
const dash byte = '-'

// UUID v1/v2 storage.
var (
	storageMutex  sync.Mutex
	storageOnce   sync.Once
	epochFunc     = unixTimeFunc
	clockSequence uint16
	lastTime      uint64
	hardwareAddr  [6]byte
	posixUID      = uint32(os.Getuid())
	posixGID      = uint32(os.Getgid())
)

// String parse helpers.
var (
	urnPrefix  = []byte("urn:uuid:")
	byteGroups = []int{8, 4, 4, 4, 12}
)

func initClockSequence() {
	buf := make([]byte, 2)
	safeRandom(buf)
	clockSequence = binary.BigEndian.Uint16(buf)
}

func initHardwareAddr() {
	interfaces, err := net.Interfaces()
	if err == nil {
		for _, iface := range interfaces {
			if len(iface.HardwareAddr) >= 6 {
				copy(hardwareAddr[:], iface.HardwareAddr)
				return
			}
		}
	}

	// Initialize hardwareAddr randomly in case
	// of real network interfaces absence
	safeRandom(hardwareAddr[:])

	// Set multicast bit as recommended in RFC 4122
	hardwareAddr[0] |= 0x01
}

func initStorage() {
	initClockSequence()
	initHardwareAddr()
}

func safeRandom(dest []byte) {
	if _, err := rand.Read(dest); err != nil {
		panic(err)
	}
}

// Returns difference in 100-nanosecond intervals between
// UUID epoch (October 15, 1582) and current time.
// This is default epoch calculation function.
func unixTimeFunc() uint64 {
	return epochStart + uint64(time.Now().UnixNano()/100)
}

// ID is a UUID representation compliant with specification
// described in RFC 4122.
type ID [16]byte

// The nil UUID is special form of UUID that is specified to have all
// 128 bits set to zero.
var IDNil ID

// Predefined namespace UUIDs.
var (
	NamespaceDNS, _  = IDFromString("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	NamespaceURL, _  = IDFromString("6ba7b811-9dad-11d1-80b4-00c04fd430c8")
	NamespaceOID, _  = IDFromString("6ba7b812-9dad-11d1-80b4-00c04fd430c8")
	NamespaceX500, _ = IDFromString("6ba7b814-9dad-11d1-80b4-00c04fd430c8")
)

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
	return NullableIDFrom(id)
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
	hex.Encode(b[24:], id[10:])

	return b
}

// String returns the canonical string representation of the UUID:
//   xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
// String implements the fmt.Stringer interface.
func (id ID) String() string {
	return string(id.StringBytes())
}

// GoString returns a pseudo Go literal for the ID in the format:
//   uu.ID("xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx")
// It compiles if uu.ID is replace by uu.IDMustFromString.
func (id ID) GoString() string {
	return `uu.ID("` + id.String() + `")`
}

// PrettyPrint the ID in the format xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx.
// Implements pretty.Printer.
func (id ID) PrettyPrint(w io.Writer) {
	w.Write(id.StringBytes())
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
	return []byte(id.String()), nil
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

	var (
		newID    ID
		original = text
	)

	if bytes.HasPrefix(text, []byte("urn:uuid:")) {
		text = text[9:]
		if len(text) != 36 {
			return fmt.Errorf("uu.ID string has wrong length: %q", original)
		}
		newID, err = parseDashedFormat(text, original)
		if err != nil {
			return err
		}
		*id = newID
		return nil
	}

	if text[0] == '"' && text[len(text)-1] == '"' {
		text = text[1 : len(text)-1]
	}
	if text[0] == '{' && text[len(text)-1] == '}' {
		text = text[1 : len(text)-1]
	}

	switch len(text) {
	case 22:
		_, err = base64.RawURLEncoding.Decode(newID[:], text)
		if err != nil {
			return fmt.Errorf("uu.ID string %q base64 decoding error: %w", original, err)
		}

	case 32:
		_, err = hex.Decode(newID[:], text)
		if err != nil {
			return fmt.Errorf("uu.ID string %q hex decoding error: %w", original, err)
		}

	case 36:
		newID, err = parseDashedFormat(text, original)
		if err != nil {
			return err
		}

	default:
		return fmt.Errorf("uu.ID string has wrong length: %q", original)
	}

	*id = newID
	return nil
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

// IDFromPtr returns the dereferenced value of ptr,
// or nilVal if ptr is nil.
func IDFromPtr(ptr *ID, nilVal ID) ID {
	if ptr == nil {
		return nilVal
	}
	return *ptr
}

// IDFromBytes returns an ID converted from raw byte slice input.
// It will return error if the slice isn't 16 bytes long.
func IDFromBytes(input []byte) (id ID, err error) {
	switch len(input) {
	case 16:
		err = id.UnmarshalBinary(input)

	case 22, 32, 36:
		err = id.UnmarshalText(input)

	default:
		return IDNil, fmt.Errorf("uu.IDFromBytes expects 16, 32, or 36 bytes, but got %d", len(input))
	}

	return id, err
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
func IDFromString(input string) (id ID, err error) {
	err = id.UnmarshalText([]byte(input))
	return id, err
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
	id := idFromHash(sha1.New(), ns, name)
	id.SetVersion(5)
	id.SetVariant()
	return id
}

// Returns UUID based on hashing of namespace UUID and name.
func idFromHash(h hash.Hash, ns ID, name string) (id ID) {
	h.Write(ns[:])
	h.Write([]byte(name))
	copy(id[:], h.Sum(nil))
	return id
}

// IDCompare returns bytes.Compare result of a and b.
func IDCompare(a, b ID) int {
	return bytes.Compare(a[:], b[:])
}

// TODO test
// func idCompareOptimized(a, b ID) int {
// 	aWords := (*[2]uint64)(unsafe.Pointer(&a[0]))
// 	bWords := (*[2]uint64)(unsafe.Pointer(&b[0]))
// 	if aWords[1] < bWords[1] {
// 		return -1
// 	}
// 	if aWords[1] > bWords[1] {
// 		return +1
// 	}
// 	if aWords[0] < bWords[0] {
// 		return -1
// 	}
// 	if aWords[0] > bWords[0] {
// 		return +1
// 	}
// 	return 0
// }
