package uu

import (
	"crypto/rand"
	"encoding/binary"
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

// Predefined namespace UUIDs.
var (
	NamespaceDNS  = IDMustFromString("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	NamespaceURL  = IDMustFromString("6ba7b811-9dad-11d1-80b4-00c04fd430c8")
	NamespaceOID  = IDMustFromString("6ba7b812-9dad-11d1-80b4-00c04fd430c8")
	NamespaceX500 = IDMustFromString("6ba7b814-9dad-11d1-80b4-00c04fd430c8")
)

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
