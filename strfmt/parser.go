package strfmt

import "github.com/domonda/go-types/language"

// Parser parses a string into a normalized form, optionally
// using language hints to disambiguate locale-dependent input.
// Returns the normalized string or a parsing error.
//
// Implemented by: bank.IBANParser, bank.BICParser, date.Parser,
// date.Formatter, date.Format, money.AmountParser,
// money.CurrencyParser, money.CurrencyAmountParser, vat.IDParser.
type Parser interface {
	Parse(str string, langHints ...language.Code) (normalized string, err error)
}
