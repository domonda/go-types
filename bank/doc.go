// Package bank provides banking identifier types for Go applications.
//
// The package provides:
//   - [IBAN]: International Bank Account Number with validation
//     of the country specific length and the mod-97 check sum
//   - [BIC]: Business Identifier Code (SWIFT code) of 8 or 11 characters
//   - [NullableIBAN] and [NullableBIC]: the empty string is the null value
//   - [Account]: IBAN, BIC, currency, and holder of a bank account
//   - [IBANFinder] and [BICFinder]: find IBANs and BICs in text,
//     [IBANParser] and [BICParser] implement strfmt.Parser
//   - [CAMT53] and related structs: encoding/xml tags to decode
//     ISO 20022 CAMT.053 bank statements, without parsing logic of their own
//
// [IBAN.Normalized] removes all whitespace before validating,
// [BIC.Normalized] also removes all whitespace and appends "XXX"
// to an 8 character BIC. Valid and Validate of both normalize first,
// so they accept spaced input. Scan, JSON and XML decoding store the
// normalized form of a valid value and an invalid value unchanged,
// without validating it. Value and MarshalJSON write the normalized form,
// and Value returns an error for an invalid value. Only [BIC.Parse] and
// [BICParser] read the BIC as given.
//
// All four types write an empty value as SQL NULL, only NullableIBAN and
// NullableBIC write JSON null. Account has no driver.Valuer and no
// MarshalJSON method.
//
// See the README.md file of the package for the complete API reference,
// the serialization rules, and the known pitfalls.
package bank
