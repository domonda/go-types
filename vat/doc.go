// Package vat validates and normalizes European VAT identification numbers.
//
// [ID] is a string type. [ID.Normalized] upper-cases the input, removes
// whitespace and punctuation, then checks the length, the country prefix,
// the format of that country and, for most countries, the check digit:
// "atu 1022 3006" becomes "ATU10223006". Every rejection wraps
// [ErrInvalidID], so callers can detect invalid input with errors.Is.
// The checks run offline: they do not ask VIES whether a number is registered.
//
// Supported prefixes are the EU member states (Greece as EL, not GR),
// CH, GB, IS, LI, NO and SM, plus [NorthernIrelandVATCountryCode] (XI)
// and [MOSSSchemaVATCountryCode] (EU).
//
// The package includes:
//   - [ID] and [NullableID]; both write an empty value as SQL NULL,
//     only [NullableID] encodes null as JSON null
//   - [ID.CountryCode], [ID.Number] and [ID.IsMOSS] to inspect a VAT ID
//   - [IDFinder] to find VAT IDs in text
//   - [IDParser] for the strfmt package
//
// [ID.Valid] and [ID.Validate] accept any spelling that normalizes;
// use [ID.ValidAndNormalized] or [ID.ValidateIsNormalized] to require
// the canonical form.
//
// See the README.md of this package for the full API,
// the supported formats and the serialization rules.
package vat
