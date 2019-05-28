// Copyright (C) 2013-2015 by Maxim Bublis <b@codemonkey.ru>
//
// Permission is hereby granted, free of charge, to any person obtaining
// a copy of this software and associated documentation files (the
// "Software"), to deal in the Software without restriction, including
// without limitation the rights to use, copy, modify, merge, publish,
// distribute, sublicense, and/or sell copies of the Software, and to
// permit persons to whom the Software is furnished to do so, subject to
// the following conditions:
//
// The above copyright notice and this permission notice shall be
// included in all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
// EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
// MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
// NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
// LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION
// OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION
// WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

// Package uuid provides implementation of Universally Unique Identifier (UUID).
// Supported versions are 1, 3, 4 and 5 (as specified in RFC 4122) and
// version 2 (as specified in DCE 1.1).
package uu

import (
	"bytes"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha1"
	"database/sql/driver"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
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

func (id ID) Nullable() NullableID {
	return NullableID{ID: id}
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
// xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
// String implements the fmt.Stringer interface.
func (id ID) String() string {
	return string(id.StringBytes())
}

// Hex returns the hex representation without dashes of the UUID
func (id ID) Hex() string {
	return hex.EncodeToString(id[:])
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
// "6ba7b8109dad11d180b400c04fd430c8",
// "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
// "{6ba7b810-9dad-11d1-80b4-00c04fd430c8}",
// "urn:uuid:6ba7b810-9dad-11d1-80b4-00c04fd430c8"
func (id *ID) UnmarshalText(text []byte) (err error) {
	if len(text) < 32 {
		return fmt.Errorf("uu.ID string too short: %s", text)
	}

	if len(text) == 32 {
		_, err = hex.Decode(id[:], text)
		return err
	}

	t := text[:]
	braced := false

	if bytes.Equal(t[:9], urnPrefix) {
		t = t[9:]
	} else if t[0] == '{' {
		braced = true
		t = t[1:]
	}

	b := id[:]

	for i, byteGroup := range byteGroups {
		if i > 0 {
			if t[0] != '-' {
				return fmt.Errorf("uu.ID: invalid string format")
			}
			t = t[1:]
		}

		if len(t) < byteGroup {
			return fmt.Errorf("uu.ID string too short: %s", text)
		}

		if i == 4 && len(t) > byteGroup &&
			((braced && t[byteGroup] != '}') || len(t[byteGroup:]) > 1 || !braced) {
			return fmt.Errorf("uu.ID string too long: %s", text)
		}

		_, err = hex.Decode(b[:byteGroup/2], t[:byteGroup])
		if err != nil {
			return err
		}

		t = t[byteGroup:]
		b = b[byteGroup/2:]
	}

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
	err = id.UnmarshalBinary(input)
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
