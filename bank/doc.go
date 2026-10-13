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
// so they accept spaced input, and Value and MarshalJSON write the
// normalized form. Only [BIC.Parse] and [BICParser] read the BIC as given.
//
// The Scan methods of IBAN, BIC, NullableIBAN, and NullableBIC
// don't validate. Only NullableIBAN and NullableBIC write SQL NULL
// and JSON null, IBAN and BIC write an empty value as empty string.
// Account has no driver.Valuer and no MarshalJSON method.
//
// See the README.md file of the package for the complete API reference,
// the serialization rules, and the known pitfalls.
package bank
