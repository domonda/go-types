// Package money provides money amounts, ISO 4217 currencies, and rates
// with locale-aware parsing and SQL, JSON, and strfmt support.
//
// The package provides three amount types:
//   - [Amount]: a float64 with money methods, the most tolerant to parse
//   - [DecimalAmount]: an exact fixed-point decimal in one int64 that keeps
//     its scale, with NaN and ±Inf states and explicit [RoundingMode] values
//   - [CentAmount]: an exact int64 count of cents
//
// And around them:
//   - [Currency] and [NullableCurrency]: ISO 4217 codes with a constant per
//     code, like [EUR], normalization of symbols and English names,
//     and [Currency.MinorUnits]
//   - [CurrencyAmount] and [CurrencyDecimalAmount]: a currency with an amount
//   - [Rate]: a float64 for exchange rates and percentages
//   - [NullableAmount], [NullableCentAmount], [NullableDecimalAmount], and
//     [NullableRate]: nullable.Type wrappers for optional values
//   - [AmountParser], [CurrencyAmountParser], [CurrencyParser], and
//     [AmountFinder]: strfmt.Parser and types.Finder implementations
//
// All parse functions detect the separators with float.ParseDetails, so a
// single '.' or ',' is the decimal separator and "1,000" parses as 1.
// Pass acceptedDecimals, like ParseAmount(str, 0, 2), to reject such input.
//
// Amount and Rate round the shortest decimal representation of the float64
// half away from zero, so Amount(0.145).RoundToCents() is 0.15.
// The Amount and CentAmount splits apportion the cents with the largest
// remainder method, so the parts sum up to the rounded total.
//
// See the README.md file of the package for the complete API reference,
// the SQL and JSON forms of every type, and the known pitfalls.
package money
