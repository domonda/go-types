package email

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"encoding/xml"
	"net/mail"
	"strings"

	"github.com/domonda/go-errs"
	"github.com/domonda/go-types"
	"github.com/domonda/go-types/internal"
	"github.com/domonda/go-types/strutil"
)

// Compile-time check that AddressList implements types.NormalizableValidator[AddressList]
var _ types.NormalizableValidator[AddressList] = AddressList("")

// NormalizeAddressList parses a comma-separated email address list using lenient validation
// that fixes common malformations and normalizes address parts to lowercase.
// Duplicate addresses (based on normalized address parts) are automatically removed.
// Returns an error if any address of the list can't be parsed.
// An empty list (or an "undisclosed-recipients" placeholder)
// returns nil without an error.
func NormalizeAddressList(list string) (normalized []string, err error) {
	addrs, err := ParseAddressList(list)
	if err != nil {
		return nil, err
	}
	if len(addrs) == 0 {
		return nil, nil
	}
	appended := make(map[string]bool, len(addrs))
	normalized = make([]string, 0, len(addrs))
	for _, a := range addrs {
		if appended[a.Address] {
			continue
		}
		normalized = append(normalized, string(AddressFrom(a)))
		appended[a.Address] = true
	}
	return normalized, nil
}

// AddressList represents a comma-separated list of email addresses.
// The list is meant to contain at least one email address,
// use NullableAddressList for lists that can be empty.
// Validate and Valid accept an empty list.
//
// AddressList is a collection: every format reads and writes the list
// as stored, because normalizing would rewrite every entry. Scan and
// UnmarshalJSON also read a PostgreSQL array, a JSON array and a JSON
// object with the addresses as keys, see Scan. Value writes the list
// as a plain string that works with every database, SQL NULL for a blank
// list, and returns an error for an invalid list. MarshalJSON writes the
// list as JSON string, so MarshalJSON and UnmarshalJSON round-trip every
// list. Use Normalized to normalize the list. See "Reading and writing
// values" in the README of the module.
type AddressList string

// AddressListJoin creates an AddressList by joining multiple Address values
// with comma separators. Empty addresses are skipped.
func AddressListJoin(addrs ...Address) AddressList {
	var b strings.Builder
	for _, addr := range addrs {
		if len(addr) == 0 {
			continue
		}
		if b.Len() > 0 {
			b.WriteString(", ")
		}
		b.WriteString(string(addr))
	}
	return AddressList(b.String())
}

// AddressListJoinStrings creates an AddressList by joining multiple string values
// with comma separators. Empty strings are skipped.
func AddressListJoinStrings(addrs ...string) AddressList {
	var b strings.Builder
	for _, addr := range addrs {
		if len(addr) == 0 {
			continue
		}
		if b.Len() > 0 {
			b.WriteString(", ")
		}
		b.WriteString(addr)
	}
	return AddressList(b.String())
}

// Append adds additional addresses to the AddressList.
// Empty addresses are skipped.
func (l AddressList) Append(addrs ...Address) AddressList {
	var b strings.Builder
	b.WriteString(string(l))
	for _, addr := range addrs {
		if len(addr) == 0 {
			continue
		}
		if b.Len() > 0 {
			b.WriteString(", ")
		}
		b.WriteString(string(addr))
	}
	return AddressList(b.String())
}

// Parse converts the AddressList to a slice of *mail.Address using lenient validation.
func (l AddressList) Parse() ([]*mail.Address, error) {
	return ParseAddressList(string(l))
}

// SplitNormalized parses the AddressList and returns its addresses
// normalized, see Normalized.
//
// All successfully parsed addresses are returned even when other
// addresses in the list can't be parsed. Parsing errors are collected
// and returned joined (see ParseAddressList), so a non-nil error can
// accompany a non-empty result slice. Returns nil if the list is empty.
// Use SplitRaw to get the addresses as stored.
func (l AddressList) SplitNormalized() ([]Address, error) {
	parsed, err := l.Parse()
	if len(parsed) == 0 {
		return nil, err
	}
	a := make([]Address, len(parsed))
	for i, p := range parsed {
		a[i] = AddressFrom(p)
	}
	return a, err
}

// SplitRaw returns the entries of the AddressList as stored,
// only trimmed, see SplitAddressList.
func (l AddressList) SplitRaw() []Address {
	return SplitAddressList(string(l))
}

// UniqueAddressParts returns an AddressSet containing the unique normalized
// address parts from the AddressList. This removes duplicates based on
// the actual email addresses (ignoring display names).
func (l AddressList) UniqueAddressParts() (AddressSet, error) {
	parsed, err := l.Parse()
	if err != nil {
		return nil, err
	}
	if len(parsed) == 0 {
		return nil, nil
	}
	set := make(AddressSet, len(parsed))
	for _, addr := range parsed {
		set.Add(Address(addrSpec(addr)))
	}
	return set, nil
}

// Validate returns an error if the AddressList
// is not a valid list of email addresses.
func (l AddressList) Validate() error {
	_, err := l.Parse()
	return err
}

// Valid returns true if the AddressList is a valid list of email addresses.
func (l AddressList) Valid() bool {
	return l.Validate() == nil
}

// ValidAndNormalized returns true if the AddressList is valid and already normalized.
func (l AddressList) ValidAndNormalized() bool {
	norm, err := l.Normalized()
	return err == nil && l == norm
}

// Normalized parses the AddressList using lenient validation,
// normalizes every address part to lowercase, and returns
// the addresses re-joined with ", " separators.
func (l AddressList) Normalized() (AddressList, error) {
	parsed, err := l.Parse()
	if err != nil {
		return l, err
	}
	b := strings.Builder{}
	b.Grow(len(l))
	for i, p := range parsed {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(string(AddressFrom(p)))
	}
	return AddressList(b.String()), nil
}

// ASCII returns the normalized list, but with
// non-ASCII names RFC 2047 encoded like Address.ASCII,
// or the unchanged list if it can't be normalized.
// A list without addresses is returned as "" for a message header,
// which BuildRawMessage doesn't write for it.
func (l AddressList) ASCII() string {
	parsed, err := l.Parse()
	if err != nil {
		return string(l)
	}
	return asciiAddressList(parsed)
}

// asciiAddressList joins the addresses formatted by asciiAddress
// with ", " separators.
func asciiAddressList(addrs []*mail.Address) string {
	formatted := make([]string, len(addrs))
	for i, addr := range addrs {
		formatted[i] = asciiAddress(addr)
	}
	return strings.Join(formatted, ", ")
}

// headerAddresses formats addrs for a message header with
// asciiAddressList, or like AddressFrom if an address part is not ASCII,
// because the MIME encoder then encodes the whole header value,
// which would encode the RFC 2047 encoded names a second time.
func headerAddresses(addrs ...*mail.Address) string {
	for _, addr := range addrs {
		if !isASCII(addr.Address) {
			formatted := make([]string, len(addrs))
			for i, a := range addrs {
				formatted[i] = string(AddressFrom(a))
			}
			return strings.Join(formatted, ", ")
		}
	}
	return asciiAddressList(addrs)
}

// Value implements the driver database/sql/driver.Valuer interface.
// It returns the list as stored with the surrounding whitespace trimmed,
// SQL NULL for a blank list, or the error of Validate for an invalid list.
func (l AddressList) Value() (driver.Value, error) {
	list := strutil.TrimSpace(l)
	if list == "" {
		return nil, nil
	}
	if err := list.Validate(); err != nil {
		return nil, err
	}
	return string(list), nil
}

// MarshalJSON implements encoding/json.Marshaler
// by writing the list as stored as JSON string.
func (l AddressList) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(l))
}

// UnmarshalJSON implements encoding/json.Unmarshaler.
// A JSON string is stored unchanged, so MarshalJSON and UnmarshalJSON
// round-trip every list. A JSON array of strings or an object with the
// addresses as keys is joined with ", ". JSON null leaves the AddressList
// unchanged like encoding/json does for a string type,
// and any other JSON value is an error.
func (l *AddressList) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, []byte(`null`)) {
		return nil
	}
	c, err := unmarshalJSONAddressCollection(data, "AddressList")
	if err != nil {
		return err
	}
	l.setCollection(c, false)
	return nil
}

// MarshalText implements encoding.TextMarshaler by returning the list
// as stored, so that encoding/xml and strfmt write it unchanged.
func (l AddressList) MarshalText() ([]byte, error) {
	return []byte(l), nil
}

// MarshalXMLAttr implements encoding/xml.MarshalerAttr by writing the
// text of MarshalText, and no attribute for a nil pointer, because
// encoding/xml calls MarshalText through a nil pointer attribute field
// and panics.
func (l *AddressList) MarshalXMLAttr(name xml.Name) (xml.Attr, error) {
	return internal.MarshalXMLAttrText(l, name)
}

// Nullable converts the AddressList to a NullableAddressList type.
func (l AddressList) Nullable() NullableAddressList {
	return NullableAddressList(l)
}

// Scan implements the database/sql.Scanner interface.
// A plain address list is stored as given with the surrounding whitespace
// trimmed, without parsing or validating it, so that one invalid value
// doesn't fail a whole query; Validate reports it. The elements of a
// PostgreSQL array (from a text[] column) or a JSON array and the keys of
// a JSON object (from a json or jsonb column) are joined with ", ",
// and a JSON string is read as address list. SQL NULL, a blank string
// and "null" are stored as "". Other SQL types are an error.
func (l *AddressList) Scan(value any) error {
	var text string
	switch s := value.(type) {
	case nil:
		*l = ""
		return nil
	case string:
		text = s
	case []byte:
		text = string(s)
	default:
		return errs.Errorf("can't scan SQL value of type %T as email.AddressList", value)
	}
	c, err := readAddressCollection(text, "AddressList")
	if err != nil {
		return err
	}
	l.setCollection(c, true)
	return nil
}

// setCollection stores the address collection c read by Scan or
// UnmarshalJSON, with the entries joined with ", " and a plain
// list trimmed if trim is true.
func (l *AddressList) setCollection(c addressCollection, trim bool) {
	switch {
	case c.null:
		*l = ""
	case c.entries != nil:
		*l = AddressListJoinStrings(c.entries...)
	case trim:
		*l = AddressList(strutil.TrimSpace(c.list))
	default:
		*l = AddressList(c.list)
	}
}
