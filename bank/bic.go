package bank

import (
	"database/sql/driver"
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/invopop/jsonschema"

	"github.com/domonda/go-types"
	"github.com/domonda/go-types/country"
	"github.com/domonda/go-types/internal"
	"github.com/domonda/go-types/strutil"
)

// Compile-time check that BIC implements types.NormalizableValidator[BIC]
var _ types.NormalizableValidator[BIC] = BIC("")

// NormalizeBIC returns the passed string as BIC normalized to a length of 11 characters
// by removing all whitespace like IBAN.Normalized
// and appending "XXX" in case of a valid length of 8 characters.
// Returns the string unchanged as BIC in case of an error.
func NormalizeBIC(str string) (BIC, error) {
	return BIC(str).Normalized()
}

// StringIsBIC returns true if the string can be parsed as a valid BIC.
func StringIsBIC(str string) bool {
	return BIC(str).Valid()
}

// BIC represents a SWIFT Business Identifier Code (also known as SWIFT-Code).
// See NullableBIC for a nullable version of this type.
//
// Scan, UnmarshalJSON, UnmarshalXML and ScanString without validate trim
// surrounding whitespace and uppercase the letters. A blank BIC is "".
// An invalid BIC is kept after that prettification, without an error,
// and an 8-character BIC is not extended to 11. Value, MarshalJSON and
// MarshalText write the same text, and Value writes SQL NULL for a blank
// BIC. They do not reject an invalid BIC. ScanString with validate stores
// the normalized BIC or returns the error of Normalized. fmt prints the
// BIC as stored. BIC has no UnmarshalText, so encoding/json reads and
// writes the keys of a map[BIC]V as given. See "Reading and writing
// values" in the README of the module.
type BIC string

// ScanString implements the strfmt.Scannable interface.
// Without validate it stores source like Scan, with validate it stores
// the normalized BIC or returns the error of Normalized,
// also for a blank source.
func (bic *BIC) ScanString(source string, validate bool) error {
	if validate {
		return internal.ScanStringNormalized(bic, source, true, false, BIC.Normalized, nil)
	}
	return internal.ScanStringNormalized(bic, source, false, false, internal.TrimUpper[BIC], nil)
}

// Valid returns true if this is a valid SWIFT Business Identifier Code.
func (bic BIC) Valid() bool {
	return bic.Validate() == nil
}

// ValidAndNormalized returns true if the BIC is valid and already normalized.
func (bic BIC) ValidAndNormalized() bool {
	norm, err := bic.Normalized()
	return err == nil && bic == norm
}

// Validate returns an error if this is not a valid SWIFT Business Identifier Code
// after removing all whitespace like Normalized.
// Checks length, format, and validates against a list of known invalid BICs.
func (bic BIC) Validate() error {
	_, err := bic.Normalized()
	return err
}

// validateCompact checks the BIC as given, without removing whitespace.
func (bic BIC) validateCompact() error {
	length := len(bic)
	if length != BICMinLength && length != BICMaxLength {
		return fmt.Errorf("invalid BIC %q length: %d", string(bic), length)
	}

	if isFalseBIC(bic) {
		return fmt.Errorf("BIC %q is in list of invalid BICs", string(bic))
	}
	_, _, _, isValid := bic.Parse()
	if !isValid {
		return fmt.Errorf("invalid BIC %q", string(bic))
	}
	return nil
}

// Normalized returns the BIC normalized to a length of 11 characters
// by removing all whitespace like IBAN.Normalized
// and appending "XXX" in case of a valid length of 8 characters.
// Returns the BIC unchanged in case of an error.
func (bic BIC) Normalized() (BIC, error) {
	norm := BIC(strutil.RemoveRunesString(string(bic), strutil.IsSpace))
	if err := norm.validateCompact(); err != nil {
		return bic, err
	}
	if len(norm) == 8 {
		norm += "XXX"
	}
	return norm, nil
}

// NormalizedShort returns the BIC normalized to a length of 8 or 11 characters
// by removing all whitespace like IBAN.Normalized
// and trimming the "XXX" suffix in case of a valid length of 8 characters.
// Returns the BIC unchanged in case of an error.
func (bic BIC) NormalizedShort() (BIC, error) {
	norm := BIC(strutil.RemoveRunesString(string(bic), strutil.IsSpace))
	if err := norm.validateCompact(); err != nil {
		return bic, err
	}
	if len(norm) == 11 && strings.HasSuffix(string(norm), "XXX") {
		norm = norm[:8]
	}
	return norm, nil
}

// NormalizedOrNull returns the BIC as NullableBIC, normalized if possible.
// Returns BICNull if normalization fails.
func (bic BIC) NormalizedOrNull() NullableBIC {
	normalized, err := bic.Normalized()
	if err != nil {
		return BICNull
	}
	return NullableBIC(normalized)
}

// Nullable returns the BIC as NullableBIC.
func (bic BIC) Nullable() NullableBIC {
	return NullableBIC(bic)
}

// Parse parses the BIC into its components: bank code, country code, branch code.
// Returns isValid as false if the BIC is invalid or in the list of false BICs.
// The BIC is parsed as given, so a BIC containing whitespace is invalid,
// use Normalized first.
func (bic BIC) Parse() (bankCode string, countryCode country.Code, branchCode string, isValid bool) {
	if isFalseBIC(bic) {
		return "", "", "", false
	}

	length := len(bic)
	if !(length == BICMinLength || length == BICMaxLength) {
		return "", "", "", false
	}

	strBIC := string(bic)

	// bankCode
	for i := range 4 {
		if !isUpperAZ(strBIC[i]) {
			return "", "", "", false
		}
	}
	bankCode = strBIC[0:4]

	// countryCode
	for i := 4; i < 6; i++ {
		if !isUpperAZ(strBIC[i]) {
			return "", "", "", false
		}
	}
	countryCode = country.Code(strBIC[4:6])

	if !countryCode.Valid() {
		return "", "", "", false
	}

	// location
	if !(isUpperAZ(strBIC[6]) || isInRange(strBIC[6], '2', '9')) {
		return "", "", "", false
	}

	if !(isInRange(strBIC[7], 'A', 'N') || isInRange(strBIC[7], 'P', 'Z') || isNum(strBIC[7])) {
		return "", "", "", false
	}

	// branchCode
	if length == BICMaxLength {
		branchCode = strBIC[8:]

		if branchCode != "XXX" {
			if !(isInRange(branchCode[0], 'A', 'W') || isInRange(branchCode[0], 'Y', 'Z') || isNum(branchCode[0])) {
				return "", "", "", false
			}

			if !isUpperAZ0to9(branchCode[1]) || !isUpperAZ0to9(branchCode[2]) {
				return "", "", "", false
			}
		}
	}

	return bankCode, countryCode, branchCode, true
}

// CountryCode returns the country code of the normalized BIC.
// May be invalid if the BIC is invalid.
func (bic BIC) CountryCode() country.Code {
	norm, _ := bic.Normalized()
	_, cc, _, _ := norm.Parse()
	return cc
}

// TrimBranchCode returns the normalized BIC with the branch code removed
// (first 8 characters), or the first 8 characters of the unchanged BIC
// if it can't be normalized.
func (bic BIC) TrimBranchCode() BIC {
	norm, err := bic.Normalized()
	if err == nil {
		bic = norm
	}
	if len(bic) <= 8 {
		return bic
	}
	return bic[:8]
}

// IsTestBIC returns true if this is a test BIC
// (8th character of the normalized BIC is '0').
func (bic BIC) IsTestBIC() bool {
	norm, err := bic.Normalized()
	return err == nil && norm[7] == '0'
}

// IsPassiveSWIFT returns true if this is a passive SWIFT BIC
// (8th character of the normalized BIC is '1').
func (bic BIC) IsPassiveSWIFT() bool {
	norm, err := bic.Normalized()
	return err == nil && norm[7] == '1'
}

// ReceiverPaysFees returns true if the receiver pays fees
// (8th character of the normalized BIC is '2').
func (bic BIC) ReceiverPaysFees() bool {
	norm, err := bic.Normalized()
	return err == nil && norm[7] == '2'
}

// Scan implements the database/sql.Scanner interface.
// Surrounding whitespace is trimmed and letters are uppercased.
// SQL NULL and a blank string are stored as "". An invalid BIC is
// kept after that prettification, without an error.
func (bic *BIC) Scan(value any) error {
	return internal.ScanNormalized(bic, value, internal.TrimUpper[BIC], nil)
}

// MarshalJSON implements encoding/json.Marshaler by writing the
// trimmed and uppercased BIC, or "" for a blank BIC.
func (bic BIC) MarshalJSON() ([]byte, error) {
	return internal.MarshalJSONNormalized(bic, false, internal.TrimUpper[BIC], nil)
}

// UnmarshalJSON implements encoding/json.Unmarshaler.
// A JSON string is stored like Scan stores it, JSON null leaves the BIC
// unchanged like encoding/json does for a string type, and any other
// JSON value is an error.
func (bic *BIC) UnmarshalJSON(data []byte) error {
	return internal.UnmarshalJSONNormalized(bic, data, false, internal.TrimUpper[BIC], nil)
}

// MarshalText implements encoding.TextMarshaler like MarshalJSON,
// so that encoding/xml and strfmt write the trimmed and uppercased BIC.
// encoding/json writes map keys of string types as they are
// and doesn't call MarshalText for them.
func (bic BIC) MarshalText() ([]byte, error) {
	return internal.MarshalTextNormalized(bic, internal.TrimUpper[BIC], nil)
}

// MarshalXMLAttr implements encoding/xml.MarshalerAttr by writing the
// text of MarshalText, and no attribute for a nil pointer, because
// encoding/xml calls MarshalText through a nil pointer attribute field
// and panics.
func (bic *BIC) MarshalXMLAttr(name xml.Name) (xml.Attr, error) {
	return internal.MarshalXMLAttrText(bic, name)
}

// UnmarshalXML implements encoding/xml.Unmarshaler
// by storing the text of the element like Scan stores it.
func (bic *BIC) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	return internal.UnmarshalXMLNormalized(bic, d, start, internal.TrimUpper[BIC], nil)
}

// UnmarshalXMLAttr implements encoding/xml.UnmarshalerAttr
// by storing the attribute value like Scan stores it.
func (bic *BIC) UnmarshalXMLAttr(attr xml.Attr) error {
	return internal.UnmarshalXMLAttrNormalized(bic, attr, internal.TrimUpper[BIC], nil)
}

// Value implements the driver database/sql/driver.Valuer interface.
// It returns the trimmed and uppercased BIC, or SQL NULL for a blank one.
// An invalid BIC is written after that prettification, not rejected.
func (bic BIC) Value() (driver.Value, error) {
	return internal.ValueNormalized(bic, internal.TrimUpper[BIC], nil)
}

// JSONSchema returns the JSON schema definition for the BIC type.
func (BIC) JSONSchema() *jsonschema.Schema {
	return &jsonschema.Schema{
		Title:   "BIC/SWIFT-Code",
		Type:    "string",
		Pattern: BICRegex,
	}
}

// isFalseBIC returns if bic is in falseBICs, also as 11 character BIC
// with the "XXX" branch code appended to an 8 character entry,
// like "AUTOBANKXXX", which Normalized writes for "AUTOBANK".
func isFalseBIC(bic BIC) bool {
	if _, isFalse := falseBICs[bic]; isFalse {
		return true
	}
	if len(bic) == BICMaxLength && strings.HasSuffix(string(bic), "XXX") {
		_, isFalse := falseBICs[bic[:BICMinLength]]
		return isFalse
	}
	return false
}

// falseBICs contains a list of known invalid BIC codes.
var falseBICs = map[BIC]struct{}{
	"AMTSGERICHT": {},
	"AUTOBANK":    {},
	"DEUTSCHLAND": {},
	"DIENSTGEBER": {},
	"DOCUMENT":    {},
	"DOKUMENT":    {},
	"FACILITY":    {},
	"GELISTET":    {},
	"GESAMTNETTO": {},
}
